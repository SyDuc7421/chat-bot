package services

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sashabaranov/go-openai"
	"github.com/stretchr/testify/assert"
	"hsduc.com/rag/config"
	"hsduc.com/rag/models"
)

func TestGetChatbotResponse_MissingAPIKey(t *testing.T) {
	config.App = &config.Config{
		OpenAIApiKey: "",
	}

	messages := []models.Message{
		{Role: "user", Content: "Hello"},
	}

	reply, err := GetChatbotResponse(messages, nil)
	assert.Error(t, err)
	assert.Equal(t, "missing OpenAI API Key", err.Error())
	assert.Empty(t, reply)
}

func TestGetChatbotResponse_Success(t *testing.T) {
	// Create a mock OpenAI server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/chat/completions", r.URL.Path)
		assert.Equal(t, "Bearer test-key", r.Header.Get("Authorization"))

		response := openai.ChatCompletionResponse{
			Choices: []openai.ChatCompletionChoice{
				{
					Message: openai.ChatCompletionMessage{
						Role:    "assistant",
						Content: "Hi there! I am a mock AI.",
					},
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer mockServer.Close()

	config.App = &config.Config{
		OpenAIApiKey:  "test-key",
		OpenAIBaseURL: mockServer.URL,
	}

	messages := []models.Message{
		{Role: "user", Content: "Hello"},
		{Role: "assistant", Content: "Hi"},
		{Role: "user", Content: "How are you?"},
	}
	docs := []string{"Doc1 content", "Doc2 content"}

	reply, err := GetChatbotResponse(messages, docs)
	assert.NoError(t, err)
	assert.Equal(t, "Hi there! I am a mock AI.", reply)
}
