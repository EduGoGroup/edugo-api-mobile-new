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
// PUT /v1/progress
func (h *ProgressHandler) Upsert(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UpsertProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.Upsert(c.Request.Context(), userID, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
