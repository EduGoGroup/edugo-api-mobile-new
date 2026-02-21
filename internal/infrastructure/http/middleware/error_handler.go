package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-shared/common/errors"
	"github.com/EduGoGroup/edugo-shared/logger"
)

// ErrorResponse is the standard error JSON response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code"`
	Details string `json:"details,omitempty"`
}

// ErrorHandler is a middleware that recovers from panics and returns a JSON error.
func ErrorHandler(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered", "panic", r)
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Error: "internal server error",
					Code:  "INTERNAL_ERROR",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}

// HandleError writes the appropriate HTTP response for an error.
// It should be called from handlers when an error occurs.
func HandleError(c *gin.Context, err error) {
	if appErr, ok := errors.GetAppError(err); ok {
		resp := ErrorResponse{
			Error:   appErr.Message,
			Code:    string(appErr.Code),
			Details: appErr.Details,
		}
		c.JSON(appErr.StatusCode, resp)
		return
	}

	c.JSON(http.StatusInternalServerError, ErrorResponse{
		Error: "internal server error",
		Code:  "INTERNAL_ERROR",
	})
}
