package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	pgrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/persistence/postgres/repository"
)

func TestProgressUpsert(t *testing.T) {
	db := getTestDB(t)
	repo := pgrepo.NewProgressRepository(db)
	materialRepo := pgrepo.NewMaterialRepository(db)
	ctx := context.Background()

	schoolID := uuid.New()
	teacherID := uuid.New()
	userID := uuid.New()
	materialID := uuid.New()

	// Setup foreign keys
	setupTestSchool(t, db, schoolID)
	setupTestUser(t, db, teacherID, schoolID)
	setupTestUser(t, db, userID, schoolID)

	// Create material
	err := materialRepo.Create(ctx, &pgentities.Material{
		ID:                  materialID,
		SchoolID:            schoolID,
		UploadedByTeacherID: teacherID,
		Title:               "Progress Test Material",
		Status:              "published",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Exec("DELETE FROM progress WHERE material_id = $1", materialID)
		cleanupMaterial(t, db, materialID)
		db.Exec("DELETE FROM users WHERE id IN ($1, $2)", teacherID, userID)
		db.Exec("DELETE FROM schools WHERE id = $1", schoolID)
	})

	t.Run("Insert new progress", func(t *testing.T) {
		now := time.Now()
		progress := &pgentities.Progress{
			MaterialID:     materialID,
			UserID:         userID,
			Percentage:     25,
			LastPage:       3,
			Status:         "in_progress",
			LastAccessedAt: now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		err := repo.Upsert(ctx, progress)
		require.NoError(t, err)

		found, err := repo.GetByMaterialAndUser(ctx, materialID, userID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, 25, found.Percentage)
		assert.Equal(t, 3, found.LastPage)
		assert.Equal(t, "in_progress", found.Status)
	})

	t.Run("Update existing progress (upsert)", func(t *testing.T) {
		now := time.Now()
		progress := &pgentities.Progress{
			MaterialID:     materialID,
			UserID:         userID,
			Percentage:     100,
			LastPage:       20,
			Status:         "completed",
			LastAccessedAt: now,
			CreatedAt:      now,
			UpdatedAt:      now,
		}

		err := repo.Upsert(ctx, progress)
		require.NoError(t, err)

		found, err := repo.GetByMaterialAndUser(ctx, materialID, userID)
		require.NoError(t, err)
		require.NotNil(t, found)
		assert.Equal(t, 100, found.Percentage)
		assert.Equal(t, 20, found.LastPage)
		assert.Equal(t, "completed", found.Status)
	})

	t.Run("Not found for different user", func(t *testing.T) {
		_, err := repo.GetByMaterialAndUser(ctx, materialID, uuid.New())
		require.Error(t, err)
	})
}
