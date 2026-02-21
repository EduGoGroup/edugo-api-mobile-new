package middleware

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/EduGoGroup/edugo-shared/logger"
)

// RequestLogging logs each incoming HTTP request with method, path, status, and latency.
func RequestLogging(log logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()

		log.Info("http request",
			"method", c.Request.Method,
			"path", path,
			"status", status,
			"latency_ms", latency.Milliseconds(),
			"client_ip", c.ClientIP(),
		)
	}
}
