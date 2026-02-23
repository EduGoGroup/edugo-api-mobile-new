package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"gorm.io/gorm"
)

// HealthHandler handles health check endpoints.
type HealthHandler struct {
	db    *gorm.DB
	mongo *mongo.Client
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *gorm.DB, mongo *mongo.Client) *HealthHandler {
	return &HealthHandler{db: db, mongo: mongo}
}

// Health godoc
// @Summary Health check
// @Description Returns the health status of the API and its dependencies
// @Tags health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /health [get]
func (h *HealthHandler) Health(c *gin.Context) {
	status := "ok"
	checks := map[string]string{}

	if h.db != nil {
		sqlDB, err := h.db.DB()
		if err != nil {
			status = "degraded"
			checks["postgres"] = "unhealthy"
		} else if err := sqlDB.PingContext(c.Request.Context()); err != nil {
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
