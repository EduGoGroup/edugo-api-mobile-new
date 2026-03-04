package service

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	"github.com/EduGoGroup/edugo-shared/logger"
	"github.com/EduGoGroup/edugo-shared/messaging/events"
	"github.com/EduGoGroup/edugo-shared/messaging/rabbit"

	pgentities "github.com/EduGoGroup/edugo-infrastructure/postgres/entities"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/domain/repository"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/storage"
)

// MaterialService handles material business logic.
type MaterialService struct {
	repo      repository.MaterialRepository
	s3        *storage.S3Client
	publisher rabbit.Publisher
	log       logger.Logger
	exchange  string
}

// NewMaterialService creates a new MaterialService.
func NewMaterialService(
	repo repository.MaterialRepository,
	s3 *storage.S3Client,
	publisher rabbit.Publisher,
	log logger.Logger,
	exchange string,
) *MaterialService {
	return &MaterialService{
		repo:      repo,
		s3:        s3,
		publisher: publisher,
		log:       log,
		exchange:  exchange,
	}
}

// Create creates a new material in draft status.
func (s *MaterialService) Create(ctx context.Context, req dto.CreateMaterialRequest, schoolID, teacherID uuid.UUID) (*dto.MaterialResponse, error) {
	material := &pgentities.Material{
		ID:                  uuid.New(),
		SchoolID:            schoolID,
		UploadedByTeacherID: teacherID,
		AcademicUnitID:      req.AcademicUnitID,
		Title:               req.Title,
		Description:         req.Description,
		Subject:             req.Subject,
		Grade:               req.Grade,
		Status:              "draft",
		IsPublic:            req.IsPublic,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.repo.Create(ctx, material); err != nil {
		s.log.Error("failed to create material", "error", err, "material_id", material.ID)
		return nil, err
	}

	s.log.Info("material created", "material_id", material.ID)
	return toMaterialResponse(material), nil
}

// GetByID retrieves a material by its ID.
func (s *MaterialService) GetByID(ctx context.Context, id uuid.UUID) (*dto.MaterialResponse, error) {
	material, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return toMaterialResponse(material), nil
}

// List retrieves materials with optional filters.
func (s *MaterialService) List(ctx context.Context, req dto.ListMaterialsRequest) (*dto.PaginatedResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}
	if req.Page <= 0 {
		req.Page = 1
	}

	offset := (req.Page - 1) * req.Limit

	filter := repository.MaterialFilter{
		SchoolID: req.SchoolID,
		AuthorID: req.AuthorID,
		Status:   req.Status,
		Limit:    req.Limit,
		Offset:   offset,
		Search:   req.Search,
	}
	if req.SearchFields != "" {
		filter.SearchFields = strings.Split(req.SearchFields, ",")
	}

	materials, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]dto.MaterialResponse, len(materials))
	for i := range materials {
		items[i] = *toMaterialResponse(&materials[i])
	}

	return &dto.PaginatedResponse{
		Data:  items,
		Total: total,
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// Update modifies a material.
func (s *MaterialService) Update(ctx context.Context, id uuid.UUID, req dto.UpdateMaterialRequest) (*dto.MaterialResponse, error) {
	material, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Title != nil {
		material.Title = *req.Title
	}
	if req.Description != nil {
		material.Description = req.Description
	}
	if req.Subject != nil {
		material.Subject = req.Subject
	}
	if req.Grade != nil {
		material.Grade = req.Grade
	}
	if req.Status != nil {
		material.Status = *req.Status
	}
	if req.IsPublic != nil {
		material.IsPublic = *req.IsPublic
	}
	material.UpdatedAt = time.Now()

	if err := s.repo.Update(ctx, material); err != nil {
		return nil, err
	}

	return toMaterialResponse(material), nil
}

// GetWithVersions returns a material with its version history.
func (s *MaterialService) GetWithVersions(ctx context.Context, id uuid.UUID) (*dto.MaterialWithVersionsResponse, error) {
	material, versions, err := s.repo.GetWithVersions(ctx, id)
	if err != nil {
		return nil, err
	}

	versionResponses := make([]dto.MaterialVersionResponse, len(versions))
	for i, v := range versions {
		versionResponses[i] = dto.MaterialVersionResponse{
			ID:            v.ID,
			MaterialID:    v.MaterialID,
			VersionNumber: v.VersionNumber,
			Title:         v.Title,
			ContentURL:    v.ContentURL,
			ChangedBy:     v.ChangedBy,
			CreatedAt:     v.CreatedAt,
		}
	}

	return &dto.MaterialWithVersionsResponse{
		Material: *toMaterialResponse(material),
		Versions: versionResponses,
	}, nil
}

// NotifyUploadComplete updates the material with file info and publishes an event.
func (s *MaterialService) NotifyUploadComplete(ctx context.Context, id uuid.UUID, req dto.UploadCompleteRequest) (*dto.MaterialResponse, error) {
	material, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	material.FileURL = req.FileURL
	material.FileType = req.FileType
	material.FileSizeBytes = req.FileSizeBytes
	material.Status = "uploaded"
	now := time.Now()
	material.UpdatedAt = now

	if err := s.repo.Update(ctx, material); err != nil {
		return nil, err
	}

	// Publish material.uploaded event
	if s.publisher != nil {
		payload := events.MaterialUploadedPayload{
			MaterialID:    material.ID.String(),
			SchoolID:      material.SchoolID.String(),
			TeacherID:     material.UploadedByTeacherID.String(),
			FileURL:       material.FileURL,
			FileSizeBytes: uint64(material.FileSizeBytes),
			FileType:      material.FileType,
		}

		event, eventErr := events.NewMaterialUploadedEvent(
			uuid.New().String(),
			"material.uploaded",
			"1.0",
			payload,
		)
		if eventErr != nil {
			s.log.Error("failed to create material uploaded event", "error", eventErr)
		} else {
			if pubErr := s.publisher.Publish(ctx, s.exchange, "material.uploaded", event); pubErr != nil {
				s.log.Error("failed to publish material uploaded event", "error", pubErr, "material_id", id)
			} else {
				s.log.Info("material uploaded event published", "material_id", id)
			}
		}
	}

	return toMaterialResponse(material), nil
}

// GenerateUploadURL generates a presigned S3 upload URL.
func (s *MaterialService) GenerateUploadURL(ctx context.Context, materialID uuid.UUID) (*dto.PresignedURLResponse, error) {
	if s.s3 == nil {
		return nil, errors.NewInternalError("S3 client not configured", nil)
	}

	key := "materials/" + materialID.String() + "/original"
	url, expiresAt, err := s.s3.GenerateUploadURL(ctx, key)
	if err != nil {
		s.log.Error("failed to generate upload URL", "error", err, "material_id", materialID)
		return nil, errors.NewInternalError("failed to generate upload URL", err)
	}

	return &dto.PresignedURLResponse{URL: url, ExpiresAt: expiresAt}, nil
}

// GenerateDownloadURL generates a presigned S3 download URL.
func (s *MaterialService) GenerateDownloadURL(ctx context.Context, materialID uuid.UUID) (*dto.PresignedURLResponse, error) {
	material, err := s.repo.GetByID(ctx, materialID)
	if err != nil {
		return nil, err
	}

	if s.s3 == nil {
		return nil, errors.NewInternalError("S3 client not configured", nil)
	}

	if material.FileURL == "" {
		return nil, errors.NewBusinessRuleError("material has no file uploaded")
	}

	key := "materials/" + materialID.String() + "/original"
	url, expiresAt, err := s.s3.GenerateDownloadURL(ctx, key)
	if err != nil {
		s.log.Error("failed to generate download URL", "error", err, "material_id", materialID)
		return nil, errors.NewInternalError("failed to generate download URL", err)
	}

	return &dto.PresignedURLResponse{URL: url, ExpiresAt: expiresAt}, nil
}

func toMaterialResponse(m *pgentities.Material) *dto.MaterialResponse {
	return &dto.MaterialResponse{
		ID:                  m.ID,
		SchoolID:            m.SchoolID,
		UploadedByTeacherID: m.UploadedByTeacherID,
		AcademicUnitID:      m.AcademicUnitID,
		Title:               m.Title,
		Description:         m.Description,
		Subject:             m.Subject,
		Grade:               m.Grade,
		FileURL:             m.FileURL,
		FileType:            m.FileType,
		FileSizeBytes:       m.FileSizeBytes,
		Status:              m.Status,
		IsPublic:            m.IsPublic,
		CreatedAt:           m.CreatedAt,
		UpdatedAt:           m.UpdatedAt,
	}
}
