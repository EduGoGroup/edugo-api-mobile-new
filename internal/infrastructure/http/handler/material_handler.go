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
// @Summary List materials
// @Tags materials
// @Accept json
// @Produce json
// @Param subject_id query string false "Filter by subject ID"
// @Param unit_id query string false "Filter by unit ID"
// @Param page query int false "Page number"
// @Param limit query int false "Items per page"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials [get]
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
// @Summary Get material by ID
// @Tags materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id} [get]
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
// @Summary Get material with all its versions
// @Tags materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id}/versions [get]
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
// @Summary Create a new material
// @Tags materials
// @Accept json
// @Produce json
// @Param request body dto.CreateMaterialRequest true "Material data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials [post]
func (h *MaterialHandler) Create(c *gin.Context) {
	var req dto.CreateMaterialRequest
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
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
// @Summary Update a material
// @Tags materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Param request body dto.UpdateMaterialRequest true "Material update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id} [put]
func (h *MaterialHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UpdateMaterialRequest
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
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
// @Summary Generate presigned URL for file upload
// @Tags materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id}/upload-url [post]
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
// @Summary Notify that file upload is complete
// @Tags materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Param request body dto.UploadCompleteRequest true "Upload completion data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id}/upload-complete [post]
func (h *MaterialHandler) UploadComplete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UploadCompleteRequest
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
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
// @Summary Generate presigned URL for file download
// @Tags materials
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id}/download-url [get]
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
