package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type GetProjectByIDUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewGetProjectByIDUseCase(svc *service.ProjectService, l logger.Logger) *GetProjectByIDUseCase {
	return &GetProjectByIDUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *GetProjectByIDUseCase) Execute(ctx context.Context, id string, nodeID string, accountIDStr string) (*project.ProjectResponse, error) {
	projectID, err := utils.ParseID(id, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	accountID, err := utils.ParseID(accountIDStr, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	proj, err := uc.projectService.GetProjectByID(ctx, projectID, nodeID, accountID)
	if err != nil {
		return nil, err
	}

	return &project.ProjectResponse{
		ID:        utils.ShortUUIDWithPrefix(proj.ID, entity.ProjectIDPrefix),
		Name:      proj.Name,
		Config:    proj.Config,
		CreatedAt: proj.CreatedAt,
		UpdatedAt: proj.UpdatedAt,
	}, nil
}
