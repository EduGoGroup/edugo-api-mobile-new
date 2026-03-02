package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateAssessmentRequest is the payload for creating a manual assessment.
type CreateAssessmentRequest struct {
	Title            string     `json:"title" binding:"required"`
	MaterialID       *uuid.UUID `json:"material_id"`
	PassThreshold    *int       `json:"pass_threshold" binding:"omitempty,min=0,max=100"`
	MaxAttempts      *int       `json:"max_attempts" binding:"omitempty,min=1"`
	TimeLimitMinutes *int       `json:"time_limit_minutes" binding:"omitempty,min=1"`
}

// UpdateAssessmentRequest is the payload for updating an assessment.
type UpdateAssessmentRequest struct {
	Title            *string `json:"title"`
	PassThreshold    *int    `json:"pass_threshold" binding:"omitempty,min=0,max=100"`
	MaxAttempts      *int    `json:"max_attempts" binding:"omitempty,min=1"`
	TimeLimitMinutes *int    `json:"time_limit_minutes" binding:"omitempty,min=1"`
}

// AssessmentManagementResponse is the API response for assessment management endpoints.
type AssessmentManagementResponse struct {
	ID               uuid.UUID  `json:"id"`
	MaterialID       *uuid.UUID `json:"material_id,omitempty"`
	Title            string     `json:"title"`
	QuestionsCount   int        `json:"questions_count"`
	PassThreshold    *int       `json:"pass_threshold,omitempty"`
	MaxAttempts      *int       `json:"max_attempts,omitempty"`
	TimeLimitMinutes *int       `json:"time_limit_minutes,omitempty"`
	Status           string     `json:"status"`
	SchoolID         *uuid.UUID `json:"school_id,omitempty"`
	CreatedByUserID  *uuid.UUID `json:"created_by_user_id,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
}

// ListAssessmentsRequest holds query parameters for listing assessments.
type ListAssessmentsRequest struct {
	Status       *string `form:"status"`
	Limit        int     `form:"limit,default=20"`
	Offset       int     `form:"offset,default=0"`
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
	QuestionText  string           `json:"question_text" binding:"required"`
	QuestionType  string           `json:"question_type" binding:"required,oneof=multiple_choice true_false open"`
	Options       []OptionRequest  `json:"options"`
	CorrectAnswer string           `json:"correct_answer" binding:"required"`
	Explanation   string           `json:"explanation"`
	Points        int              `json:"points" binding:"required,min=1"`
	Difficulty    string           `json:"difficulty" binding:"required,oneof=easy medium hard"`
	Tags          []string         `json:"tags"`
}

// OptionRequest is an option in a question create/update request.
type OptionRequest struct {
	OptionID   string `json:"option_id"`
	OptionText string `json:"option_text" binding:"required"`
}

// UpdateQuestionRequest is the payload for updating a question.
type UpdateQuestionRequest struct {
	QuestionText  *string          `json:"question_text"`
	QuestionType  *string          `json:"question_type" binding:"omitempty,oneof=multiple_choice true_false open"`
	Options       []OptionRequest  `json:"options"`
	CorrectAnswer *string          `json:"correct_answer"`
	Explanation   *string          `json:"explanation"`
	Points        *int             `json:"points" binding:"omitempty,min=1"`
	Difficulty    *string          `json:"difficulty" binding:"omitempty,oneof=easy medium hard"`
	Tags          []string         `json:"tags"`
}
