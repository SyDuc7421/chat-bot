package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
	"hsduc.com/rag/dtos"
	"hsduc.com/rag/models"
	"hsduc.com/rag/services"
)

// @Summary      Create Message
// @Description  Create a new message
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        body body dtos.CreateMessageRequest true "Message Request"
// @Success      201  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/messages [post]
func CreateMessage(c *gin.Context) {
	var bodyInterface dtos.CreateMessageRequest

	if err := c.ShouldBindJSON(&bodyInterface); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	input := models.Message{
		ConversationID: bodyInterface.ConversationID,
		Role:           bodyInterface.Role,
		Content:        bodyInterface.Content,
	}

	// Verify conversation exists
	userID := c.MustGet("userID").(uint)
	var conversation models.Conversation
	if err := database.DB.Where("id = ? AND user_id = ?", input.ConversationID, userID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	if err := database.DB.Create(&input).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	// Cache the last message in Redis for demonstration
	msgBytes, _ := json.Marshal(input)
	database.Redis.Set(context.Background(), "last_message", msgBytes, 10*time.Minute)

	// If the message is from the user, trigger the AI response
	if input.Role == "user" {
		// Fetch last 10 messages for context
		var previousMessages []models.Message
		database.DB.Where("conversation_id = ?", input.ConversationID).
			Order("created_at desc").
			Limit(10).
			Find(&previousMessages)

		// Reverse to pass in chronological order
		for i, j := 0, len(previousMessages)-1; i < j; i, j = i+1, j-1 {
			previousMessages[i], previousMessages[j] = previousMessages[j], previousMessages[i]
		}

		replyContent, err := services.GetChatbotResponse(previousMessages, []string{})
		if err == nil && replyContent != "" {
			assistantMsg := models.Message{
				ConversationID: input.ConversationID,
				Role:           "assistant",
				Content:        replyContent,
			}
			database.DB.Create(&assistantMsg)

			c.JSON(http.StatusCreated, gin.H{
				"user_message":      input,
				"assistant_message": assistantMsg,
			})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{"user_message": input})
}

// @Summary      Get Messages
// @Description  Get messages for a conversation
// @Tags         Messages
// @Produce      json
// @Param        conversation_id query string true "Conversation ID"
// @Success      200  {array}   models.Message
// @Security     BearerAuth
// @Router       /api/v1/messages [get]
func GetMessages(c *gin.Context) {
	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
		return
	}

	userID := c.MustGet("userID").(uint)

	// Ensure user owns the conversation
	var conversation models.Conversation
	if err := database.DB.Where("id = ? AND user_id = ?", conversationID, userID).First(&conversation).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	var messages []models.Message
	if err := database.DB.Where("conversation_id = ?", conversationID).Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, messages)
}

// @Summary      Get Message
// @Description  Get a message by id
// @Tags         Messages
// @Produce      json
// @Param        id   path      string  true  "Message ID"
// @Success      200  {object}  models.Message
// @Security     BearerAuth
// @Router       /api/v1/messages/{id} [get]
func GetMessage(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)
	var message models.Message
	if err := database.DB.Joins("JOIN conversations on messages.conversation_id = conversations.id").
		Where("messages.id = ? AND conversations.user_id = ?", id, userID).
		First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// @Summary      Update Message
// @Description  Update a message by id
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Message ID"
// @Param        body body dtos.UpdateMessageRequest true "Message Request"
// @Success      200  {object}  models.Message
// @Security     BearerAuth
// @Router       /api/v1/messages/{id} [put]
func UpdateMessage(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)
	var message models.Message
	if err := database.DB.Joins("JOIN conversations on messages.conversation_id = conversations.id").
		Where("messages.id = ? AND conversations.user_id = ?", id, userID).
		First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	var input dtos.UpdateMessageRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := database.DB.Model(&message).Update("Content", input.Content).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	c.JSON(http.StatusOK, message)
}

// @Summary      Delete Message
// @Description  Delete a message by id
// @Tags         Messages
// @Produce      json
// @Param        id   path      string  true  "Message ID"
// @Success      200  {object}  map[string]interface{}
// @Security     BearerAuth
// @Router       /api/v1/messages/{id} [delete]
func DeleteMessage(c *gin.Context) {
	id := c.Param("id")
	userID := c.MustGet("userID").(uint)
	var message models.Message
	if err := database.DB.Joins("JOIN conversations on messages.conversation_id = conversations.id").
		Where("messages.id = ? AND conversations.user_id = ?", id, userID).
		First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	database.DB.Delete(&message)
	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}
