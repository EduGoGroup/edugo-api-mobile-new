package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	"github.com/EduGoGroup/edugo-shared/logger"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/services"
)

// AssessmentService handles assessment and attempt business logic.
type AssessmentService struct {
	assessmentRepo      repository.AssessmentRepository
	attemptRepo         repository.AttemptRepository
	mongoAssessmentRepo repository.MongoAssessmentRepository
	log                 logger.Logger
}

// NewAssessmentService creates a new AssessmentService.
func NewAssessmentService(
	assessmentRepo repository.AssessmentRepository,
	attemptRepo repository.AttemptRepository,
	mongoAssessmentRepo repository.MongoAssessmentRepository,
	log logger.Logger,
) *AssessmentService {
	return &AssessmentService{
		assessmentRepo:      assessmentRepo,
		attemptRepo:         attemptRepo,
		mongoAssessmentRepo: mongoAssessmentRepo,
		log:                 log,
	}
}

// GetAssessmentByMaterialID returns the assessment for a material, with sanitized questions
// (correct answers removed).
func (s *AssessmentService) GetAssessmentByMaterialID(ctx context.Context, materialID uuid.UUID) (*dto.AssessmentResponse, error) {
	assessment, err := s.assessmentRepo.GetByMaterialID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	// Fetch questions from MongoDB
	mongoDoc, err := s.mongoAssessmentRepo.GetByMaterialID(ctx, materialID.String())
	if err != nil {
		s.log.Warn("failed to fetch questions from MongoDB, returning assessment without questions",
			"error", err, "material_id", materialID)
		return &dto.AssessmentResponse{
			ID:             assessment.ID,
			MaterialID:     assessment.MaterialID,
			Title:          assessment.Title,
			QuestionsCount: assessment.QuestionsCount,
			PassThreshold:  assessment.PassThreshold,
			MaxAttempts:    assessment.MaxAttempts,
			TimeLimitMin:   assessment.TimeLimitMinutes,
			Status:         assessment.Status,
			Questions:      []dto.QuestionResponse{},
		}, nil
	}

	// Sanitize: strip correct_answer and explanation from response
	questions := make([]dto.QuestionResponse, len(mongoDoc.Questions))
	for i, q := range mongoDoc.Questions {
		options := make([]dto.OptionResponse, len(q.Options))
		for j, opt := range q.Options {
			options[j] = dto.OptionResponse{
				OptionID:   opt.OptionID,
				OptionText: opt.OptionText,
			}
		}
		questions[i] = dto.QuestionResponse{
			QuestionID:   q.QuestionID,
			QuestionText: q.QuestionText,
			QuestionType: q.QuestionType,
			Options:      options,
			Points:       q.Points,
			Difficulty:   q.Difficulty,
		}
	}

	return &dto.AssessmentResponse{
		ID:             assessment.ID,
		MaterialID:     assessment.MaterialID,
		Title:          assessment.Title,
		QuestionsCount: assessment.QuestionsCount,
		PassThreshold:  assessment.PassThreshold,
		MaxAttempts:    assessment.MaxAttempts,
		TimeLimitMin:   assessment.TimeLimitMinutes,
		Status:         assessment.Status,
		Questions:      questions,
	}, nil
}

// CreateAttempt validates student answers against MongoDB, computes the score,
// and persists the attempt with all answers in PostgreSQL.
func (s *AssessmentService) CreateAttempt(
	ctx context.Context,
	materialID, studentID uuid.UUID,
	req dto.CreateAttemptRequest,
) (*dto.AttemptResponse, error) {
	assessment, err := s.assessmentRepo.GetByMaterialID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	// Enforce max attempts limit
	if assessment.MaxAttempts != nil {
		count, cErr := s.attemptRepo.CountByAssessmentAndStudent(ctx, assessment.ID, studentID)
		if cErr != nil {
			return nil, cErr
		}
		if count >= *assessment.MaxAttempts {
			return nil, errors.NewBusinessRuleError("maximum number of attempts reached")
		}
	}

	// Fetch questions from MongoDB for grading
	mongoDoc, err := s.mongoAssessmentRepo.GetByMaterialID(ctx, materialID.String())
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve questions for grading", err)
	}

	if len(req.Answers) > len(mongoDoc.Questions) {
		return nil, errors.NewValidationError("more answers submitted than questions available")
	}

	now := time.Now()
	attemptID := uuid.New()

	var totalScore, totalMaxScore float64
	answers := make([]pgentities.AssessmentAttemptAnswer, len(req.Answers))

	for i, ans := range req.Answers {
		if ans.QuestionIndex < 0 || ans.QuestionIndex >= len(mongoDoc.Questions) {
			return nil, errors.NewValidationError("invalid question_index")
		}

		question := mongoDoc.Questions[ans.QuestionIndex]
		result := services.GradeAnswer(question, ans.Answer)

		totalScore += result.PointsEarned
		totalMaxScore += result.MaxPoints

		isCorrect := result.IsCorrect
		pointsEarned := result.PointsEarned
		maxPoints := result.MaxPoints
		studentAnswer := ans.Answer

		answers[i] = pgentities.AssessmentAttemptAnswer{
			ID:               uuid.New(),
			AttemptID:        attemptID,
			QuestionIndex:    ans.QuestionIndex,
			StudentAnswer:    &studentAnswer,
			IsCorrect:        &isCorrect,
			PointsEarned:     &pointsEarned,
			MaxPoints:        &maxPoints,
			TimeSpentSeconds: ans.TimeSpentSeconds,
			AnsweredAt:       now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
	}

	var percentage float64
	if totalMaxScore > 0 {
		percentage = (totalScore / totalMaxScore) * 100
	}

	attempt := &pgentities.AssessmentAttempt{
		ID:             attemptID,
		AssessmentID:   assessment.ID,
		StudentID:      studentID,
		StartedAt:      now,
		CompletedAt:    &now,
		Score:          &totalScore,
		MaxScore:       &totalMaxScore,
		Percentage:     &percentage,
		IdempotencyKey: req.IdempotencyKey,
		Status:         "completed",
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.attemptRepo.Create(ctx, attempt, answers); err != nil {
		return nil, err
	}

	s.log.Info("attempt created",
		"attempt_id", attemptID,
		"assessment_id", assessment.ID,
		"student_id", studentID,
		"score", totalScore,
		"percentage", percentage,
	)

	return &dto.AttemptResponse{
		ID:           attempt.ID,
		AssessmentID: attempt.AssessmentID,
		StudentID:    attempt.StudentID,
		Score:        attempt.Score,
		MaxScore:     attempt.MaxScore,
		Percentage:   attempt.Percentage,
		Status:       attempt.Status,
		StartedAt:    attempt.StartedAt,
		CompletedAt:  attempt.CompletedAt,
	}, nil
}

// GetAttemptResult returns a detailed attempt result with correct answers and explanations.
func (s *AssessmentService) GetAttemptResult(ctx context.Context, attemptID uuid.UUID) (*dto.AttemptResultResponse, error) {
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	pgAnswers, err := s.attemptRepo.GetAnswersByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	// Get the assessment to find the material_id for MongoDB lookup
	assessment, err := s.assessmentRepo.GetByID(ctx, attempt.AssessmentID)
	if err != nil {
		return nil, err
	}

	// Fetch questions from MongoDB for explanations and correct answers
	mongoDoc, mongoErr := s.mongoAssessmentRepo.GetByMaterialID(ctx, assessment.MaterialID.String())

	answersResponse := make([]dto.AnswerResultResponse, len(pgAnswers))
	for i, a := range pgAnswers {
		ar := dto.AnswerResultResponse{
			QuestionIndex: a.QuestionIndex,
			StudentAnswer: a.StudentAnswer,
			IsCorrect:     a.IsCorrect,
			PointsEarned:  a.PointsEarned,
			MaxPoints:     a.MaxPoints,
		}

		// Enrich with MongoDB data if available
		if mongoErr == nil && mongoDoc != nil && a.QuestionIndex < len(mongoDoc.Questions) {
			q := mongoDoc.Questions[a.QuestionIndex]
			ar.QuestionText = q.QuestionText
			ar.CorrectAnswer = q.CorrectAnswer
			ar.Explanation = q.Explanation
		}

		answersResponse[i] = ar
	}

	return &dto.AttemptResultResponse{
		ID:           attempt.ID,
		AssessmentID: attempt.AssessmentID,
		StudentID:    attempt.StudentID,
		Score:        attempt.Score,
		MaxScore:     attempt.MaxScore,
		Percentage:   attempt.Percentage,
		Status:       attempt.Status,
		StartedAt:    attempt.StartedAt,
		CompletedAt:  attempt.CompletedAt,
		Answers:      answersResponse,
	}, nil
}

// ListAttemptsByUser returns a paginated list of attempts for a user.
func (s *AssessmentService) ListAttemptsByUser(ctx context.Context, userID uuid.UUID, limit, offset int) (*dto.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	attempts, total, err := s.attemptRepo.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, err
	}

	items := make([]dto.AttemptResponse, len(attempts))
	for i, a := range attempts {
		items[i] = dto.AttemptResponse{
			ID:           a.ID,
			AssessmentID: a.AssessmentID,
			StudentID:    a.StudentID,
			Score:        a.Score,
			MaxScore:     a.MaxScore,
			Percentage:   a.Percentage,
			Status:       a.Status,
			StartedAt:    a.StartedAt,
			CompletedAt:  a.CompletedAt,
		}
	}

	return &dto.PaginatedResponse{
		Data:   items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
