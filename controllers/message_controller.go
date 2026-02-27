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
)

// @Summary      Create Message
// @Description  Create a new message
// @Tags         Messages
// @Accept       json
// @Produce      json
// @Param        body body dtos.CreateMessageRequest true "Message Request"
// @Success      201  {object}  models.Message
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
	var conversation models.Conversation
	if err := database.DB.First(&conversation, input.ConversationID).Error; err != nil {
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

	c.JSON(http.StatusCreated, input)
}

// @Summary      Get Messages
// @Description  Get messages for a conversation
// @Tags         Messages
// @Produce      json
// @Param        conversation_id query string true "Conversation ID"
// @Success      200  {array}   models.Message
// @Router       /api/v1/messages [get]
func GetMessages(c *gin.Context) {
	conversationID := c.Query("conversation_id")
	if conversationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "conversation_id is required"})
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
// @Router       /api/v1/messages/{id} [get]
func GetMessage(c *gin.Context) {
	id := c.Param("id")
	var message models.Message

	if err := database.DB.First(&message, id).Error; err != nil {
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
// @Router       /api/v1/messages/{id} [put]
func UpdateMessage(c *gin.Context) {
	id := c.Param("id")
	var message models.Message

	if err := database.DB.First(&message, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	var input dtos.UpdateMessageRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&message).Update("Content", input.Content)

	c.JSON(http.StatusOK, message)
}

// @Summary      Delete Message
// @Description  Delete a message by id
// @Tags         Messages
// @Produce      json
// @Param        id   path      string  true  "Message ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/messages/{id} [delete]
func DeleteMessage(c *gin.Context) {
	id := c.Param("id")
	var message models.Message

	if err := database.DB.First(&message, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	database.DB.Delete(&message)
	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}
