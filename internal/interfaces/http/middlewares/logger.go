package middlewares

import (
	"encoding/json"
	"strings"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
)

func FiberLoggerMiddleware(log logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {

		if c.Path() == "/healthz" || strings.HasPrefix(c.Path(), "/api/v2/healthz") {
			return c.Next()
		}

		start := time.Now()
		path := c.Path()

		// Check if this is an upload endpoint
		isUploadEndpoint := path == "/api/files/avatar/presign" || path == "/api/files/avatar/upload"

		// For upload endpoints, log detailed information
		if isUploadEndpoint {
			return loggerUploadRequest(c, log, start)
		}

		// Standard logging for other endpoints
		err := c.Next()
		duration := time.Since(start)

		log.Info("http_request", map[string]interface{}{
			"method":  c.Method(),
			"path":    c.Path(),
			"status":  c.Response().StatusCode(),
			"latency": duration.String(),
			"ip":      c.IP(),
		})

		if err != nil {
			log.Error("http_error", map[string]interface{}{
				"error": err.Error(),
				"path":  c.Path(),
			})
		}

		return err
	}
}

// loggerUploadRequest handles detailed logging for upload endpoints
func loggerUploadRequest(c *fiber.Ctx, log logger.Logger, start time.Time) error {
	path := c.Path()

	// Extract JWT claims for account_id and node_id
	var accountID, nodeID string
	if jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{}); ok {
		if aid, ok := jwtClaims["AccountId"].(string); ok {
			accountID = aid
		}
		if nid, ok := jwtClaims["NodeId"].(string); ok {
			nodeID = nid
		}
	}

	// Read request body for logging
	var requestBody map[string]interface{}
	if c.Method() == "POST" {
		_ = json.Unmarshal(c.Body(), &requestBody)
	}

	// Execute the handler
	err := c.Next()

	// Calculate duration
	duration := time.Since(start)
	status := c.Response().StatusCode()

	// Prepare log fields
	logFields := map[string]interface{}{
		"method":     c.Method(),
		"path":       path,
		"status":     status,
		"duration":   duration.String(),
		"timestamp":  start.Format(time.RFC3339),
		"account_id": accountID,
		"node_id":    nodeID,
	}

	// Add endpoint-specific fields
	if path == "/api/files/avatar/presign" {
		if fileName, ok := requestBody["file_name"].(string); ok {
			logFields["file_name"] = fileName
		}
		if fileSize, ok := requestBody["file_size"].(float64); ok {
			logFields["size"] = int64(fileSize)
		}
		if contentType, ok := requestBody["content_type"].(string); ok {
			logFields["content_type"] = contentType
		}

		// Try to extract key from response if successful
		if status >= 200 && status < 300 {
			var responseData map[string]interface{}
			if json.Unmarshal(c.Response().Body(), &responseData) == nil {
				if data, ok := responseData["data"].(map[string]interface{}); ok {
					if key, ok := data["key"].(string); ok {
						logFields["key"] = key
					}
				}
			}
		}
	} else if path == "/api/files/avatar/upload" {
		if key, ok := requestBody["key"].(string); ok {
			logFields["key"] = key
		}

		// Try to extract URL from response if successful
		if status >= 200 && status < 300 {
			var responseData map[string]interface{}
			if json.Unmarshal(c.Response().Body(), &responseData) == nil {
				if data, ok := responseData["data"].(map[string]interface{}); ok {
					if url, ok := data["url"].(string); ok {
						logFields["url"] = url
					}
				}
			}
		}
	}

	// Log based on status
	if status >= 200 && status < 300 {
		if path == "/api/files/avatar/presign" {
			log.Info("file_presign_success", logFields)
		} else {
			log.Info("file_complete_success", logFields)
		}
	} else {
		if err != nil {
			logFields["error"] = err.Error()
		}
		if path == "/api/files/avatar/presign" {
			log.Error("file_presign_failed", logFields)
		} else {
			log.Error("file_complete_failed", logFields)
		}
	}

	return err
}
