package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func TestProgressService_Upsert(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	materialID := uuid.New()

	tests := []struct {
		name       string
		req        dto.UpsertProgressRequest
		setupRepo  func(m *mock.MockProgressRepository)
		wantErr    bool
		wantStatus string
	}{
		{
			name: "in progress",
			req:  dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 50, LastPage: 5},
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, p *pgentities.Progress) error {
					assert.Equal(t, "in_progress", p.Status)
					assert.Equal(t, 50, p.Percentage)
					assert.Equal(t, 5, p.LastPage)
					return nil
				}
			},
			wantStatus: "in_progress",
		},
		{
			name: "completed at 100 percent",
			req:  dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 100, LastPage: 20},
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, p *pgentities.Progress) error {
					assert.Equal(t, "completed", p.Status)
					return nil
				}
			},
			wantStatus: "completed",
		},
		{
			name: "not started at 0 percent",
			req:  dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 0, LastPage: 0},
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, p *pgentities.Progress) error {
					assert.Equal(t, "not_started", p.Status)
					return nil
				}
			},
			wantStatus: "not_started",
		},
		{
			name: "repository error",
			req:  dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 10},
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, _ *pgentities.Progress) error {
					return errors.NewDatabaseError("upsert", nil)
				}
			},
			wantErr: true,
		},
		{
			name: "over 100 percent still completes",
			req:  dto.UpsertProgressRequest{MaterialID: materialID, Percentage: 150},
			setupRepo: func(m *mock.MockProgressRepository) {
				m.UpsertFn = func(_ context.Context, p *pgentities.Progress) error {
					assert.Equal(t, "completed", p.Status)
					return nil
				}
			},
			wantStatus: "completed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockProgressRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewProgressService(repo, mock.MockLogger{})
			resp, err := svc.Upsert(ctx, userID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantStatus, resp.Status)
			assert.Equal(t, materialID, resp.MaterialID)
			assert.Equal(t, userID, resp.UserID)
		})
	}
}
