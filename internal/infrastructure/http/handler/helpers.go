package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/EduGoGroup/edugo-shared/auth"
	"github.com/EduGoGroup/edugo-shared/common/errors"

	mw "github.com/EduGoGroup/edugo-api-mobile-new/internal/infrastructure/http/middleware"
)

// getUserID extracts the authenticated user's ID from the Gin context.
func getUserID(c *gin.Context) (uuid.UUID, error) {
	raw, exists := c.Get(mw.ContextKeyUserID)
	if !exists {
		return uuid.Nil, errors.NewUnauthorizedError("user_id not found in context")
	}
	idStr, ok := raw.(string)
	if !ok {
		return uuid.Nil, errors.NewUnauthorizedError("user_id has invalid type")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, errors.NewUnauthorizedError("user_id is not a valid UUID")
	}
	return id, nil
}

// getSchoolID extracts the school ID from the user's active context.
func getSchoolID(c *gin.Context) (uuid.UUID, error) {
	raw, exists := c.Get(mw.ContextKeyActiveContext)
	if !exists {
		return uuid.Nil, errors.NewUnauthorizedError("active_context not found")
	}
	ctx, ok := raw.(*auth.UserContext)
	if !ok || ctx == nil {
		return uuid.Nil, errors.NewUnauthorizedError("active_context has invalid type")
	}
	if ctx.SchoolID == "" {
		return uuid.Nil, errors.NewForbiddenError("no school in active context")
	}
	id, err := uuid.Parse(ctx.SchoolID)
	if err != nil {
		return uuid.Nil, errors.NewUnauthorizedError("school_id is not a valid UUID")
	}
	return id, nil
}

// parseUUIDParam parses a UUID from a URL path parameter.
func parseUUIDParam(c *gin.Context, name string) (uuid.UUID, error) {
	raw := c.Param(name)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, errors.NewValidationError(name + " must be a valid UUID")
	}
	return id, nil
}
