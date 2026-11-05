package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/invitation"
	"github.com/FrostBitzX/smart-task-ai/internal/application/invitation/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type InvitationHandler struct {
	CreateInvitationUC       *usecase.CreateInvitationUseCase
	AcceptInvitationUC       *usecase.AcceptInvitationUseCase
	RejectInvitationUC       *usecase.RejectInvitationUseCase
	CancelInvitationUC       *usecase.CancelInvitationUseCase
	ListMyInvitationsUC      *usecase.ListMyInvitationsUseCase
	ListProjectInvitationsUC *usecase.ListProjectInvitationsUseCase
	logger                   logger.Logger
}

func NewInvitationHandler(
	create *usecase.CreateInvitationUseCase,
	accept *usecase.AcceptInvitationUseCase,
	reject *usecase.RejectInvitationUseCase,
	cancel *usecase.CancelInvitationUseCase,
	listMy *usecase.ListMyInvitationsUseCase,
	listProject *usecase.ListProjectInvitationsUseCase,
	l logger.Logger,
) *InvitationHandler {
	return &InvitationHandler{
		CreateInvitationUC:       create,
		AcceptInvitationUC:       accept,
		RejectInvitationUC:       reject,
		CancelInvitationUC:       cancel,
		ListMyInvitationsUC:      listMy,
		ListProjectInvitationsUC: listProject,
		logger:                   l,
	}
}

// CreateInvitation handles POST /api/projects/:projectId/invite
func (h *InvitationHandler) CreateInvitation(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	req, err := requests.ParseAndValidate[invitation.CreateInvitationRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

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

	// Set IDs from path and JWT
	req.ProjectID = projectID
	req.InviterAccountID = accountID

	data, err := h.CreateInvitationUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Invitation created successfully")
}

// AcceptInvitation handles POST /api/projects/:projectId/invite/:inviteeAccountId/accept
func (h *InvitationHandler) AcceptInvitation(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	inviteeAccountID := c.Params("inviteeAccountId")
	if inviteeAccountID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing inviteeAccountId", "MISSING_INVITEE_ACCOUNT_ID", nil))
	}

	// Get NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	req := &invitation.AcceptInvitationRequest{
		ProjectID:        projectID,
		InviteeAccountID: inviteeAccountID,
	}

	data, err := h.AcceptInvitationUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Invitation accepted successfully")
}

// RejectInvitation handles POST /api/projects/:projectId/invite/:inviteeAccountId/reject
func (h *InvitationHandler) RejectInvitation(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	inviteeAccountID := c.Params("inviteeAccountId")
	if inviteeAccountID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing inviteeAccountId", "MISSING_INVITEE_ACCOUNT_ID", nil))
	}

	req := &invitation.RejectInvitationRequest{
		ProjectID:        projectID,
		InviteeAccountID: inviteeAccountID,
	}

	data, err := h.RejectInvitationUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Invitation rejected successfully")
}

// CancelInvitation handles DELETE /api/projects/:projectId/invite/:inviteeAccountId
func (h *InvitationHandler) CancelInvitation(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	inviteeAccountID := c.Params("inviteeAccountId")
	if inviteeAccountID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing inviteeAccountId", "MISSING_INVITEE_ACCOUNT_ID", nil))
	}

	// Get AccountID from JWT claims (the canceller)
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

	req := &invitation.CancelInvitationRequest{
		ProjectID:        projectID,
		InviteeAccountID: inviteeAccountID,
		CancellerID:      accountID,
	}

	data, err := h.CancelInvitationUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Invitation cancelled successfully")
}

// ListMyInvitations handles GET /api/accounts/me/invite
func (h *InvitationHandler) ListMyInvitations(c *fiber.Ctx) error {
	// Get AccountID from JWT claims
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

	req := &invitation.ListMyInvitationsRequest{
		InviteeAccountID: accountID,
	}

	data, err := h.ListMyInvitationsUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Invitations retrieved successfully")
}

// ListProjectInvitations handles GET /api/projects/:projectId/invite
func (h *InvitationHandler) ListProjectInvitations(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	// Get AccountID from JWT claims
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

	req := &invitation.ListProjectInvitationsRequest{
		ProjectID: projectID,
		AccountID: accountID,
	}

	data, err := h.ListProjectInvitationsUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Project invitations retrieved successfully")
}
