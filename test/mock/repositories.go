package mock

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	mongoentities "github.com/EduGoGroup/edugo-infrastructure/mongodb/entities"
	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
	"github.com/EduGoGroup/edugo-shared/logger"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// ---------------------------------------------------------------------------
// MaterialRepository mock
// ---------------------------------------------------------------------------

// MockMaterialRepository is a mock implementation of repository.MaterialRepository.
type MockMaterialRepository struct {
	CreateFn          func(ctx context.Context, material *pgentities.Material) error
	GetByIDFn         func(ctx context.Context, id uuid.UUID) (*pgentities.Material, error)
	ListFn            func(ctx context.Context, filter repository.MaterialFilter) ([]pgentities.Material, int, error)
	UpdateFn          func(ctx context.Context, material *pgentities.Material) error
	GetWithVersionsFn func(ctx context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error)
}

func (m *MockMaterialRepository) Create(ctx context.Context, material *pgentities.Material) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, material)
	}
	return nil
}

func (m *MockMaterialRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Material, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockMaterialRepository) List(ctx context.Context, filter repository.MaterialFilter) ([]pgentities.Material, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *MockMaterialRepository) Update(ctx context.Context, material *pgentities.Material) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, material)
	}
	return nil
}

func (m *MockMaterialRepository) GetWithVersions(ctx context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error) {
	if m.GetWithVersionsFn != nil {
		return m.GetWithVersionsFn(ctx, id)
	}
	return nil, nil, nil
}

// ---------------------------------------------------------------------------
// AssessmentRepository mock
// ---------------------------------------------------------------------------

// MockAssessmentRepository is a mock implementation of repository.AssessmentRepository.
type MockAssessmentRepository struct {
	GetByIDFn             func(ctx context.Context, id uuid.UUID) (*pgentities.Assessment, error)
	GetByMaterialIDFn     func(ctx context.Context, materialID uuid.UUID) (*pgentities.Assessment, error)
	ListFn                func(ctx context.Context, filter repository.AssessmentFilter) ([]pgentities.Assessment, int, error)
	CreateFn              func(ctx context.Context, assessment *pgentities.Assessment) error
	UpdateFn              func(ctx context.Context, assessment *pgentities.Assessment) error
	SoftDeleteFn          func(ctx context.Context, id uuid.UUID) error
	UpdateQuestionsCountFn func(ctx context.Context, id uuid.UUID, count int) error
}

func (m *MockAssessmentRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Assessment, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockAssessmentRepository) GetByMaterialID(ctx context.Context, materialID uuid.UUID) (*pgentities.Assessment, error) {
	if m.GetByMaterialIDFn != nil {
		return m.GetByMaterialIDFn(ctx, materialID)
	}
	return nil, nil
}

func (m *MockAssessmentRepository) List(ctx context.Context, filter repository.AssessmentFilter) ([]pgentities.Assessment, int, error) {
	if m.ListFn != nil {
		return m.ListFn(ctx, filter)
	}
	return nil, 0, nil
}

func (m *MockAssessmentRepository) Create(ctx context.Context, assessment *pgentities.Assessment) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, assessment)
	}
	return nil
}

func (m *MockAssessmentRepository) Update(ctx context.Context, assessment *pgentities.Assessment) error {
	if m.UpdateFn != nil {
		return m.UpdateFn(ctx, assessment)
	}
	return nil
}

func (m *MockAssessmentRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if m.SoftDeleteFn != nil {
		return m.SoftDeleteFn(ctx, id)
	}
	return nil
}

func (m *MockAssessmentRepository) UpdateQuestionsCount(ctx context.Context, id uuid.UUID, count int) error {
	if m.UpdateQuestionsCountFn != nil {
		return m.UpdateQuestionsCountFn(ctx, id, count)
	}
	return nil
}

// ---------------------------------------------------------------------------
// AssessmentMaterialRepository mock
// ---------------------------------------------------------------------------

// MockAssessmentMaterialRepository is a mock implementation of repository.AssessmentMaterialRepository.
type MockAssessmentMaterialRepository struct {
	ReplaceForAssessmentFn         func(ctx context.Context, assessmentID uuid.UUID, materialIDs []uuid.UUID) error
	GetByAssessmentFn              func(ctx context.Context, assessmentID uuid.UUID) ([]pgentities.AssessmentMaterial, error)
	GetMaterialTitlesFn            func(ctx context.Context, materialIDs []uuid.UUID) (map[uuid.UUID]string, error)
	GetAssessmentIDsByMaterialIDFn func(ctx context.Context, materialID uuid.UUID) ([]uuid.UUID, error)
}

func (m *MockAssessmentMaterialRepository) ReplaceForAssessment(ctx context.Context, assessmentID uuid.UUID, materialIDs []uuid.UUID) error {
	if m.ReplaceForAssessmentFn != nil {
		return m.ReplaceForAssessmentFn(ctx, assessmentID, materialIDs)
	}
	return nil
}

func (m *MockAssessmentMaterialRepository) GetByAssessment(ctx context.Context, assessmentID uuid.UUID) ([]pgentities.AssessmentMaterial, error) {
	if m.GetByAssessmentFn != nil {
		return m.GetByAssessmentFn(ctx, assessmentID)
	}
	return nil, nil
}

func (m *MockAssessmentMaterialRepository) GetMaterialTitles(ctx context.Context, materialIDs []uuid.UUID) (map[uuid.UUID]string, error) {
	if m.GetMaterialTitlesFn != nil {
		return m.GetMaterialTitlesFn(ctx, materialIDs)
	}
	return map[uuid.UUID]string{}, nil
}

func (m *MockAssessmentMaterialRepository) GetAssessmentIDsByMaterialID(ctx context.Context, materialID uuid.UUID) ([]uuid.UUID, error) {
	if m.GetAssessmentIDsByMaterialIDFn != nil {
		return m.GetAssessmentIDsByMaterialIDFn(ctx, materialID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// AttemptRepository mock
// ---------------------------------------------------------------------------

// MockAttemptRepository is a mock implementation of repository.AttemptRepository.
type MockAttemptRepository struct {
	CreateFn                      func(ctx context.Context, attempt *pgentities.AssessmentAttempt, answers []pgentities.AssessmentAttemptAnswer) error
	GetByIDFn                     func(ctx context.Context, id uuid.UUID) (*pgentities.AssessmentAttempt, error)
	GetAnswersByAttemptIDFn       func(ctx context.Context, attemptID uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error)
	ListByUserIDFn                func(ctx context.Context, userID uuid.UUID, limit, offset int, filters sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error)
	CountByAssessmentAndStudentFn func(ctx context.Context, assessmentID, studentID uuid.UUID) (int, error)
	CreateAttemptOnlyFn                    func(ctx context.Context, attempt *pgentities.AssessmentAttempt) error
	GetInProgressByStudentAndAssessmentFn  func(ctx context.Context, studentID, assessmentID uuid.UUID) (*pgentities.AssessmentAttempt, error)
	UpsertAnswerFn                         func(ctx context.Context, answer *pgentities.AssessmentAttemptAnswer) error
	UpdateAttemptFn                        func(ctx context.Context, attempt *pgentities.AssessmentAttempt) error
	UpdateAnswersFn                        func(ctx context.Context, answers []pgentities.AssessmentAttemptAnswer) error
}

func (m *MockAttemptRepository) Create(ctx context.Context, attempt *pgentities.AssessmentAttempt, answers []pgentities.AssessmentAttemptAnswer) error {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, attempt, answers)
	}
	return nil
}

func (m *MockAttemptRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.AssessmentAttempt, error) {
	if m.GetByIDFn != nil {
		return m.GetByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *MockAttemptRepository) GetAnswersByAttemptID(ctx context.Context, attemptID uuid.UUID) ([]pgentities.AssessmentAttemptAnswer, error) {
	if m.GetAnswersByAttemptIDFn != nil {
		return m.GetAnswersByAttemptIDFn(ctx, attemptID)
	}
	return nil, nil
}

func (m *MockAttemptRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int, filters sharedrepo.ListFilters) ([]pgentities.AssessmentAttempt, int, error) {
	if m.ListByUserIDFn != nil {
		return m.ListByUserIDFn(ctx, userID, limit, offset, filters)
	}
	return nil, 0, nil
}

func (m *MockAttemptRepository) CountByAssessmentAndStudent(ctx context.Context, assessmentID, studentID uuid.UUID) (int, error) {
	if m.CountByAssessmentAndStudentFn != nil {
		return m.CountByAssessmentAndStudentFn(ctx, assessmentID, studentID)
	}
	return 0, nil
}

func (m *MockAttemptRepository) CreateAttemptOnly(ctx context.Context, attempt *pgentities.AssessmentAttempt) error {
	if m.CreateAttemptOnlyFn != nil {
		return m.CreateAttemptOnlyFn(ctx, attempt)
	}
	return nil
}

func (m *MockAttemptRepository) GetInProgressByStudentAndAssessment(ctx context.Context, studentID, assessmentID uuid.UUID) (*pgentities.AssessmentAttempt, error) {
	if m.GetInProgressByStudentAndAssessmentFn != nil {
		return m.GetInProgressByStudentAndAssessmentFn(ctx, studentID, assessmentID)
	}
	return nil, nil
}

func (m *MockAttemptRepository) UpsertAnswer(ctx context.Context, answer *pgentities.AssessmentAttemptAnswer) error {
	if m.UpsertAnswerFn != nil {
		return m.UpsertAnswerFn(ctx, answer)
	}
	return nil
}

func (m *MockAttemptRepository) UpdateAttempt(ctx context.Context, attempt *pgentities.AssessmentAttempt) error {
	if m.UpdateAttemptFn != nil {
		return m.UpdateAttemptFn(ctx, attempt)
	}
	return nil
}

func (m *MockAttemptRepository) UpdateAnswers(ctx context.Context, answers []pgentities.AssessmentAttemptAnswer) error {
	if m.UpdateAnswersFn != nil {
		return m.UpdateAnswersFn(ctx, answers)
	}
	return nil
}

// ---------------------------------------------------------------------------
// ProgressRepository mock
// ---------------------------------------------------------------------------

// MockProgressRepository is a mock implementation of repository.ProgressRepository.
type MockProgressRepository struct {
	UpsertFn               func(ctx context.Context, progress *pgentities.Progress) error
	GetByMaterialAndUserFn func(ctx context.Context, materialID, userID uuid.UUID) (*pgentities.Progress, error)
}

func (m *MockProgressRepository) Upsert(ctx context.Context, progress *pgentities.Progress) error {
	if m.UpsertFn != nil {
		return m.UpsertFn(ctx, progress)
	}
	return nil
}

func (m *MockProgressRepository) GetByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*pgentities.Progress, error) {
	if m.GetByMaterialAndUserFn != nil {
		return m.GetByMaterialAndUserFn(ctx, materialID, userID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// ScreenRepository mock
// ---------------------------------------------------------------------------

// MockScreenRepository is a mock implementation of repository.ScreenRepository.
type MockScreenRepository struct {
	GetScreenByKeyFn          func(ctx context.Context, screenKey string) (*repository.ScreenComposed, error)
	GetScreensByResourceKeyFn func(ctx context.Context, resourceKey string) ([]repository.ScreenComposed, error)
	GetNavigationFn           func(ctx context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error)
	UpsertPreferencesFn       func(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error
}

func (m *MockScreenRepository) GetScreenByKey(ctx context.Context, screenKey string) (*repository.ScreenComposed, error) {
	if m.GetScreenByKeyFn != nil {
		return m.GetScreenByKeyFn(ctx, screenKey)
	}
	return nil, nil
}

func (m *MockScreenRepository) GetScreensByResourceKey(ctx context.Context, resourceKey string) ([]repository.ScreenComposed, error) {
	if m.GetScreensByResourceKeyFn != nil {
		return m.GetScreensByResourceKeyFn(ctx, resourceKey)
	}
	return nil, nil
}

func (m *MockScreenRepository) GetNavigation(ctx context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
	if m.GetNavigationFn != nil {
		return m.GetNavigationFn(ctx, scope)
	}
	return nil, nil, nil
}

func (m *MockScreenRepository) UpsertPreferences(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error {
	if m.UpsertPreferencesFn != nil {
		return m.UpsertPreferencesFn(ctx, screenKey, userID, preferences)
	}
	return nil
}

// ---------------------------------------------------------------------------
// StatsRepository mock
// ---------------------------------------------------------------------------

// MockStatsRepository is a mock implementation of repository.StatsRepository.
type MockStatsRepository struct {
	CountMaterialsFn         func(ctx context.Context, schoolID *uuid.UUID) (int, error)
	CountCompletedProgressFn func(ctx context.Context, schoolID *uuid.UUID) (int, error)
	AverageAttemptScoreFn    func(ctx context.Context, schoolID *uuid.UUID) (float64, error)
	MaterialStatsFn          func(ctx context.Context, materialID uuid.UUID) (*repository.MaterialStatsResult, error)
}

func (m *MockStatsRepository) CountMaterials(ctx context.Context, schoolID *uuid.UUID) (int, error) {
	if m.CountMaterialsFn != nil {
		return m.CountMaterialsFn(ctx, schoolID)
	}
	return 0, nil
}

func (m *MockStatsRepository) CountCompletedProgress(ctx context.Context, schoolID *uuid.UUID) (int, error) {
	if m.CountCompletedProgressFn != nil {
		return m.CountCompletedProgressFn(ctx, schoolID)
	}
	return 0, nil
}

func (m *MockStatsRepository) AverageAttemptScore(ctx context.Context, schoolID *uuid.UUID) (float64, error) {
	if m.AverageAttemptScoreFn != nil {
		return m.AverageAttemptScoreFn(ctx, schoolID)
	}
	return 0, nil
}

func (m *MockStatsRepository) MaterialStats(ctx context.Context, materialID uuid.UUID) (*repository.MaterialStatsResult, error) {
	if m.MaterialStatsFn != nil {
		return m.MaterialStatsFn(ctx, materialID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// MongoAssessmentRepository mock
// ---------------------------------------------------------------------------

// MockMongoAssessmentRepository is a mock implementation of repository.MongoAssessmentRepository.
type MockMongoAssessmentRepository struct {
	GetByMaterialIDFn  func(ctx context.Context, materialID string) (*mongoentities.MaterialAssessment, error)
	GetByObjectIDFn    func(ctx context.Context, objectID string) (*mongoentities.MaterialAssessment, error)
	CreateFn           func(ctx context.Context, doc *mongoentities.MaterialAssessment) (string, error)
	DeleteFn           func(ctx context.Context, objectID string) error
	ReplaceQuestionsFn func(ctx context.Context, objectID string, questions []mongoentities.Question, totalPoints int) error
}

func (m *MockMongoAssessmentRepository) GetByMaterialID(ctx context.Context, materialID string) (*mongoentities.MaterialAssessment, error) {
	if m.GetByMaterialIDFn != nil {
		return m.GetByMaterialIDFn(ctx, materialID)
	}
	return nil, nil
}

func (m *MockMongoAssessmentRepository) GetByObjectID(ctx context.Context, objectID string) (*mongoentities.MaterialAssessment, error) {
	if m.GetByObjectIDFn != nil {
		return m.GetByObjectIDFn(ctx, objectID)
	}
	return nil, nil
}

func (m *MockMongoAssessmentRepository) Create(ctx context.Context, doc *mongoentities.MaterialAssessment) (string, error) {
	if m.CreateFn != nil {
		return m.CreateFn(ctx, doc)
	}
	return "000000000000000000000000", nil
}

func (m *MockMongoAssessmentRepository) Delete(ctx context.Context, objectID string) error {
	if m.DeleteFn != nil {
		return m.DeleteFn(ctx, objectID)
	}
	return nil
}

func (m *MockMongoAssessmentRepository) ReplaceQuestions(ctx context.Context, objectID string, questions []mongoentities.Question, totalPoints int) error {
	if m.ReplaceQuestionsFn != nil {
		return m.ReplaceQuestionsFn(ctx, objectID, questions, totalPoints)
	}
	return nil
}

// ---------------------------------------------------------------------------
// MongoSummaryRepository mock
// ---------------------------------------------------------------------------

// MockMongoSummaryRepository is a mock implementation of repository.MongoSummaryRepository.
type MockMongoSummaryRepository struct {
	GetByMaterialIDFn func(ctx context.Context, materialID string) (*mongoentities.MaterialSummary, error)
}

func (m *MockMongoSummaryRepository) GetByMaterialID(ctx context.Context, materialID string) (*mongoentities.MaterialSummary, error) {
	if m.GetByMaterialIDFn != nil {
		return m.GetByMaterialIDFn(ctx, materialID)
	}
	return nil, nil
}

// ---------------------------------------------------------------------------
// Publisher mock
// ---------------------------------------------------------------------------

// MockPublisher is a mock implementation of rabbit.Publisher.
type MockPublisher struct {
	PublishFn             func(ctx context.Context, exchange, routingKey string, body interface{}) error
	PublishWithPriorityFn func(ctx context.Context, exchange, routingKey string, body interface{}, priority uint8) error
	CloseFn               func() error
}

func (m *MockPublisher) Publish(ctx context.Context, exchange, routingKey string, body interface{}) error {
	if m.PublishFn != nil {
		return m.PublishFn(ctx, exchange, routingKey, body)
	}
	return nil
}

func (m *MockPublisher) PublishWithPriority(ctx context.Context, exchange, routingKey string, body interface{}, priority uint8) error {
	if m.PublishWithPriorityFn != nil {
		return m.PublishWithPriorityFn(ctx, exchange, routingKey, body, priority)
	}
	return nil
}

func (m *MockPublisher) Close() error {
	if m.CloseFn != nil {
		return m.CloseFn()
	}
	return nil
}

// ---------------------------------------------------------------------------
// Logger mock
// ---------------------------------------------------------------------------

// MockLogger is a no-op mock implementation of logger.Logger for tests.
type MockLogger struct{}

func (MockLogger) Debug(_ string, _ ...interface{})      {}
func (MockLogger) Info(_ string, _ ...interface{})       {}
func (MockLogger) Warn(_ string, _ ...interface{})       {}
func (MockLogger) Error(_ string, _ ...interface{})      {}
func (MockLogger) Fatal(_ string, _ ...interface{})      {}
func (l MockLogger) With(_ ...interface{}) logger.Logger { return l }
func (MockLogger) Sync() error                           { return nil }
