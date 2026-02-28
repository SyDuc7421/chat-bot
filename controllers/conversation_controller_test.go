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
	"hsduc.com/rag/dtos"
	"hsduc.com/rag/models"
)

func TestCreateConversation(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() []byte
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Create Conversation",
			setup: func() []byte {
				payload := dtos.CreateConversationRequest{Title: "New Conversation"}
				body, _ := json.Marshal(payload)
				return body
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var conversation models.Conversation
				err := json.Unmarshal(w.Body.Bytes(), &conversation)
				assert.NoError(t, err)
				assert.Equal(t, "New Conversation", conversation.Title)
				assert.NotZero(t, conversation.ID)
				assert.Equal(t, uint(1), conversation.UserID) // based on mocked context
			},
		},
		{
			name: "Error - Invalid JSON",
			setup: func() []byte {
				return []byte(`{"title":`)
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.NotEmpty(t, response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.POST("/conversations", CreateConversation)

			payload := tt.setup()
			req, _ := http.NewRequest("POST", "/conversations", bytes.NewBuffer(payload))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestGetConversations(t *testing.T) {
	tests := []struct {
		name           string
		setup          func()
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Get all conversations",
			setup: func() {
				database.DB.Create(&models.Conversation{Title: "Chat 1", UserID: 1})
				database.DB.Create(&models.Conversation{Title: "Chat 2", UserID: 1})
				database.DB.Create(&models.Conversation{Title: "Chat 3", UserID: 99}) // Should not be retrieved for User 1
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var conversations []models.Conversation
				err := json.Unmarshal(w.Body.Bytes(), &conversations)
				assert.NoError(t, err)
				assert.Len(t, conversations, 2)
				assert.Equal(t, "Chat 1", conversations[0].Title)
				assert.Equal(t, "Chat 2", conversations[1].Title)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			tt.setup()

			r := GetTestRouter()
			r.GET("/conversations", GetConversations)

			req, _ := http.NewRequest("GET", "/conversations", nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestGetConversation(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() string
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Get existing conversation",
			setup: func() string {
				conversation := models.Conversation{Title: "Chat with messages", UserID: 1}
				database.DB.Create(&conversation)
				message := models.Message{ConversationID: conversation.ID, Role: "user", Content: "Hello"}
				database.DB.Create(&message)
				return fmt.Sprintf("/conversations/%d", conversation.ID)
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result models.Conversation
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "Chat with messages", result.Title)
				assert.Len(t, result.Messages, 1)
				assert.Equal(t, "Hello", result.Messages[0].Content)
			},
		},
		{
			name: "Error - Conversation Not Found",
			setup: func() string {
				return "/conversations/999"
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation not found", response["error"])
			},
		},
		{
			name: "Error - Unauthorized (belongs to another user)",
			setup: func() string {
				conversation := models.Conversation{Title: "Hacked Chat", UserID: 999}
				database.DB.Create(&conversation)
				return fmt.Sprintf("/conversations/%d", conversation.ID)
			},
			expectedStatus: http.StatusNotFound, // Same because it ensures user_id matches
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation not found", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.GET("/conversations/:id", GetConversation)

			path := tt.setup()
			req, _ := http.NewRequest("GET", path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w)
		})
	}
}

func TestUpdateConversation(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (string, []byte)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder)
	}{
		{
			name: "Success - Update conversation",
			setup: func() (string, []byte) {
				conversation := models.Conversation{Title: "Old Title", UserID: 1}
				database.DB.Create(&conversation)

				payload := dtos.UpdateConversationRequest{Title: "New Title"}
				body, _ := json.Marshal(payload)
				return fmt.Sprintf("/conversations/%d", conversation.ID), body
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var result models.Conversation
				err := json.Unmarshal(w.Body.Bytes(), &result)
				assert.NoError(t, err)
				assert.Equal(t, "New Title", result.Title)
			},
		},
		{
			name: "Error - Update Non-Existent",
			setup: func() (string, []byte) {
				payload := dtos.UpdateConversationRequest{Title: "New Title"}
				body, _ := json.Marshal(payload)
				return "/conversations/999", body
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation not found", response["error"])
			},
		},
		{
			name: "Error - Invalid JSON",
			setup: func() (string, []byte) {
				conversation := models.Conversation{Title: "Old Title", UserID: 1}
				database.DB.Create(&conversation)
				return fmt.Sprintf("/conversations/%d", conversation.ID), []byte(`{"title":`)
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
			r.PUT("/conversations/:id", UpdateConversation)

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

func TestDeleteConversation(t *testing.T) {
	tests := []struct {
		name           string
		setup          func() (string, *models.Conversation)
		expectedStatus int
		checkResponse  func(t *testing.T, w *httptest.ResponseRecorder, conv *models.Conversation)
	}{
		{
			name: "Success - Delete Conversation",
			setup: func() (string, *models.Conversation) {
				conversation := models.Conversation{Title: "To Be Deleted", UserID: 1}
				database.DB.Create(&conversation)
				return fmt.Sprintf("/conversations/%d", conversation.ID), &conversation
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, conv *models.Conversation) {
				// Verify soft delete in DB
				var check models.Conversation
				err := database.DB.Where("id = ?", conv.ID).First(&check).Error
				assert.Error(t, err)

				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation deleted", response["message"])
			},
		},
		{
			name: "Error - Delete Non-Existent",
			setup: func() (string, *models.Conversation) {
				return "/conversations/999", nil
			},
			expectedStatus: http.StatusNotFound,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder, conv *models.Conversation) {
				var response map[string]string
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, "Conversation not found", response["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetupTestDB()
			r := GetTestRouter()
			r.DELETE("/conversations/:id", DeleteConversation)

			path, conv := tt.setup()
			req, _ := http.NewRequest("DELETE", path, nil)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.checkResponse(t, w, conv)
		})
	}
}
