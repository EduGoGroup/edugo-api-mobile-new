package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateAssessmentRequest is the payload for creating a manual assessment.
// Numeric fields use float64 to tolerate both integer and decimal JSON numbers.
// When writing to the database, only MaxAttempts is truncated to int; PassThreshold and TimeLimitMinutes retain decimal precision.
type CreateAssessmentRequest struct {
	Title              string     `json:"title" binding:"required"`
	Description        string     `json:"description"`
	MaterialIDs        []string   `json:"material_ids"`
	PassThreshold      *float64   `json:"pass_threshold" binding:"omitempty,min=0,max=100"`
	MaxAttempts        *float64   `json:"max_attempts" binding:"omitempty,min=1"`
	TimeLimitMinutes   *float64   `json:"time_limit_minutes" binding:"omitempty,min=1"`
	IsTimed            *bool      `json:"is_timed"`
	ShuffleQuestions   *bool      `json:"shuffle_questions"`
	ShowCorrectAnswers *bool      `json:"show_correct_answers"`
	AvailableFrom      *time.Time `json:"available_from"`
	AvailableUntil     *time.Time `json:"available_until"`
}

// UpdateAssessmentRequest is the payload for updating an assessment.
type UpdateAssessmentRequest struct {
	Title              *string    `json:"title,omitempty"`
	Description        *string    `json:"description,omitempty"`
	MaterialIDs        []string   `json:"material_ids,omitempty"`
	PassThreshold      *float64   `json:"pass_threshold,omitempty" binding:"omitempty,min=0,max=100"`
	MaxAttempts        *float64   `json:"max_attempts,omitempty" binding:"omitempty,min=1"`
	TimeLimitMinutes   *float64   `json:"time_limit_minutes,omitempty" binding:"omitempty,min=1"`
	IsTimed            *bool      `json:"is_timed,omitempty"`
	ShuffleQuestions   *bool      `json:"shuffle_questions,omitempty"`
	ShowCorrectAnswers *bool      `json:"show_correct_answers,omitempty"`
	AvailableFrom      *time.Time `json:"available_from,omitempty"`
	AvailableUntil     *time.Time `json:"available_until,omitempty"`
}

// ToInt truncates a float64 pointer to an int pointer (nil-safe).
func ToInt(f *float64) *int {
	if f == nil {
		return nil
	}
	v := int(*f)
	return &v
}

// MaterialSummaryDTO is a lightweight representation of a material (id + title).
type MaterialSummaryDTO struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// AssessmentManagementResponse is the API response for assessment management endpoints.
type AssessmentManagementResponse struct {
	ID                 uuid.UUID            `json:"id"`
	Title              string               `json:"title"`
	Description        string               `json:"description"`
	QuestionsCount     int                  `json:"questions_count"`
	MaterialIDs        []string             `json:"material_ids"`
	Materials          []MaterialSummaryDTO `json:"materials"`
	PassThreshold      *float64             `json:"pass_threshold,omitempty"`
	MaxAttempts        *int                 `json:"max_attempts,omitempty"`
	TimeLimitMinutes   *float64             `json:"time_limit_minutes,omitempty"`
	IsTimed            bool                 `json:"is_timed"`
	ShuffleQuestions   bool                 `json:"shuffle_questions"`
	ShowCorrectAnswers bool                 `json:"show_correct_answers"`
	AvailableFrom      *time.Time           `json:"available_from,omitempty"`
	AvailableUntil     *time.Time           `json:"available_until,omitempty"`
	Status             string               `json:"status"`
	SchoolID           *uuid.UUID           `json:"school_id,omitempty"`
	CreatedByUserID    *uuid.UUID           `json:"created_by_user_id,omitempty"`
	CreatedAt          time.Time            `json:"created_at"`
	UpdatedAt          time.Time            `json:"updated_at"`
}

// ListAssessmentsRequest holds query parameters for listing assessments.
type ListAssessmentsRequest struct {
	Status       *string `form:"status"`
	Page         int     `form:"page,default=1"`
	Limit        int     `form:"limit,default=20"`
	Search       string  `form:"search"`
	SearchFields string  `form:"search_fields"`
}

// AssessmentDetailResponse is the full assessment detail with questions (for teachers).
type AssessmentDetailResponse struct {
	AssessmentManagementResponse
	Questions []TeacherQuestionResponse `json:"questions"`
}

// TeacherQuestionResponse is a question with correct answer visible (for teachers).
type TeacherQuestionResponse struct {
	QuestionID    string           `json:"question_id"`
	QuestionText  string           `json:"question_text"`
	QuestionType  string           `json:"question_type"`
	Options       []OptionResponse `json:"options,omitempty"`
	CorrectAnswer string           `json:"correct_answer"`
	Explanation   string           `json:"explanation"`
	Points        int              `json:"points"`
	Difficulty    string           `json:"difficulty"`
}

// CreateQuestionRequest is the payload for adding a question.
type CreateQuestionRequest struct {
	QuestionText  string          `json:"question_text" binding:"required"`
	QuestionType  string          `json:"question_type" binding:"required,oneof=multiple_choice true_false open"`
	Options       []OptionRequest `json:"options"`
	CorrectAnswer string          `json:"correct_answer" binding:"required"`
	Explanation   string          `json:"explanation"`
	Points        int             `json:"points" binding:"required,min=1"`
	Difficulty    string          `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Tags          []string        `json:"tags"`
}

// OptionRequest is an option in a question create/update request.
type OptionRequest struct {
	OptionID   string `json:"option_id"`
	OptionText string `json:"option_text" binding:"required"`
}

// UpdateQuestionRequest is the payload for updating a question.
type UpdateQuestionRequest struct {
	QuestionText  *string         `json:"question_text"`
	QuestionType  *string         `json:"question_type" binding:"omitempty,oneof=multiple_choice true_false open"`
	Options       []OptionRequest `json:"options"`
	CorrectAnswer *string         `json:"correct_answer"`
	Explanation   *string         `json:"explanation"`
	Points        *int            `json:"points" binding:"omitempty,min=1"`
	Difficulty    *string         `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	Tags          []string        `json:"tags"`
}
