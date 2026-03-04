package service

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func newTestMgmtService(
	assessRepo *mock.MockAssessmentRepository,
	mongoRepo *mock.MockMongoAssessmentRepository,
) *AssessmentManagementService {
	return NewAssessmentManagementService(assessRepo, mongoRepo, mock.MockLogger{})
}

func TestMgmt_CreateAssessment(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()
	userID := uuid.New()

	tests := []struct {
		name       string
		req        dto.CreateAssessmentRequest
		setupMongo func(m *mock.MockMongoAssessmentRepository)
		setupPG    func(m *mock.MockAssessmentRepository)
		wantErr    bool
	}{
		{
			name: "happy path - manual assessment",
			req:  dto.CreateAssessmentRequest{Title: "Test Assessment"},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.CreateFn = func(_ context.Context, doc *mongoentities.MaterialAssessment) (string, error) {
					assert.Equal(t, "manual", doc.AIModel)
					assert.Empty(t, doc.Questions)
					return "aabbccddee0011223344ff00", nil
				}
			},
			setupPG: func(m *mock.MockAssessmentRepository) {
				m.CreateFn = func(_ context.Context, a *pgentities.Assessment) error {
					assert.Equal(t, "draft", a.Status)
					assert.Equal(t, "aabbccddee0011223344ff00", a.MongoDocumentID)
					assert.Equal(t, &schoolID, a.SchoolID)
					assert.Equal(t, &userID, a.CreatedByUserID)
					assert.Nil(t, a.MaterialID)
					return nil
				}
			},
		},
		{
			name: "mongo create fails",
			req:  dto.CreateAssessmentRequest{Title: "Fail"},
			setupMongo: func(m *mock.MockMongoAssessmentRepository) {
				m.CreateFn = func(_ context.Context, _ *mongoentities.MaterialAssessment) (string, error) {
					return "", errors.NewDatabaseError("create mongo", nil)
				}
			},
			setupPG: func(_ *mock.MockAssessmentRepository) {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{}
			mongoRepo := &mock.MockMongoAssessmentRepository{}
			tt.setupMongo(mongoRepo)
			tt.setupPG(assessRepo)

			svc := newTestMgmtService(assessRepo, mongoRepo)
			resp, err := svc.CreateAssessment(ctx, tt.req, schoolID, userID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotNil(t, resp)
			assert.Equal(t, "draft", resp.Status)
			assert.Equal(t, "Test Assessment", resp.Title)
		})
	}
}

func TestMgmt_UpdateAssessment_DraftOnly(t *testing.T) {
	ctx := context.Background()
	assessmentID := uuid.New()

	tests := []struct {
		name    string
		status  string
		wantErr bool
		errCode errors.ErrorCode
	}{
		{name: "draft can be edited", status: "draft"},
		{name: "published cannot be edited", status: "published", wantErr: true, errCode: errors.ErrorCodeBusinessRule},
		{name: "archived cannot be edited", status: "archived", wantErr: true, errCode: errors.ErrorCodeBusinessRule},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			title := "Original"
			assessRepo := &mock.MockAssessmentRepository{
				GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, Title: &title, Status: tt.status}, nil
				},
			}
			mongoRepo := &mock.MockMongoAssessmentRepository{}

			svc := newTestMgmtService(assessRepo, mongoRepo)
			newTitle := "Updated"
			resp, err := svc.UpdateAssessment(ctx, assessmentID, dto.UpdateAssessmentRequest{Title: &newTitle})

			if tt.wantErr {
				require.Error(t, err)
				appErr, ok := errors.GetAppError(err)
				require.True(t, ok)
				assert.Equal(t, tt.errCode, appErr.Code)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "Updated", resp.Title)
		})
	}
}

func TestMgmt_PublishAssessment(t *testing.T) {
	ctx := context.Background()
	assessmentID := uuid.New()

	tests := []struct {
		name    string
		status  string
		qCount  int
		wantErr bool
	}{
		{name: "draft with questions can be published", status: "draft", qCount: 3},
		{name: "generated with questions can be published", status: "generated", qCount: 5},
		{name: "draft with no questions cannot be published", status: "draft", qCount: 0, wantErr: true},
		{name: "published cannot be published again", status: "published", qCount: 3, wantErr: true},
		{name: "archived cannot be published", status: "archived", qCount: 3, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{
				GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, Status: tt.status, QuestionsCount: tt.qCount}, nil
				},
			}

			svc := newTestMgmtService(assessRepo, nil)
			resp, err := svc.PublishAssessment(ctx, assessmentID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "published", resp.Status)
		})
	}
}

func TestMgmt_ArchiveAssessment(t *testing.T) {
	ctx := context.Background()
	assessmentID := uuid.New()

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "published can be archived", status: "published"},
		{name: "draft cannot be archived", status: "draft", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{
				GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, Status: tt.status}, nil
				},
			}

			svc := newTestMgmtService(assessRepo, nil)
			resp, err := svc.ArchiveAssessment(ctx, assessmentID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "archived", resp.Status)
		})
	}
}

func TestMgmt_DeleteAssessment_DraftOnly(t *testing.T) {
	ctx := context.Background()
	assessmentID := uuid.New()

	tests := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{name: "draft can be deleted", status: "draft"},
		{name: "published cannot be deleted", status: "published", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{
				GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{ID: assessmentID, Status: tt.status}, nil
				},
			}

			svc := newTestMgmtService(assessRepo, nil)
			err := svc.DeleteAssessment(ctx, assessmentID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestMgmt_AddQuestion(t *testing.T) {
	ctx := context.Background()
	assessmentID := uuid.New()
	mongoDocID := "aabbccddee0011223344ff00"

	tests := []struct {
		name    string
		status  string
		wantErr bool
		wantLen int
	}{
		{name: "add question to draft", status: "draft", wantLen: 2},
		{name: "cannot add to published", status: "published", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{
				GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{
						ID: assessmentID, Status: tt.status,
						MongoDocumentID: mongoDocID, QuestionsCount: 1,
					}, nil
				},
			}
			mongoRepo := &mock.MockMongoAssessmentRepository{
				GetByObjectIDFn: func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionID: "q1", QuestionText: "Existing", QuestionType: "open", Points: 5},
						},
					}, nil
				},
			}

			svc := newTestMgmtService(assessRepo, mongoRepo)
			result, err := svc.AddQuestion(ctx, assessmentID, dto.CreateQuestionRequest{
				QuestionText:  "New Question",
				QuestionType:  "multiple_choice",
				CorrectAnswer: "A",
				Points:        10,
				Difficulty:    "medium",
			})

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, result, tt.wantLen)
		})
	}
}

func TestMgmt_DeleteQuestion(t *testing.T) {
	ctx := context.Background()
	assessmentID := uuid.New()
	mongoDocID := "aabbccddee0011223344ff00"

	tests := []struct {
		name    string
		idx     int
		wantErr bool
		wantLen int
	}{
		{name: "delete first question", idx: 0, wantLen: 1},
		{name: "index out of range", idx: 5, wantErr: true},
		{name: "negative index", idx: -1, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assessRepo := &mock.MockAssessmentRepository{
				GetByIDFn: func(_ context.Context, _ uuid.UUID) (*pgentities.Assessment, error) {
					return &pgentities.Assessment{
						ID: assessmentID, Status: "draft",
						MongoDocumentID: mongoDocID, QuestionsCount: 2,
					}, nil
				},
			}
			mongoRepo := &mock.MockMongoAssessmentRepository{
				GetByObjectIDFn: func(_ context.Context, _ string) (*mongoentities.MaterialAssessment, error) {
					return &mongoentities.MaterialAssessment{
						Questions: []mongoentities.Question{
							{QuestionID: "q1", QuestionText: "Q1", Points: 5},
							{QuestionID: "q2", QuestionText: "Q2", Points: 10},
						},
					}, nil
				},
			}

			svc := newTestMgmtService(assessRepo, mongoRepo)
			result, err := svc.DeleteQuestion(ctx, assessmentID, tt.idx)

			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Len(t, result, tt.wantLen)
		})
	}
}

func TestMgmt_ListAssessments(t *testing.T) {
	ctx := context.Background()
	schoolID := uuid.New()

	t.Run("happy path returns correct total", func(t *testing.T) {
		assessRepo := &mock.MockAssessmentRepository{
			ListFn: func(_ context.Context, filter repository.AssessmentFilter) ([]pgentities.Assessment, int, error) {
				assert.Equal(t, &schoolID, filter.SchoolID)
				title := "Test"
				return []pgentities.Assessment{
					{ID: uuid.New(), Title: &title, Status: "draft"},
				}, 1, nil
			},
		}

		svc := newTestMgmtService(assessRepo, nil)
		resp, err := svc.ListAssessments(ctx, schoolID, dto.ListAssessmentsRequest{Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Total)
	})

	t.Run("page defaults to 1 when zero", func(t *testing.T) {
		assessRepo := &mock.MockAssessmentRepository{
			ListFn: func(_ context.Context, filter repository.AssessmentFilter) ([]pgentities.Assessment, int, error) {
				assert.Equal(t, 0, filter.Offset) // page 1 → offset 0
				return nil, 0, nil
			},
		}

		svc := newTestMgmtService(assessRepo, nil)
		resp, err := svc.ListAssessments(ctx, schoolID, dto.ListAssessmentsRequest{Page: 0, Limit: 20})
		require.NoError(t, err)
		assert.Equal(t, 1, resp.Page) // defaulted to 1
	})

	t.Run("page 2 produces correct offset", func(t *testing.T) {
		const limit = 20
		assessRepo := &mock.MockAssessmentRepository{
			ListFn: func(_ context.Context, filter repository.AssessmentFilter) ([]pgentities.Assessment, int, error) {
				assert.Equal(t, limit, filter.Limit)
				assert.Equal(t, limit, filter.Offset) // (2-1)*20 = 20
				return nil, 0, nil
			},
		}

		svc := newTestMgmtService(assessRepo, nil)
		resp, err := svc.ListAssessments(ctx, schoolID, dto.ListAssessmentsRequest{Page: 2, Limit: limit})
		require.NoError(t, err)
		assert.Equal(t, 2, resp.Page)
	})
}
