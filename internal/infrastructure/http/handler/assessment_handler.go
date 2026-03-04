package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
	sharedrepo "github.com/EduGoGroup/edugo-shared/repository"
)

// AssessmentHandler handles assessment and attempt HTTP endpoints.
type AssessmentHandler struct {
	svc *service.AssessmentService
}

// NewAssessmentHandler creates a new AssessmentHandler.
func NewAssessmentHandler(svc *service.AssessmentService) *AssessmentHandler {
	return &AssessmentHandler{svc: svc}
}

// GetAssessment godoc
// @Summary Get assessment for a material
// @Tags assessments
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id}/assessment [get]
func (h *AssessmentHandler) GetAssessment(c *gin.Context) {
	materialID, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetAssessmentByMaterialID(c.Request.Context(), materialID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// CreateAttempt godoc
// @Summary Submit an assessment attempt
// @Tags assessments
// @Accept json
// @Produce json
// @Param id path string true "Material ID (UUID)"
// @Param request body dto.CreateAttemptRequest true "Attempt data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /materials/{id}/assessment/attempts [post]
func (h *AssessmentHandler) CreateAttempt(c *gin.Context) {
	materialID, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.CreateAttemptRequest
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.CreateAttempt(c.Request.Context(), materialID, userID, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// GetAttemptResult godoc
// @Summary Get result of an assessment attempt
// @Tags assessments
// @Accept json
// @Produce json
// @Param id path string true "Attempt ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /attempts/{id}/results [get]
func (h *AssessmentHandler) GetAttemptResult(c *gin.Context) {
	attemptID, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetAttemptResult(c.Request.Context(), attemptID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// ListUserAttempts godoc
// @Summary List all attempts for the current user
// @Tags assessments
// @Accept json
// @Produce json
// @Param limit query int false "Limit results"
// @Param page query int false "Page number (1-based)"
// @Param search query string false "Search term (ILIKE)"
// @Param search_fields query string false "Comma-separated fields to search"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /users/me/attempts [get]
func (h *AssessmentHandler) ListUserAttempts(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	page := 1
	limit := 20
	if v, exists := c.GetQuery("page"); exists {
		if parsed, err := parseInt(v); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if v, exists := c.GetQuery("limit"); exists {
		if parsed, err := parseInt(v); err == nil {
			limit = parsed
		}
	}
	var filters sharedrepo.ListFilters
	if search := c.Query("search"); search != "" {
		filters.Search = search
		if fields := c.Query("search_fields"); fields != "" {
			filters.SearchFields = strings.Split(fields, ",")
		}
	}

	result, err := h.svc.ListAttemptsByUser(c.Request.Context(), userID, page, limit, filters)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func parseInt(s string) (int, error) {
	var n int
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, &invalidIntError{s}
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

type invalidIntError struct {
	s string
}

func (e *invalidIntError) Error() string {
	return "invalid integer: " + e.s
}
