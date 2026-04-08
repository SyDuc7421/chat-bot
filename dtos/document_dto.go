package dtos

type CreateDocumentRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
}

type UpdateDocumentRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}
