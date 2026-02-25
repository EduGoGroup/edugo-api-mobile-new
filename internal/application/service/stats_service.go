package service

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/logger"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// StatsService handles statistics retrieval.
type StatsService struct {
	repo repository.StatsRepository
	log  logger.Logger
}

// NewStatsService creates a new StatsService.
func NewStatsService(repo repository.StatsRepository, log logger.Logger) *StatsService {
	return &StatsService{repo: repo, log: log}
}

// GetGlobalStats retrieves global statistics using parallel queries.
func (s *StatsService) GetGlobalStats(ctx context.Context, schoolID *uuid.UUID) (*dto.GlobalStatsResponse, error) {
	var (
		wg             sync.WaitGroup
		mu             sync.Mutex
		totalMaterials int
		completedProg  int
		avgScore       float64
		firstErr       error
	)

	wg.Add(3)

	go func() {
		defer wg.Done()
		count, err := s.repo.CountMaterials(ctx, schoolID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
			return
		}
		totalMaterials = count
	}()

	go func() {
		defer wg.Done()
		count, err := s.repo.CountCompletedProgress(ctx, schoolID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
			return
		}
		completedProg = count
	}()

	go func() {
		defer wg.Done()
		avg, err := s.repo.AverageAttemptScore(ctx, schoolID)
		mu.Lock()
		defer mu.Unlock()
		if err != nil && firstErr == nil {
			firstErr = err
			return
		}
		avgScore = avg
	}()

	wg.Wait()

	if firstErr != nil {
		return nil, firstErr
	}

	return &dto.GlobalStatsResponse{
		TotalMaterials:      totalMaterials,
		CompletedProgress:   completedProg,
		AverageAttemptScore: avgScore,
	}, nil
}

// GetMaterialStats retrieves statistics for a specific material.
func (s *StatsService) GetMaterialStats(ctx context.Context, materialID uuid.UUID) (*dto.MaterialStatsResponse, error) {
	result, err := s.repo.MaterialStats(ctx, materialID)
	if err != nil {
		return nil, err
	}

	return &dto.MaterialStatsResponse{
		TotalAttempts:  result.TotalAttempts,
		AverageScore:   result.AverageScore,
		CompletionRate: result.CompletionRate,
		UniqueStudents: result.UniqueStudents,
	}, nil
}
