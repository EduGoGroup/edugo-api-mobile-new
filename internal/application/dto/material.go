package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateMaterialRequest is the payload for creating a material.
type CreateMaterialRequest struct {
	Title          string     `json:"title" binding:"required"`
	Description    *string    `json:"description"`
	Subject        *string    `json:"subject"`
	Grade          *string    `json:"grade"`
	AcademicUnitID *uuid.UUID `json:"academic_unit_id"`
	IsPublic       bool       `json:"is_public"`
}

// UpdateMaterialRequest is the payload for updating a material.
type UpdateMaterialRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Subject     *string `json:"subject"`
	Grade       *string `json:"grade"`
	Status      *string `json:"status"`
	IsPublic    *bool   `json:"is_public"`
}

// UploadCompleteRequest notifies that a file upload finished.
type UploadCompleteRequest struct {
	FileURL       string `json:"file_url" binding:"required"`
	FileType      string `json:"file_type" binding:"required"`
	FileSizeBytes int64  `json:"file_size_bytes" binding:"required,min=1"`
}

// MaterialResponse is the API response for a material.
type MaterialResponse struct {
	ID                  uuid.UUID  `json:"id"`
	SchoolID            uuid.UUID  `json:"school_id"`
	UploadedByTeacherID uuid.UUID  `json:"uploaded_by_teacher_id"`
	AcademicUnitID      *uuid.UUID `json:"academic_unit_id,omitempty"`
	Title               string     `json:"title"`
	Description         *string    `json:"description,omitempty"`
	Subject             *string    `json:"subject,omitempty"`
	Grade               *string    `json:"grade,omitempty"`
	FileURL             string     `json:"file_url"`
	FileType            string     `json:"file_type"`
	FileSizeBytes       int64      `json:"file_size_bytes"`
	Status              string     `json:"status"`
	IsPublic            bool       `json:"is_public"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

// MaterialVersionResponse is the API response for a material version.
type MaterialVersionResponse struct {
	ID            uuid.UUID `json:"id"`
	MaterialID    uuid.UUID `json:"material_id"`
	VersionNumber int       `json:"version_number"`
	Title         string    `json:"title"`
	ContentURL    string    `json:"content_url"`
	ChangedBy     uuid.UUID `json:"changed_by"`
	CreatedAt     time.Time `json:"created_at"`
}

// MaterialWithVersionsResponse wraps a material with its version history.
type MaterialWithVersionsResponse struct {
	Material MaterialResponse          `json:"material"`
	Versions []MaterialVersionResponse `json:"versions"`
}

// PresignedURLResponse is returned for upload/download URL requests.
type PresignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// ListMaterialsRequest holds query parameters for listing materials.
type ListMaterialsRequest struct {
	SchoolID *uuid.UUID `form:"school_id"`
	AuthorID *uuid.UUID `form:"author_id"`
	Status   *string    `form:"status"`
	Limit    int        `form:"limit,default=20"`
	Offset   int        `form:"offset,default=0"`
}

// PaginatedResponse wraps a list with pagination metadata.
type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int         `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}
