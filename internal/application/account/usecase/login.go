package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/account"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type LoginUseCase struct {
	accountService *service.AccountService
	logger         logger.Logger
}

func NewLoginUseCase(accSvc *service.AccountService, l logger.Logger) *LoginUseCase {
	return &LoginUseCase{
		accountService: accSvc,
		logger:         l,
	}
}

func (u *LoginUseCase) Execute(ctx context.Context, req *account.LoginRequest) (*account.LoginResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	token, err := u.accountService.Login(ctx, req)

	if err != nil {
		return nil, err
	}

	res := &account.LoginResponse{
		Token: token,
	}

	return res, nil
}
