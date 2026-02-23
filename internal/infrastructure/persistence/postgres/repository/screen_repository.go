package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	sharedErrors "github.com/EduGoGroup/edugo-shared/common/errors"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"
	"gorm.io/gorm"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// ScreenRepository implements repository.ScreenRepository using PostgreSQL.
type ScreenRepository struct {
	db *gorm.DB
}

// NewScreenRepository creates a new ScreenRepository.
func NewScreenRepository(db *gorm.DB) *ScreenRepository {
	return &ScreenRepository{db: db}
}

// GetScreenByKey retrieves a composed screen (template + instance) by screen_key.
func (r *ScreenRepository) GetScreenByKey(ctx context.Context, screenKey string) (*repository.ScreenComposed, error) {
	var inst pgentities.ScreenInstance
	var tmpl pgentities.ScreenTemplate

	row := r.db.WithContext(ctx).Raw(queryScreenByKey, screenKey).Row()
	err := row.Scan(
		&inst.ID, &inst.ScreenKey, &inst.TemplateID, &inst.Name, &inst.Description,
		&inst.SlotData, &inst.Actions, &inst.DataEndpoint, &inst.DataConfig,
		&inst.Scope, &inst.RequiredPermission, &inst.HandlerKey, &inst.IsActive,
		&inst.CreatedAt, &inst.UpdatedAt,
		&tmpl.ID, &tmpl.Pattern, &tmpl.Name, &tmpl.Description, &tmpl.Version,
		&tmpl.Definition, &tmpl.IsActive, &tmpl.CreatedAt, &tmpl.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, sharedErrors.NewNotFoundError("screen")
		}
		return nil, sharedErrors.NewDatabaseError("get screen by key", err)
	}

	return &repository.ScreenComposed{Instance: inst, Template: tmpl}, nil
}

// GetScreensByResourceKey returns all screens associated with a resource key.
func (r *ScreenRepository) GetScreensByResourceKey(ctx context.Context, resourceKey string) ([]repository.ScreenComposed, error) {
	rows, err := r.db.WithContext(ctx).Raw(queryScreensByResourceKey, resourceKey).Rows()
	if err != nil {
		return nil, sharedErrors.NewDatabaseError("get screens by resource", err)
	}
	defer func() { _ = rows.Close() }()

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
			return nil, sharedErrors.NewDatabaseError("scan screen", err)
		}
		result = append(result, repository.ScreenComposed{Instance: inst, Template: tmpl})
	}
	return result, rows.Err()
}

// GetNavigation retrieves resources and their screen mappings for a given scope.
func (r *ScreenRepository) GetNavigation(ctx context.Context, scope string) ([]pgentities.Resource, []pgentities.ResourceScreen, error) {
	var resources []pgentities.Resource
	if err := r.db.WithContext(ctx).Where("is_active = true AND is_menu_visible = true AND scope = ?", scope).Order("sort_order").Find(&resources).Error; err != nil {
		return nil, nil, sharedErrors.NewDatabaseError("get navigation resources", err)
	}

	var resourceScreens []pgentities.ResourceScreen
	if err := r.db.WithContext(ctx).Table("ui_config.resource_screens").Where("is_active = true").Order("sort_order").Find(&resourceScreens).Error; err != nil {
		return nil, nil, sharedErrors.NewDatabaseError("get resource screens", err)
	}

	return resources, resourceScreens, nil
}

// UpsertPreferences saves or updates user preferences for a screen.
func (r *ScreenRepository) UpsertPreferences(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error {
	if err := r.db.WithContext(ctx).Exec(queryUpsertPreferences, screenKey, preferences).Error; err != nil {
		return sharedErrors.NewDatabaseError("upsert preferences", err)
	}
	return nil
}
