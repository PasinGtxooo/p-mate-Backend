package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
)

type UpdateProjectUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewUpdateProjectUseCase(svc *service.ProjectService, l logger.Logger) *UpdateProjectUseCase {
	return &UpdateProjectUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *UpdateProjectUseCase) Execute(ctx context.Context, req *project.UpdateProjectRequest, nodeID string) (*project.UpdateProjectResponse, error) {
	proj, err := uc.projectService.UpdateProject(ctx, req, nodeID)
	if err != nil {
		return nil, err
	}

	return &project.UpdateProjectResponse{
		ID:        utils.ShortUUIDWithPrefix(proj.ID, entity.ProjectIDPrefix),
		Name:      proj.Name,
		Config:    proj.Config,
		CreatedAt: proj.CreatedAt,
		UpdatedAt: proj.UpdatedAt,
	}, nil
}
