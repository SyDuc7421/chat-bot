package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	ConversationID uint           `gorm:"not null" json:"conversation_id"`
	Role           string         `gorm:"size:50;not null" json:"role"` // e.g., "user", "assistant"
	Content        string         `gorm:"type:text;not null" json:"content"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
}
