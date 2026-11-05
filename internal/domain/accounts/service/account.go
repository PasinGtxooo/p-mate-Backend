package service

import (
	"context"
	"os"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/account"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
)

type AccountService struct {
	repo accounts.AccountRepository
}

func NewAccountService(repo accounts.AccountRepository) *AccountService {
	return &AccountService{repo: repo}
}

func (s *AccountService) CreateAccount(ctx context.Context, req *account.CreateAccountRequest) (*entity.Account, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// Check if username or email already exists
	exists, err := s.repo.ExistsAccount(ctx, req.Username, req.Email)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to check account existence", "CHECK_ACCOUNT_EXISTS_ERROR", err)
	}
	if exists {
		return nil, apperror.NewBadRequestError("username or email already exists", "USERNAME_OR_EMAIL_EXISTS", nil)
	}

	if req.Password != req.ConfirmPassword {
		return nil, apperror.NewBadRequestError("password and confirm password does not match", "PASSWORD_DOES_NOT_MATCH_ERROR", nil)
	}

	// hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to hash password", "HASH_PASSWORD_ERROR", err)
	}

	// create domain entity
	now := time.Now()
	acc := &entity.Account{
		ID:        uuid.New(),
		NodeID:    uuid.New(),
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		State:     "active",
		CreatedAt: now,
		UpdatedAt: now,
	}

	// persist account to database
	if err := s.repo.CreateAccount(ctx, acc); err != nil {
		return nil, apperror.NewInternalServerError("failed to create account", "CREATE_ACCOUNT_ERROR", err)
	}

	return acc, nil
}

func (s *AccountService) ListAccounts(ctx context.Context, limit, offset int) ([]*entity.Account, int, error) {
	accounts, total, err := s.repo.ListAccounts(ctx, limit, offset)
	if err != nil {
		return nil, 0, apperror.NewInternalServerError("failed to list accounts", "LIST_ACCOUNTS_ERROR", err)
	}
	return accounts, total, nil
}

func (s *AccountService) Login(ctx context.Context, req *account.LoginRequest) (string, error) {
	acc, err := s.repo.GetByUsername(ctx, req.Username)
	if err != nil {
		return "", apperror.NewBadRequestError("user does not exist", "LOGIN_ERROR", nil)
	}

	// compare password
	if err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte(req.Password)); err != nil {
		return "", apperror.NewBadRequestError("invalid username or password", "LOGIN_ERROR", nil)
	}

	expirationTime := time.Now().Add(time.Hour * 72).Unix()

	claims := jwt.MapClaims{
		"AccountId": acc.ID.String(),
		"NodeId":    acc.NodeID.String(),
		"Email":     acc.Email,
		"Username":  acc.Username,
		"Exp":       expirationTime,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return "", apperror.NewInternalServerError(
			"JWT_SECRET environment variable not set",
			"JWT_SECRET_MISSING",
			nil,
		)
	}

	t, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", apperror.NewInternalServerError(
			"failed to sign jwt",
			"JWT_SIGN_ERROR",
			err,
		)
	}

	return t, nil
}
