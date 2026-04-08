package controllers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
	"hsduc.com/rag/dtos"
	"hsduc.com/rag/models"
)

// @Summary      Create Document
// @Description  Create a new document
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Param        body body dtos.CreateDocumentRequest true "Create Document Request"
// @Success      201  {object}  models.Document
// @Security     BearerAuth
// @Router       /api/v1/documents [post]
func CreateDocument(c *gin.Context) {
	userID := c.GetUint("userID")

	var input dtos.CreateDocumentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	doc := models.Document{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
	}
	if err := database.DB.Create(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document"})
		return
	}

	c.JSON(http.StatusCreated, doc)
}

// @Summary      Get Documents
// @Description  Get all documents for the authenticated user
// @Tags         Documents
// @Produce      json
// @Success      200  {array}   models.Document
// @Security     BearerAuth
// @Router       /api/v1/documents [get]
func GetDocuments(c *gin.Context) {
	userID := c.GetUint("userID")

	var docs []models.Document
	if err := database.DB.Where("user_id = ?", userID).Preload("Files").Find(&docs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve documents"})
		return
	}

	c.JSON(http.StatusOK, docs)
}

// @Summary      Get Document
// @Description  Get a document by ID with its files
// @Tags         Documents
// @Produce      json
// @Param        id   path      int  true  "Document ID"
// @Success      200  {object}  models.Document
// @Security     BearerAuth
// @Router       /api/v1/documents/{id} [get]
func GetDocument(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var doc models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Files").First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// @Summary      Update Document
// @Description  Update a document's title or description
// @Tags         Documents
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "Document ID"
// @Param        body body dtos.UpdateDocumentRequest true "Update Document Request"
// @Success      200  {object}  models.Document
// @Security     BearerAuth
// @Router       /api/v1/documents/{id} [put]
func UpdateDocument(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var doc models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	var input dtos.UpdateDocumentRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if input.Title != "" {
		doc.Title = input.Title
	}
	if input.Description != "" {
		doc.Description = input.Description
	}

	if err := database.DB.Save(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}

	c.JSON(http.StatusOK, doc)
}

// @Summary      Delete Document
// @Description  Delete a document and all its files
// @Tags         Documents
// @Produce      json
// @Param        id   path      int  true  "Document ID"
// @Success      200  {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/documents/{id} [delete]
func DeleteDocument(c *gin.Context) {
	userID := c.GetUint("userID")
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	var doc models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", id, userID).Preload("Files").First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	// Delete all associated files from MinIO
	for _, f := range doc.Files {
		_ = deleteDocumentFileFromStorage(c.Request.Context(), f.ObjectKey)
	}

	if err := database.DB.Delete(&doc).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}
