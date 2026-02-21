package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
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
// GET /v1/materials/:id/assessment
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
// POST /v1/materials/:id/assessment/attempts
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
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
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
// GET /v1/attempts/:id/results
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
// GET /v1/users/me/attempts
func (h *AssessmentHandler) ListUserAttempts(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	limit := 20
	offset := 0
	if v, exists := c.GetQuery("limit"); exists {
		if parsed, err := parseInt(v); err == nil {
			limit = parsed
		}
	}
	if v, exists := c.GetQuery("offset"); exists {
		if parsed, err := parseInt(v); err == nil {
			offset = parsed
		}
	}

	result, err := h.svc.ListAttemptsByUser(c.Request.Context(), userID, limit, offset)
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
