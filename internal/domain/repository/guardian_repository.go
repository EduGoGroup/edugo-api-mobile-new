package repository

import (
	"context"

	"github.com/google/uuid"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// GuardianRelationJoined is a guardian relation with resolved user names.
type GuardianRelationJoined struct {
	pgentities.GuardianRelation
	GuardianName string
	StudentName  string
}

// ChildJoined is a student user linked to a guardian.
type ChildJoined struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}

// ChildProgressResult holds computed progress for a single child.
type ChildProgressResult struct {
	TotalMaterials int
	Completed      int
	AvgScore       float64
	CompletionRate float64
}

// GuardianRepository defines operations for guardian relations.
type GuardianRepository interface {
	Create(ctx context.Context, relation *pgentities.GuardianRelation) error
	GetByID(ctx context.Context, id uuid.UUID) (*pgentities.GuardianRelation, error)
	FindByGuardianAndStudent(ctx context.Context, guardianID, studentID uuid.UUID) (*pgentities.GuardianRelation, error)
	ListPendingBySchool(ctx context.Context, schoolID uuid.UUID) ([]GuardianRelationJoined, error)
	ListActiveChildrenByGuardian(ctx context.Context, guardianID uuid.UUID) ([]ChildJoined, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) error
	FindUserByEmail(ctx context.Context, email string) (*UserBasic, error)
	GetChildProgress(ctx context.Context, childID uuid.UUID) (*ChildProgressResult, error)
}

// UserBasic is a minimal user representation for lookups.
type UserBasic struct {
	ID        uuid.UUID
	FirstName string
	LastName  string
	Email     string
}
