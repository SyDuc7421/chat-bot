package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
	"hsduc.com/rag/dtos"
	"hsduc.com/rag/models"
)

// @Summary      Create Conversation
// @Description  Create a new conversation
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Param        body body dtos.CreateConversationRequest true "Conversation Request"
// @Success      201  {object}  models.Conversation
// @Router       /api/v1/conversations [post]
func CreateConversation(c *gin.Context) {
	var input dtos.CreateConversationRequest

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

// @Summary      Get Conversations
// @Description  Get all conversations
// @Tags         Conversations
// @Produce      json
// @Success      200  {array}   models.Conversation
// @Router       /api/v1/conversations [get]
func GetConversations(c *gin.Context) {
	var conversations []models.Conversation
	if err := database.DB.Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

// @Summary      Get Conversation
// @Description  Get a conversation by id
// @Tags         Conversations
// @Produce      json
// @Param        id   path      string  true  "Conversation ID"
// @Success      200  {object}  models.Conversation
// @Router       /api/v1/conversations/{id} [get]
func GetConversation(c *gin.Context) {
	id := c.Param("id")
	var conversation models.Conversation

	if err := database.DB.Preload("Messages").First(&conversation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	c.JSON(http.StatusOK, conversation)
}

// @Summary      Update Conversation
// @Description  Update a conversation by id
// @Tags         Conversations
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Conversation ID"
// @Param        body body dtos.UpdateConversationRequest true "Conversation Request"
// @Success      200  {object}  models.Conversation
// @Router       /api/v1/conversations/{id} [put]
func UpdateConversation(c *gin.Context) {
	id := c.Param("id")
	var conversation models.Conversation

	if err := database.DB.First(&conversation, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	var input dtos.UpdateConversationRequest

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	database.DB.Model(&conversation).Update("Title", input.Title)

	c.JSON(http.StatusOK, conversation)
}

// @Summary      Delete Conversation
// @Description  Delete a conversation by id
// @Tags         Conversations
// @Produce      json
// @Param        id   path      string  true  "Conversation ID"
// @Success      200  {object}  map[string]interface{}
// @Router       /api/v1/conversations/{id} [delete]
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
