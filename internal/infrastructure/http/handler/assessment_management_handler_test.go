package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/EduGoGroup/edugo-shared/auth"
	"github.com/EduGoGroup/edugo-shared/common/errors"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	mw "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func setupManagementHandler(
	assessRepo *mock.MockAssessmentRepository,
	mongoRepo *mock.MockMongoAssessmentRepository,
) (*AssessmentManagementHandler, *gin.Engine) {
	svc := service.NewAssessmentManagementService(assessRepo, mongoRepo, mock.MockLogger{})
	h := NewAssessmentManagementHandler(svc)
	r := gin.New()
	return h, r
}

// setManagementAuthContext sets user_id and active_context with school in gin.Context for tests.
func setManagementAuthContext(c *gin.Context, userID, schoolID uuid.UUID) {
	c.Set(mw.ContextKeyUserID, userID.String())
	c.Set(mw.ContextKeyActiveContext, &auth.UserContext{
		SchoolID: schoolID.String(),
		RoleName: "teacher",
	})
}

func TestManagementHandler_List(t *testing.T) {
	schoolID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		setAuth    bool
		query      string
		setupPG    func(m *mock.MockAssessmentRepository)
		wantStatus int
	}{
		{
			name:    "200 - happy path",
			setAuth: true,
			query:   "?limit=10&offset=0",
			setupPG: func(m *mock.MockAssessmentRepository) {
				m.ListFn = func(_ context.Context, filter repository.AssessmentFilter) ([]pgentities.Assessment, int, error) {
					title := "Test"
					return []pgentities.Assessment{
						{ID: uuid.New(), Title: &title, Status: "draft"},
					}, 1, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "401 - no auth context",
			setAuth:    false,
			query:      "?limit=10",
			setupPG:    func(_ *mock.MockAssessmentRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupPG != nil {
				tt.setupPG(assessRepo)
			}
			h, r := setupManagementHandler(assessRepo, mongoRepo)

			r.GET("/v1/management/assessments", func(c *gin.Context) {
				if tt.setAuth {
					setManagementAuthContext(c, userID, schoolID)
				}
				c.Next()
			}, h.List)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/management/assessments"+tt.query, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestManagementHandler_Create(t *testing.T) {
	schoolID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		setAuth    bool
		body       interface{}
		setupPG    func(m *mock.MockAssessmentRepository)
		setupMongo func(m *mock.MockMongoAssessmentRepository)
		wantStatus int
	}{
		{
			name:    "201 - happy path",
			setAuth: true,
			body:    dto.CreateAssessmentRequest{Title: "New Assessment"},
			setupPG: func(m *mock.MockAssessmentRepository) {
				m.CreateFn = func(_ context.Context, _ *pgentities.Assessment) error {
					return nil
				}
			},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.CreateFn = func(_ context.Context, _ *mongoentities.MaterialAssessment) (string, error) {
					return "aabbccddee0011223344ff00", nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "400 - invalid body",
			setAuth:    true,
			body:       map[string]interface{}{"title": ""}, // title required
			setupPG:    func(_ *mock.MockAssessmentRepository) {},
			setupMongo: func(_ *mock.MockMongoAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupPG != nil {
				tt.setupPG(assessRepo)
			}
			if tt.setupMongo != nil {
				tt.setupMongo(mongoRepo)
			}
			h, r := setupManagementHandler(assessRepo, mongoRepo)

			r.POST("/v1/management/assessments", func(c *gin.Context) {
				if tt.setAuth {
					setManagementAuthContext(c, userID, schoolID)
				}
				c.Next()
			}, h.Create)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/management/assessments", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestManagementHandler_Update_NonDraft(t *testing.T) {
	assessmentID := uuid.New()
	schoolID := uuid.New()
	userID := uuid.New()

	// Updating a non-draft assessment should return 422 (business rule error)
	assessRepo := &mock.MockAssessmentRepository{
		GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
			title := "Published Assessment"
			return &pgentities.Assessment{
				ID:     assessmentID,
				Title:  &title,
				Status: "published",
			}, nil
		},
	}
	mongoRepo := &mock.MockMongoAssessmentRepository{}

	h, r := setupManagementHandler(assessRepo, mongoRepo)

	r.PUT("/v1/management/assessments/:id", func(c *gin.Context) {
		setManagementAuthContext(c, userID, schoolID)
		c.Next()
	}, h.Update)

	newTitle := "Updated Title"
	body, _ := json.Marshal(dto.UpdateAssessmentRequest{Title: &newTitle})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPut, "/v1/management/assessments/"+assessmentID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, string(errors.ErrorCodeBusinessRule), resp["code"])
}

func TestManagementHandler_Delete(t *testing.T) {
	assessmentID := uuid.New()
	schoolID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupPG    func(m *mock.MockAssessmentRepository)
		wantStatus int
	}{
		{
			name:  "204 - happy path",
			param: assessmentID.String(),
			setupPG: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, Status: "draft"}, nil
				}
				m.SoftDeleteFn = func(_ context.Context, _ uuid.UUID) error {
					return nil
				}
			},
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "400 - invalid uuid",
			param:      "not-a-uuid",
			setupPG:    func(_ *mock.MockAssessmentRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "404 - not found",
			param: uuid.New().String(),
			setupPG: func(m *mock.MockAssessmentRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return nil, errors.NewNotFoundError("assessment")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			if tt.setupPG != nil {
				tt.setupPG(assessRepo)
			}
			h, r := setupManagementHandler(assessRepo, mongoRepo)

			r.DELETE("/v1/management/assessments/:id", func(c *gin.Context) {
				setManagementAuthContext(c, userID, schoolID)
				c.Next()
			}, h.Delete)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodDelete, "/v1/management/assessments/"+tt.param, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
