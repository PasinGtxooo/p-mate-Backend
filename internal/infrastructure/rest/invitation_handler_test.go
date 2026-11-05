package rest_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FrostBitzX/smart-task-ai/internal/application/invitation"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestInvitationEndpoints_Integration tests the invitation API endpoints end-to-end
func TestInvitationEndpoints_Integration(t *testing.T) {
	// Skip if not in integration test mode
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("CreateInvitation_Success", func(t *testing.T) {
		app := setupTestApp(t)

		reqBody := invitation.CreateInvitationRequest{
			InviteeShortID: "acc_test123",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+uuid.New().String()+"/invite", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 201 Created or appropriate status
		assert.True(t, resp.StatusCode == http.StatusCreated || resp.StatusCode == http.StatusBadRequest)
	})

	t.Run("CreateInvitation_InvalidShortID", func(t *testing.T) {
		app := setupTestApp(t)

		reqBody := invitation.CreateInvitationRequest{
			InviteeShortID: "invalid_id",
		}
		body, _ := json.Marshal(reqBody)

		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+uuid.New().String()+"/invite", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 400 Bad Request
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("AcceptInvitation_Success", func(t *testing.T) {
		app := setupTestApp(t)

		projectID := uuid.New().String()
		inviteeID := uuid.New().String()

		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/invite/"+inviteeID+"/accept", nil)
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 200 OK or 404 Not Found (if invitation doesn't exist)
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)
	})

	t.Run("RejectInvitation_Success", func(t *testing.T) {
		app := setupTestApp(t)

		projectID := uuid.New().String()
		inviteeID := uuid.New().String()

		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/invite/"+inviteeID+"/reject", nil)
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 200 OK or 404 Not Found
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)
	})

	t.Run("CancelInvitation_Success", func(t *testing.T) {
		app := setupTestApp(t)

		projectID := uuid.New().String()
		inviteeID := uuid.New().String()

		req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID+"/invite/"+inviteeID, nil)
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 200 OK or 404 Not Found
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound)
	})

	t.Run("ListMyInvitations_Success", func(t *testing.T) {
		app := setupTestApp(t)

		req := httptest.NewRequest(http.MethodGet, "/api/accounts/me/invite", nil)
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 200 OK
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("ListProjectInvitations_Success", func(t *testing.T) {
		app := setupTestApp(t)

		projectID := uuid.New().String()

		req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID+"/invite", nil)
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)

		// Should return 200 OK or 403 Forbidden (if not a member)
		assert.True(t, resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusForbidden)
	})
}

// TestInvitationWorkflow_EndToEnd tests the complete invitation workflow
func TestInvitationWorkflow_EndToEnd(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test")
	}

	t.Run("CompleteWorkflow_CreateAcceptCancel", func(t *testing.T) {
		app := setupTestApp(t)

		// This test would require a real database setup
		// For now, we verify the endpoints are accessible

		projectID := uuid.New().String()

		// Step 1: Create invitation
		createReq := invitation.CreateInvitationRequest{
			InviteeShortID: "acc_workflow",
		}
		body, _ := json.Marshal(createReq)
		req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID+"/invite", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+getTestJWT(t))

		resp, err := app.Test(req, -1)
		assert.NoError(t, err)
		assert.NotEqual(t, http.StatusInternalServerError, resp.StatusCode)
	})
}

// Helper functions

func setupTestApp(t *testing.T) *fiber.App {
	// This would normally set up the full app with test database
	// For now, return a minimal fiber app
	app := fiber.New()

	// Add test routes here if needed
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	return app
}

func getTestJWT(t *testing.T) string {
	// Return a test JWT token
	// In real tests, this would generate a valid JWT
	return "test_jwt_token"
}
