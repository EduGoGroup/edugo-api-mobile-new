package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// AssessmentRepository implements repository.AssessmentRepository using PostgreSQL.
type AssessmentRepository struct {
	db *gorm.DB
}

// NewAssessmentRepository creates a new AssessmentRepository.
func NewAssessmentRepository(db *gorm.DB) *AssessmentRepository {
	return &AssessmentRepository{db: db}
}

// GetByID retrieves an assessment by its ID.
func (r *AssessmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Assessment, error) {
	var a pgentities.Assessment
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("assessment")
		}
		return nil, sharedErrors.NewDatabaseError("get assessment by id", err)
	}
	return &a, nil
}

// GetByMaterialID retrieves the assessment for a given material.
func (r *AssessmentRepository) GetByMaterialID(ctx context.Context, materialID uuid.UUID) (*pgentities.Assessment, error) {
	var a pgentities.Assessment
	err := r.db.WithContext(ctx).Where("material_id = ?", materialID).Order("created_at DESC").First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("assessment")
		}
		return nil, sharedErrors.NewDatabaseError("get assessment by material", err)
	}
	return &a, nil
}
