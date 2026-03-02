package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/dto"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/application/service"
	"github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// AssessmentManagementHandler handles assessment CRUD and question management HTTP endpoints.
type AssessmentManagementHandler struct {
	svc *service.AssessmentManagementService
}

// NewAssessmentManagementHandler creates a new AssessmentManagementHandler.
func NewAssessmentManagementHandler(svc *service.AssessmentManagementService) *AssessmentManagementHandler {
	return &AssessmentManagementHandler{svc: svc}
}

// List godoc
// @Summary List assessments for the teacher's school
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param status query string false "Filter by status"
// @Param limit query int false "Limit results"
// @Param offset query int false "Offset for pagination"
// @Param search query string false "Search term"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments [get]
func (h *AssessmentManagementHandler) List(c *gin.Context) {
	schoolID, err := getSchoolID(c)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.ListAssessmentsRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.ListAssessments(c.Request.Context(), schoolID, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// GetByID godoc
// @Summary Get assessment detail with questions
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id} [get]
func (h *AssessmentManagementHandler) GetByID(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetAssessmentDetail(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Create godoc
// @Summary Create a manual assessment
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param request body dto.CreateAssessmentRequest true "Assessment data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments [post]
func (h *AssessmentManagementHandler) Create(c *gin.Context) {
	var req dto.CreateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
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

	result, err := h.svc.CreateAssessment(c.Request.Context(), req, schoolID, userID)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, result)
}

// Update godoc
// @Summary Update an assessment (draft only)
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Param request body dto.UpdateAssessmentRequest true "Assessment update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id} [put]
func (h *AssessmentManagementHandler) Update(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.UpdateAssessmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.UpdateAssessment(c.Request.Context(), id, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Publish godoc
// @Summary Publish an assessment
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id}/publish [patch]
func (h *AssessmentManagementHandler) Publish(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.PublishAssessment(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Archive godoc
// @Summary Archive an assessment
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id}/archive [patch]
func (h *AssessmentManagementHandler) Archive(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.ArchiveAssessment(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, result)
}

// Delete godoc
// @Summary Delete an assessment (draft only, soft delete)
// @Tags assessment-management
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Success 204
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id} [delete]
func (h *AssessmentManagementHandler) Delete(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	if err := h.svc.DeleteAssessment(c.Request.Context(), id); err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// --- Question Management ---

// GetQuestions godoc
// @Summary Get all questions for an assessment
// @Tags assessment-questions
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id}/questions [get]
func (h *AssessmentManagementHandler) GetQuestions(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	result, err := h.svc.GetQuestions(c.Request.Context(), id)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": result})
}

// AddQuestion godoc
// @Summary Add a question to an assessment
// @Tags assessment-questions
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Param request body dto.CreateQuestionRequest true "Question data"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id}/questions [post]
func (h *AssessmentManagementHandler) AddQuestion(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	var req dto.CreateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.AddQuestion(c.Request.Context(), id, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"questions": result})
}

// UpdateQuestion godoc
// @Summary Update a question by index
// @Tags assessment-questions
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Param idx path int true "Question index (0-based)"
// @Param request body dto.UpdateQuestionRequest true "Question update data"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id}/questions/{idx} [put]
func (h *AssessmentManagementHandler) UpdateQuestion(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	idx, err := strconv.Atoi(c.Param("idx"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idx must be a valid integer", "code": "VALIDATION_ERROR"})
		return
	}

	var req dto.UpdateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.UpdateQuestion(c.Request.Context(), id, idx, req)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": result})
}

// DeleteQuestion godoc
// @Summary Delete a question by index
// @Tags assessment-questions
// @Accept json
// @Produce json
// @Param id path string true "Assessment ID (UUID)"
// @Param idx path int true "Question index (0-based)"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 422 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Security BearerAuth
// @Router /assessments/{id}/questions/{idx} [delete]
func (h *AssessmentManagementHandler) DeleteQuestion(c *gin.Context) {
	id, err := parseUUIDParam(c, "id")
	if err != nil {
		middleware.HandleError(c, err)
		return
	}

	idx, err := strconv.Atoi(c.Param("idx"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "idx must be a valid integer", "code": "VALIDATION_ERROR"})
		return
	}

	result, err := h.svc.DeleteQuestion(c.Request.Context(), id, idx)
	if err != nil {
		middleware.HandleError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"questions": result})
}
