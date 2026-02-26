package controllers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
)

func CreateMessage(c *gin.Context) {
	var input models.Message

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
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

func GetMessage(c *gin.Context) {
	id := c.Param("id")
	var message models.Message

	if err := database.DB.First(&message, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	c.JSON(http.StatusOK, message)
}

func UpdateMessage(c *gin.Context) {
	id := c.Param("id")
	var message models.Message

	if err := database.DB.First(&message, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	var input struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&message).Update("Content", input.Content)

	c.JSON(http.StatusOK, message)
}

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
