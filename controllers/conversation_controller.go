package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
)

func CreateConversation(c *gin.Context) {
	var input struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conversation := models.Conversation{Title: input.Title}
	if err := database.DB.Create(&conversation).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	c.JSON(http.StatusCreated, conversation)
}

func GetConversations(c *gin.Context) {
	var conversations []models.Conversation
	if err := database.DB.Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func GetConversation(c *gin.Context) {
	id := c.Param("id")
	var conversation models.Conversation

	if err := database.DB.Preload("Messages").First(&conversation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

func UpdateConversation(c *gin.Context) {
	id := c.Param("id")
	var conversation models.Conversation

	if err := database.DB.First(&conversation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	var input struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&conversation).Update("Title", input.Title)

	c.JSON(http.StatusOK, conversation)
}

func DeleteConversation(c *gin.Context) {
	id := c.Param("id")
	var conversation models.Conversation

	if err := database.DB.First(&conversation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	database.DB.Delete(&conversation)

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}
