package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/account"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type CreateAccountUseCase struct {
	accountService *service.AccountService
	logger         logger.Logger
}

func NewCreateAccountUseCase(svc *service.AccountService, l logger.Logger) *CreateAccountUseCase {
	return &CreateAccountUseCase{
		accountService: svc,
		logger:         l,
	}
}

func (uc *CreateAccountUseCase) Execute(ctx context.Context, req *account.CreateAccountRequest) (*account.CreateAccountResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	acc, err := uc.accountService.CreateAccount(ctx, req)

	if err != nil {
		return nil, err
	}

	// Convert UUID to string with prefix
	accountID := utils.ShortUUIDWithPrefix(acc.ID, entity.AccountIDPrefix)

	res := &account.CreateAccountResponse{
		AccountID: accountID,
	}
	return res, nil
}
