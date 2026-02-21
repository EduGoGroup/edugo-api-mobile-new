package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func TestSummaryService_GetByMaterialID(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name       string
		materialID string
		setupRepo  func(m *mock.MockMongoSummaryRepository)
		wantErr    bool
		errCode    errors.ErrorCode
		check      func(t *testing.T, resp interface{})
	}{
		{
			name:       "happy path",
			materialID: "550e8400-e29b-41d4-a716-446655440000",
			setupRepo: func(m *mock.MockMongoSummaryRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, mid string) (*mongoentities.MaterialSummary, error) {
					return &mongoentities.MaterialSummary{
						MaterialID: mid,
						Summary:    "This is a summary of the material.",
						KeyPoints:  []string{"Point 1", "Point 2"},
						Language:   "en",
						WordCount:  100,
						Version:    1,
						AIModel:    "gpt-4",
						CreatedAt:  time.Now(),
					}, nil
				}
			},
		},
		{
			name:       "not found",
			materialID: "nonexistent-id",
			setupRepo: func(m *mock.MockMongoSummaryRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialSummary, error) {
					return nil, errors.NewNotFoundError("summary")
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeNotFound,
		},
		{
			name:       "database error",
			materialID: "some-id",
			setupRepo: func(m *mock.MockMongoSummaryRepository) {
				m.GetByMaterialIDFn = func(_ context.Context, _ string) (*mongoentities.MaterialSummary, error) {
					return nil, errors.NewDatabaseError("query", nil)
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMongoSummaryRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewSummaryService(repo, mock.MockLogger{})
			resp, err := svc.GetByMaterialID(ctx, tt.materialID)

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
			assert.Equal(t, tt.materialID, resp.MaterialID)
			assert.NotEmpty(t, resp.Summary)
			assert.NotEmpty(t, resp.KeyPoints)
			assert.NotEmpty(t, resp.AIModel)
		})
	}
}
