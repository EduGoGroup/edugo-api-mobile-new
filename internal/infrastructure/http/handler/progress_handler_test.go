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
	mw "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func setupProgressHandler(repo *mock.MockProgressRepository) (*ProgressHandler, *gin.Engine) {
	svc := service.NewProgressService(repo, mock.MockLogger{})
	h := NewProgressHandler(svc)
	r := gin.New()
	return h, r
}

func TestProgressHandler_Upsert(t *testing.T) {
	userID := uuid.New()
	materialID := uuid.New()

	tests := []struct {
		name       string
		body       interface{}
		setAuth    bool
		setupRepo  func(m *mock.MockProgressRepository)
		wantStatus int
	}{
		{
			name:    "happy path",
			body:    dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 50, LastPage: 5},
			setAuth: true,
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, _ *pgentities.Progress) error {
					return nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "no auth context",
			body:       dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 10},
			setAuth:    false,
			setupRepo:  func(_ *mock.MockProgressRepository) {},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "missing material_id",
			body:       map[string]int{"percentage": 50},
			setAuth:    true,
			setupRepo:  func(_ *mock.MockProgressRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "repository error",
			body:    dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 10},
			setAuth: true,
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, _ *pgentities.Progress) error {
					return errors.NewDatabaseError("upsert", nil)
				}
			},
			wantStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockProgressRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupProgressHandler(repo)

			r.PUT("/v1/progress", func(c *gin.Context) {
				if tt.setAuth {
					c.Set(mw.ContextKeyUserID, userID.String())
				}
				c.Next()
			}, h.Upsert)

			bodyBytes, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodPut, "/v1/progress", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
