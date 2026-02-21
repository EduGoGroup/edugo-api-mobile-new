package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func setupStatsHandler(repo *mock.MockStatsRepository) (*StatsHandler, *gin.Engine) {
	svc := service.NewStatsService(repo, mock.MockLogger{})
	h := NewStatsHandler(svc)
	r := gin.New()
	return h, r
}

func TestStatsHandler_GetGlobalStats(t *testing.T) {
	tests := []struct {
		name       string
		query      string
		setupRepo  func(m *mock.MockStatsRepository)
		wantStatus int
	}{
		{
			name:  "happy path without school_id",
			query: "",
			setupRepo: func(m *mock.MockStatsRepository) {
				m.CountMaterialsFn = func(_ context.Context, _ *uuid.UUID) (int, error) { return 10, nil }
				m.CountCompletedProgressFn = func(_ context.Context, _ *uuid.UUID) (int, error) { return 5, nil }
				m.AverageAttemptScoreFn = func(_ context.Context, _ *uuid.UUID) (float64, error) { return 80.0, nil }
			},
			wantStatus: http.StatusOK,
		},
		{
			name:  "with valid school_id",
			query: "?school_id=" + uuid.New().String(),
			setupRepo: func(m *mock.MockStatsRepository) {
				m.CountMaterialsFn = func(_ context.Context, _ *uuid.UUID) (int, error) { return 3, nil }
				m.CountCompletedProgressFn = func(_ context.Context, _ *uuid.UUID) (int, error) { return 2, nil }
				m.AverageAttemptScoreFn = func(_ context.Context, _ *uuid.UUID) (float64, error) { return 90.0, nil }
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid school_id",
			query:      "?school_id=not-a-uuid",
			setupRepo:  func(_ *mock.MockStatsRepository) {},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStatsRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupStatsHandler(repo)
			r.GET("/v1/stats/global", h.GetGlobalStats)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/stats/global"+tt.query, nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestStatsHandler_GetMaterialStats(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupRepo  func(m *mock.MockStatsRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			setupRepo: func(m *mock.MockStatsRepository) {
				m.MaterialStatsFn = func(_ context.Context, _ uuid.UUID) (*repository.MaterialStatsResult, error) {
					return &repository.MaterialStatsResult{
						TotalAttempts:  5,
						AverageScore:   72.5,
						CompletionRate: 60.0,
						UniqueStudents: 3,
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "bad",
			setupRepo:  func(_ *mock.MockStatsRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "not found",
			param: materialID.String(),
			setupRepo: func(m *mock.MockStatsRepository) {
				m.MaterialStatsFn = func(_ context.Context, _ uuid.UUID) (*repository.MaterialStatsResult, error) {
					return nil, errors.NewNotFoundError("stats")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStatsRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupStatsHandler(repo)
			r.GET("/v1/materials/:id/stats", h.GetMaterialStats)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials/"+tt.param+"/stats", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
