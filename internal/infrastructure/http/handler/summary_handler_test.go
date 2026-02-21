package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func setupSummaryHandler(repo *mock.MockMongoSummaryRepository) (*SummaryHandler, *gin.Engine) {
	svc := service.NewSummaryService(repo, mock.MockLogger{})
	h := NewSummaryHandler(svc)
	r := gin.New()
	return h, r
}

func TestSummaryHandler_GetSummary(t *testing.T) {
	materialID := uuid.New()

	tests := []struct {
		name       string
		param      string
		setupRepo  func(m *mock.MockMongoSummaryRepository)
		wantStatus int
	}{
		{
			name:  "happy path",
			param: materialID.String(),
			setupRepo: func(m *mock.MockMongoSummaryRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, mid string) (*mongoentities.MaterialSummary, error) {
					return &mongoentities.MaterialSummary{
						MaterialID: mid,
						Summary:    "A concise summary",
						KeyPoints:  []string{"Point 1"},
						Language:   "en",
						WordCount:  50,
						Version:    1,
						AIModel:    "gpt-4",
						CreatedAt:  time.Now(),
					}, nil
				}
			},
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid uuid",
			param:      "bad-uuid",
			setupRepo:  func(_ *mock.MockMongoSummaryRepository) {},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "not found",
			param: materialID.String(),
			setupRepo: func(m *mock.MockMongoSummaryRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialSummary, error) {
					return nil, errors.NewNotFoundError("summary")
				}
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMongoSummaryRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			h, r := setupSummaryHandler(repo)
			r.GET("/v1/materials/:id/summary", h.GetSummary)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/v1/materials/"+tt.param+"/summary", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}
