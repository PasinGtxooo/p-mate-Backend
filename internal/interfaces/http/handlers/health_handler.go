package handlers

import (
	"context"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/database"
	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles health check endpoints
type HealthHandler struct {
	dbConnector *database.DBConnector
	startTime   time.Time
}

// NewHealthHandler creates a new HealthHandler
func NewHealthHandler(dbConnector *database.DBConnector) *HealthHandler {
	return &HealthHandler{
		dbConnector: dbConnector,
		startTime:   time.Now(),
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string                 `json:"status"`
	Timestamp string                 `json:"timestamp"`
	Uptime    string                 `json:"uptime"`
	Version   string                 `json:"version"`
	Checks    map[string]HealthCheck `json:"checks"`
}

// HealthCheck represents individual health check result
type HealthCheck struct {
	Status  string                 `json:"status"`
	Message string                 `json:"message,omitempty"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// Health performs a basic health check
// @Summary Health check
// @Description Returns the health status of the API
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	response := HealthResponse{
		Status:    "healthy",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Uptime:    time.Since(h.startTime).String(),
		Version:   "1.0.0",
		Checks:    make(map[string]HealthCheck),
	}

	// Check database connection
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()

	if err := h.dbConnector.HealthCheck(ctx); err != nil {
		response.Status = "unhealthy"
		response.Checks["database"] = HealthCheck{
			Status:  "unhealthy",
			Message: err.Error(),
		}
		return c.Status(fiber.StatusServiceUnavailable).JSON(response)
	}

	response.Checks["database"] = HealthCheck{
		Status:  "healthy",
		Message: "Database connection is active",
		Details: h.dbConnector.GetStats(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
