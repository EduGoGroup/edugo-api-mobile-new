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

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	mw "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func setupScreenHandler(repo *mock.MockScreenRepository) (*ScreenHandler, *gin.Engine) {
	svc := service.NewScreenService(repo, nil, nil, mock.MockLogger{})
	h := NewScreenHandler(svc)
	r := gin.New()
	return h, r
}

func TestScreenHandler_GetScreen(t *testing.T) {
	tests := []struct {
		name       string
		screenKey  string
		setupRepo  func(m *mock.MockScreenRepository)
		wantStatus int
	}{
		{
			name:      "happy path",
			screenKey: "home",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreenByKeyFn = func(_ context.Context, _ string) (*repository.ScreenComposed, error) {
					return &repository.ScreenComposed{
						Instance: pgentities.ScreenInstance{ScreenKey: "home", Name: "Home", IsActive: true, SlotData: json.RawMessage(`{}`)},
						Template: pgentities.ScreenTemplate{Pattern: "list", Definition: json.RawMessage(`{}`)},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:      "not found",
			screenKey: "missing",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreenByKeyFn = func(_ context.Context, _ string) (*repository.ScreenComposed, error) {
					return nil, errors.NewNotFoundError("screen")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupScreenHandler(repo)
			r.GET("/v1/screens/:screenKey", h.GetScreen)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/screens/"+tt.screenKey, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestScreenHandler_GetScreensByResource(t *testing.T) {
	tests := []struct {
		name        string
		resourceKey string
		setupRepo   func(m *mock.MockScreenRepository)
		wantStatus  int
	}{
		{
			name:        "happy path",
			resourceKey: "materials",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreensByResourceKeyFn = func(_ context.Context, _ string) ([]repository.ScreenComposed, error) {
					return []repository.ScreenComposed{
						{
							Instance: pgentities.ScreenInstance{ScreenKey: "mat-list", Name: "List", IsActive: true, SlotData: json.RawMessage(`{}`)},
							Template: pgentities.ScreenTemplate{Pattern: "list", Definition: json.RawMessage(`{}`)},
						},
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:        "repository error",
			resourceKey: "fail",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreensByResourceKeyFn = func(_ context.Context, _ string) ([]repository.ScreenComposed, error) {
					return nil, errors.NewDatabaseError("query", nil)
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupScreenHandler(repo)
			r.GET("/v1/screens/resource/:resourceKey", h.GetScreensByResource)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/screens/resource/"+tt.resourceKey, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestScreenHandler_GetNavigation(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setupRepo  func(m *mock.MockScreenRepository)
		wantStatus int
	}{
		{
			name:  "default scope",
			query: "",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetNavigationFn = func(_ context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
					assert.Equal(t, "system", scope)
					return nil, nil, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "custom scope",
			query: "?scope=school",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetNavigationFn = func(_ context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
					assert.Equal(t, "school", scope)
					return nil, nil, nil
				}
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupScreenHandler(repo)
			r.GET("/v1/screens/navigation", h.GetNavigation)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/screens/navigation"+tt.query, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestScreenHandler_SavePreferences(t *testing.T) {
	userID := uuid.New()

	tests := []struct {
		name       string
		screenKey  string
		body       interface{}
		setAuth    bool
		setupRepo  func(m *mock.MockScreenRepository)
		wantStatus int
	}{
		{
			name:      "happy path",
			screenKey: "home",
			body:      dto.SavePreferencesRequest{Preferences: json.RawMessage(`{"theme":"dark"}`)},
			setAuth:   true,
			setupRepo: func(m *mock.MockScreenRepository) {
				m.UpsertPreferencesFn = func(_ context.Context, _, _ string, _ json.RawMessage) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no auth",
			screenKey:  "home",
			body:       dto.SavePreferencesRequest{Preferences: json.RawMessage(`{}`)},
			setAuth:    false,
			setupRepo:  func(_ *mock.MockScreenRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing preferences",
			screenKey:  "home",
			body:       map[string]string{},
			setAuth:    true,
			setupRepo:  func(_ *mock.MockScreenRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupScreenHandler(repo)

			r.PUT("/v1/screens/:screenKey/preferences", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.SavePreferences)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/v1/screens/"+tt.screenKey+"/preferences", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
