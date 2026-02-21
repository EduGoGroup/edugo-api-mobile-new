package service

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func TestScreenService_GetScreen(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		screenKey string
		setupRepo func(m *mock.MockScreenRepository)
		wantErr   bool
		errCode   errors.ErrorCode
		check     func(t *testing.T, resp interface{})
	}{
		{
			name:      "happy path",
			screenKey: "home-screen",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreenByKeyFn = func(_ context.Context, key string) (*repository.ScreenComposed, error) {
					assert.Equal(t, "home-screen", key)
					return &repository.ScreenComposed{
						Instance: pgentities.ScreenInstance{
							ScreenKey: "home-screen",
							Name:      "Home",
							IsActive:  true,
							SlotData:  json.RawMessage(`{"title":"Welcome"}`),
						},
						Template: pgentities.ScreenTemplate{
							Pattern:    "list",
							Definition: json.RawMessage(`{}`),
						},
					}, nil
				}
			},
		},
		{
			name:      "not found",
			screenKey: "nonexistent",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreenByKeyFn = func(_ context.Context, _ string) (*repository.ScreenComposed, error) {
					return nil, errors.NewNotFoundError("screen")
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewScreenService(repo, mock.MockLogger{})
			resp, err := svc.GetScreen(ctx, tt.screenKey)

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
			assert.Equal(t, "home-screen", resp.ScreenKey)
			assert.Equal(t, "Home", resp.Name)
			assert.True(t, resp.IsActive)
		})
	}
}

func TestScreenService_GetScreensByResourceKey(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		resourceKey string
		setupRepo   func(m *mock.MockScreenRepository)
		wantErr     bool
		wantCount   int
	}{
		{
			name:        "happy path with multiple screens",
			resourceKey: "materials",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreensByResourceKeyFn = func(_ context.Context, _ string) ([]repository.ScreenComposed, error) {
					return []repository.ScreenComposed{
						{
							Instance: pgentities.ScreenInstance{ScreenKey: "mat-list", Name: "Material List", IsActive: true},
							Template: pgentities.ScreenTemplate{Pattern: "list", Definition: json.RawMessage(`{}`)},
						},
						{
							Instance: pgentities.ScreenInstance{ScreenKey: "mat-detail", Name: "Material Detail", IsActive: true},
							Template: pgentities.ScreenTemplate{Pattern: "detail", Definition: json.RawMessage(`{}`)},
						},
					}, nil
				}
			},
			wantCount: 2,
		},
		{
			name:        "empty result",
			resourceKey: "unknown",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetScreensByResourceKeyFn = func(_ context.Context, _ string) ([]repository.ScreenComposed, error) {
					return nil, nil
				}
			},
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewScreenService(repo, mock.MockLogger{})
			resp, err := svc.GetScreensByResourceKey(ctx, tt.resourceKey)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp, tt.wantCount)
		})
	}
}

func TestScreenService_GetNavigation(t *testing.T) {
	ctx := context.Background()
	rootID := uuid.New()
	childID := uuid.New()

	tests := []struct {
		name      string
		scope     string
		setupRepo func(m *mock.MockScreenRepository)
		wantErr   bool
		wantRoots int
	}{
		{
			name:  "happy path with hierarchy",
			scope: "system",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetNavigationFn = func(_ context.Context, _ string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
					return []pgentities.Resource{
							{ID: rootID, Key: "dashboard", DisplayName: "Dashboard", IsMenuVisible: true, IsActive: true, SortOrder: 1},
							{ID: childID, Key: "materials", DisplayName: "Materials", ParentID: &rootID, IsMenuVisible: true, IsActive: true, SortOrder: 1},
						}, []pgentities.ResourceScreen{
							{ResourceKey: "dashboard", ScreenKey: "dash-main", ScreenType: "main", IsDefault: true, SortOrder: 0},
						}, nil
				}
			},
			wantRoots: 1, // only dashboard is root; materials is a child
		},
		{
			name:  "inactive resources are filtered",
			scope: "system",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetNavigationFn = func(_ context.Context, _ string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
					return []pgentities.Resource{
						{ID: uuid.New(), Key: "active", DisplayName: "Active", IsMenuVisible: true, IsActive: true},
						{ID: uuid.New(), Key: "inactive", DisplayName: "Inactive", IsMenuVisible: true, IsActive: false},
						{ID: uuid.New(), Key: "hidden", DisplayName: "Hidden", IsMenuVisible: false, IsActive: true},
					}, nil, nil
				}
			},
			wantRoots: 1, // only the active + menu visible one
		},
		{
			name:  "repository error",
			scope: "system",
			setupRepo: func(m *mock.MockScreenRepository) {
				m.GetNavigationFn = func(_ context.Context, _ string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
					return nil, nil, errors.NewDatabaseError("query", nil)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewScreenService(repo, mock.MockLogger{})
			resp, err := svc.GetNavigation(ctx, tt.scope)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Len(t, resp, tt.wantRoots)
		})
	}
}

func TestScreenService_SavePreferences(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()

	tests := []struct {
		name      string
		screenKey string
		prefs     json.RawMessage
		setupRepo func(m *mock.MockScreenRepository)
		wantErr   bool
	}{
		{
			name:      "happy path",
			screenKey: "home",
			prefs:     json.RawMessage(`{"theme":"dark"}`),
			setupRepo: func(m *mock.MockScreenRepository) {
				m.UpsertPreferencesFn = func(_ context.Context, key, uid string, prefs json.RawMessage) error {
					assert.Equal(t, "home", key)
					assert.Equal(t, userID.String(), uid)
					assert.JSONEq(t, `{"theme":"dark"}`, string(prefs))
					return nil
				}
			},
		},
		{
			name:      "repository error",
			screenKey: "home",
			prefs:     json.RawMessage(`{}`),
			setupRepo: func(m *mock.MockScreenRepository) {
				m.UpsertPreferencesFn = func(_ context.Context, _, _ string, _ json.RawMessage) error {
					return errors.NewDatabaseError("upsert", nil)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockScreenRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewScreenService(repo, mock.MockLogger{})
			err := svc.SavePreferences(ctx, tt.screenKey, userID.String(), tt.prefs)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
