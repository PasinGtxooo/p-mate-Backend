package routes

import (
	accUC "github.com/FrostBitzX/smart-task-ai/internal/application/account/usecase"
	accDomain "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	repo "github.com/FrostBitzX/smart-task-ai/internal/infrastructure/persistence"
	accHandler "github.com/FrostBitzX/smart-task-ai/internal/infrastructure/rest"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func RegisterPublicRoutes(app fiber.Router, db *gorm.DB, log logger.Logger) {
	api := app.Group("/api")

	accountRepository := repo.NewAccountRepository(db)
	accountService := accDomain.NewAccountService(accountRepository)
	accountSignUpUC := accUC.NewCreateAccountUseCase(accountService, log)
	accountLoginUC := accUC.NewLoginUseCase(accountService, log)
	listAccountUC := accUC.NewListAccountUseCase(accountService, log)
	accountHandler := accHandler.NewAccountHandler(accountSignUpUC, listAccountUC, accountLoginUC, log)

	api.Post("/signup", accountHandler.CreateAccount)
	api.Post("/login", accountHandler.Login)
	api.Get("/accounts", accountHandler.ListAccounts)
}
