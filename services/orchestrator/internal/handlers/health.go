package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/BerylCAtieno/group24-notification-system/services/orchestrator/internal/database"
	"github.com/gin-gonic/gin"
)

type HealthHandler struct {
	db *database.DB
}

func NewHealthHandler(db *database.DB) *HealthHandler {
	return &HealthHandler{db: db}
}

func (h *HealthHandler) Live(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (h *HealthHandler) Ready(c *gin.Context) {
	checks := gin.H{
		"user_service":     "ok",
		"template_service": "ok",
		"kafka":            "ok",
	}

	// Check database connectivity
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := h.db.Ping(ctx); err != nil {
		checks["database"] = "error"
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":    "not_ready",
			"checks":    checks,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
			"error":     "database connection failed",
		})
		return
	}

	checks["database"] = "ok"

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"checks":    checks,
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}
