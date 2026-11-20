package models

import (
	"time"

	"github.com/google/uuid"
)

type Media struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	EntityType string     `json:"entity_type" db:"entity_type"`
	EntityID   uuid.UUID  `json:"entity_id" db:"entity_id"`
	FilePath   string     `json:"-" db:"file_path"`
	FileURL    string     `json:"file_url"`
	FileName   string     `json:"file_name" db:"file_name"`
	MediaType  string     `json:"media_type" db:"media_type"`
	MimeType   *string    `json:"mime_type" db:"mime_type"`
	FileSize   int64      `json:"file_size" db:"file_size"`
	UploadedBy *uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// ValidEntityTypes lists the allowed values for entity_type.
var ValidEntityTypes = map[string]bool{
	"vehicle":     true,
	"maintenance": true,
	"service":     true,
	"insurance":   true,
	"puc":         true,
	"tyre":        true,
	"battery":     true,
	"expense":     true,
}
