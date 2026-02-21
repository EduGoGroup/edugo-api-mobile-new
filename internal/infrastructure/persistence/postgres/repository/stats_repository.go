package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	domrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// StatsRepository implements repository.StatsRepository using PostgreSQL.
type StatsRepository struct {
	db *sql.DB
}

// NewStatsRepository creates a new StatsRepository.
func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// CountMaterials returns the total number of active materials, optionally filtered by school.
func (r *StatsRepository) CountMaterials(ctx context.Context, schoolID *uuid.UUID) (int, error) {
	var count int
	var err error

	if schoolID != nil {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM materials WHERE deleted_at IS NULL AND school_id = $1`,
			*schoolID,
		).Scan(&count)
	} else {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM materials WHERE deleted_at IS NULL`,
		).Scan(&count)
	}

	if err != nil {
		return 0, errors.NewDatabaseError("count materials", err)
	}
	return count, nil
}

// CountCompletedProgress returns the number of progress records with status 'completed'.
func (r *StatsRepository) CountCompletedProgress(ctx context.Context, schoolID *uuid.UUID) (int, error) {
	var count int
	var err error

	if schoolID != nil {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM progress p
			 JOIN materials m ON m.id = p.material_id
			 WHERE p.status = 'completed' AND m.school_id = $1 AND m.deleted_at IS NULL`,
			*schoolID,
		).Scan(&count)
	} else {
		err = r.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM progress WHERE status = 'completed'`,
		).Scan(&count)
	}

	if err != nil {
		return 0, errors.NewDatabaseError("count completed progress", err)
	}
	return count, nil
}

// AverageAttemptScore returns the average percentage across all completed attempts.
func (r *StatsRepository) AverageAttemptScore(ctx context.Context, schoolID *uuid.UUID) (float64, error) {
	var avg sql.NullFloat64
	var err error

	if schoolID != nil {
		err = r.db.QueryRowContext(ctx,
			`SELECT AVG(aa.percentage) FROM assessment_attempt aa
			 JOIN assessment a ON a.id = aa.assessment_id
			 JOIN materials m ON m.id = a.material_id
			 WHERE aa.status = 'completed' AND m.school_id = $1 AND m.deleted_at IS NULL`,
			*schoolID,
		).Scan(&avg)
	} else {
		err = r.db.QueryRowContext(ctx,
			`SELECT AVG(percentage) FROM assessment_attempt WHERE status = 'completed'`,
		).Scan(&avg)
	}

	if err != nil {
		return 0, errors.NewDatabaseError("average attempt score", err)
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}

// MaterialStats returns computed statistics for a specific material.
func (r *StatsRepository) MaterialStats(ctx context.Context, materialID uuid.UUID) (*domrepo.MaterialStatsResult, error) {
	query := `
		SELECT
			COUNT(*) as total_attempts,
			COALESCE(AVG(aa.percentage), 0) as average_score,
			COUNT(DISTINCT aa.student_id) as unique_students
		FROM assessment_attempt aa
		JOIN assessment a ON a.id = aa.assessment_id
		WHERE a.material_id = $1 AND aa.status = 'completed'`

	var result domrepo.MaterialStatsResult
	err := r.db.QueryRowContext(ctx, query, materialID).Scan(
		&result.TotalAttempts,
		&result.AverageScore,
		&result.UniqueStudents,
	)
	if err != nil {
		return nil, errors.NewDatabaseError("material stats", err)
	}

	// Compute completion rate from progress
	completionQuery := `
		SELECT
			CASE WHEN COUNT(*) > 0
				THEN CAST(SUM(CASE WHEN status = 'completed' THEN 1 ELSE 0 END) AS FLOAT) / COUNT(*)
				ELSE 0
			END
		FROM progress
		WHERE material_id = $1`

	if err := r.db.QueryRowContext(ctx, completionQuery, materialID).Scan(&result.CompletionRate); err != nil {
		return nil, errors.NewDatabaseError("material completion rate", err)
	}

	return &result, nil
}
