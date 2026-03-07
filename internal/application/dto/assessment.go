package dto

import (
	"time"

	"github.com/google/uuid"
)

// QuestionResponse is a sanitized question without the correct answer.
type QuestionResponse struct {
	QuestionID   string           `json:"question_id"`
	QuestionText string           `json:"question_text"`
	QuestionType string           `json:"question_type"`
	Options      []OptionResponse `json:"options,omitempty"`
	Points       int              `json:"points"`
	Difficulty   string           `json:"difficulty"`
}

// OptionResponse is a single answer option.
type OptionResponse struct {
	OptionID   string `json:"option_id"`
	OptionText string `json:"option_text"`
}

// AssessmentResponse is the assessment payload returned to students.
type AssessmentResponse struct {
	ID               uuid.UUID          `json:"id"`
	Title            *string            `json:"title,omitempty"`
	QuestionsCount   int                `json:"questions_count"`
	PassThreshold    *float64           `json:"pass_threshold,omitempty"`
	MaxAttempts      *int               `json:"max_attempts,omitempty"`
	TimeLimitMin     *float64           `json:"time_limit_minutes,omitempty"`
	IsTimed          bool               `json:"is_timed"`
	ShuffleQuestions bool               `json:"shuffle_questions"`
	Status           string             `json:"status"`
	Questions        []QuestionResponse `json:"questions"`
}

// CreateAttemptRequest is the payload for submitting answers.
type CreateAttemptRequest struct {
	Answers        []AnswerSubmission `json:"answers" binding:"required,min=1"`
	IdempotencyKey *string            `json:"idempotency_key"`
}

// AnswerSubmission is a single answer in an attempt.
type AnswerSubmission struct {
	QuestionIndex    int    `json:"question_index"`
	Answer           string `json:"answer"`
	TimeSpentSeconds *int   `json:"time_spent_seconds"`
}

// AttemptResponse is the API response after creating an attempt.
type AttemptResponse struct {
	ID           uuid.UUID  `json:"id"`
	AssessmentID uuid.UUID  `json:"assessment_id"`
	StudentID    uuid.UUID  `json:"student_id"`
	Score        *float64   `json:"score,omitempty"`
	MaxScore     *float64   `json:"max_score,omitempty"`
	Percentage   *float64   `json:"percentage,omitempty"`
	Status       string     `json:"status"`
	StartedAt    time.Time  `json:"started_at"`
	CompletedAt  *time.Time `json:"completed_at,omitempty"`
}

// AttemptResultResponse is the detailed result of an attempt with feedback.
type AttemptResultResponse struct {
	ID           uuid.UUID              `json:"id"`
	AssessmentID uuid.UUID              `json:"assessment_id"`
	StudentID    uuid.UUID              `json:"student_id"`
	Score        *float64               `json:"score"`
	MaxScore     *float64               `json:"max_score"`
	Percentage   *float64               `json:"percentage"`
	Status       string                 `json:"status"`
	StartedAt    time.Time              `json:"started_at"`
	CompletedAt  *time.Time             `json:"completed_at,omitempty"`
	Answers      []AnswerResultResponse `json:"answers"`
}

// AnswerResultResponse is a single answer with feedback.
type AnswerResultResponse struct {
	QuestionIndex int      `json:"question_index"`
	QuestionText  string   `json:"question_text"`
	StudentAnswer *string  `json:"student_answer"`
	CorrectAnswer string   `json:"correct_answer"`
	IsCorrect     *bool    `json:"is_correct"`
	PointsEarned  *float64 `json:"points_earned"`
	MaxPoints     *float64 `json:"max_points"`
	Explanation   string   `json:"explanation"`
}

// StartAttemptResponse is the response after starting a progressive attempt.
type StartAttemptResponse struct {
	AttemptID         uuid.UUID          `json:"attempt_id"`
	AssessmentID      uuid.UUID          `json:"assessment_id"`
	Title             *string            `json:"title,omitempty"`
	QuestionsCount    int                `json:"questions_count"`
	PassThreshold     *float64           `json:"pass_threshold,omitempty"`
	MaxAttempts       *int               `json:"max_attempts,omitempty"`
	RemainingAttempts *int               `json:"remaining_attempts,omitempty"`
	TimeLimitMin      *float64           `json:"time_limit_minutes,omitempty"`
	IsTimed           bool               `json:"is_timed"`
	ShuffleQuestions  bool               `json:"shuffle_questions"`
	Questions         []QuestionResponse `json:"questions"`
	StartedAt         time.Time          `json:"started_at"`
}

// SaveAnswerRequest is the payload for saving a single answer progressively.
type SaveAnswerRequest struct {
	Answer           string `json:"answer" binding:"required"`
	TimeSpentSeconds *int   `json:"time_spent_seconds"`
}

// SaveAnswerResponse is the response after saving a single answer.
type SaveAnswerResponse struct {
	QuestionIndex int       `json:"question_index"`
	Saved         bool      `json:"saved"`
	AnsweredAt    time.Time `json:"answered_at"`
}

// SubmitAttemptRequest is the payload for finalizing an attempt.
type SubmitAttemptRequest struct {
	Answers []AnswerSubmission `json:"answers,omitempty"`
}
