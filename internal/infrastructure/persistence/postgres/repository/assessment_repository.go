package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	domainrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// AssessmentRepository implements repository.AssessmentRepository using PostgreSQL.
type AssessmentRepository struct {
	db *gorm.DB
}

// NewAssessmentRepository creates a new AssessmentRepository.
func NewAssessmentRepository(db *gorm.DB) *AssessmentRepository {
	return &AssessmentRepository{db: db}
}

// GetByID retrieves an assessment by its ID, preloading its material associations.
func (r *AssessmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Assessment, error) {
	var a pgentities.Assessment
	err := r.db.WithContext(ctx).Preload("Materials", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).First(&a, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("assessment")
		}
		return nil, sharedErrors.NewDatabaseError("get assessment by id", err)
	}
	return &a, nil
}

// GetByMaterialID retrieves the assessment for a given material via the junction table.
func (r *AssessmentRepository) GetByMaterialID(ctx context.Context, materialID uuid.UUID) (*pgentities.Assessment, error) {
	var a pgentities.Assessment
	err := r.db.WithContext(ctx).
		Joins("JOIN assessment.assessment_materials am ON am.assessment_id = assessment.assessment.id").
		Where("am.material_id = ?", materialID).
		Order("assessment.assessment.created_at DESC").
		Preload("Materials", func(db *gorm.DB) *gorm.DB {
			return db.Order("sort_order ASC")
		}).
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("assessment")
		}
		return nil, sharedErrors.NewDatabaseError("get assessment by material", err)
	}
	return &a, nil
}

// allowedAssessmentSearchFields defines the whitelist of columns that can be
// used in search queries to prevent SQL injection.
var allowedAssessmentSearchFields = map[string]bool{
	"title":  true,
	"status": true,
}

// List retrieves assessments matching the given filter.
func (r *AssessmentRepository) List(ctx context.Context, filter domainrepo.AssessmentFilter) ([]pgentities.Assessment, int, error) {
	query := r.db.WithContext(ctx).Model(&pgentities.Assessment{})

	if filter.SchoolID != nil {
		query = query.Where("school_id = ?", *filter.SchoolID)
	}
	if filter.Status != nil {
		query = query.Where("status = ?", *filter.Status)
	}
	if filter.Search != "" {
		searchFields := filter.SearchFields
		if len(searchFields) == 0 {
			searchFields = []string{"title"}
		}
		var clauses []string
		var args []interface{}
		for _, field := range searchFields {
			if !allowedAssessmentSearchFields[field] {
				continue
			}
			clauses = append(clauses, field+" ILIKE ?")
			args = append(args, "%"+filter.Search+"%")
		}
		if len(clauses) == 0 {
			// Fallback to title if no valid fields
			clauses = []string{"title ILIKE ?"}
			args = []interface{}{"%" + filter.Search + "%"}
		}
		query = query.Where(strings.Join(clauses, " OR "), args...)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, sharedErrors.NewDatabaseError("count assessments", err)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var assessments []pgentities.Assessment
	err := query.Preload("Materials", func(db *gorm.DB) *gorm.DB {
		return db.Order("sort_order ASC")
	}).Order("created_at DESC").Limit(limit).Offset(filter.Offset).Find(&assessments).Error
	if err != nil {
		return nil, 0, sharedErrors.NewDatabaseError("list assessments", err)
	}
	return assessments, int(total), nil
}

// Create inserts a new assessment.
func (r *AssessmentRepository) Create(ctx context.Context, assessment *pgentities.Assessment) error {
	if err := r.db.WithContext(ctx).Create(assessment).Error; err != nil {
		return sharedErrors.NewDatabaseError("create assessment", err)
	}
	return nil
}

// Update updates an existing assessment.
func (r *AssessmentRepository) Update(ctx context.Context, assessment *pgentities.Assessment) error {
	if err := r.db.WithContext(ctx).Save(assessment).Error; err != nil {
		return sharedErrors.NewDatabaseError("update assessment", err)
	}
	return nil
}

// SoftDelete marks an assessment as deleted.
func (r *AssessmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Model(&pgentities.Assessment{}).Where("id = ?", id).Update("deleted_at", time.Now())
	if result.Error != nil {
		return sharedErrors.NewDatabaseError("soft delete assessment", result.Error)
	}
	if result.RowsAffected == 0 {
		return sharedErrors.NewNotFoundError("assessment")
	}
	return nil
}

// UpdateQuestionsCount updates the questions_count field.
func (r *AssessmentRepository) UpdateQuestionsCount(ctx context.Context, id uuid.UUID, count int) error {
	result := r.db.WithContext(ctx).Model(&pgentities.Assessment{}).Where("id = ?", id).Update("questions_count", count)
	if result.Error != nil {
		return sharedErrors.NewDatabaseError("update questions count", result.Error)
	}
	return nil
}
