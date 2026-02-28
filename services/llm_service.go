package services

import (
	"context"
	"errors"
	"log"
	"strings"

	"github.com/sashabaranov/go-openai"
	"hsduc.com/rag/config"
	"hsduc.com/rag/models"
)

// GetChatbotResponse calls the LLM with the context of the previous messages and document contexts
func GetChatbotResponse(previousMessages []models.Message, documents []string) (string, error) {
	if config.App == nil || config.App.OpenAIApiKey == "" {
		return "", errors.New("missing OpenAI API Key")
	}

	cfg := openai.DefaultConfig(config.App.OpenAIApiKey)
	if config.App.OpenAIBaseURL != "" {
		cfg.BaseURL = config.App.OpenAIBaseURL
	}
	client := openai.NewClientWithConfig(cfg)

	// Build message history
	var chatMessages []openai.ChatCompletionMessage

	// Add system prompt
	systemPrompt := "You are a helpful and polite chatbot assistant."
	if len(documents) > 0 {
		systemPrompt += "\n\nPlease use the following context from documents to answer the user's question:\n" + strings.Join(documents, "\n---\n")
	}

	chatMessages = append(chatMessages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: systemPrompt,
	})

	// Add the history up to the last 10 messages
	for _, m := range previousMessages {
		var role string
		switch m.Role {
		case "assistant":
			role = openai.ChatMessageRoleAssistant
		case "system":
			role = openai.ChatMessageRoleSystem
		default:
			role = openai.ChatMessageRoleUser
		}

		chatMessages = append(chatMessages, openai.ChatCompletionMessage{
			Role:    role,
			Content: m.Content,
		})
	}

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini, // Default fallback model, "gpt-4o-mini" acts as a fast/lightweight endpoint
			Messages: chatMessages,
		},
	)

	if err != nil {
		log.Printf("ChatCompletion error: %v\n", err)
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}
