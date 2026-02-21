package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// AttemptRepository implements repository.AttemptRepository using PostgreSQL.
type AttemptRepository struct {
	db *sql.DB
}

// NewAttemptRepository creates a new AttemptRepository.
func NewAttemptRepository(db *sql.DB) *AttemptRepository {
	return &AttemptRepository{db: db}
}

// Create inserts an attempt and its answers in a single transaction.
func (r *AttemptRepository) Create(ctx context.Context, attempt *pgentities.AssessmentAttempt, answers []pgentities.AssessmentAttemptAnswer) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.NewDatabaseError("begin transaction", err)
	}
	defer func() { _ = tx.Rollback() }()

	attemptQuery := `
		INSERT INTO assessment_attempt (
			id, assessment_id, student_id, started_at, completed_at,
			score, max_score, percentage, time_spent_seconds,
			idempotency_key, status, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`

	_, err = tx.ExecContext(ctx, attemptQuery,
		attempt.ID, attempt.AssessmentID, attempt.StudentID,
		attempt.StartedAt, attempt.CompletedAt,
		attempt.Score, attempt.MaxScore, attempt.Percentage,
		attempt.TimeSpentSeconds, attempt.IdempotencyKey,
		attempt.Status, attempt.CreatedAt, attempt.UpdatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("insert attempt", err)
	}

	answerQuery := `
		INSERT INTO assessment_attempt_answer (
			id, attempt_id, question_index, student_answer,
			is_correct, points_earned, max_points,
			time_spent_seconds, answered_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)`

	for i, a := range answers {
		_, err = tx.ExecContext(ctx, answerQuery,
			a.ID, a.AttemptID, a.QuestionIndex, a.StudentAnswer,
			a.IsCorrect, a.PointsEarned, a.MaxPoints,
			a.TimeSpentSeconds, a.AnsweredAt, a.CreatedAt, a.UpdatedAt,
		)
		if err != nil {
			return errors.NewDatabaseError(fmt.Sprintf("insert answer %d", i), err)
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.NewDatabaseError("commit transaction", err)
	}
	return nil
}

// GetByID retrieves an attempt by its ID.
func (r *AttemptRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.AssessmentAttempt, error) {
	query := `
		SELECT id, assessment_id, student_id, started_at, completed_at,
			score, max_score, percentage, time_spent_seconds,
			idempotency_key, status, created_at, updated_at
		FROM assessment_attempt
		WHERE id = $1`

	var a pgentities.AssessmentAttempt
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&a.ID, &a.AssessmentID, &a.StudentID, &a.StartedAt, &a.CompletedAt,
		&a.Score, &a.MaxScore, &a.Percentage, &a.TimeSpentSeconds,
		&a.IdempotencyKey, &a.Status, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("attempt")
		}
		return nil, errors.NewDatabaseError("get attempt", err)
	}
	return &a, nil
}

// GetAnswersByAttemptID returns all answers for an attempt.
func (r *AttemptRepository) GetAnswersByAttemptID(ctx context.Context, attemptID uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
	query := `
		SELECT id, attempt_id, question_index, student_answer,
			is_correct, points_earned, max_points,
			time_spent_seconds, answered_at, created_at, updated_at
		FROM assessment_attempt_answer
		WHERE attempt_id = $1
		ORDER BY question_index`

	rows, err := r.db.QueryContext(ctx, query, attemptID)
	if err != nil {
		return nil, errors.NewDatabaseError("list answers", err)
	}
	defer rows.Close()

	var answers []pgentities.AssessmentAttemptAnswer
	for rows.Next() {
		var a pgentities.AssessmentAttemptAnswer
		if err := rows.Scan(
			&a.ID, &a.AttemptID, &a.QuestionIndex, &a.StudentAnswer,
			&a.IsCorrect, &a.PointsEarned, &a.MaxPoints,
			&a.TimeSpentSeconds, &a.AnsweredAt, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, errors.NewDatabaseError("scan answer", err)
		}
		answers = append(answers, a)
	}
	return answers, rows.Err()
}

// ListByUserID returns paginated attempts for a user.
func (r *AttemptRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]pgentities.AssessmentAttempt, int, error) {
	countQuery := `SELECT COUNT(*) FROM assessment_attempt WHERE student_id = $1`
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.NewDatabaseError("count attempts", err)
	}

	query := `
		SELECT id, assessment_id, student_id, started_at, completed_at,
			score, max_score, percentage, time_spent_seconds,
			idempotency_key, status, created_at, updated_at
		FROM assessment_attempt
		WHERE student_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("list attempts by user", err)
	}
	defer rows.Close()

	var attempts []pgentities.AssessmentAttempt
	for rows.Next() {
		var a pgentities.AssessmentAttempt
		if err := rows.Scan(
			&a.ID, &a.AssessmentID, &a.StudentID, &a.StartedAt, &a.CompletedAt,
			&a.Score, &a.MaxScore, &a.Percentage, &a.TimeSpentSeconds,
			&a.IdempotencyKey, &a.Status, &a.CreatedAt, &a.UpdatedAt,
		); err != nil {
			return nil, 0, errors.NewDatabaseError("scan attempt", err)
		}
		attempts = append(attempts, a)
	}
	return attempts, total, rows.Err()
}

// CountByAssessmentAndStudent returns the number of attempts a student has made for an assessment.
func (r *AttemptRepository) CountByAssessmentAndStudent(ctx context.Context, assessmentID, studentID uuid.UUID) (int, error) {
	query := `SELECT COUNT(*) FROM assessment_attempt WHERE assessment_id = $1 AND student_id = $2`
	var count int
	if err := r.db.QueryRowContext(ctx, query, assessmentID, studentID).Scan(&count); err != nil {
		return 0, errors.NewDatabaseError("count attempts", err)
	}
	return count, nil
}
