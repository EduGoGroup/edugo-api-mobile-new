package service

import (
	"context"

	"github.com/EduGoGroup/edugo-shared/logger"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// SummaryService handles material summary retrieval from MongoDB.
type SummaryService struct {
	repo repository.MongoSummaryRepository
	log  logger.Logger
}

// NewSummaryService creates a new SummaryService.
func NewSummaryService(repo repository.MongoSummaryRepository, log logger.Logger) *SummaryService {
	return &SummaryService{repo: repo, log: log}
}

// GetByMaterialID retrieves the AI-generated summary for a material.
func (s *SummaryService) GetByMaterialID(ctx context.Context, materialID string) (*dto.SummaryResponse, error) {
	summary, err := s.repo.GetByMaterialID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	return &dto.SummaryResponse{
		MaterialID: summary.MaterialID,
		Summary:    summary.Summary,
		KeyPoints:  summary.KeyPoints,
		Language:   summary.Language,
		WordCount:  summary.WordCount,
		Version:    summary.Version,
		AIModel:    summary.AIModel,
		CreatedAt:  summary.CreatedAt,
	}, nil
}
