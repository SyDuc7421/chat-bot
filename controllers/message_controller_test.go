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
	"hsduc.com/rag/dtos"
	"hsduc.com/rag/models"
)

func TestCreateMessage(t *testing.T) {
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

	tests := []struct {
		name           string
		setup          func() []byte
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - User Message with LLM response",
			setup: func() []byte {
				conversation := models.Conversation{Title: "Test Conversation", UserID: 1}
				database.DB.Create(&conversation)
				payload := dtos.CreateMessageRequest{
					ConversationID: conversation.ID,
					Role:           "user",
					Content:        "Hello Assistant",
				}
				body, _ := json.Marshal(payload)
				return body
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]models.Message
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)

				userMsg, exists := response["user_message"]
				assert.True(t, exists)
				assert.Equal(t, "user", userMsg.Role)
				assert.Equal(t, "Hello Assistant", userMsg.Content)
				assert.NotZero(t, userMsg.ID)

				assistantMsg, aExists := response["assistant_message"]
				assert.True(t, aExists)
				assert.Equal(t, "assistant", assistantMsg.Role)
				assert.Equal(t, "Mocked assistant response", assistantMsg.Content)
				assert.NotZero(t, assistantMsg.ID)
			},
		},
		{
			name: "Error - Conversation Not Found",
			setup: func() []byte {
				payload := dtos.CreateMessageRequest{
					ConversationID: 9999, // Non-existent
					Role:           "user",
					Content:        "Lost Message",
				}
				body, _ := json.Marshal(payload)
				return body
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation not found", response["error"])
			},
		},
		{
			name: "Error - Invalid JSON format",
			setup: func() []byte {
				return []byte(`{"conversation_id": "invalid"}`) // Invalid format
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.NotEmpty(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.POST("/messages", CreateMessage)

			payload := tt.setup()
			req, _ := http.NewRequest("POST", "/messages", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestGetMessages(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Get messages from conversation",
			setup: func() string {
				conversation := models.Conversation{Title: "Chat", UserID: 1}
				database.DB.Create(&conversation)
				database.DB.Create(&models.Message{ConversationID: conversation.ID, Role: "user", Content: "Hello"})
				database.DB.Create(&models.Message{ConversationID: conversation.ID, Role: "assistant", Content: "Hi there"})
				return fmt.Sprintf("/messages?conversation_id=%d", conversation.ID)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var messages []models.Message
				err := json.Unmarshal(w.Body.Bytes(), &messages)
				assert.NoError(t, err)
				assert.Len(t, messages, 2)
				assert.Equal(t, "Hello", messages[0].Content)
				assert.Equal(t, "Hi there", messages[1].Content)
			},
		},
		{
			name: "Error - Conversation not found or belongs to another user",
			setup: func() string {
				// Another user's conversation
				conversation := models.Conversation{Title: "Chat", UserID: 999}
				database.DB.Create(&conversation)
				return fmt.Sprintf("/messages?conversation_id=%d", conversation.ID)
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation not found", response["error"])
			},
		},
		{
			name: "Error - Missing conversation_id",
			setup: func() string {
				return "/messages" // No query param
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "conversation_id is required", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.GET("/messages", GetMessages)

			path := tt.setup()
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestGetMessage(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (string, *models.Message)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message)
	}{
		{
			name: "Success - Get single message",
			setup: func() (string, *models.Message) {
				conversation := models.Conversation{Title: "Chat", UserID: 1}
				database.DB.Create(&conversation)
				message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Retrieve Me"}
				database.DB.Create(&message)
				return fmt.Sprintf("/messages/%d", message.ID), &message
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message) {
				var result models.Message
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Retrieve Me", result.Content)
				assert.Equal(t, msg.ID, result.ID)
			},
		},
		{
			name: "Error - Message Not Found",
			setup: func() (string, *models.Message) {
				return "/messages/9999", nil
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Message not found", response["error"])
			},
		},
		{
			name: "Error - Message belongs to other user",
			setup: func() (string, *models.Message) {
				conversation := models.Conversation{Title: "Chat", UserID: 999}
				database.DB.Create(&conversation)
				message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Not Yours"}
				database.DB.Create(&message)
				return fmt.Sprintf("/messages/%d", message.ID), nil
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Message not found", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.GET("/messages/:id", GetMessage)

			path, msg := tt.setup()
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w, msg)
		})
	}
}

func TestUpdateMessage(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (string, []byte)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Update message",
			setup: func() (string, []byte) {
				conversation := models.Conversation{Title: "Chat", UserID: 1}
				database.DB.Create(&conversation)
				message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Old Text"}
				database.DB.Create(&message)

				payload := dtos.UpdateMessageRequest{Content: "New Text"}
				body, _ := json.Marshal(payload)
				return fmt.Sprintf("/messages/%d", message.ID), body
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result models.Message
				json.Unmarshal(w.Body.Bytes(), &result)
				assert.Equal(t, "New Text", result.Content)
			},
		},
		{
			name: "Error - Message not found",
			setup: func() (string, []byte) {
				payload := dtos.UpdateMessageRequest{Content: "New Text"}
				body, _ := json.Marshal(payload)
				return "/messages/9999", body
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Message not found", response["error"])
			},
		},
		{
			name: "Error - Invalid JSON format",
			setup: func() (string, []byte) {
				conversation := models.Conversation{Title: "Chat", UserID: 1}
				database.DB.Create(&conversation)
				message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Old Text"}
				database.DB.Create(&message)
				return fmt.Sprintf("/messages/%d", message.ID), []byte(`{"content":}`)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.NotEmpty(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.PUT("/messages/:id", UpdateMessage)

			path, body := tt.setup()
			req, _ := http.NewRequest("PUT", path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestDeleteMessage(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (string, *models.Message)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message)
	}{
		{
			name: "Success - Delete message",
			setup: func() (string, *models.Message) {
				conversation := models.Conversation{Title: "Chat", UserID: 1}
				database.DB.Create(&conversation)
				message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "To Be Deleted"}
				database.DB.Create(&message)

				return fmt.Sprintf("/messages/%d", message.ID), &message
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message) {
				// Verify soft delete
				var check models.Message
				err := database.DB.Where("id = ?", msg.ID).First(&check).Error
				assert.Error(t, err)

				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Message deleted", response["message"])
			},
		},
		{
			name: "Error - Message not found",
			setup: func() (string, *models.Message) {
				return "/messages/9999", nil
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, msg *models.Message) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Message not found", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.DELETE("/messages/:id", DeleteMessage)

			path, msg := tt.setup()
			req, _ := http.NewRequest("DELETE", path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w, msg)
		})
	}
}
