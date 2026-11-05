package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/invitation"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
)

type CreateInvitationUseCase struct {
	invitationService *service.InvitationService
	logger            logger.Logger
}

func NewCreateInvitationUseCase(svc *service.InvitationService, l logger.Logger) *CreateInvitationUseCase {
	return &CreateInvitationUseCase{
		invitationService: svc,
		logger:            l,
	}
}

func (uc *CreateInvitationUseCase) Execute(ctx context.Context, req *invitation.CreateInvitationRequest, nodeID string) (*invitation.CreateInvitationResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	projectID, err := utils.ParseID(req.ProjectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	inviterAccountID, err := utils.ParseID(req.InviterAccountID, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid inviter account ID format", "INVALID_INVITER_ID", err)
	}

	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	inv, err := uc.invitationService.CreateInvitation(ctx, projectID, inviterAccountID, req.InviteeShortID, string(entity.RoleMember), nodeUUID)
	if err != nil {
		return nil, err
	}

	response := &invitation.CreateInvitationResponse{
		InvitationID:     utils.ShortUUIDWithPrefix(inv.ID, entity.InvitationIDPrefix),
		ProjectID:        utils.ShortUUIDWithPrefix(inv.Project.ID, entity.ProjectIDPrefix),
		ProjectName:      inv.Project.Name,
		InviterAccountID: utils.ShortUUIDWithPrefix(inv.Inviter.ID, "acc"),
		InviterName:      inv.Inviter.Username,
		InviteeAccountID: utils.ShortUUIDWithPrefix(inv.Invitee.ID, "acc"),
		InviteeShortID:   utils.ShortUUIDWithPrefix(inv.Invitee.ID, "acc"),
		InviteeName:      inv.Invitee.Username,
		Role:             inv.Role,
		Status:           inv.Status,
		CreatedAt:        inv.CreatedAt,
		ExpiresAt:        inv.ExpiresAt,
	}

	return response, nil
}
