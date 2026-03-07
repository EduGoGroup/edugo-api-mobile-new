package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/audit"
	"github.com/EduGoGroup/edugo-shared/common/errors"
	"github.com/EduGoGroup/edugo-shared/logger"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
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
	auditLogger         audit.AuditLogger
}

// NewAssessmentService creates a new AssessmentService.
func NewAssessmentService(
	assessmentRepo repository.AssessmentRepository,
	attemptRepo repository.AttemptRepository,
	mongoAssessmentRepo repository.MongoAssessmentRepository,
	log logger.Logger,
	auditLogger audit.AuditLogger,
) *AssessmentService {
	return &AssessmentService{
		assessmentRepo:      assessmentRepo,
		attemptRepo:         attemptRepo,
		mongoAssessmentRepo: mongoAssessmentRepo,
		log:                 log,
		auditLogger:         auditLogger,
	}
}

// GetAssessmentByMaterialID returns the assessment for a material, with sanitized questions
// (correct answers removed).
func (s *AssessmentService) GetAssessmentByMaterialID(ctx context.Context, materialID uuid.UUID) (*dto.AssessmentResponse, error) {
	assessment, err := s.assessmentRepo.GetByMaterialID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	// Fetch questions from MongoDB using the mongo_document_id
	var mongoDoc *mongoentities.MaterialAssessment
	if assessment.MongoDocumentID != "" {
		mongoDoc, err = s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
		if err != nil {
			s.log.Warn("failed to fetch questions from MongoDB, returning assessment without questions",
				"error", err, "material_id", materialID)
		}
	}

	if mongoDoc == nil {
		return &dto.AssessmentResponse{
			ID:               assessment.ID,
			Title:            assessment.Title,
			QuestionsCount:   assessment.QuestionsCount,
			PassThreshold:    assessment.PassThreshold,
			MaxAttempts:      assessment.MaxAttempts,
			TimeLimitMin:     assessment.TimeLimitMinutes,
			IsTimed:          assessment.IsTimed,
			ShuffleQuestions: assessment.ShuffleQuestions,
			Status:           assessment.Status,
			Questions:        []dto.QuestionResponse{},
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
		ID:               assessment.ID,
		Title:            assessment.Title,
		QuestionsCount:   assessment.QuestionsCount,
		PassThreshold:    assessment.PassThreshold,
		MaxAttempts:      assessment.MaxAttempts,
		TimeLimitMin:     assessment.TimeLimitMinutes,
		IsTimed:          assessment.IsTimed,
		ShuffleQuestions: assessment.ShuffleQuestions,
		Status:           assessment.Status,
		Questions:        questions,
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

	// Fetch questions from MongoDB for grading using the mongo_document_id
	if assessment.MongoDocumentID == "" {
		return nil, errors.NewInternalError("assessment has no questions document", nil)
	}
	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
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

	// Audit: assessment attempt taken
	if s.auditLogger != nil {
		if err := s.auditLogger.Log(ctx, audit.AuditEvent{
			Action:       "take",
			ResourceType: "assessment",
			ResourceID:   assessment.ID.String(),
			Severity:     audit.SeverityInfo,
			Category:     audit.CategoryData,
		}); err != nil {
			s.log.Error("failed to emit audit event for assessment attempt",
				"error", err,
				"assessment_id", assessment.ID,
				"attempt_id", attemptID,
				"student_id", studentID,
			)
		}
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

	// Get the assessment to find the MongoDB reference
	assessment, err := s.assessmentRepo.GetByID(ctx, attempt.AssessmentID)
	if err != nil {
		return nil, err
	}

	// Fetch questions from MongoDB for explanations and correct answers.
	var mongoDoc *mongoentities.MaterialAssessment
	var mongoErr error
	if assessment.MongoDocumentID != "" {
		mongoDoc, mongoErr = s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	}
	if mongoErr != nil {
		s.log.Warn("failed to fetch questions from MongoDB for attempt result enrichment",
			"error", mongoErr, "attempt_id", attemptID, "assessment_id", assessment.ID)
	}

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
func (s *AssessmentService) ListAttemptsByUser(ctx context.Context, userID uuid.UUID, page, limit int, filters sharedrepo.ListFilters) (*dto.PaginatedResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if page <= 0 {
		page = 1
	}

	offset := (page - 1) * limit

	attempts, total, err := s.attemptRepo.ListByUserID(ctx, userID, limit, offset, filters)
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
		Data:  items,
		Total: total,
		Page:  page,
		Limit: limit,
	}, nil
}

// StartAttempt starts a new progressive attempt or returns an existing in-progress one.
func (s *AssessmentService) StartAttempt(ctx context.Context, assessmentID, studentID uuid.UUID) (*dto.StartAttemptResponse, error) {
	assessment, err := s.assessmentRepo.GetByID(ctx, assessmentID)
	if err != nil {
		return nil, err
	}

	if assessment.Status != "published" {
		return nil, errors.NewBusinessRuleError("assessment is not published")
	}

	now := time.Now()
	if assessment.AvailableFrom != nil && now.Before(*assessment.AvailableFrom) {
		return nil, errors.NewBusinessRuleError("assessment is not yet available")
	}
	if assessment.AvailableUntil != nil && now.After(*assessment.AvailableUntil) {
		return nil, errors.NewBusinessRuleError("assessment is no longer available")
	}

	// Check for existing in-progress attempt FIRST. If one exists, return it
	// without checking MaxAttempts. Otherwise MaxAttempts=1 with an in-progress
	// attempt would return an error instead of the existing attempt.
	existing, err := s.attemptRepo.GetInProgressByStudentAndAssessment(ctx, studentID, assessment.ID)
	if err != nil {
		return nil, err
	}

	// Load questions from MongoDB
	if assessment.MongoDocumentID == "" {
		return nil, errors.NewInternalError("assessment has no questions document", nil)
	}
	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve questions", err)
	}

	questions := s.sanitizeQuestions(mongoDoc)

	// TODO: ShuffleQuestions is disabled because the shuffled order is not persisted
	// in the attempt. Since answers are graded by questionIndex against the MongoDB
	// document order, shuffling here would cause grading mismatches when the student
	// submits answers. To properly support shuffling, we need to persist the shuffled
	// question order in the attempt record and use it during grading.
	// See: https://github.com/EduGoGroup/edugo-api-mobile-new/pull/23
	// if assessment.ShuffleQuestions {
	// 	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	// 	rng.Shuffle(len(questions), func(i, j int) {
	// 		questions[i], questions[j] = questions[j], questions[i]
	// 	})
	// }

	var remainingAttempts *int
	if existing != nil {
		// Compute remaining attempts for the response
		if assessment.MaxAttempts != nil {
			count, cErr := s.attemptRepo.CountByAssessmentAndStudent(ctx, assessment.ID, studentID)
			if cErr != nil {
				return nil, cErr
			}
			rem := *assessment.MaxAttempts - count
			remainingAttempts = &rem
		}
		return &dto.StartAttemptResponse{
			AttemptID:         existing.ID,
			AssessmentID:      existing.AssessmentID,
			Title:             assessment.Title,
			QuestionsCount:    assessment.QuestionsCount,
			PassThreshold:     assessment.PassThreshold,
			MaxAttempts:       assessment.MaxAttempts,
			RemainingAttempts: remainingAttempts,
			TimeLimitMin:      assessment.TimeLimitMinutes,
			IsTimed:           assessment.IsTimed,
			ShuffleQuestions:  assessment.ShuffleQuestions,
			Questions:         questions,
			StartedAt:         existing.StartedAt,
		}, nil
	}

	// Check max attempts only when creating a new attempt
	if assessment.MaxAttempts != nil {
		count, cErr := s.attemptRepo.CountByAssessmentAndStudent(ctx, assessment.ID, studentID)
		if cErr != nil {
			return nil, cErr
		}
		if count >= *assessment.MaxAttempts {
			return nil, errors.NewBusinessRuleError("maximum number of attempts reached")
		}
		rem := *assessment.MaxAttempts - count
		remainingAttempts = &rem
	}

	attemptID := uuid.New()
	attempt := &pgentities.AssessmentAttempt{
		ID:           attemptID,
		AssessmentID: assessment.ID,
		StudentID:    studentID,
		StartedAt:    now,
		Status:       "in_progress",
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.attemptRepo.CreateAttemptOnly(ctx, attempt); err != nil {
		return nil, err
	}

	// Decrement remaining after creating
	if remainingAttempts != nil {
		rem := *remainingAttempts - 1
		remainingAttempts = &rem
	}

	s.log.Info("progressive attempt started",
		"attempt_id", attemptID,
		"assessment_id", assessment.ID,
		"student_id", studentID,
	)

	return &dto.StartAttemptResponse{
		AttemptID:         attemptID,
		AssessmentID:      assessment.ID,
		Title:             assessment.Title,
		QuestionsCount:    assessment.QuestionsCount,
		PassThreshold:     assessment.PassThreshold,
		MaxAttempts:       assessment.MaxAttempts,
		RemainingAttempts: remainingAttempts,
		TimeLimitMin:      assessment.TimeLimitMinutes,
		IsTimed:           assessment.IsTimed,
		ShuffleQuestions:  assessment.ShuffleQuestions,
		Questions:         questions,
		StartedAt:         now,
	}, nil
}

// SaveAnswer saves a single answer for a progressive attempt.
func (s *AssessmentService) SaveAnswer(ctx context.Context, attemptID uuid.UUID, questionIndex int, studentID uuid.UUID, req dto.SaveAnswerRequest) (*dto.SaveAnswerResponse, error) {
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	if attempt.StudentID != studentID {
		return nil, errors.NewForbiddenError("attempt does not belong to this student")
	}

	if attempt.Status != "in_progress" {
		return nil, errors.NewBusinessRuleError("attempt is not in progress")
	}

	// Check time limit
	if err := s.checkTimeLimit(ctx, attempt); err != nil {
		return nil, err
	}

	// Validate questionIndex against the assessment's questions
	assessment, err := s.assessmentRepo.GetByID(ctx, attempt.AssessmentID)
	if err != nil {
		return nil, err
	}
	if assessment.MongoDocumentID == "" {
		return nil, errors.NewInternalError("assessment has no questions document", nil)
	}
	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve questions", err)
	}
	if questionIndex < 0 || questionIndex >= len(mongoDoc.Questions) {
		return nil, errors.NewValidationError("invalid question_index")
	}

	now := time.Now()
	answer := &pgentities.AssessmentAttemptAnswer{
		ID:               uuid.New(),
		AttemptID:        attemptID,
		QuestionIndex:    questionIndex,
		StudentAnswer:    &req.Answer,
		TimeSpentSeconds: req.TimeSpentSeconds,
		AnsweredAt:       now,
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	if err := s.attemptRepo.UpsertAnswer(ctx, answer); err != nil {
		return nil, errors.NewInternalError("failed to save answer", err)
	}

	return &dto.SaveAnswerResponse{
		QuestionIndex: questionIndex,
		Saved:         true,
		AnsweredAt:    now,
	}, nil
}

// SubmitAttempt finalizes a progressive attempt, grades all answers, and returns the result.
func (s *AssessmentService) SubmitAttempt(ctx context.Context, attemptID, studentID uuid.UUID, req dto.SubmitAttemptRequest) (*dto.AttemptResultResponse, error) {
	attempt, err := s.attemptRepo.GetByID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	if attempt.StudentID != studentID {
		return nil, errors.NewForbiddenError("attempt does not belong to this student")
	}

	if attempt.Status != "in_progress" {
		return nil, errors.NewBusinessRuleError("attempt is not in progress")
	}

	// Check time limit before saving/grading
	if err := s.checkTimeLimit(ctx, attempt); err != nil {
		return nil, err
	}

	// Load assessment for MongoDB reference (needed for validation and grading)
	assessment, err := s.assessmentRepo.GetByID(ctx, attempt.AssessmentID)
	if err != nil {
		return nil, err
	}

	if assessment.MongoDocumentID == "" {
		return nil, errors.NewInternalError("assessment has no questions document", nil)
	}
	mongoDoc, err := s.mongoAssessmentRepo.GetByObjectID(ctx, assessment.MongoDocumentID)
	if err != nil {
		return nil, errors.NewInternalError("failed to retrieve questions for grading", err)
	}

	// Validate QuestionIndex in submitted answers before persisting
	for _, ans := range req.Answers {
		if ans.QuestionIndex < 0 || ans.QuestionIndex >= len(mongoDoc.Questions) {
			return nil, errors.NewValidationError("invalid question_index in submitted answers")
		}
	}

	now := time.Now()

	// Save any remaining answers from the request body
	for _, ans := range req.Answers {
		studentAnswer := ans.Answer
		answer := &pgentities.AssessmentAttemptAnswer{
			ID:               uuid.New(),
			AttemptID:        attemptID,
			QuestionIndex:    ans.QuestionIndex,
			StudentAnswer:    &studentAnswer,
			TimeSpentSeconds: ans.TimeSpentSeconds,
			AnsweredAt:       now,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if err := s.attemptRepo.UpsertAnswer(ctx, answer); err != nil {
			return nil, errors.NewInternalError("failed to save answer during submit", err)
		}
	}

	// Load all saved answers
	pgAnswers, err := s.attemptRepo.GetAnswersByAttemptID(ctx, attemptID)
	if err != nil {
		return nil, err
	}

	// Grade each answer
	var totalScore, totalMaxScore float64
	for i := range pgAnswers {
		a := &pgAnswers[i]
		if a.QuestionIndex < 0 || a.QuestionIndex >= len(mongoDoc.Questions) {
			continue
		}
		question := mongoDoc.Questions[a.QuestionIndex]
		studentAnswer := ""
		if a.StudentAnswer != nil {
			studentAnswer = *a.StudentAnswer
		}
		result := services.GradeAnswer(question, studentAnswer)

		isCorrect := result.IsCorrect
		pointsEarned := result.PointsEarned
		maxPoints := result.MaxPoints

		a.IsCorrect = &isCorrect
		a.PointsEarned = &pointsEarned
		a.MaxPoints = &maxPoints
		a.UpdatedAt = now

		totalScore += result.PointsEarned
		totalMaxScore += result.MaxPoints
	}

	// Update answers with grading results
	if len(pgAnswers) > 0 {
		if err := s.attemptRepo.UpdateAnswers(ctx, pgAnswers); err != nil {
			return nil, err
		}
	}

	// Calculate percentage
	var percentage float64
	if totalMaxScore > 0 {
		percentage = (totalScore / totalMaxScore) * 100
	}

	// Update attempt
	attempt.Status = "completed"
	attempt.Score = &totalScore
	attempt.MaxScore = &totalMaxScore
	attempt.Percentage = &percentage
	attempt.CompletedAt = &now
	attempt.UpdatedAt = now

	if err := s.attemptRepo.UpdateAttempt(ctx, attempt); err != nil {
		return nil, err
	}

	// Audit log
	if s.auditLogger != nil {
		if err := s.auditLogger.Log(ctx, audit.AuditEvent{
			Action:       "submit",
			ResourceType: "assessment",
			ResourceID:   assessment.ID.String(),
			Severity:     audit.SeverityInfo,
			Category:     audit.CategoryData,
		}); err != nil {
			s.log.Error("failed to emit audit event for assessment submit",
				"error", err,
				"assessment_id", assessment.ID,
				"attempt_id", attemptID,
				"student_id", studentID,
			)
		}
	}

	s.log.Info("progressive attempt submitted",
		"attempt_id", attemptID,
		"assessment_id", assessment.ID,
		"student_id", studentID,
		"score", totalScore,
		"percentage", percentage,
	)

	// Build result response
	answersResponse := make([]dto.AnswerResultResponse, len(pgAnswers))
	for i, a := range pgAnswers {
		ar := dto.AnswerResultResponse{
			QuestionIndex: a.QuestionIndex,
			StudentAnswer: a.StudentAnswer,
			IsCorrect:     a.IsCorrect,
			PointsEarned:  a.PointsEarned,
			MaxPoints:     a.MaxPoints,
		}
		if a.QuestionIndex >= 0 && a.QuestionIndex < len(mongoDoc.Questions) {
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

// sanitizeQuestions strips correct answers and explanations from questions.
func (s *AssessmentService) sanitizeQuestions(mongoDoc *mongoentities.MaterialAssessment) []dto.QuestionResponse {
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
	return questions
}

// checkTimeLimit validates that a timed attempt hasn't exceeded its time limit.
func (s *AssessmentService) checkTimeLimit(ctx context.Context, attempt *pgentities.AssessmentAttempt) error {
	assessment, err := s.assessmentRepo.GetByID(ctx, attempt.AssessmentID)
	if err != nil {
		return err
	}
	if assessment.IsTimed && assessment.TimeLimitMinutes != nil {
		elapsed := time.Since(attempt.StartedAt).Minutes()
		if elapsed > *assessment.TimeLimitMinutes {
			return errors.NewBusinessRuleError("time limit exceeded")
		}
	}
	return nil
}
