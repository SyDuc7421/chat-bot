package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
)

func TestCreateConversation(t *testing.T) {
	SetupTestDB()
	r := GetTestRouter()
	r.POST("/conversations", CreateConversation)

	// Test valid input
	payload := []byte(`{"title":"New Conversation"}`)
	req, _ := http.NewRequest("POST", "/conversations", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var conversation models.Conversation
	err := json.Unmarshal(w.Body.Bytes(), &conversation)
	assert.NoError(t, err)
	assert.Equal(t, "New Conversation", conversation.Title)
	assert.NotZero(t, conversation.ID)

	// Test invalid input
	payloadInvalid := []byte(`{"invalid_field":"test"}`)
	reqInvalid, _ := http.NewRequest("POST", "/conversations", bytes.NewBuffer(payloadInvalid))
	reqInvalid.Header.Set("Content-Type", "application/json")
	wInvalid := httptest.NewRecorder()

	r.ServeHTTP(wInvalid, reqInvalid)
	assert.Equal(t, http.StatusBadRequest, wInvalid.Code)
}

func TestGetConversations(t *testing.T) {
	SetupTestDB()
	// Insert dummy data
	database.DB.Create(&models.Conversation{Title: "Chat 1"})
	database.DB.Create(&models.Conversation{Title: "Chat 2"})

	r := GetTestRouter()
	r.GET("/conversations", GetConversations)

	req, _ := http.NewRequest("GET", "/conversations", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var conversations []models.Conversation
	err := json.Unmarshal(w.Body.Bytes(), &conversations)
	assert.NoError(t, err)
	assert.Len(t, conversations, 2)
	assert.Equal(t, "Chat 1", conversations[0].Title)
	assert.Equal(t, "Chat 2", conversations[1].Title)
}

func TestGetConversation(t *testing.T) {
	SetupTestDB()
	conversation := models.Conversation{Title: "Chat with messages"}
	database.DB.Create(&conversation)
	message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Hello"}
	database.DB.Create(&message)

	r := GetTestRouter()
	// need the :id parameter route
	r.GET("/conversations/:id", GetConversation)

	req, _ := http.NewRequest("GET", fmt.Sprintf("/conversations/%d", conversation.ID), nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result models.Conversation
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, "Chat with messages", result.Title)
	assert.Len(t, result.Messages, 1)
	assert.Equal(t, "Hello", result.Messages[0].Content)
}
