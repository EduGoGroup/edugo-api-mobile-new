package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-shared/auth"

	"github.com/EduGoGroup/edugo-api-mobile-new/internal/client"
)

// Context key constants for values injected by the auth middleware.
const (
	ContextKeyUserID        = "user_id"
	ContextKeyEmail         = "email"
	ContextKeyRole          = "role"
	ContextKeyActiveContext = "active_context"
	ContextKeyClaims        = "jwt_claims"
)

// RemoteAuthMiddleware validates the JWT token from the Authorization header
// using the AuthClient (local + optional remote fallback), then injects
// user info into the Gin context.
func RemoteAuthMiddleware(authClient *client.AuthClient) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "authorization header required",
				"code":  "MISSING_AUTH_HEADER",
			})
			c.Abort()
			return
		}

		if len(authHeader) < 8 || authHeader[:7] != "Bearer " {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format, expected 'Bearer {token}'",
				"code":  "INVALID_AUTH_FORMAT",
			})
			c.Abort()
			return
		}

		tokenString := authHeader[7:]

		info, err := authClient.ValidateToken(c.Request.Context(), tokenString)
		if err != nil || !info.Valid {
			errMsg := "invalid or expired token"
			if info != nil && info.Error != "" {
				errMsg = info.Error
			}
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": errMsg,
				"code":  "INVALID_TOKEN",
			})
			c.Abort()
			return
		}

		// Inject user info into context
		c.Set(ContextKeyUserID, info.UserID)
		c.Set(ContextKeyEmail, info.Email)
		if info.ActiveContext != nil {
			c.Set(ContextKeyRole, info.ActiveContext.RoleName)
			c.Set(ContextKeyActiveContext, info.ActiveContext)
		}

		// Set jwt_claims so sharedmw.RequirePermission can read permissions
		c.Set(ContextKeyClaims, &auth.Claims{
			UserID:        info.UserID,
			Email:         info.Email,
			ActiveContext: info.ActiveContext,
		})

		c.Next()
	}
}
