package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// StatsHandler handles statistics HTTP endpoints.
type StatsHandler struct {
	svc *service.StatsService
}

// NewStatsHandler creates a new StatsHandler.
func NewStatsHandler(svc *service.StatsService) *StatsHandler {
	return &StatsHandler{svc: svc}
}

// GetGlobalStats godoc
// GET /v1/stats/global
func (h *StatsHandler) GetGlobalStats(c *gin.Context) {
	var schoolID *uuid.UUID
	if raw := c.Query("school_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "school_id must be a valid UUID", "code": "VALIDATION_ERROR"})
			return
		}
		schoolID = &id
	}

	result, err := h.svc.GetGlobalStats(c.Request.Context(), schoolID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetMaterialStats godoc
// GET /v1/materials/:id/stats
func (h *StatsHandler) GetMaterialStats(c *gin.Context) {
	materialID, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetMaterialStats(c.Request.Context(), materialID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
