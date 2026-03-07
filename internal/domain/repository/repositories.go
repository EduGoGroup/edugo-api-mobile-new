package repository

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"
)

// MaterialRepository defines operations for materials in PostgreSQL.
type MaterialRepository interface {
	Create(ctx context.Context, material *pgentities.Material) error
	GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Material, error)
	List(ctx context.Context, filter MaterialFilter) ([]pgentities.Material, int, error)
	Update(ctx context.Context, material *pgentities.Material) error
	GetWithVersions(ctx context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error)
}

// MaterialFilter holds query parameters for listing materials.
type MaterialFilter struct {
	SchoolID     *uuid.UUID
	AuthorID     *uuid.UUID
	Status       *string
	Limit        int
	Offset       int
	Search       string
	SearchFields []string
}

// AssessmentFilter holds query parameters for listing assessments.
type AssessmentFilter struct {
	SchoolID     *uuid.UUID
	Status       *string
	Limit        int
	Offset       int
	Search       string
	SearchFields []string
}

// AssessmentRepository defines operations for assessments in PostgreSQL.
type AssessmentRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Assessment, error)
	GetByMaterialID(ctx context.Context, materialID uuid.UUID) (*pgentities.Assessment, error)
	List(ctx context.Context, filter AssessmentFilter) ([]pgentities.Assessment, int, error)
	Create(ctx context.Context, assessment *pgentities.Assessment) error
	Update(ctx context.Context, assessment *pgentities.Assessment) error
	SoftDelete(ctx context.Context, id uuid.UUID) error
	UpdateQuestionsCount(ctx context.Context, id uuid.UUID, count int) error
}

// AssessmentMaterialRepository defines operations for the N:N assessment-material junction table.
type AssessmentMaterialRepository interface {
	ReplaceForAssessment(ctx context.Context, assessmentID uuid.UUID, materialIDs []uuid.UUID) error
	GetByAssessment(ctx context.Context, assessmentID uuid.UUID) ([]pgentities.AssessmentMaterial, error)
	GetMaterialTitles(ctx context.Context, materialIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetAssessmentIDsByMaterialID(ctx context.Context, materialID uuid.UUID) ([]uuid.UUID, error)
}

// AttemptRepository defines operations for assessment attempts in PostgreSQL.
type AttemptRepository interface {
	Create(ctx context.Context, attempt *pgentities.AssessmentAttempt, answers []pgentities.AssessmentAttemptAnswer) error
	CreateAttemptOnly(ctx context.Context, attempt *pgentities.AssessmentAttempt) error
	GetByID(ctx context.Context, id uuid.UUID) (*pgentities.AssessmentAttempt, error)
	GetInProgressByStudentAndAssessment(ctx context.Context, studentID, assessmentID uuid.UUID) (*pgentities.AssessmentAttempt, error)
	GetAnswersByAttemptID(ctx context.Context, attemptID uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, filters sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error)
	CountByAssessmentAndStudent(ctx context.Context, assessmentID, studentID uuid.UUID) (int, error)
	UpsertAnswer(ctx context.Context, answer *pgentities.AssessmentAttemptAnswer) error
	UpdateAttempt(ctx context.Context, attempt *pgentities.AssessmentAttempt) error
	UpdateAnswers(ctx context.Context, answers []pgentities.AssessmentAttemptAnswer) error
}

// ProgressRepository defines operations for reading progress in PostgreSQL.
type ProgressRepository interface {
	Upsert(ctx context.Context, progress *pgentities.Progress) error
	GetByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*pgentities.Progress, error)
}

// ScreenRepository defines operations for dynamic UI screens in PostgreSQL.
type ScreenRepository interface {
	GetScreenByKey(ctx context.Context, screenKey string) (*ScreenComposed, error)
	GetScreensByResourceKey(ctx context.Context, resourceKey string) ([]ScreenComposed, error)
	GetNavigation(ctx context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error)
	UpsertPreferences(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error
}

// ScreenComposed is a joined view of screen template + instance.
type ScreenComposed struct {
	Instance pgentities.ScreenInstance
	Template pgentities.ScreenTemplate
}

// StatsRepository defines operations for statistics queries.
type StatsRepository interface {
	CountMaterials(ctx context.Context, schoolID *uuid.UUID) (int, error)
	CountCompletedProgress(ctx context.Context, schoolID *uuid.UUID) (int, error)
	AverageAttemptScore(ctx context.Context, schoolID *uuid.UUID) (float64, error)
	MaterialStats(ctx context.Context, materialID uuid.UUID) (*MaterialStatsResult, error)
}

// MaterialStatsResult holds computed statistics for a single material.
type MaterialStatsResult struct {
	TotalAttempts  int     `json:"total_attempts"`
	AverageScore   float64 `json:"average_score"`
	CompletionRate float64 `json:"completion_rate"`
	UniqueStudents int     `json:"unique_students"`
}

// MongoAssessmentRepository defines operations for assessment questions in MongoDB.
type MongoAssessmentRepository interface {
	GetByMaterialID(ctx context.Context, materialID string) (*mongoentities.MaterialAssessment, error)
	GetByObjectID(ctx context.Context, objectID string) (*mongoentities.MaterialAssessment, error)
	Create(ctx context.Context, doc *mongoentities.MaterialAssessment) (string, error)
	Delete(ctx context.Context, objectID string) error
	ReplaceQuestions(ctx context.Context, objectID string, questions []mongoentities.Question, totalPoints int) error
}

// MongoSummaryRepository defines operations for material summaries in MongoDB.
type MongoSummaryRepository interface {
	GetByMaterialID(ctx context.Context, materialID string) (*mongoentities.MaterialSummary, error)
}
