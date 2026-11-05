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

type ListProjectMembersUseCase struct {
	projectService *service.ProjectService
	logger         logger.Logger
}

func NewListProjectMembersUseCase(svc *service.ProjectService, l logger.Logger) *ListProjectMembersUseCase {
	return &ListProjectMembersUseCase{
		projectService: svc,
		logger:         l,
	}
}

func (uc *ListProjectMembersUseCase) Execute(ctx context.Context, projectIDStr string, nodeID string, accountIDStr string) (*project.ListProjectMembersResponse, error) {
	projectID, err := utils.ParseID(projectIDStr, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	accountID, err := utils.ParseID(accountIDStr, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	members, err := uc.projectService.ListProjectMembers(ctx, projectID, nodeID, accountID)
	if err != nil {
		return nil, err
	}

	memberResponses := make([]project.ProjectMemberResponse, len(members))
	for i, member := range members {
		memberResponses[i] = project.ProjectMemberResponse{
			AccountID: utils.ShortUUIDWithPrefix(member.AccountID, accountEntity.AccountIDPrefix),
			Role:      member.Role.String(),
			CreatedAt: member.CreatedAt,
		}
	}

	return &project.ListProjectMembersResponse{
		Members: memberResponses,
	}, nil
}
