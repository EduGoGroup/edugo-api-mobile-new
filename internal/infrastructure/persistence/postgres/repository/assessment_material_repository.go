package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// AssessmentMaterialRepository implements repository.AssessmentMaterialRepository using PostgreSQL.
type AssessmentMaterialRepository struct {
	db *gorm.DB
}

// NewAssessmentMaterialRepository creates a new AssessmentMaterialRepository.
func NewAssessmentMaterialRepository(db *gorm.DB) *AssessmentMaterialRepository {
	return &AssessmentMaterialRepository{db: db}
}

// ReplaceForAssessment deletes all existing materials for an assessment and inserts the new ones.
func (r *AssessmentMaterialRepository) ReplaceForAssessment(ctx context.Context, assessmentID uuid.UUID, materialIDs []uuid.UUID) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete all existing associations
		if err := tx.Where("assessment_id = ?", assessmentID).Delete(&pgentities.AssessmentMaterial{}).Error; err != nil {
			return sharedErrors.NewDatabaseError("delete assessment materials", err)
		}

		// Insert new associations
		for i, matID := range materialIDs {
			am := pgentities.AssessmentMaterial{
				ID:           uuid.New(),
				AssessmentID: assessmentID,
				MaterialID:   matID,
				SortOrder:    i,
			}
			if err := tx.Create(&am).Error; err != nil {
				return sharedErrors.NewDatabaseError("create assessment material", err)
			}
		}

		return nil
	})
}

// GetByAssessment retrieves all material associations for a given assessment.
func (r *AssessmentMaterialRepository) GetByAssessment(ctx context.Context, assessmentID uuid.UUID) ([]pgentities.AssessmentMaterial, error) {
	var materials []pgentities.AssessmentMaterial
	err := r.db.WithContext(ctx).
		Where("assessment_id = ?", assessmentID).
		Order("sort_order ASC").
		Find(&materials).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get assessment materials", err)
	}
	return materials, nil
}

// GetMaterialTitles retrieves titles for the given material IDs from content.materials.
func (r *AssessmentMaterialRepository) GetMaterialTitles(ctx context.Context, materialIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if len(materialIDs) == 0 {
		return map[uuid.UUID]string{}, nil
	}

	type materialRow struct {
		ID    uuid.UUID
		Title string
	}

	var rows []materialRow
	err := r.db.WithContext(ctx).
		Table("content.materials").
		Select("id, title").
		Where("id IN ?", materialIDs).
		Find(&rows).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get material titles", err)
	}

	result := make(map[uuid.UUID]string, len(rows))
	for _, row := range rows {
		result[row.ID] = row.Title
	}
	return result, nil
}

// GetAssessmentIDsByMaterialID returns assessment IDs linked to a material via the junction table.
func (r *AssessmentMaterialRepository) GetAssessmentIDsByMaterialID(ctx context.Context, materialID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.WithContext(ctx).
		Model(&pgentities.AssessmentMaterial{}).
		Where("material_id = ?", materialID).
		Pluck("assessment_id", &ids).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get assessment ids by material", err)
	}
	return ids, nil
}
