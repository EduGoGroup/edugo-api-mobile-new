package integration

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	pgrepo "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/persistence/postgres/repository"
)

func TestMaterialCRUD(t *testing.T) {
	db := getTestDB(t)
	repo := pgrepo.NewMaterialRepository(db)
	ctx := context.Background()

	schoolID := uuid.New()
	teacherID := uuid.New()
	materialID := uuid.New()

	// Setup foreign keys
	setupTestSchool(t, db, schoolID)
	setupTestUser(t, db, teacherID, schoolID)

	t.Cleanup(func() {
		cleanupMaterial(t, db, materialID)
		db.Exec("DELETE FROM users WHERE id = ?", teacherID)
		db.Exec("DELETE FROM schools WHERE id = ?", schoolID)
	})

	t.Run("Create material", func(t *testing.T) {
		desc := "A test description"
		material := &pgentities.Material{
			ID:                  materialID,
			SchoolID:            schoolID,
			UploadedByTeacherID: teacherID,
			Title:               "Integration Test Material",
			Description:         &desc,
			Status:              "draft",
			IsPublic:            true,
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}

		err := repo.Create(ctx, material)
		require.NoError(t, err)
	})

	t.Run("Get by ID", func(t *testing.T) {
		material, err := repo.GetByID(ctx, materialID)
		require.NoError(t, err)
		require.NotNil(t, material)
		assert.Equal(t, "Integration Test Material", material.Title)
		assert.Equal(t, schoolID, material.SchoolID)
		assert.Equal(t, teacherID, material.UploadedByTeacherID)
		assert.Equal(t, "draft", material.Status)
		assert.True(t, material.IsPublic)
	})

	t.Run("Update material", func(t *testing.T) {
		material, err := repo.GetByID(ctx, materialID)
		require.NoError(t, err)

		material.Title = "Updated Title"
		material.Status = "published"
		material.UpdatedAt = time.Now()

		err = repo.Update(ctx, material)
		require.NoError(t, err)

		updated, err := repo.GetByID(ctx, materialID)
		require.NoError(t, err)
		assert.Equal(t, "Updated Title", updated.Title)
		assert.Equal(t, "published", updated.Status)
	})

	t.Run("List with filter", func(t *testing.T) {
		filter := repository.MaterialFilter{
			SchoolID: &schoolID,
			Limit:    10,
			Offset:   0,
		}
		materials, total, err := repo.List(ctx, filter)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, 1)
		assert.GreaterOrEqual(t, len(materials), 1)
	})

	t.Run("Get by ID not found", func(t *testing.T) {
		_, err := repo.GetByID(ctx, uuid.New())
		require.Error(t, err)
	})
}
