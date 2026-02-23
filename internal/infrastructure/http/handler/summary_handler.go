package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// SummaryHandler handles material summary HTTP endpoints.
type SummaryHandler struct {
	svc *service.SummaryService
}

// NewSummaryHandler creates a new SummaryHandler.
func NewSummaryHandler(svc *service.SummaryService) *SummaryHandler {
	return &SummaryHandler{svc: svc}
}

// GetSummary godoc
// @Summary Get AI-generated summary for a material
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
// @Router /materials/{id}/summary [get]
func (h *SummaryHandler) GetSummary(c *gin.Context) {
	materialID, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetByMaterialID(c.Request.Context(), materialID.String())
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
