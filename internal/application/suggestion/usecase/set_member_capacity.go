package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type SetMemberCapacityUseCase struct {
	service *suggestion.SuggestionService
	logger  logger.Logger
}

func NewSetMemberCapacityUseCase(s *suggestion.SuggestionService, l logger.Logger) *SetMemberCapacityUseCase {
	return &SetMemberCapacityUseCase{service: s, logger: l}
}

func (uc *SetMemberCapacityUseCase) Execute(ctx context.Context, projectIDStr, targetAccountIDStr, nodeID, callerAccountIDStr string, req *suggestion.SetMemberCapacityRequest) error {
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
	return uc.service.SetMemberCapacity(ctx, projectID, targetAccountID, nodeID, callerAccountID, req.CapacityMDPerYear)
}
