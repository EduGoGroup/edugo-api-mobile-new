package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
)

// ProgressRepository implements repository.ProgressRepository using PostgreSQL.
type ProgressRepository struct {
	db *sql.DB
}

// NewProgressRepository creates a new ProgressRepository.
func NewProgressRepository(db *sql.DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// Upsert inserts or updates a progress record (INSERT ON CONFLICT).
func (r *ProgressRepository) Upsert(ctx context.Context, p *pgentities.Progress) error {
	query := `
		INSERT INTO progress (material_id, user_id, percentage, last_page, status, last_accessed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (material_id, user_id)
		DO UPDATE SET
			percentage = EXCLUDED.percentage,
			last_page = EXCLUDED.last_page,
			status = EXCLUDED.status,
			last_accessed_at = EXCLUDED.last_accessed_at,
			updated_at = EXCLUDED.updated_at`

	_, err := r.db.ExecContext(ctx, query,
		p.MaterialID, p.UserID, p.Percentage, p.LastPage,
		p.Status, p.LastAccessedAt, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return errors.NewDatabaseError("upsert progress", err)
	}
	return nil
}

// GetByMaterialAndUser retrieves a progress record for a material/user pair.
func (r *ProgressRepository) GetByMaterialAndUser(ctx context.Context, materialID, userID uuid.UUID) (*pgentities.Progress, error) {
	query := `
		SELECT material_id, user_id, percentage, last_page, status, last_accessed_at, created_at, updated_at
		FROM progress
		WHERE material_id = $1 AND user_id = $2`

	var p pgentities.Progress
	err := r.db.QueryRowContext(ctx, query, materialID, userID).Scan(
		&p.MaterialID, &p.UserID, &p.Percentage, &p.LastPage,
		&p.Status, &p.LastAccessedAt, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("progress")
		}
		return nil, errors.NewDatabaseError("get progress", err)
	}
	return &p, nil
}
