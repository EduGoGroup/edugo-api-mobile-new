package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	domrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// StatsRepository implements repository.StatsRepository using PostgreSQL.
type StatsRepository struct {
	db *gorm.DB
}

// NewStatsRepository creates a new StatsRepository.
func NewStatsRepository(db *gorm.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// CountMaterials returns the total number of active materials, optionally filtered by school.
func (r *StatsRepository) CountMaterials(ctx context.Context, schoolID *uuid.UUID) (int, error) {
	query := r.db.WithContext(ctx).Model(&pgentities.Material{})
	if schoolID != nil {
		query = query.Where("school_id = ?", *schoolID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return 0, sharedErrors.NewDatabaseError("count materials", err)
	}
	return int(count), nil
}

// CountCompletedProgress returns the number of progress records with status 'completed'.
func (r *StatsRepository) CountCompletedProgress(ctx context.Context, schoolID *uuid.UUID) (int, error) {
	var count int64

	query := r.db.WithContext(ctx).Model(&pgentities.Progress{}).Where("content.progress.status = 'completed'")
	if schoolID != nil {
		query = query.
			Joins("JOIN content.materials m ON m.id = content.progress.material_id").
			Where("m.school_id = ? AND m.deleted_at IS NULL", *schoolID)
	}

	if err := query.Count(&count).Error; err != nil {
		return 0, sharedErrors.NewDatabaseError("count completed progress", err)
	}
	return int(count), nil
}

// AverageAttemptScore returns the average percentage across all completed attempts.
func (r *StatsRepository) AverageAttemptScore(ctx context.Context, schoolID *uuid.UUID) (float64, error) {
	var avg sql.NullFloat64
	var err error

	if schoolID != nil {
		err = r.db.WithContext(ctx).Raw(queryAvgAttemptScoreBySchool, *schoolID).Scan(&avg).Error
	} else {
		err = r.db.WithContext(ctx).
			Model(&pgentities.AssessmentAttempt{}).
			Where("status = 'completed'").
			Select("AVG(percentage)").
			Scan(&avg).Error
	}

	if err != nil {
		return 0, sharedErrors.NewDatabaseError("average attempt score", err)
	}
	if !avg.Valid {
		return 0, nil
	}
	return avg.Float64, nil
}

// MaterialStats returns computed statistics for a specific material.
func (r *StatsRepository) MaterialStats(ctx context.Context, materialID uuid.UUID) (*domrepo.MaterialStatsResult, error) {
	var result domrepo.MaterialStatsResult
	if err := r.db.WithContext(ctx).Raw(queryMaterialAttemptStats, materialID).Scan(&result).Error; err != nil {
		return nil, sharedErrors.NewDatabaseError("material stats", err)
	}

	if err := r.db.WithContext(ctx).Raw(queryMaterialCompletionRate, materialID).Scan(&result.CompletionRate).Error; err != nil {
		return nil, sharedErrors.NewDatabaseError("material completion rate", err)
	}

	return &result, nil
}
