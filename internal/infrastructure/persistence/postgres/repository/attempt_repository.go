package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// AttemptRepository implements repository.AttemptRepository using PostgreSQL.
type AttemptRepository struct {
	db *gorm.DB
}

// NewAttemptRepository creates a new AttemptRepository.
func NewAttemptRepository(db *gorm.DB) *AttemptRepository {
	return &AttemptRepository{db: db}
}

// Create inserts an attempt and its answers in a single transaction.
func (r *AttemptRepository) Create(ctx context.Context, attempt *pgentities.AssessmentAttempt, answers []pgentities.AssessmentAttemptAnswer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(attempt).Error; err != nil {
			return sharedErrors.NewDatabaseError("insert attempt", err)
		}
		if len(answers) > 0 {
			if err := tx.Create(&answers).Error; err != nil {
				return sharedErrors.NewDatabaseError("insert answers", err)
			}
		}
		return nil
	})
}

// GetByID retrieves an attempt by its ID.
func (r *AttemptRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.AssessmentAttempt, error) {
	var a pgentities.AssessmentAttempt
	err := r.db.WithContext(ctx).First(&a, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, sharedErrors.NewNotFoundError("attempt")
		}
		return nil, sharedErrors.NewDatabaseError("get attempt", err)
	}
	return &a, nil
}

// GetAnswersByAttemptID returns all answers for an attempt.
func (r *AttemptRepository) GetAnswersByAttemptID(ctx context.Context, attemptID uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
	var answers []pgentities.AssessmentAttemptAnswer
	if err := r.db.WithContext(ctx).Where("attempt_id = ?", attemptID).Order("question_index").Find(&answers).Error; err != nil {
		return nil, sharedErrors.NewDatabaseError("list answers", err)
	}
	return answers, nil
}

// ListByUserID returns paginated attempts for a user.
func (r *AttemptRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, filters sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error) {
	query := r.db.WithContext(ctx).Model(&pgentities.AssessmentAttempt{}).Where("student_id = ?", userID)
	query = filters.ApplySearch(query)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, sharedErrors.NewDatabaseError("count attempts", err)
	}

	var attempts []pgentities.AssessmentAttempt
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&attempts).Error; err != nil {
		return nil, 0, sharedErrors.NewDatabaseError("list attempts by user", err)
	}

	return attempts, int(total), nil
}

// CountByAssessmentAndStudent returns the number of attempts a student has made for an assessment.
func (r *AttemptRepository) CountByAssessmentAndStudent(ctx context.Context, assessmentID, studentID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&pgentities.AssessmentAttempt{}).Where("assessment_id = ? AND student_id = ?", assessmentID, studentID).Count(&count).Error; err != nil {
		return 0, sharedErrors.NewDatabaseError("count attempts", err)
	}
	return int(count), nil
}

// CreateAttemptOnly inserts an attempt without answers.
func (r *AttemptRepository) CreateAttemptOnly(ctx context.Context, attempt *pgentities.AssessmentAttempt) error {
	if err := r.db.WithContext(ctx).Create(attempt).Error; err != nil {
		return sharedErrors.NewDatabaseError("insert attempt", err)
	}
	return nil
}

// GetInProgressByStudentAndAssessment returns an in-progress attempt for a student+assessment, or nil if none.
func (r *AttemptRepository) GetInProgressByStudentAndAssessment(ctx context.Context, studentID, assessmentID uuid.UUID) (*pgentities.AssessmentAttempt, error) {
	var a pgentities.AssessmentAttempt
	err := r.db.WithContext(ctx).
		Where("student_id = ? AND assessment_id = ? AND status = ?", studentID, assessmentID, "in_progress").
		First(&a).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, sharedErrors.NewDatabaseError("get in-progress attempt", err)
	}
	return &a, nil
}

// UpsertAnswer inserts or updates an answer keyed by attempt_id + question_index.
func (r *AttemptRepository) UpsertAnswer(ctx context.Context, answer *pgentities.AssessmentAttemptAnswer) error {
	return r.db.WithContext(ctx).
		Where("attempt_id = ? AND question_index = ?", answer.AttemptID, answer.QuestionIndex).
		Assign(pgentities.AssessmentAttemptAnswer{
			StudentAnswer:    answer.StudentAnswer,
			TimeSpentSeconds: answer.TimeSpentSeconds,
			AnsweredAt:       answer.AnsweredAt,
			UpdatedAt:        answer.UpdatedAt,
		}).
		FirstOrCreate(answer).Error
}

// UpdateAttempt updates an existing attempt.
func (r *AttemptRepository) UpdateAttempt(ctx context.Context, attempt *pgentities.AssessmentAttempt) error {
	if err := r.db.WithContext(ctx).Save(attempt).Error; err != nil {
		return sharedErrors.NewDatabaseError("update attempt", err)
	}
	return nil
}

// UpdateAnswers batch-updates answers.
func (r *AttemptRepository) UpdateAnswers(ctx context.Context, answers []pgentities.AssessmentAttemptAnswer) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i := range answers {
			if err := tx.Save(&answers[i]).Error; err != nil {
				return sharedErrors.NewDatabaseError("update answer", err)
			}
		}
		return nil
	})
}
