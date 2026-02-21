package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db    *sql.DB
	mongo *mongo.Client
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *sql.DB, mongo *mongo.Client) *HealthHandler {
	return &HealthHandler{db: db, mongo: mongo}
}

// Health returns the health status of the API and its dependencies.
func (h *HealthHandler) Health(c *gin.Context) {
	status := "ok"
	checks := map[string]string{}

	if h.db != nil {
		if err := h.db.PingContext(c.Request.Context()); err != nil {
			status = "degraded"
			checks["postgres"] = "unhealthy"
		} else {
			checks["postgres"] = "healthy"
		}
	}

	if h.mongo != nil {
		if err := h.mongo.Ping(c.Request.Context(), nil); err != nil {
			status = "degraded"
			checks["mongodb"] = "unhealthy"
		} else {
			checks["mongodb"] = "healthy"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status": status,
		"checks": checks,
	})
}
