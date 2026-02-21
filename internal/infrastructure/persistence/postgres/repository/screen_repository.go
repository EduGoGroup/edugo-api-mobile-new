package repository

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// ScreenRepository implements repository.ScreenRepository using PostgreSQL.
type ScreenRepository struct {
	db *sql.DB
}

// NewScreenRepository creates a new ScreenRepository.
func NewScreenRepository(db *sql.DB) *ScreenRepository {
	return &ScreenRepository{db: db}
}

// GetScreenByKey retrieves a composed screen (template + instance) by screen_key.
func (r *ScreenRepository) GetScreenByKey(ctx context.Context, screenKey string) (*repository.ScreenComposed, error) {
	query := `
		SELECT
			si.id, si.screen_key, si.template_id, si.name, si.description,
			si.slot_data, si.actions, si.data_endpoint, si.data_config,
			si.scope, si.required_permission, si.handler_key, si.is_active,
			si.created_at, si.updated_at,
			st.id, st.pattern, st.name, st.description, st.version,
			st.definition, st.is_active, st.created_at, st.updated_at
		FROM ui_config.screen_instances si
		JOIN ui_config.screen_templates st ON st.id = si.template_id
		WHERE si.screen_key = $1 AND si.is_active = true`

	var inst pgentities.ScreenInstance
	var tmpl pgentities.ScreenTemplate

	err := r.db.QueryRowContext(ctx, query, screenKey).Scan(
		&inst.ID, &inst.ScreenKey, &inst.TemplateID, &inst.Name, &inst.Description,
		&inst.SlotData, &inst.Actions, &inst.DataEndpoint, &inst.DataConfig,
		&inst.Scope, &inst.RequiredPermission, &inst.HandlerKey, &inst.IsActive,
		&inst.CreatedAt, &inst.UpdatedAt,
		&tmpl.ID, &tmpl.Pattern, &tmpl.Name, &tmpl.Description, &tmpl.Version,
		&tmpl.Definition, &tmpl.IsActive, &tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.NewNotFoundError("screen")
		}
		return nil, errors.NewDatabaseError("get screen by key", err)
	}

	return &repository.ScreenComposed{Instance: inst, Template: tmpl}, nil
}

// GetScreensByResourceKey returns all screens associated with a resource key.
func (r *ScreenRepository) GetScreensByResourceKey(ctx context.Context, resourceKey string) ([]repository.ScreenComposed, error) {
	query := `
		SELECT
			si.id, si.screen_key, si.template_id, si.name, si.description,
			si.slot_data, si.actions, si.data_endpoint, si.data_config,
			si.scope, si.required_permission, si.handler_key, si.is_active,
			si.created_at, si.updated_at,
			st.id, st.pattern, st.name, st.description, st.version,
			st.definition, st.is_active, st.created_at, st.updated_at
		FROM ui_config.resource_screens rs
		JOIN ui_config.screen_instances si ON si.screen_key = rs.screen_key
		JOIN ui_config.screen_templates st ON st.id = si.template_id
		WHERE rs.resource_key = $1 AND rs.is_active = true AND si.is_active = true
		ORDER BY rs.sort_order`

	rows, err := r.db.QueryContext(ctx, query, resourceKey)
	if err != nil {
		return nil, errors.NewDatabaseError("get screens by resource", err)
	}
	defer rows.Close()

	var result []repository.ScreenComposed
	for rows.Next() {
		var inst pgentities.ScreenInstance
		var tmpl pgentities.ScreenTemplate
		if err := rows.Scan(
			&inst.ID, &inst.ScreenKey, &inst.TemplateID, &inst.Name, &inst.Description,
			&inst.SlotData, &inst.Actions, &inst.DataEndpoint, &inst.DataConfig,
			&inst.Scope, &inst.RequiredPermission, &inst.HandlerKey, &inst.IsActive,
			&inst.CreatedAt, &inst.UpdatedAt,
			&tmpl.ID, &tmpl.Pattern, &tmpl.Name, &tmpl.Description, &tmpl.Version,
			&tmpl.Definition, &tmpl.IsActive, &tmpl.CreatedAt, &tmpl.UpdatedAt,
		); err != nil {
			return nil, errors.NewDatabaseError("scan screen", err)
		}
		result = append(result, repository.ScreenComposed{Instance: inst, Template: tmpl})
	}
	return result, rows.Err()
}

// GetNavigation retrieves resources and their screen mappings for a given scope.
func (r *ScreenRepository) GetNavigation(ctx context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
	resourceQuery := `
		SELECT id, key, display_name, description, icon, parent_id,
			sort_order, is_menu_visible, scope, is_active, created_at, updated_at
		FROM resources
		WHERE is_active = true AND is_menu_visible = true AND scope = $1
		ORDER BY sort_order`

	rows, err := r.db.QueryContext(ctx, resourceQuery, scope)
	if err != nil {
		return nil, nil, errors.NewDatabaseError("get navigation resources", err)
	}
	defer rows.Close()

	var resources []pgentities.Resource
	for rows.Next() {
		var res pgentities.Resource
		if err := rows.Scan(
			&res.ID, &res.Key, &res.DisplayName, &res.Description, &res.Icon, &res.ParentID,
			&res.SortOrder, &res.IsMenuVisible, &res.Scope, &res.IsActive, &res.CreatedAt, &res.UpdatedAt,
		); err != nil {
			return nil, nil, errors.NewDatabaseError("scan resource", err)
		}
		resources = append(resources, res)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, errors.NewDatabaseError("iterate resources", err)
	}

	rsQuery := `
		SELECT id, resource_id, resource_key, screen_key, screen_type,
			is_default, sort_order, is_active, created_at, updated_at
		FROM ui_config.resource_screens
		WHERE is_active = true
		ORDER BY sort_order`

	rsRows, err := r.db.QueryContext(ctx, rsQuery)
	if err != nil {
		return nil, nil, errors.NewDatabaseError("get resource screens", err)
	}
	defer rsRows.Close()

	var resourceScreens []pgentities.ResourceScreen
	for rsRows.Next() {
		var rs pgentities.ResourceScreen
		if err := rsRows.Scan(
			&rs.ID, &rs.ResourceID, &rs.ResourceKey, &rs.ScreenKey, &rs.ScreenType,
			&rs.IsDefault, &rs.SortOrder, &rs.IsActive, &rs.CreatedAt, &rs.UpdatedAt,
		); err != nil {
			return nil, nil, errors.NewDatabaseError("scan resource screen", err)
		}
		resourceScreens = append(resourceScreens, rs)
	}
	return resources, resourceScreens, rsRows.Err()
}

// UpsertPreferences saves or updates user preferences for a screen.
func (r *ScreenRepository) UpsertPreferences(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error {
	query := `
		INSERT INTO ui_config.screen_instances (screen_key, slot_data, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (screen_key)
		DO UPDATE SET slot_data = EXCLUDED.slot_data, updated_at = NOW()`

	// For a proper per-user preferences system, you'd have a separate table.
	// This is a simplified version that updates the instance slot_data.
	_, err := r.db.ExecContext(ctx, query, screenKey, preferences)
	if err != nil {
		return errors.NewDatabaseError("upsert preferences", err)
	}
	return nil
}
