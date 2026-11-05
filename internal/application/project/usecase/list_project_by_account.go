package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/common"
	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type ListProjectByAccountUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewListProjectByAccountUseCase(svc *service.ProjectService, l logger.Logger) *ListProjectByAccountUseCase {
	return &ListProjectByAccountUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *ListProjectByAccountUseCase) Execute(ctx context.Context, req *project.ListProjectRequest, nodeID string) (*project.ListProjectResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	accountID, err := utils.ParseID(req.AccountID, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Set pagination
	limit, offset := common.ValidatePagination(req.Limit, req.Offset)

	// Get projects from service
	projs, total, err := uc.projectService.ListProjectByAccountID(ctx, accountID, nodeID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert entities to DTOs
	items := make([]project.ProjectResponse, len(projs))
	for i, p := range projs {
		role := string(entity.RoleMember) // Member
		if p.OwnerID == accountID {
			role = string(entity.RoleOwner) // Owner
		}

		items[i] = project.ProjectResponse{
			ID:        utils.ShortUUIDWithPrefix(p.ID, entity.ProjectIDPrefix),
			Name:      p.Name,
			Role:      role,
			Config:    p.Config,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		}
	}

	// Calculate pagination info
	hasMore := common.CalculateHasMore(offset, limit, total)

	// Return response
	return &project.ListProjectResponse{
		Items: items,
		Pagination: common.Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}, nil
}
