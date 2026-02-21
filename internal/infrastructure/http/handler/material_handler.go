package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// MaterialHandler handles material HTTP endpoints.
type MaterialHandler struct {
	svc *service.MaterialService
}

// NewMaterialHandler creates a new MaterialHandler.
func NewMaterialHandler(svc *service.MaterialService) *MaterialHandler {
	return &MaterialHandler{svc: svc}
}

// List godoc
// GET /v1/materials
func (h *MaterialHandler) List(c *gin.Context) {
	var req dto.ListMaterialsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.List(c.Request.Context(), req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// GET /v1/materials/:id
func (h *MaterialHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetWithVersions godoc
// GET /v1/materials/:id/versions
func (h *MaterialHandler) GetWithVersions(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetWithVersions(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Create godoc
// POST /v1/materials
func (h *MaterialHandler) Create(c *gin.Context) {
	var req dto.CreateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	schoolID, err := getSchoolID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.Create(c.Request.Context(), req, schoolID, userID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Update godoc
// PUT /v1/materials/:id
func (h *MaterialHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UpdateMaterialRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.Update(c.Request.Context(), id, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GenerateUploadURL godoc
// POST /v1/materials/:id/upload-url
func (h *MaterialHandler) GenerateUploadURL(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GenerateUploadURL(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// UploadComplete godoc
// POST /v1/materials/:id/upload-complete
func (h *MaterialHandler) UploadComplete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UploadCompleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.NotifyUploadComplete(c.Request.Context(), id, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GenerateDownloadURL godoc
// GET /v1/materials/:id/download-url
func (h *MaterialHandler) GenerateDownloadURL(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GenerateDownloadURL(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
