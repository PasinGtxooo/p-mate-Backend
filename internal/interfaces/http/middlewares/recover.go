package middlewares

import (
	"fmt"
	"runtime/debug"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
)

// RecoverConfig holds configuration for the recover middleware
type RecoverConfig struct {
	// EnableStackTrace enables stack trace logging on panic
	EnableStackTrace bool
}

// DefaultRecoverConfig returns the default recover configuration
func DefaultRecoverConfig() RecoverConfig {
	return RecoverConfig{
		EnableStackTrace: true,
	}
}

// RecoverMiddleware returns a middleware that recovers from panics
func RecoverMiddleware(log logger.Logger, config ...RecoverConfig) fiber.Handler {
	cfg := DefaultRecoverConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				err, ok := r.(error)
				if !ok {
					err = fmt.Errorf("%v", r)
				}

				// Log the panic with stack trace
				fields := map[string]interface{}{
					"error":  err.Error(),
					"path":   c.Path(),
					"method": c.Method(),
				}

				if cfg.EnableStackTrace {
					fields["stack"] = string(debug.Stack())
				}

				log.Error("Panic recovered", fields)

				// Return internal server error
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"success": false,
					"message": "Internal server error",
					"data":    nil,
					"error": fiber.Map{
						"code":    fiber.StatusInternalServerError,
						"message": "INTERNAL_SERVER_ERROR",
						"details": "An unexpected error occurred",
					},
				})
			}
		}()

		return c.Next()
	}
}
