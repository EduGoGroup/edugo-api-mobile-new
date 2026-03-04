package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// GuardianHandler handles guardian HTTP endpoints.
type GuardianHandler struct {
	svc *service.GuardianService
}

// NewGuardianHandler creates a new GuardianHandler.
func NewGuardianHandler(svc *service.GuardianService) *GuardianHandler {
	return &GuardianHandler{svc: svc}
}

// RequestRelation godoc
// @Summary Request a guardian-student link
// @Tags guardian-relations
// @Accept json
// @Produce json
// @Param request body dto.GuardianRelationRequestDTO true "Relation request"
// @Success 201 {object} dto.GuardianRelationResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardian-relations/request [post]
func (h *GuardianHandler) RequestRelation(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.GuardianRelationRequestDTO
	if err := bindJSON(c, &req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.RequestRelation(c.Request.Context(), userID, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// ListPendingRequests godoc
// @Summary List pending guardian relation requests for a school
// @Tags guardian-relations
// @Produce json
// @Success 200 {array} dto.GuardianRelationResponse
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardian-relations/pending [get]
func (h *GuardianHandler) ListPendingRequests(c *gin.Context) {
	schoolID, err := getSchoolID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	results, err := h.svc.ListPendingRequests(c.Request.Context(), schoolID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, results)
}

// ApproveRequest godoc
// @Summary Approve a pending guardian relation request
// @Tags guardian-relations
// @Produce json
// @Param id path string true "Relation ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardian-relations/{id}/approve [post]
func (h *GuardianHandler) ApproveRequest(c *gin.Context) {
	schoolID, err := getSchoolID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	if err := h.svc.ApproveRequest(c.Request.Context(), schoolID, id); err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "relation approved"})
}

// RejectRequest godoc
// @Summary Reject a pending guardian relation request
// @Tags guardian-relations
// @Produce json
// @Param id path string true "Relation ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardian-relations/{id}/reject [post]
func (h *GuardianHandler) RejectRequest(c *gin.Context) {
	schoolID, err := getSchoolID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	if err := h.svc.RejectRequest(c.Request.Context(), schoolID, id); err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "relation rejected"})
}

// ListMyChildren godoc
// @Summary List a guardian's active children
// @Tags guardians
// @Produce json
// @Success 200 {array} dto.ChildResponse
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardians/me/children [get]
func (h *GuardianHandler) ListMyChildren(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	results, err := h.svc.ListMyChildren(c.Request.Context(), userID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, results)
}

// GetChildProgress godoc
// @Summary Get academic progress for a guardian's child
// @Tags guardians
// @Produce json
// @Param childId path string true "Child User ID (UUID)"
// @Success 200 {object} dto.ChildProgressResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 403 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardians/me/children/{childId}/progress [get]
func (h *GuardianHandler) GetChildProgress(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	childID, err := parseUUIDParam(c, "childId")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetChildProgress(c.Request.Context(), userID, childID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetMyStats godoc
// @Summary Get aggregate stats for all children of a guardian
// @Tags guardians
// @Produce json
// @Success 200 {object} dto.GuardianStatsResponse
// @Failure 401 {object} map[string]interface{}
// @Security BearerAuth
// @Router /guardians/me/stats [get]
func (h *GuardianHandler) GetMyStats(c *gin.Context) {
	userID, err := getUserID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetMyStats(c.Request.Context(), userID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}
