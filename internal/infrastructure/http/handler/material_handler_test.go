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
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/auth"
	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	mw "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func setupMaterialHandler(repo *mock.MockMaterialRepository) (*MaterialHandler, *gin.Engine) {
	svc := service.NewMaterialService(repo, nil, nil, mock.MockLogger{}, "test")
	h := NewMaterialHandler(svc)
	r := gin.New()
	return h, r
}

// setAuthContext sets user_id and active_context in gin.Context for tests.
func setAuthContext(c *gin.Context, userID, schoolID uuid.UUID) {
	c.Set(mw.ContextKeyUserID, userID.String())
	c.Set(mw.ContextKeyActiveContext, &auth.UserContext{
		SchoolID: schoolID.String(),
		RoleName: "teacher",
	})
}

func TestMaterialHandler_GetByID(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, Title: "Test", Status: "draft", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "not-a-uuid",
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "not found",
			param: materialID.String(),
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return nil, errors.NewNotFoundError("material")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)
			r.GET("/v1/materials/:id", h.GetByID)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials/"+tt.param, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_List(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			query: "?limit=10&offset=0",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, _ repository.MaterialFilter) ([]pgentities.Material, int, error) {
					return []pgentities.Material{{ID: uuid.New(), Title: "M1", CreatedAt: time.Now(), UpdatedAt: time.Now()}}, 1, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "empty list",
			query: "",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, _ repository.MaterialFilter) ([]pgentities.Material, int, error) {
					return nil, 0, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "repository error",
			query: "",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, _ repository.MaterialFilter) ([]pgentities.Material, int, error) {
					return nil, 0, errors.NewDatabaseError("list", nil)
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)
			r.GET("/v1/materials", h.List)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials"+tt.query, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_Create(t *testing.T) {
	userID := uuid.New()
	schoolID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setAuth    bool
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:    "happy path",
			body:    dto.CreateMaterialRequest{Title: "New Material", IsPublic: true},
			setAuth: true,
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.CreateFn = func(_ context.Context, _ *pgentities.Material) error {
					return nil
				}
			},
			wantStatus: http.StatusCreated,
		},
		{
			name:       "missing title - validation error",
			body:       dto.CreateMaterialRequest{},
			setAuth:    true,
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "no auth context",
			body:       dto.CreateMaterialRequest{Title: "Material"},
			setAuth:    false,
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)

			r.POST("/v1/materials", func(c *gin.Context) {
				if tt.setAuth {
					setAuthContext(c, userID, schoolID)
				}
				c.Next()
			}, h.Create)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/materials", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_Update(t *testing.T) {
	materialID := uuid.New()
	title := "Updated"

	tests := []struct {
		name       string
		param      string
		body       interface{}
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			body:  dto.UpdateMaterialRequest{Title: &title},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, Title: "Old", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
				}
				m.UpdateFn = func(_ context.Context, _ *pgentities.Material) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "bad-id",
			body:       dto.UpdateMaterialRequest{Title: &title},
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)
			r.PUT("/v1/materials/:id", h.Update)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/v1/materials/"+tt.param, bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_GetWithVersions(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetWithVersionsFn = func(_ context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error) {
					return &pgentities.Material{ID: id, Title: "Mat", CreatedAt: time.Now(), UpdatedAt: time.Now()},
						[]pgentities.MaterialVersion{
							{ID: uuid.New(), MaterialID: id, VersionNumber: 1, Title: "v1", ChangedBy: uuid.New(), CreatedAt: time.Now()},
						}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "not-valid",
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)
			r.GET("/v1/materials/:id/versions", h.GetWithVersions)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials/"+tt.param+"/versions", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_GenerateUploadURL(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		wantStatus int
	}{
		{
			name:       "s3 not configured",
			param:      materialID.String(),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			h, r := setupMaterialHandler(repo)
			r.POST("/v1/materials/:id/upload-url", h.GenerateUploadURL)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/materials/"+tt.param+"/upload-url", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_UploadComplete(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		body       interface{}
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			body:  dto.UploadCompleteRequest{FileURL: "s3://bucket/key", FileType: "pdf", FileSizeBytes: 1024},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, SchoolID: uuid.New(), UploadedByTeacherID: uuid.New(), CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
				}
				m.UpdateFn = func(_ context.Context, _ *pgentities.Material) error { return nil }
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing required fields",
			param:      materialID.String(),
			body:       map[string]string{},
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)
			r.POST("/v1/materials/:id/upload-complete", h.UploadComplete)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPost, "/v1/materials/"+tt.param+"/upload-complete", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestMaterialHandler_GenerateDownloadURL(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupRepo  func(m *mock.MockMaterialRepository)
		wantStatus int
	}{
		{
			name:  "s3 nil checked before file url",
			param: materialID.String(),
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, FileURL: ""}, nil
				}
			},
			wantStatus: http.StatusInternalServerError, // S3 nil is checked before FileURL
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			setupRepo:  func(_ *mock.MockMaterialRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupMaterialHandler(repo)
			r.GET("/v1/materials/:id/download-url", h.GenerateDownloadURL)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials/"+tt.param+"/download-url", nil)
			r.ServeHTTP(w, req)

			require.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
