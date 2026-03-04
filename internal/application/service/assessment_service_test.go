package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func newTestAssessmentService(
	assessRepo *mock.MockAssessmentRepository,
	attemptRepo *mock.MockAttemptRepository,
	mongoRepo *mock.MockMongoAssessmentRepository,
) *AssessmentService {
	return NewAssessmentService(assessRepo, attemptRepo, mongoRepo, mock.MockLogger{})
}

func TestAssessmentService_GetAssessmentByMaterialID(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()
	assessmentID := uuid.New()

	tests := []struct {
		name           string
		setupAssess    func(m *mock.MockAssessmentRepository)
		setupMongo     func(m *mock.MockMongoAssessmentRepository)
		wantErr        bool
		errCode        errors.ErrorCode
		checkSanitized bool
		wantQuestions  int
	}{
		{
			name: "happy path with sanitized questions",
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, mid uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{
						ID:             assessmentID,
						MaterialID:     &mid,
						QuestionsCount: 2,
						Status:         "published",
					}, nil
				}
			},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, mid string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						MaterialID: mid,
						Questions: []mongoentities.Question{
							{
								QuestionID:    "q1",
								QuestionText:  "What is 2+2?",
								QuestionType:  "multiple_choice",
								CorrectAnswer: "B",
								Explanation:   "Simple math",
								Points:        1,
								Difficulty:    "easy",
								Options: []mongoentities.Option{
									{OptionID: "A", OptionText: "3"},
									{OptionID: "B", OptionText: "4"},
								},
							},
							{
								QuestionID:    "q2",
								QuestionText:  "Is the sky blue?",
								QuestionType:  "true_false",
								CorrectAnswer: "true",
								Explanation:   "It is blue",
								Points:        1,
								Difficulty:    "easy",
							},
						},
					}, nil
				}
			},
			checkSanitized: true,
			wantQuestions:  2,
		},
		{
			name: "assessment not found in postgres",
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return nil, errors.NewNotFoundError("assessment")
				}
			},
			setupMongo: func(_ *mock.MockMongoAssessmentRepository) {},
			wantErr:    true,
			errCode:    errors.ErrorCodeNotFound,
		},
		{
			name: "mongodb fetch fails - returns assessment without questions",
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{
						ID:             assessmentID,
						MaterialID:     &materialID,
						QuestionsCount: 1,
						Status:         "published",
					}, nil
				}
			},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return nil, errors.NewNotFoundError("mongo assessment")
				}
			},
			wantQuestions: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupAssess != nil {
				tt.setupAssess(assessRepo)
			}
			if tt.setupMongo != nil {
				tt.setupMongo(mongoRepo)
			}

			svc := newTestAssessmentService(assessRepo, nil, mongoRepo)
			resp, err := svc.GetAssessmentByMaterialID(ctx, materialID)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := errors.GetAppError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, appErr.Code)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Questions, tt.wantQuestions)

			if tt.checkSanitized {
				// Verify that correct_answer and explanation are NOT exposed
				for _, q := range resp.Questions {
					assert.NotEmpty(t, q.QuestionText)
					assert.NotEmpty(t, q.QuestionType)
					// CorrectAnswer and Explanation should NOT be in QuestionResponse
					// (they are not fields in the DTO)
				}
			}
		})
	}
}

func TestAssessmentService_CreateAttempt(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()
	studentID := uuid.New()
	assessmentID := uuid.New()
	maxAttempts := 3

	tests := []struct {
		name         string
		req          dto.CreateAttemptRequest
		setupAssess  func(m *mock.MockAssessmentRepository)
		setupAttempt func(m *mock.MockAttemptRepository)
		setupMongo   func(m *mock.MockMongoAssessmentRepository)
		wantErr      bool
		errCode      errors.ErrorCode
		errMsg       string
		check        func(t *testing.T, resp *dto.AttemptResponse)
	}{
		{
			name: "happy path - all correct",
			req: dto.CreateAttemptRequest{
				Answers: []dto.AnswerSubmission{
					{QuestionIndex: 0, Answer: "B"},
					{QuestionIndex: 1, Answer: "true"},
				},
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MaterialID: &materialID}, nil
				}
			},
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.CreateFn = func(_ context.Context, a *pgentities.AssessmentAttempt, answers []pgentities.AssessmentAttemptAnswer) error {
					assert.Equal(t, assessmentID, a.AssessmentID)
					assert.Equal(t, studentID, a.StudentID)
					assert.Equal(t, "completed", a.Status)
					assert.Len(t, answers, 2)
					return nil
				}
			},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
							{QuestionType: "true_false", CorrectAnswer: "true", Points: 1},
						},
					}, nil
				}
			},
			check: func(t *testing.T, resp *dto.AttemptResponse) {
				require.NotNil(t, resp.Score)
				require.NotNil(t, resp.MaxScore)
				require.NotNil(t, resp.Percentage)
				assert.InDelta(t, 2.0, *resp.Score, 0.001)
				assert.InDelta(t, 2.0, *resp.MaxScore, 0.001)
				assert.InDelta(t, 100.0, *resp.Percentage, 0.001)
			},
		},
		{
			name: "partially correct answers",
			req: dto.CreateAttemptRequest{
				Answers: []dto.AnswerSubmission{
					{QuestionIndex: 0, Answer: "B"},
					{QuestionIndex: 1, Answer: "false"}, // wrong
				},
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MaterialID: &materialID}, nil
				}
			},
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.CreateFn = func(_ context.Context, _ *pgentities.AssessmentAttempt, _ []pgentities.AssessmentAttemptAnswer) error {
					return nil
				}
			},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
							{QuestionType: "true_false", CorrectAnswer: "true", Points: 1},
						},
					}, nil
				}
			},
			check: func(t *testing.T, resp *dto.AttemptResponse) {
				require.NotNil(t, resp.Score)
				require.NotNil(t, resp.Percentage)
				assert.InDelta(t, 1.0, *resp.Score, 0.001)
				assert.InDelta(t, 50.0, *resp.Percentage, 0.001)
			},
		},
		{
			name: "max attempts reached",
			req: dto.CreateAttemptRequest{
				Answers: []dto.AnswerSubmission{{QuestionIndex: 0, Answer: "A"}},
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{
						ID:          assessmentID,
						MaterialID:  &materialID,
						MaxAttempts: &maxAttempts,
					}, nil
				}
			},
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.CountByAssessmentAndStudentFn = func(_ context.Context, _, _ uuid.UUID) (int, error) {
					return 3, nil // already at max
				}
			},
			setupMongo: func(_ *mock.MockMongoAssessmentRepository) {},
			wantErr:    true,
			errCode:    errors.ErrorCodeBusinessRule,
			errMsg:     "maximum number of attempts reached",
		},
		{
			name: "more answers than questions",
			req: dto.CreateAttemptRequest{
				Answers: []dto.AnswerSubmission{
					{QuestionIndex: 0, Answer: "A"},
					{QuestionIndex: 1, Answer: "B"},
					{QuestionIndex: 2, Answer: "C"}, // extra
				},
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MaterialID: &materialID}, nil
				}
			},
			setupAttempt: func(_ *mock.MockAttemptRepository) {},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "A", Points: 1},
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
						},
					}, nil
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeValidation,
		},
		{
			name: "invalid question index",
			req: dto.CreateAttemptRequest{
				Answers: []dto.AnswerSubmission{
					{QuestionIndex: 99, Answer: "A"}, // out of bounds
				},
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MaterialID: &materialID}, nil
				}
			},
			setupAttempt: func(_ *mock.MockAttemptRepository) {},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "A", Points: 1},
						},
					}, nil
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeValidation,
		},
		{
			name: "mongo fetch fails",
			req: dto.CreateAttemptRequest{
				Answers: []dto.AnswerSubmission{{QuestionIndex: 0, Answer: "A"}},
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MaterialID: &materialID}, nil
				}
			},
			setupAttempt: func(_ *mock.MockAttemptRepository) {},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return nil, errors.NewNotFoundError("questions")
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupAssess != nil {
				tt.setupAssess(assessRepo)
			}
			if tt.setupAttempt != nil {
				tt.setupAttempt(attemptRepo)
			}
			if tt.setupMongo != nil {
				tt.setupMongo(mongoRepo)
			}

			svc := newTestAssessmentService(assessRepo, attemptRepo, mongoRepo)
			resp, err := svc.CreateAttempt(ctx, materialID, studentID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errCode != "" {
					appErr, ok := errors.GetAppError(err)
					require.True(t, ok, "expected AppError but got: %T - %v", err, err)
					assert.Equal(t, tt.errCode, appErr.Code)
				}
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestAssessmentService_GetAttemptResult(t *testing.T) {
	ctx := context.Background()
	attemptID := uuid.New()
	assessmentID := uuid.New()
	materialID := uuid.New()
	studentID := uuid.New()
	score := 1.0
	maxScore := 2.0
	percentage := 50.0
	now := time.Now()

	tests := []struct {
		name         string
		setupAssess  func(m *mock.MockAssessmentRepository)
		setupAttempt func(m *mock.MockAttemptRepository)
		setupMongo   func(m *mock.MockMongoAssessmentRepository)
		wantErr      bool
		wantAnswers  int
	}{
		{
			name: "happy path with mongo enrichment",
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return &pgentities.AssessmentAttempt{
						ID:           attemptID,
						AssessmentID: assessmentID,
						StudentID:    studentID,
						Score:        &score,
						MaxScore:     &maxScore,
						Percentage:   &percentage,
						Status:       "completed",
						StartedAt:    now,
						CompletedAt:  &now,
					}, nil
				}
				isCorrect := true
				isWrong := false
				p1 := 1.0
				p0 := 0.0
				mp := 1.0
				ans1 := "B"
				ans2 := "false"
				m.GetAnswersByAttemptIDFn = func(_ context.Context, _ uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
					return []pgentities.AssessmentAttemptAnswer{
						{QuestionIndex: 0, StudentAnswer: &ans1, IsCorrect: &isCorrect, PointsEarned: &p1, MaxPoints: &mp},
						{QuestionIndex: 1, StudentAnswer: &ans2, IsCorrect: &isWrong, PointsEarned: &p0, MaxPoints: &mp},
					}, nil
				}
			},
			setupAssess: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MaterialID: &materialID}, nil
				}
			},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionText: "What is 2+2?", CorrectAnswer: "B", Explanation: "Math"},
							{QuestionText: "Is sky blue?", CorrectAnswer: "true", Explanation: "Yes it is"},
						},
					}, nil
				}
			},
			wantAnswers: 2,
		},
		{
			name: "attempt not found",
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return nil, errors.NewNotFoundError("attempt")
				}
			},
			setupAssess: func(_ *mock.MockAssessmentRepository) {},
			setupMongo:  func(_ *mock.MockMongoAssessmentRepository) {},
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupAssess != nil {
				tt.setupAssess(assessRepo)
			}
			if tt.setupAttempt != nil {
				tt.setupAttempt(attemptRepo)
			}
			if tt.setupMongo != nil {
				tt.setupMongo(mongoRepo)
			}

			svc := newTestAssessmentService(assessRepo, attemptRepo, mongoRepo)
			resp, err := svc.GetAttemptResult(ctx, attemptID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Answers, tt.wantAnswers)
			// Verify enrichment: check that QuestionText is populated from MongoDB
			if tt.wantAnswers > 0 {
				assert.NotEmpty(t, resp.Answers[0].QuestionText)
				assert.NotEmpty(t, resp.Answers[0].CorrectAnswer)
			}
		})
	}
}

func TestAssessmentService_ListAttemptsByUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name         string
		page         int
		limit        int
		setupAttempt func(m *mock.MockAttemptRepository)
		wantErr      bool
		wantTotal    int
		wantLimit    int
		wantPage     int
	}{
		{
			name:  "happy path with defaults",
			page:  0,
			limit: 0,
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.ListByUserIDFn = func(_ context.Context, _ uuid.UUID, limit, _ int, _ sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error) {
					assert.Equal(t, 20, limit)
					return []pgentities.AssessmentAttempt{
						{ID: uuid.New(), AssessmentID: uuid.New(), StudentID: userID, Status: "completed", StartedAt: time.Now()},
					}, 1, nil
				}
			},
			wantTotal: 1,
			wantLimit: 20,
			wantPage:  1,
		},
		{
			name:  "limit capped at 100",
			page:  1,
			limit: 200,
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.ListByUserIDFn = func(_ context.Context, _ uuid.UUID, limit, _ int, _ sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error) {
					assert.Equal(t, 100, limit)
					return nil, 0, nil
				}
			},
			wantTotal: 0,
			wantLimit: 100,
			wantPage:  1,
		},
		{
			name:  "repository error",
			page:  1,
			limit: 10,
			setupAttempt: func(m *mock.MockAttemptRepository) {
				m.ListByUserIDFn = func(_ context.Context, _ uuid.UUID, _, _ int, _ sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error) {
					return nil, 0, errors.NewDatabaseError("list", nil)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attemptRepo := &mock.MockAttemptRepository{}
			if tt.setupAttempt != nil {
				tt.setupAttempt(attemptRepo)
			}

			svc := newTestAssessmentService(nil, attemptRepo, nil)
			resp, err := svc.ListAttemptsByUser(ctx, userID, tt.page, tt.limit, sharedrepo.ListFilters{})

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantTotal, resp.Total)
			assert.Equal(t, tt.wantLimit, resp.Limit)
			assert.Equal(t, tt.wantPage, resp.Page)
		})
	}
}
