package dtos

type CreateMessageRequest struct {
	ConversationID uint   `json:"conversation_id" binding:"required"`
	Role           string `json:"role" binding:"required,oneof=user assistant system"`
	Content        string `json:"content" binding:"required"`
}

type UpdateMessageRequest struct {
	Content string `json:"content" binding:"required"`
}
