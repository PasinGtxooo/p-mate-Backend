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

type AcceptInvitationUseCase struct {
	invitationService *service.InvitationService
	logger            logger.Logger
}

func NewAcceptInvitationUseCase(svc *service.InvitationService, l logger.Logger) *AcceptInvitationUseCase {
	return &AcceptInvitationUseCase{
		invitationService: svc,
		logger:            l,
	}
}

func (uc *AcceptInvitationUseCase) Execute(ctx context.Context, req *invitation.AcceptInvitationRequest, nodeID string) (*invitation.AcceptInvitationResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	projectID, err := utils.ParseID(req.ProjectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	inviteeAccountID, err := utils.ParseID(req.InviteeAccountID, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid invitee account ID format", "INVALID_INVITEE_ID", err)
	}

	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	err = uc.invitationService.AcceptInvitation(ctx, projectID, inviteeAccountID, nodeUUID)
	if err != nil {
		return nil, err
	}

	return &invitation.AcceptInvitationResponse{
		Message: "Invitation accepted successfully",
	}, nil
}
