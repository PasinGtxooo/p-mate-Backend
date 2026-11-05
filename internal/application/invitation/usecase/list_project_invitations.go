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

type ListProjectInvitationsUseCase struct {
	invitationService *service.InvitationService
	logger            logger.Logger
}

func NewListProjectInvitationsUseCase(svc *service.InvitationService, l logger.Logger) *ListProjectInvitationsUseCase {
	return &ListProjectInvitationsUseCase{
		invitationService: svc,
		logger:            l,
	}
}

func (uc *ListProjectInvitationsUseCase) Execute(ctx context.Context, req *invitation.ListProjectInvitationsRequest) (*invitation.ListProjectInvitationsResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	projectID, err := utils.ParseID(req.ProjectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	accountID, err := utils.ParseID(req.AccountID, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	invitations, err := uc.invitationService.ListProjectInvitations(ctx, projectID, accountID)
	if err != nil {
		return nil, err
	}

	if len(invitations) == 0 {
		return &invitation.ListProjectInvitationsResponse{
			Invitations: []invitation.InvitationResponse{},
			Total:       0,
		}, nil
	}

	responses := make([]invitation.InvitationResponse, 0, len(invitations))
	for _, inv := range invitations {
		response := invitation.InvitationResponse{
			InvitationID:     utils.ShortUUIDWithPrefix(inv.ID, entity.InvitationIDPrefix),
			ProjectID:        utils.ShortUUIDWithPrefix(inv.ProjectID, entity.ProjectIDPrefix),
			ProjectName:      inv.Project.Name,
			InviterAccountID: utils.ShortUUIDWithPrefix(inv.InviterAccountID, "acc"),
			InviterName:      inv.Inviter.Username,
			InviteeAccountID: utils.ShortUUIDWithPrefix(inv.InviteeAccountID, "acc"),
			InviteeShortID:   utils.ShortUUIDWithPrefix(inv.Invitee.ID, "acc"),
			InviteeName:      inv.Invitee.Username,
			Role:             inv.Role,
			Status:           inv.Status,
			CreatedAt:        inv.CreatedAt,
			ExpiresAt:        inv.ExpiresAt,
			RespondedAt:      inv.RespondedAt,
		}
		responses = append(responses, response)
	}

	return &invitation.ListProjectInvitationsResponse{
		Invitations: responses,
		Total:       len(responses),
	}, nil
}
