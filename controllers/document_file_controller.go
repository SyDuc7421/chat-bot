package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"hsduc.com/rag/database"
	"hsduc.com/rag/models"
	"hsduc.com/rag/services"
)

// @Summary      Upload File to Document
// @Description  Upload a file and attach it to a document
// @Tags         Documents
// @Accept       multipart/form-data
// @Produce      json
// @Param        id    path      int   true  "Document ID"
// @Param        file  formData  file  true  "File to upload"
// @Success      201   {object}  models.DocumentFile
// @Security     BearerAuth
// @Router       /api/v1/documents/{id}/files [post]
func UploadDocumentFile(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}

	// Verify the document belongs to the user
	var doc models.Document
	if err := database.DB.Where("id = ? AND user_id = ?", docID, userID).First(&doc).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "File is required"})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer file.Close()

	contentType := fileHeader.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	objectKey := services.BuildObjectKey(uint(docID), fileHeader.Filename)
	if err := services.UploadFile(c.Request.Context(), objectKey, contentType, file, fileHeader.Size); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to upload file"})
		return
	}

	docFile := models.DocumentFile{
		DocumentID:  uint(docID),
		FileName:    fileHeader.Filename,
		ObjectKey:   objectKey,
		ContentType: contentType,
		Size:        fileHeader.Size,
	}
	if err := database.DB.Create(&docFile).Error; err != nil {
		// Clean up the uploaded object if DB insert fails
		_ = services.DeleteFile(c.Request.Context(), objectKey)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
		return
	}

	c.JSON(http.StatusCreated, docFile)
}

// @Summary      Get Download URL
// @Description  Get a presigned download URL for a document file (valid 15 minutes)
// @Tags         Documents
// @Produce      json
// @Param        id      path  int  true  "Document ID"
// @Param        fileId  path  int  true  "File ID"
// @Success      200     {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/documents/{id}/files/{fileId}/download [get]
func GetDocumentFileDownloadURL(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	fileID, err := strconv.Atoi(c.Param("fileId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var docFile models.DocumentFile
	if err := database.DB.
		Joins("JOIN documents ON documents.id = document_files.document_id").
		Where("document_files.id = ? AND document_files.document_id = ? AND documents.user_id = ?", fileID, docID, userID).
		First(&docFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	url, err := services.GetPresignedURL(c.Request.Context(), docFile.ObjectKey, 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate download URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"url": url, "expires_in": "15m"})
}

// @Summary      Delete Document File
// @Description  Delete a file from a document
// @Tags         Documents
// @Produce      json
// @Param        id      path  int  true  "Document ID"
// @Param        fileId  path  int  true  "File ID"
// @Success      200     {object}  map[string]string
// @Security     BearerAuth
// @Router       /api/v1/documents/{id}/files/{fileId} [delete]
func DeleteDocumentFile(c *gin.Context) {
	userID := c.GetUint("userID")
	docID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid document ID"})
		return
	}
	fileID, err := strconv.Atoi(c.Param("fileId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file ID"})
		return
	}

	var docFile models.DocumentFile
	if err := database.DB.
		Joins("JOIN documents ON documents.id = document_files.document_id").
		Where("document_files.id = ? AND document_files.document_id = ? AND documents.user_id = ?", fileID, docID, userID).
		First(&docFile).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	_ = deleteDocumentFileFromStorage(c.Request.Context(), docFile.ObjectKey)

	if err := database.DB.Delete(&docFile).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// deleteDocumentFileFromStorage is a shared helper used by document and file controllers.
func deleteDocumentFileFromStorage(ctx context.Context, objectKey string) error {
	return services.DeleteFile(ctx, objectKey)
}
