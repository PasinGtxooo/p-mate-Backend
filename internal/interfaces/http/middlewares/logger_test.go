package middlewares

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

// MockLogger implements logger.Logger for testing
type MockLogger struct {
	InfoCalls  []LogCall
	ErrorCalls []LogCall
	WarnCalls  []LogCall
	DebugCalls []LogCall
}

type LogCall struct {
	Message string
	Fields  map[string]interface{}
}

func (m *MockLogger) Info(msg string, fields ...map[string]interface{}) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.InfoCalls = append(m.InfoCalls, call)
}

func (m *MockLogger) Warn(msg string, fields ...map[string]interface{}) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.WarnCalls = append(m.WarnCalls, call)
}

func (m *MockLogger) Error(msg string, fields ...map[string]interface{}) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.ErrorCalls = append(m.ErrorCalls, call)
}

func (m *MockLogger) Debug(msg string, fields ...map[string]interface{}) {
	call := LogCall{Message: msg}
	if len(fields) > 0 {
		call.Fields = fields[0]
	}
	m.DebugCalls = append(m.DebugCalls, call)
}

func (m *MockLogger) With(fields map[string]interface{}) logger.Logger {
	return m
}

func TestFiberLoggerMiddleware_UploadPresignSuccess(t *testing.T) {
	app := fiber.New()
	mockLogger := &MockLogger{}

	// Setup JWT middleware mock
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("jwt_claims", map[string]interface{}{
			"AccountId": "test-account-123",
			"NodeId":    "test-node-456",
		})
		return c.Next()
	})

	// Apply logger middleware
	app.Use(FiberLoggerMiddleware(mockLogger))

	// Setup test handler
	app.Post("/api/files/avatar/presign", func(c *fiber.Ctx) error {
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"presigned_url":    "https://example.com/presigned",
				"key":              "avatars/test-file.jpg",
				"expires_datetime": "2025-02-12T14:30:00Z",
			},
		}
		return c.JSON(response)
	})

	// Create test request
	requestBody := map[string]interface{}{
		"file_name":    "test.jpg",
		"file_size":    1024,
		"content_type": "image/jpeg",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/files/avatar/presign", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify logging
	assert.Len(t, mockLogger.InfoCalls, 1)
	assert.Equal(t, "file_presign_success", mockLogger.InfoCalls[0].Message)

	fields := mockLogger.InfoCalls[0].Fields
	assert.Equal(t, "test-account-123", fields["account_id"])
	assert.Equal(t, "test-node-456", fields["node_id"])
	assert.Equal(t, "test.jpg", fields["file_name"])
	assert.Equal(t, int64(1024), fields["size"])
	assert.Equal(t, "image/jpeg", fields["content_type"])
	assert.Equal(t, "avatars/test-file.jpg", fields["key"])
	assert.Equal(t, 200, fields["status"])
}

func TestFiberLoggerMiddleware_UploadCompleteSuccess(t *testing.T) {
	app := fiber.New()
	mockLogger := &MockLogger{}

	// Setup JWT middleware mock
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("jwt_claims", map[string]interface{}{
			"AccountId": "test-account-123",
			"NodeId":    "test-node-456",
		})
		return c.Next()
	})

	// Apply logger middleware
	app.Use(FiberLoggerMiddleware(mockLogger))

	// Setup test handler
	app.Post("/api/files/avatar/upload", func(c *fiber.Ctx) error {
		response := map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"url":        "https://example.com/avatars/test-file.jpg",
				"updated_at": "2024-01-15T10:30:00Z",
			},
		}
		return c.JSON(response)
	})

	// Create test request
	requestBody := map[string]interface{}{
		"key": "avatars/test-file.jpg",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/files/avatar/upload", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify logging
	assert.Len(t, mockLogger.InfoCalls, 1)
	assert.Equal(t, "file_complete_success", mockLogger.InfoCalls[0].Message)

	fields := mockLogger.InfoCalls[0].Fields
	assert.Equal(t, "test-account-123", fields["account_id"])
	assert.Equal(t, "test-node-456", fields["node_id"])
	assert.Equal(t, "avatars/test-file.jpg", fields["key"])
	assert.Equal(t, "https://example.com/avatars/test-file.jpg", fields["url"])
	assert.Equal(t, 200, fields["status"])
}

func TestFiberLoggerMiddleware_UploadPresignFailure(t *testing.T) {
	app := fiber.New()
	mockLogger := &MockLogger{}

	// Setup JWT middleware mock
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("jwt_claims", map[string]interface{}{
			"AccountId": "test-account-123",
			"NodeId":    "test-node-456",
		})
		return c.Next()
	})

	// Apply logger middleware
	app.Use(FiberLoggerMiddleware(mockLogger))

	// Setup test handler that returns error
	app.Post("/api/files/avatar/presign", func(c *fiber.Ctx) error {
		return c.Status(400).JSON(map[string]interface{}{
			"status": "error",
			"error": map[string]interface{}{
				"code":    "FILE_SIZE_EXCEEDED",
				"message": "File too large",
			},
		})
	})

	// Create test request
	requestBody := map[string]interface{}{
		"file_name":    "large.jpg",
		"file_size":    10485760, // 10MB
		"content_type": "image/jpeg",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/files/avatar/presign", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 400, resp.StatusCode)

	// Verify error logging
	assert.Len(t, mockLogger.ErrorCalls, 1)
	assert.Equal(t, "file_presign_failed", mockLogger.ErrorCalls[0].Message)

	fields := mockLogger.ErrorCalls[0].Fields
	assert.Equal(t, "test-account-123", fields["account_id"])
	assert.Equal(t, "test-node-456", fields["node_id"])
	assert.Equal(t, "large.jpg", fields["file_name"])
	assert.Equal(t, int64(10485760), fields["size"])
	assert.Equal(t, 400, fields["status"])
}

func TestFiberLoggerMiddleware_StandardEndpoint(t *testing.T) {
	app := fiber.New()
	mockLogger := &MockLogger{}

	// Apply logger middleware
	app.Use(FiberLoggerMiddleware(mockLogger))

	// Setup test handler for non-upload endpoint
	app.Get("/api/profiles", func(c *fiber.Ctx) error {
		return c.JSON(map[string]interface{}{"status": "success"})
	})

	req := httptest.NewRequest("GET", "/api/profiles", nil)

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify standard logging occurred (not upload-specific)
	assert.Len(t, mockLogger.InfoCalls, 1)
	assert.Equal(t, "http_request", mockLogger.InfoCalls[0].Message)
	assert.Len(t, mockLogger.ErrorCalls, 0)
}

func TestFiberLoggerMiddleware_UploadMissingJWTClaims(t *testing.T) {
	app := fiber.New()
	mockLogger := &MockLogger{}

	// Apply logger middleware without JWT claims
	app.Use(FiberLoggerMiddleware(mockLogger))

	// Setup test handler
	app.Post("/api/files/avatar/presign", func(c *fiber.Ctx) error {
		return c.JSON(map[string]interface{}{
			"status": "success",
			"data": map[string]interface{}{
				"key": "avatars/test.jpg",
			},
		})
	})

	// Create test request
	requestBody := map[string]interface{}{
		"file_name":    "test.jpg",
		"file_size":    1024,
		"content_type": "image/jpeg",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest("POST", "/api/files/avatar/presign", strings.NewReader(string(bodyBytes)))
	req.Header.Set("Content-Type", "application/json")

	// Execute request
	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	// Verify logging still occurs but with empty account_id and node_id
	assert.Len(t, mockLogger.InfoCalls, 1)
	fields := mockLogger.InfoCalls[0].Fields
	assert.Equal(t, "", fields["account_id"])
	assert.Equal(t, "", fields["node_id"])
}
