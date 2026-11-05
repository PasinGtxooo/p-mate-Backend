package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/chat"
	"github.com/FrostBitzX/smart-task-ai/internal/application/chat/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// ChatHandler handles chat-related HTTP requests
type ChatHandler struct {
	sendMessageUC *usecase.SendMessageUseCase
	logger        logger.Logger
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(sendMessageUC *usecase.SendMessageUseCase, l logger.Logger) *ChatHandler {
	return &ChatHandler{
		sendMessageUC: sendMessageUC,
		logger:        l,
	}
}

// SendMessage handles POST /api/:projectId/chat endpoint
func (h *ChatHandler) SendMessage(c *fiber.Ctx) error {
	// Get projectId from URL parameter
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID is required", "INVALID_PROJECT_ID", nil))
	}

	req, err := requests.ParseAndValidate[chat.SendMessageRequestDTO](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	// Set projectID from URL parameter
	req.ProjectID = projectID

	// Get AccountID and NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	resp, err := h.sendMessageUC.Execute(c.Context(), accountID, nodeID, req)
	if err != nil {
		h.logger.Error("Failed to send message", map[string]interface{}{
			"error":      err.Error(),
			"account_id": accountID,
			"node_id":    nodeID,
			"project_id": req.ProjectID,
		})
		return responses.Error(c, err)
	}

	return responses.Success(c, resp, "Message sent successfully")
}
