package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// Ensure pq is imported for its side effects (PostgreSQL driver).
var _ = pq.ErrNotSupported

// MaterialRepository implements repository.MaterialRepository using PostgreSQL.
type MaterialRepository struct {
	db *sql.DB
}

// NewMaterialRepository creates a new MaterialRepository.
func NewMaterialRepository(db *sql.DB) *MaterialRepository {
	return &MaterialRepository{db: db}
}

// Create inserts a new material.
func (r *MaterialRepository) Create(ctx context.Context, m *pgentities.Material) error {
	query := `
		INSERT INTO materials (
			id, school_id, uploaded_by_teacher_id, academic_unit_id,
			title, description, subject, grade,
			file_url, file_type, file_size_bytes,
			status, is_public, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`

	_, err := r.db.ExecContext(ctx, query,
		m.ID, m.SchoolID, m.UploadedByTeacherID, m.AcademicUnitID,
		m.Title, m.Description, m.Subject, m.Grade,
		m.FileURL, m.FileType, m.FileSizeBytes,
		m.Status, m.IsPublic, m.CreatedAt, m.UpdatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("create material", err)
	}
	return nil
}

// GetByID retrieves a material by its ID.
func (r *MaterialRepository) GetByID(ctx context.Context, id uuid.UUID) (*pgentities.Material, error) {
	query := `
		SELECT id, school_id, uploaded_by_teacher_id, academic_unit_id,
			title, description, subject, grade,
			file_url, file_type, file_size_bytes,
			status, processing_started_at, processing_completed_at,
			is_public, created_at, updated_at, deleted_at
		FROM materials
		WHERE id = $1 AND deleted_at IS NULL`

	var m pgentities.Material
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&m.ID, &m.SchoolID, &m.UploadedByTeacherID, &m.AcademicUnitID,
		&m.Title, &m.Description, &m.Subject, &m.Grade,
		&m.FileURL, &m.FileType, &m.FileSizeBytes,
		&m.Status, &m.ProcessingStartedAt, &m.ProcessingCompletedAt,
		&m.IsPublic, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("material")
		}
		return nil, errors.NewDatabaseError("get material", err)
	}
	return &m, nil
}

// List retrieves materials matching the given filter.
func (r *MaterialRepository) List(ctx context.Context, filter repository.MaterialFilter) ([]pgentities.Material, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	conditions = append(conditions, "deleted_at IS NULL")

	if filter.SchoolID != nil {
		conditions = append(conditions, fmt.Sprintf("school_id = $%d", argIdx))
		args = append(args, *filter.SchoolID)
		argIdx++
	}
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("uploaded_by_teacher_id = $%d", argIdx))
		args = append(args, *filter.AuthorID)
		argIdx++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, *filter.Status)
		argIdx++
	}

	where := strings.Join(conditions, " AND ")

	// Count total
	countQuery := "SELECT COUNT(*) FROM materials WHERE " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, errors.NewDatabaseError("count materials", err)
	}

	// Fetch page
	dataQuery := fmt.Sprintf(`
		SELECT id, school_id, uploaded_by_teacher_id, academic_unit_id,
			title, description, subject, grade,
			file_url, file_type, file_size_bytes,
			status, processing_started_at, processing_completed_at,
			is_public, created_at, updated_at, deleted_at
		FROM materials
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d`, where, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, dataQuery, args...)
	if err != nil {
		return nil, 0, errors.NewDatabaseError("list materials", err)
	}
	defer rows.Close()

	var materials []pgentities.Material
	for rows.Next() {
		var m pgentities.Material
		if err := rows.Scan(
			&m.ID, &m.SchoolID, &m.UploadedByTeacherID, &m.AcademicUnitID,
			&m.Title, &m.Description, &m.Subject, &m.Grade,
			&m.FileURL, &m.FileType, &m.FileSizeBytes,
			&m.Status, &m.ProcessingStartedAt, &m.ProcessingCompletedAt,
			&m.IsPublic, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt,
		); err != nil {
			return nil, 0, errors.NewDatabaseError("scan material", err)
		}
		materials = append(materials, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, errors.NewDatabaseError("iterate materials", err)
	}

	return materials, total, nil
}

// Update modifies an existing material.
func (r *MaterialRepository) Update(ctx context.Context, m *pgentities.Material) error {
	query := `
		UPDATE materials SET
			title = $2, description = $3, subject = $4, grade = $5,
			file_url = $6, file_type = $7, file_size_bytes = $8,
			status = $9, is_public = $10, updated_at = $11
		WHERE id = $1 AND deleted_at IS NULL`

	result, err := r.db.ExecContext(ctx, query,
		m.ID, m.Title, m.Description, m.Subject, m.Grade,
		m.FileURL, m.FileType, m.FileSizeBytes,
		m.Status, m.IsPublic, m.UpdatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("update material", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.NewNotFoundError("material")
	}
	return nil
}

// GetWithVersions returns a material with its version history.
func (r *MaterialRepository) GetWithVersions(ctx context.Context, id uuid.UUID) (*pgentities.Material, []pgentities.MaterialVersion, error) {
	material, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, nil, err
	}

	query := `
		SELECT id, material_id, version_number, title, content_url, changed_by, created_at
		FROM material_versions
		WHERE material_id = $1
		ORDER BY version_number DESC`

	rows, err := r.db.QueryContext(ctx, query, id)
	if err != nil {
		return nil, nil, errors.NewDatabaseError("list material versions", err)
	}
	defer rows.Close()

	var versions []pgentities.MaterialVersion
	for rows.Next() {
		var v pgentities.MaterialVersion
		if err := rows.Scan(&v.ID, &v.MaterialID, &v.VersionNumber, &v.Title, &v.ContentURL, &v.ChangedBy, &v.CreatedAt); err != nil {
			return nil, nil, errors.NewDatabaseError("scan material version", err)
		}
		versions = append(versions, v)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, errors.NewDatabaseError("iterate material versions", err)
	}

	return material, versions, nil
}
