//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../../mocks/account_repository.go -package=mocks
package accounts

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
)

type AccountRepository interface {
	CreateAccount(ctx context.Context, acc *entity.Account) error
	ExistsAccount(ctx context.Context, username, email string) (bool, error)
	GetByUsername(ctx context.Context, username string) (*entity.Account, error)
	GetByEmail(ctx context.Context, email string) (*entity.Account, error)
	GetAccount(ctx context.Context, id string) (*entity.Account, error)
	ListAccounts(ctx context.Context, limit, offset int) ([]*entity.Account, int, error)
	ListByIDs(ctx context.Context, ids []string) ([]*entity.Account, error)
}
