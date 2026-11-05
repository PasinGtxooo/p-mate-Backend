package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type RemoveMemberUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewRemoveMemberUseCase(svc *service.ProjectService, l logger.Logger) *RemoveMemberUseCase {
	return &RemoveMemberUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *RemoveMemberUseCase) Execute(ctx context.Context, req *project.RemoveMemberRequest, nodeID string) (*project.RemoveMemberResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	projectID, err := utils.ParseID(req.ProjectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	accountID, err := utils.ParseID(req.AccountID, accountEntity.AccountIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	err = uc.projectService.RemoveMember(ctx, projectID, accountID, nodeID)
	if err != nil {
		return nil, err
	}

	return &project.RemoveMemberResponse{
		ProjectID: req.ProjectID,
		AccountID: req.AccountID,
	}, nil
}
