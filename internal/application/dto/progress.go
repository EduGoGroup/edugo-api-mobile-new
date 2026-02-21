package dto

import (
	"time"

	"github.com/google/uuid"
)

// UpsertProgressRequest is the payload for updating reading progress.
type UpsertProgressRequest struct {
	MaterialID uuid.UUID `json:"material_id" binding:"required"`
	Percentage int       `json:"percentage" binding:"min=0,max=100"`
	LastPage   int       `json:"last_page" binding:"min=0"`
}

// ProgressResponse is the API response for a progress record.
type ProgressResponse struct {
	MaterialID     uuid.UUID `json:"material_id"`
	UserID         uuid.UUID `json:"user_id"`
	Percentage     int       `json:"percentage"`
	LastPage       int       `json:"last_page"`
	Status         string    `json:"status"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}
