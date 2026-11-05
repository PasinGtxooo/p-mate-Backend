package usecase

import (
	"context"
	"errors"

	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts"
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type AddMemberUseCase struct {
	projectService *service.ProjectService
	accountRepo    accounts.AccountRepository
	logger         logger.Logger
}

func NewAddMemberUseCase(svc *service.ProjectService, accountRepo accounts.AccountRepository, l logger.Logger) *AddMemberUseCase {
	return &AddMemberUseCase{
		projectService: svc,
		accountRepo:    accountRepo,
		logger:         l,
	}
}

func (uc *AddMemberUseCase) Execute(ctx context.Context, req *project.AddMemberRequest, nodeID string) (*project.AddMemberResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	projectID, err := utils.ParseID(req.ProjectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	acc, err := uc.accountRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("user with this email address does not exist", "ACCOUNT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to look up account", "LOOKUP_ACCOUNT_ERROR", err)
	}

	role := string(entity.RoleMember) // Member

	member, err := uc.projectService.AddMember(ctx, projectID, acc.ID, role, nodeID)
	if err != nil {
		return nil, err
	}

	return &project.AddMemberResponse{
		ProjectID: utils.ShortUUIDWithPrefix(member.ProjectID, entity.ProjectIDPrefix),
		AccountID: utils.ShortUUIDWithPrefix(member.AccountID, accountEntity.AccountIDPrefix),
		Role:      member.Role.String(),
	}, nil
}
