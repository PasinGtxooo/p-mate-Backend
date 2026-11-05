package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type SetMemberRoleUseCase struct {
	service *suggestion.SuggestionService
	logger  logger.Logger
}

func NewSetMemberRoleUseCase(s *suggestion.SuggestionService, l logger.Logger) *SetMemberRoleUseCase {
	return &SetMemberRoleUseCase{service: s, logger: l}
}

func (uc *SetMemberRoleUseCase) Execute(ctx context.Context, projectIDStr, targetAccountIDStr, nodeID, callerAccountIDStr string, req *suggestion.SetMemberRoleRequest) error {
	if req == nil {
		return apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}
	projectID, err := utils.ParseID(projectIDStr, projectEntity.ProjectIDPrefix)
	if err != nil {
		return apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}
	targetAccountID, err := utils.ParseID(targetAccountIDStr, "acc")
	if err != nil {
		return apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}
	callerAccountID, err := utils.ParseID(callerAccountIDStr, "acc")
	if err != nil {
		return apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}
	return uc.service.SetMemberRole(ctx, projectID, targetAccountID, nodeID, callerAccountID, req.Role)
}
