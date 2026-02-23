package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// ProgressRepository implements repository.ProgressRepository using PostgreSQL.
type ProgressRepository struct {
	db *gorm.DB
}

// NewProgressRepository creates a new ProgressRepository.
func NewProgressRepository(db *gorm.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// Upsert inserts or updates a progress record (INSERT ON CONFLICT).
func (r *ProgressRepository) Upsert(ctx context.Context, p *pgentities.Progress) error {
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "material_id"}, {Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"percentage", "last_page", "status", "last_accessed_at", "updated_at"}),
	}).Create(p).Error
	if err != nil {
		return sharedErrors.NewDatabaseError("upsert progress", err)
	}
	return nil
}

// GetByMaterialAndUser retrieves a progress record for a material/user pair.
func (r *ProgressRepository) GetByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*pgentities.Progress, error) {
	var p pgentities.Progress
	err := r.db.WithContext(ctx).Where("material_id = ? AND user_id = ?", materialID, userID).First(&p).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("progress")
		}
		return nil, sharedErrors.NewDatabaseError("get progress", err)
	}
	return &p, nil
}
