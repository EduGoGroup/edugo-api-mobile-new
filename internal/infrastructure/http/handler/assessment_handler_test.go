package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	mw "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func setupAssessmentHandler(
	assessRepo *mock.MockAssessmentRepository,
	attemptRepo *mock.MockAttemptRepository,
	mongoRepo *mock.MockMongoAssessmentRepository,
) (*AssessmentHandler, *gin.Engine) {
	svc := service.NewAssessmentService(assessRepo, attemptRepo, mongoRepo, mock.MockLogger{}, nil)
	h := NewAssessmentHandler(svc)
	r := gin.New()
	return h, r
}

func TestAssessmentHandler_GetAssessment(t *testing.T) {
	materialID := uuid.New()
	assessmentID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupA     func(m *mock.MockAssessmentRepository)
		setupM     func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MongoDocumentID: "mongo123", Status: "published"}, nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionID: "q1", QuestionText: "Test?", QuestionType: "multiple_choice", Points: 1, Difficulty: "easy"},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "bad-uuid",
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "not found",
			param: materialID.String(),
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return nil, errors.NewNotFoundError("assessment")
				}
			},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupA != nil {
				tt.setupA(assessRepo)
			}
			if tt.setupM != nil {
				tt.setupM(mongoRepo)
			}
			h, r := setupAssessmentHandler(assessRepo, nil, mongoRepo)
			r.GET("/v1/materials/:id/assessment", h.GetAssessment)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials/"+tt.param+"/assessment", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAssessmentHandler_CreateAttempt(t *testing.T) {
	materialID := uuid.New()
	userID := uuid.New()
	assessmentID := uuid.New()

	tests := []struct {
		name       string
		param      string
		body       interface{}
		setAuth    bool
		setupA     func(m *mock.MockAssessmentRepository)
		setupAtt   func(m *mock.MockAttemptRepository)
		setupM     func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:    "happy path",
			param:   materialID.String(),
			body:    dto.CreateAttemptRequest{Answers: []dto.AnswerSubmission{{QuestionIndex: 0, Answer: "B"}}},
			setAuth: true,
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MongoDocumentID: "mongo123"}, nil
				}
			},
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.CreateFn = func(_ context.Context, _ *pgentities.AssessmentAttempt, _ []pgentities.AssessmentAttemptAnswer) error {
					return nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
						},
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			body:       dto.CreateAttemptRequest{Answers: []dto.AnswerSubmission{{QuestionIndex: 0, Answer: "A"}}},
			setAuth:    true,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no auth",
			param:      materialID.String(),
			body:       dto.CreateAttemptRequest{Answers: []dto.AnswerSubmission{{QuestionIndex: 0, Answer: "A"}}},
			setAuth:    false,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "empty answers",
			param:      materialID.String(),
			body:       map[string]interface{}{"answers": []string{}},
			setAuth:    true,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupA != nil {
				tt.setupA(assessRepo)
			}
			if tt.setupAtt != nil {
				tt.setupAtt(attemptRepo)
			}
			if tt.setupM != nil {
				tt.setupM(mongoRepo)
			}
			h, r := setupAssessmentHandler(assessRepo, attemptRepo, mongoRepo)

			r.POST("/v1/materials/:id/assessment/attempts", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.CreateAttempt)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/materials/"+tt.param+"/assessment/attempts", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAssessmentHandler_GetAttemptResult(t *testing.T) {
	attemptID := uuid.New()
	assessmentID := uuid.New()
	score := 1.0
	maxScore := 1.0
	pct := 100.0
	now := time.Now()

	tests := []struct {
		name       string
		param      string
		setupA     func(m *mock.MockAssessmentRepository)
		setupAtt   func(m *mock.MockAttemptRepository)
		setupM     func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: attemptID.String(),
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return &pgentities.AssessmentAttempt{
						ID: attemptID, AssessmentID: assessmentID, StudentID: uuid.New(),
						Score: &score, MaxScore: &maxScore, Percentage: &pct,
						Status: "completed", StartedAt: now, CompletedAt: &now,
					}, nil
				}
				m.GetAnswersByAttemptIDFn = func(_ context.Context, _ uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
					return nil, nil
				}
			},
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MongoDocumentID: "mongo123"}, nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{Questions: nil}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupA != nil {
				tt.setupA(assessRepo)
			}
			if tt.setupAtt != nil {
				tt.setupAtt(attemptRepo)
			}
			if tt.setupM != nil {
				tt.setupM(mongoRepo)
			}
			h, r := setupAssessmentHandler(assessRepo, attemptRepo, mongoRepo)
			r.GET("/v1/attempts/:id/results", h.GetAttemptResult)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/attempts/"+tt.param+"/results", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAssessmentHandler_StartAttempt(t *testing.T) {
	assessmentID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setAuth    bool
		setupA     func(m *mock.MockAssessmentRepository)
		setupAtt   func(m *mock.MockAttemptRepository)
		setupM     func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:    "happy path",
			param:   assessmentID.String(),
			setAuth: true,
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{
						ID:              assessmentID,
						MongoDocumentID: "mongo123",
						Status:          "published",
						QuestionsCount:  1,
					}, nil
				}
			},
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.GetInProgressByStudentAndAssessmentFn = func(_ context.Context, _, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return nil, nil
				}
				m.CreateAttemptOnlyFn = func(_ context.Context, _ *pgentities.AssessmentAttempt) error {
					return nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionID: "q1", QuestionText: "Test?", QuestionType: "multiple_choice", Points: 1},
						},
					}, nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			setAuth:    true,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no auth",
			param:      assessmentID.String(),
			setAuth:    false,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			tt.setupA(assessRepo)
			tt.setupAtt(attemptRepo)
			tt.setupM(mongoRepo)
			h, r := setupAssessmentHandler(assessRepo, attemptRepo, mongoRepo)

			r.POST("/v1/assessments/:id/start", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.StartAttempt)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/assessments/"+tt.param+"/start", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAssessmentHandler_SaveAnswer(t *testing.T) {
	attemptID := uuid.New()
	assessmentID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		paramID    string
		paramQI    string
		body       interface{}
		setAuth    bool
		setupA     func(m *mock.MockAssessmentRepository)
		setupAtt   func(m *mock.MockAttemptRepository)
		setupM     func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:    "happy path",
			paramID: attemptID.String(),
			paramQI: "0",
			body:    dto.SaveAnswerRequest{Answer: "B"},
			setAuth: true,
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MongoDocumentID: "mongo123"}, nil
				}
			},
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return &pgentities.AssessmentAttempt{
						ID:           attemptID,
						AssessmentID: assessmentID,
						StudentID:    userID,
						Status:       "in_progress",
						StartedAt:    time.Now(),
					}, nil
				}
				m.UpsertAnswerFn = func(_ context.Context, _ *pgentities.AssessmentAttemptAnswer) error {
					return nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid attempt uuid",
			paramID:    "bad",
			paramQI:    "0",
			body:       dto.SaveAnswerRequest{Answer: "B"},
			setAuth:    true,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid questionIndex",
			paramID:    attemptID.String(),
			paramQI:    "abc",
			body:       dto.SaveAnswerRequest{Answer: "B"},
			setAuth:    true,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no auth",
			paramID:    attemptID.String(),
			paramQI:    "0",
			body:       dto.SaveAnswerRequest{Answer: "B"},
			setAuth:    false,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			tt.setupA(assessRepo)
			tt.setupAtt(attemptRepo)
			tt.setupM(mongoRepo)
			h, r := setupAssessmentHandler(assessRepo, attemptRepo, mongoRepo)

			r.PUT("/v1/attempts/:id/answers/:questionIndex", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.SaveAnswer)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/v1/attempts/"+tt.paramID+"/answers/"+tt.paramQI, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAssessmentHandler_SubmitAttempt(t *testing.T) {
	attemptID := uuid.New()
	assessmentID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		param      string
		body       interface{}
		setAuth    bool
		setupA     func(m *mock.MockAssessmentRepository)
		setupAtt   func(m *mock.MockAttemptRepository)
		setupM     func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:    "happy path with empty body",
			param:   attemptID.String(),
			body:    nil,
			setAuth: true,
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MongoDocumentID: "mongo123"}, nil
				}
			},
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return &pgentities.AssessmentAttempt{
						ID:           attemptID,
						AssessmentID: assessmentID,
						StudentID:    userID,
						Status:       "in_progress",
						StartedAt:    time.Now(),
					}, nil
				}
				m.GetAnswersByAttemptIDFn = func(_ context.Context, _ uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
					return nil, nil
				}
				m.UpdateAttemptFn = func(_ context.Context, _ *pgentities.AssessmentAttempt) error {
					return nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:    "happy path with answers",
			param:   attemptID.String(),
			body:    dto.SubmitAttemptRequest{Answers: []dto.AnswerSubmission{{QuestionIndex: 0, Answer: "B"}}},
			setAuth: true,
			setupA: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, MongoDocumentID: "mongo123"}, nil
				}
			},
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.AssessmentAttempt, error) {
					return &pgentities.AssessmentAttempt{
						ID:           attemptID,
						AssessmentID: assessmentID,
						StudentID:    userID,
						Status:       "in_progress",
						StartedAt:    time.Now(),
					}, nil
				}
				m.UpsertAnswerFn = func(_ context.Context, _ *pgentities.AssessmentAttemptAnswer) error {
					return nil
				}
				ans := "B"
				m.GetAnswersByAttemptIDFn = func(_ context.Context, _ uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
					return []pgentities.AssessmentAttemptAnswer{
						{QuestionIndex: 0, StudentAnswer: &ans},
					}, nil
				}
				m.UpdateAnswersFn = func(_ context.Context, _ []pgentities.AssessmentAttemptAnswer) error {
					return nil
				}
				m.UpdateAttemptFn = func(_ context.Context, _ *pgentities.AssessmentAttempt) error {
					return nil
				}
			},
			setupM: func(m *mock.MockMongoAssessmentRepository) {
				m.GetByObjectIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionType: "multiple_choice", CorrectAnswer: "B", Points: 1},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			body:       nil,
			setAuth:    true,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no auth",
			param:      attemptID.String(),
			body:       nil,
			setAuth:    false,
			setupA:     func(_ *mock.MockAssessmentRepository) {},
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			setupM:     func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			attemptRepo := &mock.MockAttemptRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			tt.setupA(assessRepo)
			tt.setupAtt(attemptRepo)
			tt.setupM(mongoRepo)
			h, r := setupAssessmentHandler(assessRepo, attemptRepo, mongoRepo)

			r.POST("/v1/attempts/:id/submit", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.SubmitAttempt)

			var bodyBytes []byte
			if tt.body != nil {
				bodyBytes, _ = json.Marshal(tt.body)
			}
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/attempts/"+tt.param+"/submit", bytes.NewReader(bodyBytes))
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestAssessmentHandler_ListUserAttempts(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		query      string
		setAuth    bool
		setupAtt   func(m *mock.MockAttemptRepository)
		wantStatus int
	}{
		{
			name:    "happy path",
			query:   "?limit=10&page=1",
			setAuth: true,
			setupAtt: func(m *mock.MockAttemptRepository) {
				m.ListByUserIDFn = func(_ context.Context, _ uuid.UUID, _, _ int, _ sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error) {
					return nil, 0, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no auth",
			query:      "",
			setAuth:    false,
			setupAtt:   func(_ *mock.MockAttemptRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attemptRepo := &mock.MockAttemptRepository{}
			if tt.setupAtt != nil {
				tt.setupAtt(attemptRepo)
			}
			h, r := setupAssessmentHandler(nil, attemptRepo, nil)

			r.GET("/v1/users/me/attempts", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.ListUserAttempts)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/users/me/attempts"+tt.query, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
