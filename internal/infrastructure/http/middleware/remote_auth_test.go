package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/EduGoGroup/edugo-shared/auth"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/client"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestRemoteAuthMiddleware(t *testing.T) {
	jwtSecret := "test-secret-key-for-unit-tests"
	jwtIssuer := "edugo-central"
	jwtMgr := auth.NewJWTManager(jwtSecret, jwtIssuer)

	userID := uuid.New().String()
	email := "test@example.com"
	activeCtx := &auth.UserContext{
		RoleID:      uuid.New().String(),
		RoleName:    "teacher",
		SchoolID:    uuid.New().String(),
		SchoolName:  "Test School",
		Permissions: []string{"materials:read"},
	}

	validToken, _, err := jwtMgr.GenerateTokenWithContext(userID, email, activeCtx, time.Hour)
	require.NoError(t, err)

	authClient := client.NewAuthClient(client.AuthClientConfig{
		JWTSecret: jwtSecret,
		JWTIssuer: jwtIssuer,
	})

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
		wantCode   string
		checkCtx   func(t *testing.T, c *gin.Context)
	}{
		{
			name:       "missing authorization header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "MISSING_AUTH_HEADER",
		},
		{
			name:       "invalid format - no Bearer prefix",
			authHeader: "Basic abc123",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_AUTH_FORMAT",
		},
		{
			name:       "invalid format - too short",
			authHeader: "Bearer",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_AUTH_FORMAT",
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid.token.here",
			wantStatus: http.StatusUnauthorized,
			wantCode:   "INVALID_TOKEN",
		},
		{
			name:       "valid token",
			authHeader: "Bearer " + validToken,
			wantStatus: http.StatusOK,
			checkCtx: func(t *testing.T, c *gin.Context) {
				uid, exists := c.Get(ContextKeyUserID)
				require.True(t, exists)
				assert.Equal(t, userID, uid)

				eml, exists := c.Get(ContextKeyEmail)
				require.True(t, exists)
				assert.Equal(t, email, eml)

				role, exists := c.Get(ContextKeyRole)
				require.True(t, exists)
				assert.Equal(t, "teacher", role)

				ctx, exists := c.Get(ContextKeyActiveContext)
				require.True(t, exists)
				assert.NotNil(t, ctx)
				uc, ok := ctx.(*auth.UserContext)
				require.True(t, ok)
				assert.Equal(t, "teacher", uc.RoleName)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()

			var capturedCtx *gin.Context
			r.Use(RemoteAuthMiddleware(authClient))
			r.GET("/test", func(c *gin.Context) {
				capturedCtx = c.Copy()
				c.JSON(http.StatusOK, gin.H{"ok": true})
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantCode != "" {
				assert.Contains(t, w.Body.String(), tt.wantCode)
			}

			if tt.checkCtx != nil && capturedCtx != nil {
				tt.checkCtx(t, capturedCtx)
			}
		})
	}
}
