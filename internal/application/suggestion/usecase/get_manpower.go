package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type GetManpowerUseCase struct {
	service *suggestion.SuggestionService
	logger  logger.Logger
}

func NewGetManpowerUseCase(s *suggestion.SuggestionService, l logger.Logger) *GetManpowerUseCase {
	return &GetManpowerUseCase{service: s, logger: l}
}

func (uc *GetManpowerUseCase) Execute(ctx context.Context, projectIDStr, nodeID, accountIDStr string) (*suggestion.ManpowerResponse, error) {
	projectID, err := utils.ParseID(projectIDStr, projectEntity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}
	accountID, err := utils.ParseID(accountIDStr, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}
	return uc.service.GetManpower(ctx, projectID, nodeID, accountID)
}
