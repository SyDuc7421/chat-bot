package dtos

type CreateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}

type UpdateConversationRequest struct {
	Title string `json:"title" binding:"required"`
}
