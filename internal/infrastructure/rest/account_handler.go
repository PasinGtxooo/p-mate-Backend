package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/account"
	"github.com/FrostBitzX/smart-task-ai/internal/application/account/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/gofiber/fiber/v2"
)

type AccountHandler struct {
	CreateAccountUC *usecase.CreateAccountUseCase
	LoginUC         *usecase.LoginUseCase
	ListAccountUC   *usecase.ListAccountUseCase
	logger          logger.Logger
}

func NewAccountHandler(
	create *usecase.CreateAccountUseCase,
	listUC *usecase.ListAccountUseCase,
	login *usecase.LoginUseCase,
	l logger.Logger,
) *AccountHandler {
	return &AccountHandler{
		CreateAccountUC: create,
		ListAccountUC:   listUC,
		LoginUC:         login,
		logger:          l,
	}

}

func (h *AccountHandler) CreateAccount(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[account.CreateAccountRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	data, err := h.CreateAccountUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Account created successfully")
}

func (h *AccountHandler) ListAccounts(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateQuery[account.ListAccountsRequest](c)
	if err != nil {
		h.logger.Warn("Invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	data, err := h.ListAccountUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "List accounts successfully")
}

func (h *AccountHandler) Login(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[account.LoginRequest](c)
	if err != nil {
		h.logger.Warn("Failed to validate request", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	data, err := h.LoginUC.Execute(c.Context(), req)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Login successfully")
}
