package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion"
	taskEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type SuggestTeammatesUseCase struct {
	service *suggestion.SuggestionService
	logger  logger.Logger
}

func NewSuggestTeammatesUseCase(s *suggestion.SuggestionService, l logger.Logger) *SuggestTeammatesUseCase {
	return &SuggestTeammatesUseCase{service: s, logger: l}
}

func (uc *SuggestTeammatesUseCase) Execute(ctx context.Context, taskIDStr, nodeID, accountIDStr string) (*suggestion.SuggestTeammateResponse, error) {
	taskID, err := utils.ParseID(taskIDStr, taskEntity.TaskIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid task ID format", "INVALID_TASK_ID", err)
	}
	accountID, err := utils.ParseID(accountIDStr, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}
	return uc.service.SuggestTeammatesForTask(ctx, taskID, nodeID, accountID)
}
