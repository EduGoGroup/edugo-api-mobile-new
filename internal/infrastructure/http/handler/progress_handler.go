package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// ProgressHandler handles progress HTTP endpoints.
type ProgressHandler struct {
	svc *service.ProgressService
}

// NewProgressHandler creates a new ProgressHandler.
func NewProgressHandler(svc *service.ProgressService) *ProgressHandler {
	return &ProgressHandler{svc: svc}
}

// Upsert godoc
// @Summary Update or insert progress for a material
// @Tags progress
// @Accept json
// @Produce json
// @Param request body dto.UpsertProgressRequest true "Progress data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /progress [put]
func (h *ProgressHandler) Upsert(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UpsertProgressRequest
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.Upsert(c.Request.Context(), userID, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
