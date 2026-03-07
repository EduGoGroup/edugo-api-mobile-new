package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"

	pgrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/persistence/postgres/repository"
)

func TestAssessmentAndAttemptCRUD(t *testing.T) {
	db := getTestDB(t)
	assessRepo := pgrepo.NewAssessmentRepository(db)
	attemptRepo := pgrepo.NewAttemptRepository(db)
	materialRepo := pgrepo.NewMaterialRepository(db)
	ctx := context.Background()

	schoolID := uuid.New()
	teacherID := uuid.New()
	studentID := uuid.New()
	materialID := uuid.New()
	assessmentID := uuid.New()

	// Setup foreign keys
	setupTestSchool(t, db, schoolID)
	setupTestUser(t, db, teacherID, schoolID)
	setupTestUser(t, db, studentID, schoolID)

	t.Cleanup(func() {
		cleanupMaterial(t, db, materialID)
		db.Exec("DELETE FROM users WHERE id IN (?, ?)", teacherID, studentID)
		db.Exec("DELETE FROM schools WHERE id = ?", schoolID)
	})

	// Create material first
	t.Run("Setup material", func(t *testing.T) {
		err := materialRepo.Create(ctx, &pgentities.Material{
			ID:                  materialID,
			SchoolID:            schoolID,
			UploadedByTeacherID: teacherID,
			Title:               "Assessment Test Material",
			Status:              "published",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		})
		require.NoError(t, err)
	})

	// Create assessment
	t.Run("Create assessment", func(t *testing.T) {
		err := db.Exec(`INSERT INTO assessment.assessment (id, mongo_document_id, questions_count, status, created_at, updated_at)
			VALUES (?, ?, ?, ?, NOW(), NOW())`,
			assessmentID, "integration_test_doc", 2, "published").Error
		require.NoError(t, err)

		// Create junction table entry
		err = db.Exec(`INSERT INTO assessment.assessment_materials (id, assessment_id, material_id, sort_order, created_at)
			VALUES (?, ?, ?, 0, NOW())`,
			uuid.New(), assessmentID, materialID).Error
		require.NoError(t, err)
	})

	t.Run("Get assessment by material ID", func(t *testing.T) {
		assessment, err := assessRepo.GetByMaterialID(ctx, materialID)
		require.NoError(t, err)
		require.NotNil(t, assessment)
		assert.Equal(t, assessmentID, assessment.ID)
		assert.Equal(t, 2, assessment.QuestionsCount)
	})

	t.Run("Get assessment by ID", func(t *testing.T) {
		assessment, err := assessRepo.GetByID(ctx, assessmentID)
		require.NoError(t, err)
		require.NotNil(t, assessment)
		assert.Equal(t, assessmentID, assessment.ID)
	})

	// Create attempt with answers
	t.Run("Create attempt with answers", func(t *testing.T) {
		attemptID := uuid.New()
		now := time.Now()
		score := 1.5
		maxScore := 2.0
		pct := 75.0

		attempt := &pgentities.AssessmentAttempt{
			ID:           attemptID,
			AssessmentID: assessmentID,
			StudentID:    studentID,
			StartedAt:    now,
			CompletedAt:  &now,
			Score:        &score,
			MaxScore:     &maxScore,
			Percentage:   &pct,
			Status:       "completed",
			CreatedAt:    now,
			UpdatedAt:    now,
		}

		isCorrect := true
		isWrong := false
		p1 := 1.0
		p05 := 0.5
		mp := 1.0
		ans1 := "B"
		ans2 := "false"

		answers := []pgentities.AssessmentAttemptAnswer{
			{
				ID:            uuid.New(),
				AttemptID:     attemptID,
				QuestionIndex: 0,
				StudentAnswer: &ans1,
				IsCorrect:     &isCorrect,
				PointsEarned:  &p1,
				MaxPoints:     &mp,
				AnsweredAt:    now,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
			{
				ID:            uuid.New(),
				AttemptID:     attemptID,
				QuestionIndex: 1,
				StudentAnswer: &ans2,
				IsCorrect:     &isWrong,
				PointsEarned:  &p05,
				MaxPoints:     &mp,
				AnsweredAt:    now,
				CreatedAt:     now,
				UpdatedAt:     now,
			},
		}

		err := attemptRepo.Create(ctx, attempt, answers)
		require.NoError(t, err)

		// Get attempt by ID
		found, err := attemptRepo.GetByID(ctx, attemptID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, "completed", found.Status)
		require.NotNil(t, found.Score)
		assert.InDelta(t, 1.5, *found.Score, 0.01)

		// Get answers
		foundAnswers, err := attemptRepo.GetAnswersByAttemptID(ctx, attemptID)
		require.NoError(t, err)
		assert.Len(t, foundAnswers, 2)

		// Count attempts
		count, err := attemptRepo.CountByAssessmentAndStudent(ctx, assessmentID, studentID)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		// List by user
		attempts, total, err := attemptRepo.ListByUserID(ctx, studentID, 10, 0, sharedrepo.ListFilters{})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.GreaterOrEqual(t, len(attempts), 1)
	})
}
