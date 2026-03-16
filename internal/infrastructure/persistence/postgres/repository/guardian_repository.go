package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// GuardianRepository implements repository.GuardianRepository using PostgreSQL.
type GuardianRepository struct {
	db *gorm.DB
}

// NewGuardianRepository creates a new GuardianRepository.
func NewGuardianRepository(db *gorm.DB) *GuardianRepository {
	return &GuardianRepository{db: db}
}

// Create inserts a new guardian relation.
func (r *GuardianRepository) Create(ctx context.Context, relation *pgentities.GuardianRelation) error {
	if err := r.db.WithContext(ctx).Create(relation).Error; err != nil {
		return sharedErrors.NewDatabaseError("create guardian relation", err)
	}
	return nil
}

// GetByID retrieves a guardian relation by its ID.
func (r *GuardianRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.GuardianRelation, error) {
	var rel pgentities.GuardianRelation
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&rel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("guardian relation")
		}
		return nil, sharedErrors.NewDatabaseError("get guardian relation", err)
	}
	return &rel, nil
}

// GetByIDAndSchool retrieves a guardian relation scoped to a school (via student membership).
func (r *GuardianRepository) GetByIDAndSchool(ctx context.Context, id, schoolID uuid.UUID) (*pgentities.GuardianRelation, error) {
	var rel pgentities.GuardianRelation
	err := r.db.WithContext(ctx).
		Raw(`SELECT gr.*
			FROM academic.guardian_relations gr
			JOIN academic.memberships m ON m.user_id = gr.student_id
			WHERE gr.id = ? AND m.school_id = ?
			LIMIT 1`, id, schoolID).
		Scan(&rel).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get guardian relation by school", err)
	}
	if rel.ID == uuid.Nil {
		return nil, sharedErrors.NewNotFoundError("guardian relation")
	}
	return &rel, nil
}

// FindByGuardianAndStudent finds an existing relation between guardian and student.
func (r *GuardianRepository) FindByGuardianAndStudent(ctx context.Context, guardianID, studentID uuid.UUID) (*pgentities.GuardianRelation, error) {
	var rel pgentities.GuardianRelation
	err := r.db.WithContext(ctx).
		Where("guardian_id = ? AND student_id = ?", guardianID, studentID).
		First(&rel).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, sharedErrors.NewDatabaseError("find guardian relation", err)
	}
	return &rel, nil
}

// ListPendingBySchool returns pending relations for a school with resolved names.
func (r *GuardianRepository) ListPendingBySchool(ctx context.Context, schoolID uuid.UUID) ([]repository.GuardianRelationJoined, error) {
	var results []repository.GuardianRelationJoined

	err := r.db.WithContext(ctx).
		Raw(`SELECT gr.*,
			CONCAT(gu.first_name, ' ', gu.last_name) AS guardian_name,
			CONCAT(su.first_name, ' ', su.last_name) AS student_name
			FROM academic.guardian_relations gr
			JOIN auth.users gu ON gu.id = gr.guardian_id
			JOIN auth.users su ON su.id = gr.student_id
			JOIN academic.memberships m ON m.user_id = gr.student_id
			WHERE gr.status = ? AND gr.is_active = ? AND m.school_id = ?
			ORDER BY gr.created_at DESC`, "pending", true, schoolID).
		Scan(&results).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("list pending guardian relations", err)
	}
	return results, nil
}

// ListActiveChildrenByGuardian returns active children for a guardian.
func (r *GuardianRepository) ListActiveChildrenByGuardian(ctx context.Context, guardianID uuid.UUID) ([]repository.ChildJoined, error) {
	var results []repository.ChildJoined

	err := r.db.WithContext(ctx).
		Raw(`SELECT u.id, u.first_name, u.last_name, u.email
			FROM academic.guardian_relations gr
			JOIN auth.users u ON u.id = gr.student_id
			WHERE gr.guardian_id = ? AND gr.status = ? AND gr.is_active = ?`, guardianID, "active", true).
		Scan(&results).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("list guardian children", err)
	}
	return results, nil
}

// UpdateStatus updates the status of a guardian relation.
func (r *GuardianRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) error {
	result := r.db.WithContext(ctx).
		Model(&pgentities.GuardianRelation{}).
		Where("id = ?", id).
		Update("status", status)
	if result.Error != nil {
		return sharedErrors.NewDatabaseError("update guardian relation status", result.Error)
	}
	if result.RowsAffected == 0 {
		return sharedErrors.NewNotFoundError("guardian relation")
	}
	return nil
}

// FindUserByEmail finds a user by email.
func (r *GuardianRepository) FindUserByEmail(ctx context.Context, email string) (*repository.UserBasic, error) {
	var user repository.UserBasic
	err := r.db.WithContext(ctx).
		Table("auth.users").
		Select("id, first_name, last_name, email").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, sharedErrors.NewDatabaseError("find user by email", err)
	}
	return &user, nil
}

// GetChildProgress returns progress stats for a child.
func (r *GuardianRepository) GetChildProgress(ctx context.Context, childID uuid.UUID) (*repository.ChildProgressResult, error) {
	var result repository.ChildProgressResult

	// Total materials assigned (via memberships → academic_units → materials)
	err := r.db.WithContext(ctx).
		Table("content.progress").
		Select("COUNT(*) as total_materials, "+
			"COUNT(CASE WHEN status = 'completed' THEN 1 END) as completed, "+
			"COALESCE(COUNT(CASE WHEN status = 'completed' THEN 1 END)::float / NULLIF(COUNT(*), 0), 0) as completion_rate").
		Where("user_id = ?", childID).
		Scan(&result).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get child progress", err)
	}

	// Average score from assessment attempts
	var avgScore struct {
		AvgScore float64
	}
	err = r.db.WithContext(ctx).
		Table("assessment.assessment_attempt").
		Select("COALESCE(AVG(score), 0) as avg_score").
		Where("student_id = ?", childID).
		Scan(&avgScore).Error
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get child avg score", err)
	}
	result.AvgScore = avgScore.AvgScore

	return &result, nil
}
