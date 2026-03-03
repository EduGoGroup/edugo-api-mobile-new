package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	"github.com/EduGoGroup/edugo-shared/logger"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// AssessmentManagementService handles assessment CRUD and question management.
type AssessmentManagementService struct {
	assessmentRepo      repository.AssessmentRepository
	mongoAssessmentRepo repository.MongoAssessmentRepository
	log                 logger.Logger
}

// NewAssessmentManagementService creates a new AssessmentManagementService.
func NewAssessmentManagementService(
	assessmentRepo repository.AssessmentRepository,
	mongoAssessmentRepo repository.MongoAssessmentRepository,
	log logger.Logger,
) *AssessmentManagementService {
	return &AssessmentManagementService{
		assessmentRepo:      assessmentRepo,
		mongoAssessmentRepo: mongoAssessmentRepo,
		log:                 log,
	}
}

// ListAssessments returns a paginated list of assessments for the given school.
func (s *AssessmentManagementService) ListAssessments(ctx context.Context, schoolID uuid.UUID, req dto.ListAssessmentsRequest) (*dto.PaginatedResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit

	var searchFields []string
	if req.SearchFields != "" {
		for _, f := range strings.Split(req.SearchFields, ",") {
			trimmed := strings.TrimSpace(f)
			if trimmed != "" {
				searchFields = append(searchFields, trimmed)
			}
		}
	}

	filter := repository.AssessmentFilter{
		SchoolID:     &schoolID,
		Status:       req.Status,
		Limit:        req.Limit,
		Offset:       offset,
		Search:       req.Search,
		SearchFields: searchFields,
	}

	assessments, total, err := s.assessmentRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AssessmentManagementResponse, len(assessments))
	for i, a := range assessments {
		items[i] = toManagementResponse(&a)
	}

	return &dto.PaginatedResponse{
		Data:  items,
		Total: total,
		Page:  req.Page,
		Limit: filter.Limit,
	}, nil
}

// GetAssessmentDetail returns the full assessment detail with questions.
func (s *AssessmentManagementService) GetAssessmentDetail(ctx context.Context, id uuid.UUID) (*dto.AssessmentDetailResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	resp := &dto.AssessmentDetailResponse{
		AssessmentManagementResponse: toManagementResponse(assessment),
		Questions:                   []dto.TeacherQuestionResponse{},
	}

	if assessment.MongoDocumentID != "" {
		mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
		if err != nil {
			// For the management/teacher view, questions are essential.
			// Return error for non-NotFound failures.
			s.log.Error("failed to fetch questions from MongoDB",
				"error", err, "assessment_id", id, "mongo_document_id", assessment.MongoDocumentID)
			return nil, errors.NewInternalError("failed to fetch questions", err)
		}
		resp.Questions = toTeacherQuestions(mongoDoc.Questions)
	}

	return resp, nil
}

// CreateAssessment creates a manual assessment with an empty MongoDB document.
func (s *AssessmentManagementService) CreateAssessment(ctx context.Context, req dto.CreateAssessmentRequest, schoolID, userID uuid.UUID) (*dto.AssessmentManagementResponse, error) {
	now := time.Now()

	// Determine material_id for the MongoDB document
	materialIDStr := ""
	if req.MaterialID != nil {
		materialIDStr = req.MaterialID.String()
	}

	// Create MongoDB document first (empty questions for manual assessment)
	mongoDoc := &mongoentities.MaterialAssessment{
		MaterialID:     materialIDStr,
		Questions:      []mongoentities.Question{},
		TotalQuestions: 0,
		TotalPoints:    0,
		Version:        1,
		AIModel:        "manual",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	mongoID, err := s.mongoAssessmentRepo.Create(ctx, mongoDoc)
	if err != nil {
		return nil, errors.NewInternalError("failed to create mongo assessment document", err)
	}

	// Create PostgreSQL record
	title := req.Title
	passThreshold := 70
	if req.PassThreshold != nil {
		passThreshold = *req.PassThreshold
	}

	assessment := &pgentities.Assessment{
		ID:              uuid.New(),
		MaterialID:      req.MaterialID,
		MongoDocumentID: mongoID,
		SchoolID:        &schoolID,
		CreatedByUserID: &userID,
		QuestionsCount:  0,
		Title:           &title,
		PassThreshold:   &passThreshold,
		MaxAttempts:     req.MaxAttempts,
		TimeLimitMinutes: req.TimeLimitMinutes,
		Status:          "draft",
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.assessmentRepo.Create(ctx, assessment); err != nil {
		// Compensating delete: remove the orphaned MongoDB document
		if delErr := s.mongoAssessmentRepo.Delete(ctx, mongoID); delErr != nil {
			s.log.Error("failed to rollback mongo assessment after PG create failure",
				"error", delErr, "mongo_id", mongoID)
		}
		return nil, err
	}

	s.log.Info("manual assessment created",
		"assessment_id", assessment.ID,
		"school_id", schoolID,
		"created_by", userID,
	)

	resp := toManagementResponse(assessment)
	return &resp, nil
}

// UpdateAssessment updates an assessment (draft only).
func (s *AssessmentManagementService) UpdateAssessment(ctx context.Context, id uuid.UUID, req dto.UpdateAssessmentRequest) (*dto.AssessmentManagementResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "draft" {
		return nil, errors.NewBusinessRuleError("only draft assessments can be edited")
	}

	if req.Title != nil {
		assessment.Title = req.Title
	}
	if req.PassThreshold != nil {
		assessment.PassThreshold = req.PassThreshold
	}
	if req.MaxAttempts != nil {
		assessment.MaxAttempts = req.MaxAttempts
	}
	if req.TimeLimitMinutes != nil {
		assessment.TimeLimitMinutes = req.TimeLimitMinutes
	}

	if err := s.assessmentRepo.Update(ctx, assessment); err != nil {
		return nil, err
	}

	resp := toManagementResponse(assessment)
	return &resp, nil
}

// PublishAssessment changes status from draft/generated to published.
func (s *AssessmentManagementService) PublishAssessment(ctx context.Context, id uuid.UUID) (*dto.AssessmentManagementResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "draft" && assessment.Status != "generated" {
		return nil, errors.NewBusinessRuleError(
			fmt.Sprintf("cannot publish assessment in '%s' status, must be 'draft' or 'generated'", assessment.Status))
	}

	if assessment.QuestionsCount == 0 {
		return nil, errors.NewBusinessRuleError("cannot publish assessment with no questions")
	}

	assessment.Status = "published"
	if err := s.assessmentRepo.Update(ctx, assessment); err != nil {
		return nil, err
	}

	s.log.Info("assessment published", "assessment_id", id)
	resp := toManagementResponse(assessment)
	return &resp, nil
}

// ArchiveAssessment changes status from published to archived.
func (s *AssessmentManagementService) ArchiveAssessment(ctx context.Context, id uuid.UUID) (*dto.AssessmentManagementResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "published" {
		return nil, errors.NewBusinessRuleError("only published assessments can be archived")
	}

	assessment.Status = "archived"
	if err := s.assessmentRepo.Update(ctx, assessment); err != nil {
		return nil, err
	}

	s.log.Info("assessment archived", "assessment_id", id)
	resp := toManagementResponse(assessment)
	return &resp, nil
}

// DeleteAssessment soft-deletes a draft assessment.
func (s *AssessmentManagementService) DeleteAssessment(ctx context.Context, id uuid.UUID) error {
	assessment, err := s.assessmentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if assessment.Status != "draft" {
		return errors.NewBusinessRuleError("only draft assessments can be deleted")
	}

	return s.assessmentRepo.SoftDelete(ctx, id)
}

// --- Question Management ---

// GetQuestions returns all questions for an assessment (teacher view with correct answers).
func (s *AssessmentManagementService) GetQuestions(ctx context.Context, assessmentID uuid.UUID) ([]dto.TeacherQuestionResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	if assessment.MongoDocumentID == "" {
		return []dto.TeacherQuestionResponse{}, nil
	}

	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, err
	}

	return toTeacherQuestions(mongoDoc.Questions), nil
}

// AddQuestion adds a question to an assessment (draft only).
func (s *AssessmentManagementService) AddQuestion(ctx context.Context, assessmentID uuid.UUID, req dto.CreateQuestionRequest) ([]dto.TeacherQuestionResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "draft" {
		return nil, errors.NewBusinessRuleError("can only add questions to draft assessments")
	}

	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, err
	}

	newQuestion := mongoentities.Question{
		QuestionID:    uuid.New().String(),
		QuestionText:  req.QuestionText,
		QuestionType:  req.QuestionType,
		CorrectAnswer: req.CorrectAnswer,
		Explanation:   req.Explanation,
		Points:        req.Points,
		Difficulty:    req.Difficulty,
		Tags:          req.Tags,
	}

	if len(req.Options) > 0 {
		options := make([]mongoentities.Option, len(req.Options))
		for i, o := range req.Options {
			optID := o.OptionID
			if optID == "" {
				optID = fmt.Sprintf("opt_%d", i)
			}
			options[i] = mongoentities.Option{
				OptionID:   optID,
				OptionText: o.OptionText,
			}
		}
		newQuestion.Options = options
	}

	questions := append(mongoDoc.Questions, newQuestion)
	totalPoints := calcTotalPoints(questions)

	if err := s.mongoAssessmentRepo.ReplaceQuestions(ctx, assessment.MongoDocumentID, questions, totalPoints); err != nil {
		return nil, err
	}

	if err := s.assessmentRepo.UpdateQuestionsCount(ctx, assessmentID, len(questions)); err != nil {
		return nil, errors.NewInternalError("failed to update questions count", err)
	}

	return toTeacherQuestions(questions), nil
}

// UpdateQuestion updates a question at a given index (draft only).
func (s *AssessmentManagementService) UpdateQuestion(ctx context.Context, assessmentID uuid.UUID, idx int, req dto.UpdateQuestionRequest) ([]dto.TeacherQuestionResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "draft" {
		return nil, errors.NewBusinessRuleError("can only edit questions in draft assessments")
	}

	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, err
	}

	if idx < 0 || idx >= len(mongoDoc.Questions) {
		return nil, errors.NewValidationError("question index out of range")
	}

	q := &mongoDoc.Questions[idx]
	if req.QuestionText != nil {
		q.QuestionText = *req.QuestionText
	}
	if req.QuestionType != nil {
		q.QuestionType = *req.QuestionType
	}
	if req.CorrectAnswer != nil {
		q.CorrectAnswer = *req.CorrectAnswer
	}
	if req.Explanation != nil {
		q.Explanation = *req.Explanation
	}
	if req.Points != nil {
		q.Points = *req.Points
	}
	if req.Difficulty != nil {
		q.Difficulty = *req.Difficulty
	}
	if req.Tags != nil {
		q.Tags = req.Tags
	}
	if req.Options != nil {
		options := make([]mongoentities.Option, len(req.Options))
		for i, o := range req.Options {
			optID := o.OptionID
			if optID == "" {
				optID = fmt.Sprintf("opt_%d", i)
			}
			options[i] = mongoentities.Option{
				OptionID:   optID,
				OptionText: o.OptionText,
			}
		}
		q.Options = options
	}

	totalPoints := calcTotalPoints(mongoDoc.Questions)

	if err := s.mongoAssessmentRepo.ReplaceQuestions(ctx, assessment.MongoDocumentID, mongoDoc.Questions, totalPoints); err != nil {
		return nil, err
	}

	return toTeacherQuestions(mongoDoc.Questions), nil
}

// DeleteQuestion removes a question at a given index (draft only).
func (s *AssessmentManagementService) DeleteQuestion(ctx context.Context, assessmentID uuid.UUID, idx int) ([]dto.TeacherQuestionResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "draft" {
		return nil, errors.NewBusinessRuleError("can only delete questions from draft assessments")
	}

	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, err
	}

	if idx < 0 || idx >= len(mongoDoc.Questions) {
		return nil, errors.NewValidationError("question index out of range")
	}

	questions := append(mongoDoc.Questions[:idx], mongoDoc.Questions[idx+1:]...)
	totalPoints := calcTotalPoints(questions)

	if err := s.mongoAssessmentRepo.ReplaceQuestions(ctx, assessment.MongoDocumentID, questions, totalPoints); err != nil {
		return nil, err
	}

	if err := s.assessmentRepo.UpdateQuestionsCount(ctx, assessmentID, len(questions)); err != nil {
		return nil, errors.NewInternalError("failed to update questions count", err)
	}

	return toTeacherQuestions(questions), nil
}

// --- Helpers ---

func toManagementResponse(a *pgentities.Assessment) dto.AssessmentManagementResponse {
	title := ""
	if a.Title != nil {
		title = *a.Title
	}
	return dto.AssessmentManagementResponse{
		ID:               a.ID,
		MaterialID:       a.MaterialID,
		Title:            title,
		QuestionsCount:   a.QuestionsCount,
		PassThreshold:    a.PassThreshold,
		MaxAttempts:      a.MaxAttempts,
		TimeLimitMinutes: a.TimeLimitMinutes,
		Status:           a.Status,
		SchoolID:         a.SchoolID,
		CreatedByUserID:  a.CreatedByUserID,
		CreatedAt:        a.CreatedAt,
		UpdatedAt:        a.UpdatedAt,
	}
}

func toTeacherQuestions(questions []mongoentities.Question) []dto.TeacherQuestionResponse {
	result := make([]dto.TeacherQuestionResponse, len(questions))
	for i, q := range questions {
		options := make([]dto.OptionResponse, len(q.Options))
		for j, opt := range q.Options {
			options[j] = dto.OptionResponse{
				OptionID:   opt.OptionID,
				OptionText: opt.OptionText,
			}
		}
		result[i] = dto.TeacherQuestionResponse{
			QuestionID:    q.QuestionID,
			QuestionText:  q.QuestionText,
			QuestionType:  q.QuestionType,
			Options:       options,
			CorrectAnswer: q.CorrectAnswer,
			Explanation:   q.Explanation,
			Points:        q.Points,
			Difficulty:    q.Difficulty,
		}
	}
	return result
}

func calcTotalPoints(questions []mongoentities.Question) int {
	total := 0
	for _, q := range questions {
		total += q.Points
	}
	return total
}
