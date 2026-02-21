package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// ScreenHandler handles dynamic UI screen endpoints.
type ScreenHandler struct {
	svc *service.ScreenService
}

// NewScreenHandler creates a new ScreenHandler.
func NewScreenHandler(svc *service.ScreenService) *ScreenHandler {
	return &ScreenHandler{svc: svc}
}

// GetScreen godoc
// GET /v1/screens/:screenKey
func (h *ScreenHandler) GetScreen(c *gin.Context) {
	screenKey := c.Param("screenKey")
	if screenKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screenKey is required", "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.GetScreen(c.Request.Context(), screenKey)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetScreensByResource godoc
// GET /v1/screens/resource/:resourceKey
func (h *ScreenHandler) GetScreensByResource(c *gin.Context) {
	resourceKey := c.Param("resourceKey")
	if resourceKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "resourceKey is required", "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.GetScreensByResourceKey(c.Request.Context(), resourceKey)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetNavigation godoc
// GET /v1/screens/navigation
func (h *ScreenHandler) GetNavigation(c *gin.Context) {
	scope := c.DefaultQuery("scope", "system")

	result, err := h.svc.GetNavigation(c.Request.Context(), scope)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// SavePreferences godoc
// PUT /v1/screens/:screenKey/preferences
func (h *ScreenHandler) SavePreferences(c *gin.Context) {
	screenKey := c.Param("screenKey")
	if screenKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "screenKey is required", "code": "VALIDATION_ERROR"})
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.SavePreferencesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	if err := h.svc.SavePreferences(c.Request.Context(), screenKey, userID.String(), req.Preferences); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "preferences saved"})
}
