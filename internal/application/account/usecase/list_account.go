package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/account"
	"github.com/FrostBitzX/smart-task-ai/internal/application/common"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type ListAccountUseCase struct {
	accountService *service.AccountService
	logger         logger.Logger
}

func NewListAccountUseCase(svc *service.AccountService, l logger.Logger) *ListAccountUseCase {
	return &ListAccountUseCase{
		accountService: svc,
		logger:         l,
	}
}

func (uc *ListAccountUseCase) Execute(ctx context.Context, req *account.ListAccountsRequest) (*account.ListAccountsResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// Set pagination
	limit, offset := common.ValidatePagination(req.Limit, req.Offset)

	// Get accounts from service
	accounts, total, err := uc.accountService.ListAccounts(ctx, limit, offset)
	if err != nil {
		return nil, err
	}

	// Convert entities to DTOs
	accountDTOs := make([]account.AccountDTO, len(accounts))
	for i, acc := range accounts {
		accountDTOs[i] = account.AccountDTO{
			ID:       utils.ShortUUIDWithPrefix(acc.ID, entity.AccountIDPrefix),
			Username: acc.Username,
			Email:    acc.Email,
			Status:   acc.State,
		}
	}

	// Calculate pagination info
	hasMore := common.CalculateHasMore(offset, limit, total)

	// Build response
	response := &account.ListAccountsResponse{
		Items: accountDTOs,
		Pagination: common.Pagination{
			Total:   total,
			Limit:   limit,
			Offset:  offset,
			HasMore: hasMore,
		},
	}

	return response, nil
}
