package models

import (
	"time"

	"gorm.io/gorm"
)

type DocumentFile struct {
	ID          uint           `gorm:"primarykey" json:"id"`
	DocumentID  uint           `gorm:"not null;index" json:"document_id"`
	FileName    string         `gorm:"size:255;not null" json:"file_name"`
	ObjectKey   string         `gorm:"size:500;not null" json:"object_key"`
	ContentType string         `gorm:"size:100" json:"content_type"`
	Size        int64          `json:"size"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
