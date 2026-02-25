package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/EduGoGroup/edugo-shared/logger"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/client"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"

	rediscache "github.com/EduGoGroup/edugo-shared/cache/redis"
)

const (
	screenCacheTTL     = 15 * time.Minute
	navigationCacheTTL = 10 * time.Minute
)

// ScreenService handles dynamic UI screen logic with Redis caching.
type ScreenService struct {
	repo      repository.ScreenRepository
	iamClient *client.IAMPlatformClient
	cache     rediscache.CacheService
	log       logger.Logger
}

// NewScreenService creates a new ScreenService.
func NewScreenService(repo repository.ScreenRepository, iamClient *client.IAMPlatformClient, cache rediscache.CacheService, log logger.Logger) *ScreenService {
	return &ScreenService{repo: repo, iamClient: iamClient, cache: cache, log: log}
}

// GetScreen retrieves a composed screen by its key, with Redis caching.
func (s *ScreenService) GetScreen(ctx context.Context, screenKey string) (*dto.ScreenResponse, error) {
	// Try cache first
	if s.cache != nil {
		cacheKey := fmt.Sprintf("screen:%s", screenKey)
		var cached dto.ScreenResponse
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			return &cached, nil
		}
	}

	// Try iam-platform via HTTP if available
	if s.iamClient != nil {
		combined, err := s.iamClient.ResolveScreenByKey(ctx, screenKey, "")
		if err == nil && combined != nil {
			resp := &dto.ScreenResponse{
				ScreenKey:  combined.ScreenKey,
				Name:       combined.ScreenName,
				Pattern:    combined.Pattern,
				Definition: combined.Template,
				SlotData:   combined.SlotData,
				IsActive:   true,
			}
			if s.cache != nil {
				cacheKey := fmt.Sprintf("screen:%s", screenKey)
				_ = s.cache.Set(ctx, cacheKey, resp, screenCacheTTL)
			}
			return resp, nil
		}
		s.log.Warn("iam-platform screen fetch failed, falling back to local", "key", screenKey, "error", err)
	}

	// Fallback to local DB
	composed, err := s.repo.GetScreenByKey(ctx, screenKey)
	if err != nil {
		return nil, err
	}
	resp := toScreenResponse(composed)

	if s.cache != nil {
		cacheKey := fmt.Sprintf("screen:%s", screenKey)
		_ = s.cache.Set(ctx, cacheKey, resp, screenCacheTTL)
	}

	return resp, nil
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

// GetNavigation builds a hierarchical navigation tree.
func (s *ScreenService) GetNavigation(ctx context.Context, scope string) ([]dto.NavigationNode, error) {
	// Try cache first
	if s.cache != nil {
		cacheKey := fmt.Sprintf("navigation:%s", scope)
		var cached []dto.NavigationNode
		if err := s.cache.Get(ctx, cacheKey, &cached); err == nil {
			return cached, nil
		}
	}

	// Fallback to local DB
	resources, resourceScreens, err := s.repo.GetNavigation(ctx, scope)
	if err != nil {
		return nil, err
	}

	result := buildNavigationTree(resources, resourceScreens)

	if s.cache != nil {
		cacheKey := fmt.Sprintf("navigation:%s", scope)
		_ = s.cache.Set(ctx, cacheKey, result, navigationCacheTTL)
	}

	return result, nil
}

// SavePreferences upserts user preferences for a screen and invalidates cache.
func (s *ScreenService) SavePreferences(ctx context.Context, screenKey, userID string, preferences json.RawMessage) error {
	if err := s.repo.UpsertPreferences(ctx, screenKey, userID, preferences); err != nil {
		return err
	}
	// Invalidate cache
	if s.cache != nil {
		_ = s.cache.Delete(ctx, fmt.Sprintf("screen:%s", screenKey))
	}
	return nil
}

func toScreenResponse(c *repository.ScreenComposed) *dto.ScreenResponse {
	return &dto.ScreenResponse{
		ScreenKey:   c.Instance.ScreenKey,
		Name:        c.Instance.Name,
		Description: c.Instance.Description,
		Pattern:     c.Template.Pattern,
		Definition:  c.Template.Definition,
		SlotData:    c.Instance.SlotData,
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
