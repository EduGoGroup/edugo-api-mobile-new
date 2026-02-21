package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/logger"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// ProgressService handles reading progress business logic.
type ProgressService struct {
	repo repository.ProgressRepository
	log  logger.Logger
}

// NewProgressService creates a new ProgressService.
func NewProgressService(repo repository.ProgressRepository, log logger.Logger) *ProgressService {
	return &ProgressService{repo: repo, log: log}
}

// Upsert creates or updates a progress record. Idempotent via INSERT ON CONFLICT.
func (s *ProgressService) Upsert(ctx context.Context, userID uuid.UUID, req dto.UpsertProgressRequest) (*dto.ProgressResponse, error) {
	status := "in_progress"
	if req.Percentage >= 100 {
		status = "completed"
	} else if req.Percentage == 0 {
		status = "not_started"
	}

	now := time.Now()
	progress := &pgentities.Progress{
		MaterialID:     req.MaterialID,
		UserID:         userID,
		Percentage:     req.Percentage,
		LastPage:       req.LastPage,
		Status:         status,
		LastAccessedAt: now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.Upsert(ctx, progress); err != nil {
		return nil, err
	}

	s.log.Info("progress upserted",
		"material_id", req.MaterialID,
		"user_id", userID,
		"percentage", req.Percentage,
		"status", status,
	)

	return &dto.ProgressResponse{
		MaterialID:     progress.MaterialID,
		UserID:         progress.UserID,
		Percentage:     progress.Percentage,
		LastPage:       progress.LastPage,
		Status:         progress.Status,
		LastAccessedAt: progress.LastAccessedAt,
	}, nil
}
