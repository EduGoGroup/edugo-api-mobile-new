package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	"github.com/EduGoGroup/edugo-shared/logger"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
)

// GuardianService handles guardian relation business logic.
type GuardianService struct {
	repo repository.GuardianRepository
	log  logger.Logger
}

// NewGuardianService creates a new GuardianService.
func NewGuardianService(repo repository.GuardianRepository, log logger.Logger) *GuardianService {
	return &GuardianService{repo: repo, log: log}
}

// RequestRelation creates a pending guardian-student relation.
func (s *GuardianService) RequestRelation(ctx context.Context, guardianID uuid.UUID, req dto.GuardianRelationRequestDTO) (*dto.GuardianRelationResponse, error) {
	// Find student by email
	student, err := s.repo.FindUserByEmail(ctx, req.Identifier)
	if err != nil {
		return nil, err
	}
	if student == nil {
		return nil, errors.NewNotFoundError("student with email " + req.Identifier)
	}

	// Check for duplicate
	existing, err := s.repo.FindByGuardianAndStudent(ctx, guardianID, student.ID)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.NewConflictError(fmt.Sprintf("relation already exists with status '%s'", existing.Status))
	}

	// Create pending relation
	relation := &pgentities.GuardianRelation{
		ID:               uuid.New(),
		GuardianID:       guardianID,
		StudentID:        student.ID,
		RelationshipType: req.RelationshipType,
		IsActive:         true,
		Status:           "pending",
		CreatedBy:        &guardianID,
	}

	if err := s.repo.Create(ctx, relation); err != nil {
		return nil, err
	}

	s.log.Info("guardian relation requested",
		"guardian_id", guardianID,
		"student_id", student.ID,
		"relationship_type", req.RelationshipType,
	)

	return &dto.GuardianRelationResponse{
		ID:               relation.ID,
		GuardianID:       relation.GuardianID,
		StudentID:        relation.StudentID,
		StudentName:      student.FirstName + " " + student.LastName,
		RelationshipType: relation.RelationshipType,
		Status:           relation.Status,
		IsPrimary:        relation.IsPrimary,
		IsActive:         relation.IsActive,
		CreatedAt:        relation.CreatedAt,
		UpdatedAt:        relation.UpdatedAt,
	}, nil
}

// ListPendingRequests returns pending guardian relations for a school.
func (s *GuardianService) ListPendingRequests(ctx context.Context, schoolID uuid.UUID) ([]dto.GuardianRelationResponse, error) {
	joined, err := s.repo.ListPendingBySchool(ctx, schoolID)
	if err != nil {
		return nil, err
	}

	results := make([]dto.GuardianRelationResponse, len(joined))
	for i, j := range joined {
		results[i] = dto.GuardianRelationResponse{
			ID:               j.ID,
			GuardianID:       j.GuardianID,
			GuardianName:     j.GuardianName,
			StudentID:        j.StudentID,
			StudentName:      j.StudentName,
			RelationshipType: j.RelationshipType,
			Status:           j.Status,
			IsPrimary:        j.IsPrimary,
			IsActive:         j.IsActive,
			CreatedAt:        j.CreatedAt,
			UpdatedAt:        j.UpdatedAt,
		}
	}
	return results, nil
}

// ApproveRequest approves a pending guardian relation scoped to a school.
func (s *GuardianService) ApproveRequest(ctx context.Context, schoolID, id uuid.UUID) error {
	rel, err := s.repo.GetByIDAndSchool(ctx, id, schoolID)
	if err != nil {
		return err
	}
	if rel.Status != "pending" {
		return errors.NewValidationError(fmt.Sprintf("relation is '%s', not pending", rel.Status))
	}

	if err := s.repo.UpdateStatus(ctx, id, "active"); err != nil {
		return err
	}

	s.log.Info("guardian relation approved", "id", id, "school_id", schoolID)
	return nil
}

// RejectRequest rejects a pending guardian relation scoped to a school.
func (s *GuardianService) RejectRequest(ctx context.Context, schoolID, id uuid.UUID) error {
	rel, err := s.repo.GetByIDAndSchool(ctx, id, schoolID)
	if err != nil {
		return err
	}
	if rel.Status != "pending" {
		return errors.NewValidationError(fmt.Sprintf("relation is '%s', not pending", rel.Status))
	}

	if err := s.repo.UpdateStatus(ctx, id, "rejected"); err != nil {
		return err
	}

	s.log.Info("guardian relation rejected", "id", id, "school_id", schoolID)
	return nil
}

// ListMyChildren returns active children for a guardian.
func (s *GuardianService) ListMyChildren(ctx context.Context, guardianID uuid.UUID) ([]dto.ChildResponse, error) {
	children, err := s.repo.ListActiveChildrenByGuardian(ctx, guardianID)
	if err != nil {
		return nil, err
	}

	results := make([]dto.ChildResponse, len(children))
	for i, c := range children {
		results[i] = dto.ChildResponse{
			ID:        c.ID,
			FirstName: c.FirstName,
			LastName:  c.LastName,
			Email:     c.Email,
		}
	}
	return results, nil
}

// GetChildProgress returns progress for a specific child of a guardian.
func (s *GuardianService) GetChildProgress(ctx context.Context, guardianID, childID uuid.UUID) (*dto.ChildProgressResponse, error) {
	// Verify active relation
	rel, err := s.repo.FindByGuardianAndStudent(ctx, guardianID, childID)
	if err != nil {
		return nil, err
	}
	if rel == nil || rel.Status != "active" {
		return nil, errors.NewForbiddenError("no active relation with this child")
	}

	progress, err := s.repo.GetChildProgress(ctx, childID)
	if err != nil {
		return nil, err
	}

	// Get child name
	children, err := s.repo.ListActiveChildrenByGuardian(ctx, guardianID)
	if err != nil {
		return nil, err
	}
	var childName string
	for _, c := range children {
		if c.ID == childID {
			childName = c.FirstName + " " + c.LastName
			break
		}
	}

	return &dto.ChildProgressResponse{
		ChildID:        childID,
		ChildName:      childName,
		TotalMaterials: progress.TotalMaterials,
		Completed:      progress.Completed,
		AvgScore:       progress.AvgScore,
		CompletionRate: progress.CompletionRate,
	}, nil
}

// GetMyStats returns aggregate stats for all of a guardian's children.
func (s *GuardianService) GetMyStats(ctx context.Context, guardianID uuid.UUID) (*dto.GuardianStatsResponse, error) {
	children, err := s.repo.ListActiveChildrenByGuardian(ctx, guardianID)
	if err != nil {
		return nil, err
	}

	stats := &dto.GuardianStatsResponse{
		ChildrenCount: len(children),
	}

	if len(children) == 0 {
		return stats, nil
	}

	var totalMaterials int
	var totalScore, totalCompletion float64
	var childrenWithProgress int

	for _, child := range children {
		progress, err := s.repo.GetChildProgress(ctx, child.ID)
		if err != nil {
			s.log.Warn("failed to get progress for child", "child_id", child.ID, "error", err)
			continue
		}
		totalMaterials += progress.TotalMaterials
		totalScore += progress.AvgScore
		totalCompletion += progress.CompletionRate
		childrenWithProgress++
	}

	stats.TotalMaterials = totalMaterials
	if childrenWithProgress > 0 {
		stats.AvgScore = totalScore / float64(childrenWithProgress)
		stats.CompletionRate = totalCompletion / float64(childrenWithProgress)
	}

	return stats, nil
}
