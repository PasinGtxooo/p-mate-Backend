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

type CreateProjectUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewCreateProjectUseCase(svc *service.ProjectService, l logger.Logger) *CreateProjectUseCase {
	return &CreateProjectUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *CreateProjectUseCase) Execute(ctx context.Context, req *project.CreateProjectRequest, nodeID string) (*project.CreateProjectResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	prof, err := uc.projectService.CreateProject(ctx, req, nodeID)
	if err != nil {
		return nil, err
	}

	// Convert UUID to string with prefix
	projectID := utils.ShortUUIDWithPrefix(prof.ID, entity.ProjectIDPrefix)

	res := &project.CreateProjectResponse{
		ProjectID: projectID,
	}
	return res, nil
}
