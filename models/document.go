package models

import (
	"time"

	"gorm.io/gorm"
)

type Document struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	UserID      uint           `gorm:"not null;index" json:"user_id"`
	Title       string         `gorm:"size:255;not null" json:"title"`
	Description string         `gorm:"type:text" json:"description"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	Files       []DocumentFile `json:"files,omitempty"`
	User        User           `json:"user,omitempty"`
}
