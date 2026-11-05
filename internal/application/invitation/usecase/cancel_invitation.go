package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/invitation"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type CancelInvitationUseCase struct {
	invitationService *service.InvitationService
	logger            logger.Logger
}

func NewCancelInvitationUseCase(svc *service.InvitationService, l logger.Logger) *CancelInvitationUseCase {
	return &CancelInvitationUseCase{
		invitationService: svc,
		logger:            l,
	}
}

func (uc *CancelInvitationUseCase) Execute(ctx context.Context, req *invitation.CancelInvitationRequest) (*invitation.CancelInvitationResponse, error) {
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

	cancellerID, err := utils.ParseID(req.CancellerID, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid canceller account ID format", "INVALID_CANCELLER_ID", err)
	}

	err = uc.invitationService.CancelInvitation(ctx, projectID, inviteeAccountID, cancellerID)
	if err != nil {
		return nil, err
	}

	return &invitation.CancelInvitationResponse{
		Message: "Invitation cancelled successfully",
	}, nil
}
