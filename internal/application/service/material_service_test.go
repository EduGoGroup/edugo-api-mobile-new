package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func newTestMaterialService(repo *mock.MockMaterialRepository, pub *mock.MockPublisher) *MaterialService {
	return NewMaterialService(repo, nil, pub, mock.MockLogger{}, "test-exchange", nil)
}

func TestMaterialService_Create(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()
	teacherID := uuid.New()

	tests := []struct {
		name      string
		req       dto.CreateMaterialRequest
		setupRepo func(m *mock.MockMaterialRepository)
		wantErr   bool
		errCode   errors.ErrorCode
		check     func(t *testing.T, resp *dto.MaterialResponse)
	}{
		{
			name: "happy path",
			req: dto.CreateMaterialRequest{
				Title:    "Test Material",
				IsPublic: true,
			},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.CreateFn = func(_ context.Context, mat *pgentities.Material) error {
					assert.Equal(t, "Test Material", mat.Title)
					assert.Equal(t, schoolID, mat.SchoolID)
					assert.Equal(t, teacherID, mat.UploadedByTeacherID)
					assert.Equal(t, "draft", mat.Status)
					assert.True(t, mat.IsPublic)
					return nil
				}
			},
			check: func(t *testing.T, resp *dto.MaterialResponse) {
				assert.Equal(t, "Test Material", resp.Title)
				assert.Equal(t, "draft", resp.Status)
				assert.True(t, resp.IsPublic)
				assert.Equal(t, schoolID, resp.SchoolID)
				assert.Equal(t, teacherID, resp.UploadedByTeacherID)
			},
		},
		{
			name: "repository error",
			req: dto.CreateMaterialRequest{
				Title: "Fail Material",
			},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.CreateFn = func(_ context.Context, _ *pgentities.Material) error {
					return errors.NewDatabaseError("insert", nil)
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeDatabaseError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := newTestMaterialService(repo, nil)
			resp, err := svc.Create(ctx, tt.req, schoolID, teacherID)

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
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestMaterialService_GetByID(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()

	tests := []struct {
		name      string
		setupRepo func(m *mock.MockMaterialRepository)
		wantErr   bool
		errCode   errors.ErrorCode
	}{
		{
			name: "happy path",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, id uuid.UUID) (*pgentities.Material, error) {
					assert.Equal(t, materialID, id)
					return &pgentities.Material{
						ID:    materialID,
						Title: "Found Material",
					}, nil
				}
			},
		},
		{
			name: "not found",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return nil, errors.NewNotFoundError("material")
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := newTestMaterialService(repo, nil)
			resp, err := svc.GetByID(ctx, materialID)

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
			assert.Equal(t, materialID, resp.ID)
		})
	}
}

func TestMaterialService_List(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()

	tests := []struct {
		name      string
		req       dto.ListMaterialsRequest
		setupRepo func(m *mock.MockMaterialRepository)
		wantErr   bool
		wantTotal int
		wantLimit int
		wantPage  int
	}{
		{
			name: "happy path with defaults",
			req:  dto.ListMaterialsRequest{SchoolID: &schoolID},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, f repository.MaterialFilter) ([]pgentities.Material, int, error) {
					assert.Equal(t, 20, f.Limit)  // default limit
					assert.Equal(t, 0, f.Offset)  // page 1 → offset 0
					return []pgentities.Material{{ID: uuid.New(), Title: "M1"}}, 1, nil
				}
			},
			wantTotal: 1,
			wantLimit: 20,
			wantPage:  1,
		},
		{
			name: "page defaults to 1 when zero",
			req:  dto.ListMaterialsRequest{SchoolID: &schoolID, Page: 0, Limit: 10},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, f repository.MaterialFilter) ([]pgentities.Material, int, error) {
					assert.Equal(t, 10, f.Limit)
					assert.Equal(t, 0, f.Offset) // page 1 → offset 0
					return []pgentities.Material{{ID: uuid.New(), Title: "M1"}}, 1, nil
				}
			},
			wantTotal: 1,
			wantLimit: 10,
			wantPage:  1,
		},
		{
			name: "page 2 produces correct offset",
			req:  dto.ListMaterialsRequest{SchoolID: &schoolID, Page: 2, Limit: 10},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, f repository.MaterialFilter) ([]pgentities.Material, int, error) {
					assert.Equal(t, 10, f.Limit)
					assert.Equal(t, 10, f.Offset) // (2-1)*10 = 10
					return []pgentities.Material{{ID: uuid.New(), Title: "M2"}}, 15, nil
				}
			},
			wantTotal: 15,
			wantLimit: 10,
			wantPage:  2,
		},
		{
			name: "limit capped at 100",
			req:  dto.ListMaterialsRequest{Limit: 500},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, f repository.MaterialFilter) ([]pgentities.Material, int, error) {
					assert.Equal(t, 100, f.Limit)
					return nil, 0, nil
				}
			},
			wantTotal: 0,
			wantLimit: 100,
			wantPage:  1,
		},
		{
			name: "repository error",
			req:  dto.ListMaterialsRequest{},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.ListFn = func(_ context.Context, _ repository.MaterialFilter) ([]pgentities.Material, int, error) {
					return nil, 0, errors.NewDatabaseError("list", nil)
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := newTestMaterialService(repo, nil)
			resp, err := svc.List(ctx, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, tt.wantTotal, resp.Total)
			assert.Equal(t, tt.wantLimit, resp.Limit)
			assert.Equal(t, tt.wantPage, resp.Page)
		})
	}
}

func TestMaterialService_Update(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()
	title := "Updated Title"
	status := "published"

	tests := []struct {
		name      string
		req       dto.UpdateMaterialRequest
		setupRepo func(m *mock.MockMaterialRepository)
		wantErr   bool
		check     func(t *testing.T, resp *dto.MaterialResponse)
	}{
		{
			name: "happy path partial update",
			req:  dto.UpdateMaterialRequest{Title: &title},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{
						ID:    materialID,
						Title: "Old Title",
					}, nil
				}
				m.UpdateFn = func(_ context.Context, mat *pgentities.Material) error {
					assert.Equal(t, "Updated Title", mat.Title)
					return nil
				}
			},
			check: func(t *testing.T, resp *dto.MaterialResponse) {
				assert.Equal(t, "Updated Title", resp.Title)
			},
		},
		{
			name: "update status",
			req:  dto.UpdateMaterialRequest{Status: &status},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{
						ID:     materialID,
						Status: "draft",
					}, nil
				}
				m.UpdateFn = func(_ context.Context, mat *pgentities.Material) error {
					assert.Equal(t, "published", mat.Status)
					return nil
				}
			},
			check: func(t *testing.T, resp *dto.MaterialResponse) {
				assert.Equal(t, "published", resp.Status)
			},
		},
		{
			name: "not found on get",
			req:  dto.UpdateMaterialRequest{Title: &title},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return nil, errors.NewNotFoundError("material")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := newTestMaterialService(repo, nil)
			resp, err := svc.Update(ctx, materialID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}

func TestMaterialService_GetWithVersions(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()

	tests := []struct {
		name      string
		setupRepo func(m *mock.MockMaterialRepository)
		wantErr   bool
		wantVers  int
	}{
		{
			name: "happy path with versions",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetWithVersionsFn = func(_ context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error) {
					return &pgentities.Material{ID: id, Title: "Mat"},
						[]pgentities.MaterialVersion{
							{ID: uuid.New(), VersionNumber: 1, MaterialID: id, Title: "v1", ChangedBy: uuid.New(), CreatedAt: time.Now()},
							{ID: uuid.New(), VersionNumber: 2, MaterialID: id, Title: "v2", ChangedBy: uuid.New(), CreatedAt: time.Now()},
						}, nil
				}
			},
			wantVers: 2,
		},
		{
			name: "not found",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetWithVersionsFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error) {
					return nil, nil, errors.NewNotFoundError("material")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := newTestMaterialService(repo, nil)
			resp, err := svc.GetWithVersions(ctx, materialID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Len(t, resp.Versions, tt.wantVers)
		})
	}
}

func TestMaterialService_GenerateUploadURL(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()

	tests := []struct {
		name    string
		s3Nil   bool
		wantErr bool
		errCode errors.ErrorCode
	}{
		{
			name:    "s3 not configured",
			s3Nil:   true,
			wantErr: true,
			errCode: errors.ErrorCodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			svc := NewMaterialService(repo, nil, nil, mock.MockLogger{}, "test", nil)

			resp, err := svc.GenerateUploadURL(ctx, materialID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)
				if tt.errCode != "" {
					appErr, ok := errors.GetAppError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, appErr.Code)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestMaterialService_GenerateDownloadURL(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()

	tests := []struct {
		name      string
		setupRepo func(m *mock.MockMaterialRepository)
		wantErr   bool
		errCode   errors.ErrorCode
	}{
		{
			name: "s3 not configured",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, FileURL: "s3://bucket/key"}, nil
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeInternal,
		},
		{
			name: "material not found",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return nil, errors.NewNotFoundError("material")
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeNotFound,
		},
		{
			name: "no file uploaded but s3 nil - returns internal error first",
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, FileURL: ""}, nil
				}
			},
			wantErr: true,
			errCode: errors.ErrorCodeInternal, // S3 nil is checked before FileURL
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}

			svc := NewMaterialService(repo, nil, nil, mock.MockLogger{}, "test", nil)
			resp, err := svc.GenerateDownloadURL(ctx, materialID)

			if tt.wantErr {
				require.Error(t, err)
				assert.Nil(t, resp)
				if tt.errCode != "" {
					appErr, ok := errors.GetAppError(err)
					require.True(t, ok)
					assert.Equal(t, tt.errCode, appErr.Code)
				}
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestMaterialService_NotifyUploadComplete(t *testing.T) {
	ctx := context.Background()
	materialID := uuid.New()

	tests := []struct {
		name      string
		req       dto.UploadCompleteRequest
		setupRepo func(m *mock.MockMaterialRepository)
		setupPub  func(p *mock.MockPublisher)
		wantErr   bool
		check     func(t *testing.T, resp *dto.MaterialResponse)
	}{
		{
			name: "happy path with publisher",
			req:  dto.UploadCompleteRequest{FileURL: "s3://b/k", FileType: "pdf", FileSizeBytes: 1024},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, SchoolID: uuid.New(), UploadedByTeacherID: uuid.New()}, nil
				}
				m.UpdateFn = func(_ context.Context, mat *pgentities.Material) error {
					assert.Equal(t, "uploaded", mat.Status)
					assert.Equal(t, "s3://b/k", mat.FileURL)
					return nil
				}
			},
			setupPub: func(p *mock.MockPublisher) {
				p.PublishFn = func(_ context.Context, exchange, routingKey string, _ interface{}) error {
					assert.Equal(t, "test-exchange", exchange)
					assert.Equal(t, "material.uploaded", routingKey)
					return nil
				}
			},
			check: func(t *testing.T, resp *dto.MaterialResponse) {
				assert.Equal(t, "uploaded", resp.Status)
				assert.Equal(t, "s3://b/k", resp.FileURL)
			},
		},
		{
			name: "happy path without publisher",
			req:  dto.UploadCompleteRequest{FileURL: "s3://b/k", FileType: "pdf", FileSizeBytes: 1024},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return &pgentities.Material{ID: materialID, SchoolID: uuid.New(), UploadedByTeacherID: uuid.New()}, nil
				}
				m.UpdateFn = func(_ context.Context, _ *pgentities.Material) error {
					return nil
				}
			},
			check: func(t *testing.T, resp *dto.MaterialResponse) {
				assert.Equal(t, "uploaded", resp.Status)
			},
		},
		{
			name: "material not found",
			req:  dto.UploadCompleteRequest{FileURL: "s3://b/k", FileType: "pdf", FileSizeBytes: 1024},
			setupRepo: func(m *mock.MockMaterialRepository) {
				m.GetByIDFn = func(_ context.Context, _ uuid.UUID) (*pgentities.Material, error) {
					return nil, errors.NewNotFoundError("material")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mock.MockMaterialRepository{}
			pub := &mock.MockPublisher{}
			if tt.setupRepo != nil {
				tt.setupRepo(repo)
			}
			if tt.setupPub != nil {
				tt.setupPub(pub)
			}

			var svc *MaterialService
			if tt.setupPub != nil {
				svc = NewMaterialService(repo, nil, pub, mock.MockLogger{}, "test-exchange", nil)
			} else {
				svc = NewMaterialService(repo, nil, nil, mock.MockLogger{}, "test-exchange", nil)
			}

			resp, err := svc.NotifyUploadComplete(ctx, materialID, tt.req)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, resp)
			if tt.check != nil {
				tt.check(t, resp)
			}
		})
	}
}
