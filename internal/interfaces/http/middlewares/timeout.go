package middlewares

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TimeoutConfig holds configuration for the timeout middleware
type TimeoutConfig struct {
	// Timeout is the maximum duration for request processing
	Timeout time.Duration

	// ErrorMessage is the message returned when timeout occurs
	ErrorMessage string

	// ErrorHandler is called when timeout occurs (optional)
	ErrorHandler fiber.ErrorHandler
}

// DefaultTimeoutConfig returns the default timeout configuration
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		Timeout:      30 * time.Second,
		ErrorMessage: "Request timeout",
	}
}

// TimeoutMiddleware returns a middleware that adds timeout to requests
func TimeoutMiddleware(config ...TimeoutConfig) fiber.Handler {
	cfg := DefaultTimeoutConfig()
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(c *fiber.Ctx) error {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.UserContext(), cfg.Timeout)
		defer cancel()

		// Set the context with timeout
		c.SetUserContext(ctx)

		// Create channel to receive result
		done := make(chan error, 1)

		go func() {
			done <- c.Next()
		}()

		select {
		case err := <-done:
			return err
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				// Handle custom error handler if provided
				if cfg.ErrorHandler != nil {
					return cfg.ErrorHandler(c, fiber.ErrRequestTimeout)
				}

				return c.Status(fiber.StatusGatewayTimeout).JSON(fiber.Map{
					"success": false,
					"message": cfg.ErrorMessage,
					"data":    nil,
					"error": fiber.Map{
						"code":    fiber.StatusGatewayTimeout,
						"message": "GATEWAY_TIMEOUT",
						"details": "Request processing exceeded the time limit",
					},
				})
			}
			return ctx.Err()
		}
	}
}

// TimeoutWithDuration is a convenience function to create timeout middleware with specific duration
func TimeoutWithDuration(timeout time.Duration) fiber.Handler {
	return TimeoutMiddleware(TimeoutConfig{
		Timeout:      timeout,
		ErrorMessage: "Request timeout",
	})
}
