package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func TestStatsService_GetGlobalStats(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()

	tests := []struct {
		name      string
		schoolID  *uuid.UUID
		setupRepo func(m *mock.MockStatsRepository)
		wantErr   bool
		check     func(t *testing.T, resp interface{})
	}{
		{
			name:     "happy path with school filter",
			schoolID: &schoolID,
			setupRepo: func(m *mock.MockStatsRepository) {
				m.CountMaterialsFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 42, nil
				}
				m.CountCompletedProgressFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 10, nil
				}
				m.AverageAttemptScoreFn = func(_ context.Context, _ *uuid.UUID) (float64, error) {
					return 85.5, nil
				}
			},
		},
		{
			name:     "happy path without school filter",
			schoolID: nil,
			setupRepo: func(m *mock.MockStatsRepository) {
				m.CountMaterialsFn = func(_ context.Context, sid *uuid.UUID) (int, error) {
					assert.Nil(t, sid)
					return 100, nil
				}
				m.CountCompletedProgressFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 50, nil
				}
				m.AverageAttemptScoreFn = func(_ context.Context, _ *uuid.UUID) (float64, error) {
					return 72.0, nil
				}
			},
		},
		{
			name:     "count materials error",
			schoolID: nil,
			setupRepo: func(m *mock.MockStatsRepository) {
				m.CountMaterialsFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 0, errors.NewDatabaseError("count", nil)
				}
				m.CountCompletedProgressFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 0, nil
				}
				m.AverageAttemptScoreFn = func(_ context.Context, _ *uuid.UUID) (float64, error) {
					return 0, nil
				}
			},
			wantErr: true,
		},
		{
			name:     "average score error",
			schoolID: nil,
			setupRepo: func(m *mock.MockStatsRepository) {
				m.CountMaterialsFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 5, nil
				}
				m.CountCompletedProgressFn = func(_ context.Context, _ *uuid.UUID) (int, error) {
					return 3, nil
				}
				m.AverageAttemptScoreFn = func(_ context.Context, _ *uuid.UUID) (float64, error) {
					return 0, errors.NewDatabaseError("avg", nil)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStatsRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewStatsService(repo, mock.MockLogger{})
			resp, err := svc.GetGlobalStats(ctx, tt.schoolID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Greater(t, resp.TotalMaterials, 0)
		})
	}
}

func TestStatsService_GetMaterialStats(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()

	tests := []struct {
		name      string
		setupRepo func(m *mock.MockStatsRepository)
		wantErr   bool
		errCode   errors.ErrorCode
	}{
		{
			name: "happy path",
			setupRepo: func(m *mock.MockStatsRepository) {
				m.MaterialStatsFn = func(_ context.Context, id uuid.UUID) (*repository.MaterialStatsResult, error) {
					assert.Equal(t, materialID, id)
					return &repository.MaterialStatsResult{
						TotalAttempts:  10,
						AverageScore:   75.0,
						CompletionRate: 80.0,
						UniqueStudents: 5,
					}, nil
				}
			},
		},
		{
			name: "not found",
			setupRepo: func(m *mock.MockStatsRepository) {
				m.MaterialStatsFn = func(_ context.Context, _ uuid.UUID) (*repository.MaterialStatsResult, error) {
					return nil, errors.NewNotFoundError("material stats")
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockStatsRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewStatsService(repo, mock.MockLogger{})
			resp, err := svc.GetMaterialStats(ctx, materialID)

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
			assert.Equal(t, 10, resp.TotalAttempts)
			assert.InDelta(t, 75.0, resp.AverageScore, 0.001)
		})
	}
}
