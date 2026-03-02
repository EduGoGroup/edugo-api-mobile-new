package dto

import (
	"time"

	"github.com/google/uuid"
)

// GuardianRelationRequestDTO is the payload for requesting a guardian-student link.
type GuardianRelationRequestDTO struct {
	Identifier       string `json:"identifier" binding:"required"`        // email of the student
	RelationshipType string `json:"relationship_type" binding:"required"` // parent, guardian, tutor, other
}

// GuardianRelationResponse is the API response for a guardian relation.
type GuardianRelationResponse struct {
	ID               uuid.UUID `json:"id"`
	GuardianID       uuid.UUID `json:"guardian_id"`
	GuardianName     string    `json:"guardian_name"`
	StudentID        uuid.UUID `json:"student_id"`
	StudentName      string    `json:"student_name"`
	RelationshipType string    `json:"relationship_type"`
	Status           string    `json:"status"`
	IsPrimary        bool      `json:"is_primary"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// ChildResponse is the API response for a guardian's child.
type ChildResponse struct {
	ID        uuid.UUID `json:"id"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
	Email     string    `json:"email"`
}

// ChildProgressResponse is the API response for a child's academic progress.
type ChildProgressResponse struct {
	ChildID        uuid.UUID `json:"child_id"`
	ChildName      string    `json:"child_name"`
	TotalMaterials int       `json:"total_materials"`
	Completed      int       `json:"completed"`
	AvgScore       float64   `json:"avg_score"`
	CompletionRate float64   `json:"completion_rate"`
}

// GuardianStatsResponse is the API response for guardian aggregate stats.
type GuardianStatsResponse struct {
	ChildrenCount  int     `json:"children_count"`
	TotalMaterials int     `json:"total_materials"`
	AvgScore       float64 `json:"avg_score"`
	CompletionRate float64 `json:"completion_rate"`
}
