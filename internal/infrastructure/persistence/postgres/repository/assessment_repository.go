package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// AssessmentRepository implements repository.AssessmentRepository using PostgreSQL.
type AssessmentRepository struct {
	db *sql.DB
}

// NewAssessmentRepository creates a new AssessmentRepository.
func NewAssessmentRepository(db *sql.DB) *AssessmentRepository {
	return &AssessmentRepository{db: db}
}

// GetByID retrieves an assessment by its ID.
func (r *AssessmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Assessment, error) {
	query := `
		SELECT id, material_id, mongo_document_id, questions_count, total_questions,
			title, pass_threshold, max_attempts, time_limit_minutes,
			status, created_at, updated_at, deleted_at
		FROM assessment
		WHERE id = $1 AND deleted_at IS NULL`

	var a pgentities.Assessment
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.MaterialID, &a.MongoDocumentID, &a.QuestionsCount, &a.TotalQuestions,
		&a.Title, &a.PassThreshold, &a.MaxAttempts, &a.TimeLimitMinutes,
		&a.Status, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("assessment")
		}
		return nil, errors.NewDatabaseError("get assessment by id", err)
	}
	return &a, nil
}

// GetByMaterialID retrieves the assessment for a given material.
func (r *AssessmentRepository) GetByMaterialID(ctx context.Context, materialID uuid.UUID) (*pgentities.Assessment, error) {
	query := `
		SELECT id, material_id, mongo_document_id, questions_count, total_questions,
			title, pass_threshold, max_attempts, time_limit_minutes,
			status, created_at, updated_at, deleted_at
		FROM assessment
		WHERE material_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT 1`

	var a pgentities.Assessment
	err := r.db.QueryRowContext(ctx, query, materialID).Scan(
		&a.ID, &a.MaterialID, &a.MongoDocumentID, &a.QuestionsCount, &a.TotalQuestions,
		&a.Title, &a.PassThreshold, &a.MaxAttempts, &a.TimeLimitMinutes,
		&a.Status, &a.CreatedAt, &a.UpdatedAt, &a.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("assessment")
		}
		return nil, errors.NewDatabaseError("get assessment by material", err)
	}
	return &a, nil
}
