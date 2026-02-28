package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"hsduc.com/rag/config"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
)

func TestCreateMessage(t *testing.T) {
	SetupTestDB()
	r := GetTestRouter()
	r.POST("/messages", CreateMessage)

	// Create a conversation first to associate the message
	conversation := models.Conversation{Title: "Test Conversation", UserID: 1}
	database.DB.Create(&conversation)

	// Create a mock OpenAI server
	mockOpenAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := map[string]interface{}{
			"choices": []map[string]interface{}{
				{
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": "Mocked assistant response",
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockOpenAI.Close()

	if config.App == nil {
		config.App = &config.Config{}
	}
	config.App.OpenAIApiKey = "test-key"
	config.App.OpenAIBaseURL = mockOpenAI.URL

	// Test valid input
	payload := []byte(fmt.Sprintf(`{"conversation_id":%d, "role":"user", "content":"Hello Assistant"}`, conversation.ID))
	req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]models.Message
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	message, exists := response["user_message"]
	assert.True(t, exists)
	assert.Equal(t, conversation.ID, message.ConversationID)
	assert.Equal(t, "user", message.Role)
	assert.Equal(t, "Hello Assistant", message.Content)
	assert.NotZero(t, message.ID)

	assistantMessage, aExists := response["assistant_message"]
	assert.True(t, aExists)
	assert.Equal(t, conversation.ID, assistantMessage.ConversationID)
	assert.Equal(t, "assistant", assistantMessage.Role)
	assert.Equal(t, "Mocked assistant response", assistantMessage.Content)
	assert.NotZero(t, assistantMessage.ID)

	// Test invalid conversation ID (Not Found)
	payloadInvalid := []byte(`{"conversation_id":999, "role":"user", "content":"Lost Message"}`)
	reqInvalid, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(payloadInvalid))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()

	r.ServeHTTP(wInvalid, reqInvalid)
	assert.Equal(t, http.StatusNotFound, wInvalid.Code)
}

func TestGetMessages(t *testing.T) {
	SetupTestDB()
	r := GetTestRouter()
	r.GET("/messages", GetMessages)

	// setup data
	conversation := models.Conversation{Title: "Chat", UserID: 1}
	database.DB.Create(&conversation)
	database.DB.Create(&models.Message{ConversationID: conversation.ID, Role: "user", Content: "Hello"})
	database.DB.Create(&models.Message{ConversationID: conversation.ID, Role: "assistant", Content: "Hi there"})

	// test valid GET message by conversation_id
	req, _ := http.NewRequest("GET", fmt.Sprintf("/messages?conversation_id=%d", conversation.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var messages []models.Message
	err := json.Unmarshal(w.Body.Bytes(), &messages)
	assert.NoError(t, err)
	assert.Len(t, messages, 2)
	assert.Equal(t, "Hello", messages[0].Content)
	assert.Equal(t, "Hi there", messages[1].Content)

	// test GET missing conversation_id
	reqMissing, _ := http.NewRequest("GET", "/messages", nil)
	wMissing := httptest.NewRecorder()
	r.ServeHTTP(wMissing, reqMissing)
	assert.Equal(t, http.StatusBadRequest, wMissing.Code)
}

func TestUpdateMessage(t *testing.T) {
	SetupTestDB()
	conversation := models.Conversation{Title: "Chat", UserID: 1}
	database.DB.Create(&conversation)
	message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Old Text"}
	database.DB.Create(&message)

	r := GetTestRouter()
	r.PUT("/messages/:id", UpdateMessage)

	payload := []byte(`{"content":"New Text"}`)
	req, _ := http.NewRequest("PUT", fmt.Sprintf("/messages/%d", message.ID), bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	var result models.Message
	json.Unmarshal(w.Body.Bytes(), &result)
	assert.Equal(t, "New Text", result.Content)
}

func TestDeleteMessage(t *testing.T) {
	SetupTestDB()
	conversation := models.Conversation{Title: "Chat", UserID: 1}
	database.DB.Create(&conversation)
	message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "To Be Deleted"}
	database.DB.Create(&message)

	r := GetTestRouter()
	r.DELETE("/messages/:id", DeleteMessage)

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/messages/%d", message.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify it's actually deleted
	var check models.Message
	err := database.DB.Where("id = ?", message.ID).First(&check).Error
	assert.Error(t, err) // Should error because the soft delete set deleted_at
}
