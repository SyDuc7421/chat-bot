package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `gorm:"primarykey" json:"id"`
	Email         string         `gorm:"size:255;not null;uniqueIndex" json:"email"`
	PasswordHash  string         `gorm:"size:255;not null" json:"-"` // Hidden from JSON response
	Name          string         `gorm:"size:255" json:"name"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
	Conversations []Conversation `json:"conversations,omitempty"`
}
