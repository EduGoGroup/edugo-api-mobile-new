package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/EduGoGroup/edugo-shared/common/errors"

	"github.com/EduGoGroup/edugo-api-mobile-new/test/mock"
)

func TestErrorHandler_PanicRecovery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		handler    gin.HandlerFunc
		wantStatus int
		wantCode   string
	}{
		{
			name: "recovers from panic",
			handler: func(_ *gin.Context) {
				panic("something went wrong")
			},
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
		{
			name: "no panic - normal response",
			handler: func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"ok": true})
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.Use(ErrorHandler(mock.MockLogger{}))
			r.GET("/test", tt.handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			if tt.wantCode != "" {
				assert.Contains(t, w.Body.String(), tt.wantCode)
			}
		})
	}
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{
			name:       "validation error",
			err:        errors.NewValidationError("field is required"),
			wantStatus: http.StatusBadRequest,
			wantCode:   "VALIDATION_ERROR",
		},
		{
			name:       "not found error",
			err:        errors.NewNotFoundError("material"),
			wantStatus: http.StatusNotFound,
			wantCode:   "NOT_FOUND",
		},
		{
			name:       "unauthorized error",
			err:        errors.NewUnauthorizedError("invalid token"),
			wantStatus: http.StatusUnauthorized,
			wantCode:   "UNAUTHORIZED",
		},
		{
			name:       "forbidden error",
			err:        errors.NewForbiddenError("no permission"),
			wantStatus: http.StatusForbidden,
			wantCode:   "FORBIDDEN",
		},
		{
			name:       "business rule error",
			err:        errors.NewBusinessRuleError("max attempts reached"),
			wantStatus: http.StatusUnprocessableEntity,
			wantCode:   "BUSINESS_RULE_VIOLATION",
		},
		{
			name:       "internal error",
			err:        errors.NewInternalError("something broke", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
		{
			name:       "database error",
			err:        errors.NewDatabaseError("insert", nil),
			wantStatus: http.StatusInternalServerError,
			wantCode:   "DATABASE_ERROR",
		},
		{
			name:       "generic error - fallback to 500",
			err:        assert.AnError,
			wantStatus: http.StatusInternalServerError,
			wantCode:   "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.GET("/test", func(c *gin.Context) {
				HandleError(c, tt.err)
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantCode)
		})
	}
}
