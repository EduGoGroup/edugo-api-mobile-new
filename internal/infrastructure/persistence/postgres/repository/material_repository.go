package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// MaterialRepository implements repository.MaterialRepository using PostgreSQL.
type MaterialRepository struct {
	db *gorm.DB
}

// NewMaterialRepository creates a new MaterialRepository.
func NewMaterialRepository(db *gorm.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// Create inserts a new material.
func (r *MaterialRepository) Create(ctx context.Context, m *pgentities.Material) error {
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return sharedErrors.NewDatabaseError("create material", err)
	}
	return nil
}

// GetByID retrieves a material by its ID.
func (r *MaterialRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Material, error) {
	var m pgentities.Material
	err := r.db.WithContext(ctx).First(&m, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("material")
		}
		return nil, sharedErrors.NewDatabaseError("get material", err)
	}
	return &m, nil
}

// List retrieves materials matching the given filter.
func (r *MaterialRepository) List(ctx context.Context, filter repository.MaterialFilter) ([]pgentities.Material, int, error) {
	query := r.db.WithContext(ctx).Model(&pgentities.Material{})

	if filter.SchoolID != nil {
		query = query.Where("school_id = ?", *filter.SchoolID)
	}
	if filter.AuthorID != nil {
		query = query.Where("uploaded_by_teacher_id = ?", *filter.AuthorID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}

	// Apply search using shared ListFilters helper
	searchFilters := sharedrepo.ListFilters{Search: filter.Search, SearchFields: filter.SearchFields}
	query = searchFilters.ApplySearch(query)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, sharedErrors.NewDatabaseError("count materials", err)
	}

	var materials []pgentities.Material
	if err := query.Order("created_at DESC").Limit(filter.Limit).Offset(filter.Offset).Find(&materials).Error; err != nil {
		return nil, 0, sharedErrors.NewDatabaseError("list materials", err)
	}

	return materials, int(total), nil
}

// Update modifies an existing material.
func (r *MaterialRepository) Update(ctx context.Context, m *pgentities.Material) error {
	result := r.db.WithContext(ctx).Save(m)
	if result.Error != nil {
		return sharedErrors.NewDatabaseError("update material", result.Error)
	}
	if result.RowsAffected == 0 {
		return sharedErrors.NewNotFoundError("material")
	}
	return nil
}

// GetWithVersions returns a material with its version history.
func (r *MaterialRepository) GetWithVersions(ctx context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error) {
	material, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	var versions []pgentities.MaterialVersion
	if err := r.db.WithContext(ctx).Where("material_id = ?", id).Order("version_number DESC").Find(&versions).Error; err != nil {
		return nil, nil, sharedErrors.NewDatabaseError("list material versions", err)
	}

	return material, versions, nil
}
