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

type DeleteProjectUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewDeleteProjectUseCase(svc *service.ProjectService, l logger.Logger) *DeleteProjectUseCase {
	return &DeleteProjectUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *DeleteProjectUseCase) Execute(ctx context.Context, id string, nodeID string) (*project.DeleteProjectResponse, error) {
	projectID, err := utils.ParseID(id, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	err = uc.projectService.DeleteProject(ctx, projectID, nodeID)
	if err != nil {
		return nil, err
	}

	return &project.DeleteProjectResponse{
		ProjectID: id,
	}, nil
}
