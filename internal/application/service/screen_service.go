package service

import (
	"context"
	"encoding/json"

	"github.com/EduGoGroup/edugo-shared/logger"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// ScreenService handles dynamic UI screen logic.
type ScreenService struct {
	repo repository.ScreenRepository
	log  logger.Logger
}

// NewScreenService creates a new ScreenService.
func NewScreenService(repo repository.ScreenRepository, log logger.Logger) *ScreenService {
	return &ScreenService{repo: repo, log: log}
}

// GetScreen retrieves a composed screen by its key.
func (s *ScreenService) GetScreen(ctx context.Context, screenKey string) (*dto.ScreenResponse, error) {
	composed, err := s.repo.GetScreenByKey(ctx, screenKey)
	if err != nil {
		return nil, err
	}
	return toScreenResponse(composed), nil
}

// GetScreensByResourceKey returns all screens mapped to a resource key.
func (s *ScreenService) GetScreensByResourceKey(ctx context.Context, resourceKey string) ([]dto.ScreenResponse, error) {
	screens, err := s.repo.GetScreensByResourceKey(ctx, resourceKey)
	if err != nil {
		return nil, err
	}

	result := make([]dto.ScreenResponse, len(screens))
	for i := range screens {
		result[i] = *toScreenResponse(&screens[i])
	}
	return result, nil
}

// GetNavigation builds a hierarchical navigation tree from resources and their screens.
func (s *ScreenService) GetNavigation(ctx context.Context, scope string) ([]dto.NavigationNode, error) {
	resources, resourceScreens, err := s.repo.GetNavigation(ctx, scope)
	if err != nil {
		return nil, err
	}

	return buildNavigationTree(resources, resourceScreens), nil
}

// SavePreferences upserts user preferences for a screen.
func (s *ScreenService) SavePreferences(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error {
	return s.repo.UpsertPreferences(ctx, screenKey, userID, preferences)
}

func toScreenResponse(c *repository.ScreenComposed) *dto.ScreenResponse {
	return &dto.ScreenResponse{
		ScreenKey:   c.Instance.ScreenKey,
		Name:        c.Instance.Name,
		Description: c.Instance.Description,
		Pattern:     c.Template.Pattern,
		Definition:  c.Template.Definition,
		SlotData:    c.Instance.SlotData,
		Actions:     c.Instance.Actions,
		IsActive:    c.Instance.IsActive,
	}
}

func buildNavigationTree(resources []pgentities.Resource, resourceScreens []pgentities.ResourceScreen) []dto.NavigationNode {
	// Index screens by resource key
	screensByResource := make(map[string][]dto.NavigationScreen)
	for _, rs := range resourceScreens {
		screensByResource[rs.ResourceKey] = append(screensByResource[rs.ResourceKey], dto.NavigationScreen{
			ScreenKey:  rs.ScreenKey,
			ScreenType: rs.ScreenType,
			IsDefault:  rs.IsDefault,
			SortOrder:  rs.SortOrder,
		})
	}

	// Index resources by parent ID
	childrenByParent := make(map[string][]pgentities.Resource)
	var roots []pgentities.Resource

	for _, r := range resources {
		if r.ParentID == nil {
			roots = append(roots, r)
		} else {
			parentKey := r.ParentID.String()
			childrenByParent[parentKey] = append(childrenByParent[parentKey], r)
		}
	}

	return buildNodes(roots, childrenByParent, screensByResource)
}

func buildNodes(
	resources []pgentities.Resource,
	childrenByParent map[string][]pgentities.Resource,
	screensByResource map[string][]dto.NavigationScreen,
) []dto.NavigationNode {
	nodes := make([]dto.NavigationNode, 0, len(resources))
	for _, r := range resources {
		if !r.IsMenuVisible || !r.IsActive {
			continue
		}

		node := dto.NavigationNode{
			Key:         r.Key,
			DisplayName: r.DisplayName,
			Icon:        r.Icon,
			SortOrder:   r.SortOrder,
			Screens:     screensByResource[r.Key],
			Children:    buildNodes(childrenByParent[r.ID.String()], childrenByParent, screensByResource),
		}
		nodes = append(nodes, node)
	}
	return nodes
}
