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
// @Summary Get a screen by key
// @Tags screens
// @Accept json
// @Produce json
// @Param screenKey path string true "Screen key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /screens/{screenKey} [get]
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
// @Summary Get screens by resource key
// @Tags screens
// @Accept json
// @Produce json
// @Param resourceKey path string true "Resource key"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /screens/resource/{resourceKey} [get]
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
// @Summary Get navigation screens
// @Tags screens
// @Accept json
// @Produce json
// @Param scope query string false "Navigation scope (default: system)"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /screens/navigation [get]
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
// @Summary Save user preferences for a screen
// @Tags screens
// @Accept json
// @Produce json
// @Param screenKey path string true "Screen key"
// @Param request body dto.SavePreferencesRequest true "Preferences data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /screens/{screenKey}/preferences [put]
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
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	if err := h.svc.SavePreferences(c.Request.Context(), screenKey, userID.String(), req.Preferences); err != nil {
		middleware.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "preferences saved"})
}
