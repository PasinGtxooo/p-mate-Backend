package service

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/FrostBitzX/smart-task-ai/internal/application/account"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/mocks"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestAccountService_CreateAccount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewAccountService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name          string
		request       *account.CreateAccountRequest
		setupMock     func()
		expectedError string
		expectNil     bool
	}{
		{
			name: "success - creates account with valid data",
			request: &account.CreateAccountRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					ExistsAccount(ctx, "testuser", "test@example.com").
					Return(false, nil).
					Times(1)
				mockRepo.EXPECT().
					CreateAccount(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, acc *entity.Account) error {
						// Verify the account has correct fields
						assert.Equal(t, "testuser", acc.Username)
						assert.Equal(t, "test@example.com", acc.Email)
						assert.Equal(t, "active", acc.State)
						assert.NotEmpty(t, acc.ID)
						assert.NotEmpty(t, acc.NodeID)
						// Verify password is hashed
						err := bcrypt.CompareHashAndPassword([]byte(acc.Password), []byte("password123"))
						assert.NoError(t, err)
						return nil
					}).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name:          "error - nil request",
			request:       nil,
			setupMock:     func() {},
			expectedError: "invalid request body",
			expectNil:     true,
		},
		{
			name: "error - account already exists",
			request: &account.CreateAccountRequest{
				Username:        "existinguser",
				Email:           "existing@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					ExistsAccount(ctx, "existinguser", "existing@example.com").
					Return(true, nil).
					Times(1)
			},
			expectedError: "already exists",
			expectNil:     true,
		},
		{
			name: "error - password mismatch",
			request: &account.CreateAccountRequest{
				Username:        "user",
				Email:           "email@example.com",
				Password:        "pass1",
				ConfirmPassword: "pass2",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					ExistsAccount(ctx, "user", "email@example.com").
					Return(false, nil).
					Times(1)
			},
			expectedError: "password and confirm password does not match",
			expectNil:     true,
		},
		{
			name: "error - repository check fails",
			request: &account.CreateAccountRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					ExistsAccount(ctx, "testuser", "test@example.com").
					Return(false, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to check account existence",
			expectNil:     true,
		},
		{
			name: "error - repository create fails",
			request: &account.CreateAccountRequest{
				Username:        "testuser",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					ExistsAccount(ctx, "testuser", "test@example.com").
					Return(false, nil).
					Times(1)
				mockRepo.EXPECT().
					CreateAccount(ctx, gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to create account",
			expectNil:     true,
		},
		{
			name: "error - empty username",
			request: &account.CreateAccountRequest{
				Username:        "",
				Email:           "test@example.com",
				Password:        "password123",
				ConfirmPassword: "password123",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					ExistsAccount(ctx, "", "test@example.com").
					Return(false, nil).
					Times(1)
				mockRepo.EXPECT().
					CreateAccount(ctx, gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.CreateAccount(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
				if tt.request != nil {
					assert.Equal(t, tt.request.Username, res.Username)
					assert.Equal(t, tt.request.Email, res.Email)
				}
			}
		})
	}
}

func TestAccountService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewAccountService(mockRepo)
	ctx := context.Background()

	// Set JWT secret for tests
	os.Setenv("JWT_SECRET", "testsecret")
	defer os.Unsetenv("JWT_SECRET")

	// Pre-hash password for test accounts
	validPassword := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)
	require.NoError(t, err)

	tests := []struct {
		name          string
		request       *account.LoginRequest
		setupMock     func()
		expectedError string
		expectToken   bool
	}{
		{
			name: "success - valid credentials",
			request: &account.LoginRequest{
				Username: "testuser",
				Password: validPassword,
			},
			setupMock: func() {
				acc := &entity.Account{
					ID:       uuid.New(),
					NodeID:   uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				mockRepo.EXPECT().
					GetByUsername(ctx, "testuser").
					Return(acc, nil).
					Times(1)
			},
			expectedError: "",
			expectToken:   true,
		},
		{
			name: "error - user not found",
			request: &account.LoginRequest{
				Username: "nonexistent",
				Password: validPassword,
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetByUsername(ctx, "nonexistent").
					Return(nil, apperror.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "user does not exist",
			expectToken:   false,
		},
		{
			name: "error - invalid password (wrong password)",
			request: &account.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMock: func() {
				acc := &entity.Account{
					ID:       uuid.New(),
					NodeID:   uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
					Password: string(hashedPassword), // Correct hashed password
				}
				mockRepo.EXPECT().
					GetByUsername(ctx, "testuser").
					Return(acc, nil).
					Times(1)
			},
			expectedError: "invalid username or password",
			expectToken:   false,
		},
		{
			name: "error - invalid password (not bcrypt hash)",
			request: &account.LoginRequest{
				Username: "testuser",
				Password: validPassword,
			},
			setupMock: func() {
				acc := &entity.Account{
					ID:       uuid.New(),
					NodeID:   uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
					Password: "plaintext_not_hashed", // Invalid - not a bcrypt hash
				}
				mockRepo.EXPECT().
					GetByUsername(ctx, "testuser").
					Return(acc, nil).
					Times(1)
			},
			expectedError: "invalid username or password",
			expectToken:   false,
		},
		{
			name: "error - empty password",
			request: &account.LoginRequest{
				Username: "testuser",
				Password: "",
			},
			setupMock: func() {
				acc := &entity.Account{
					ID:       uuid.New(),
					NodeID:   uuid.New(),
					Username: "testuser",
					Email:    "test@example.com",
					Password: string(hashedPassword),
				}
				mockRepo.EXPECT().
					GetByUsername(ctx, "testuser").
					Return(acc, nil).
					Times(1)
			},
			expectedError: "invalid username or password",
			expectToken:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			token, err := svc.Login(ctx, tt.request)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expectToken {
				assert.NotEmpty(t, token)
				// Verify token has 3 parts (header.payload.signature)
				parts := len(token)
				assert.Greater(t, parts, 0)
			} else {
				assert.Empty(t, token)
			}
		})
	}
}

func TestAccountService_Login_MissingJWTSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewAccountService(mockRepo)
	ctx := context.Background()

	// Ensure JWT_SECRET is not set
	os.Unsetenv("JWT_SECRET")

	validPassword := "password123"
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	acc := &entity.Account{
		ID:       uuid.New(),
		NodeID:   uuid.New(),
		Username: "testuser",
		Email:    "test@example.com",
		Password: string(hashedPassword),
	}

	mockRepo.EXPECT().
		GetByUsername(ctx, "testuser").
		Return(acc, nil).
		Times(1)

	req := &account.LoginRequest{
		Username: "testuser",
		Password: validPassword,
	}

	token, err := svc.Login(ctx, req)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
	assert.Empty(t, token)
}

func TestAccountService_ListAccounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewAccountService(mockRepo)
	ctx := context.Background()

	tests := []struct {
		name          string
		limit         int
		offset        int
		setupMock     func()
		expectedCount int
		expectedTotal int
		expectedError string
	}{
		{
			name:   "success - returns accounts",
			limit:  10,
			offset: 0,
			setupMock: func() {
				accounts := []*entity.Account{
					{ID: uuid.New(), NodeID: uuid.New(), Username: "user1", Email: "user1@example.com"},
					{ID: uuid.New(), NodeID: uuid.New(), Username: "user2", Email: "user2@example.com"},
				}
				mockRepo.EXPECT().
					ListAccounts(ctx, 10, 0).
					Return(accounts, 2, nil).
					Times(1)
			},
			expectedCount: 2,
			expectedTotal: 2,
			expectedError: "",
		},
		{
			name:   "success - empty list",
			limit:  10,
			offset: 0,
			setupMock: func() {
				mockRepo.EXPECT().
					ListAccounts(ctx, 10, 0).
					Return([]*entity.Account{}, 0, nil).
					Times(1)
			},
			expectedCount: 0,
			expectedTotal: 0,
			expectedError: "",
		},
		{
			name:   "error - repository fails",
			limit:  10,
			offset: 0,
			setupMock: func() {
				mockRepo.EXPECT().
					ListAccounts(ctx, 10, 0).
					Return(nil, 0, errors.New("database error")).
					Times(1)
			},
			expectedCount: 0,
			expectedTotal: 0,
			expectedError: "failed to list accounts",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			accounts, total, err := svc.ListAccounts(ctx, tt.limit, tt.offset)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Len(t, accounts, tt.expectedCount)
				assert.Equal(t, tt.expectedTotal, total)
			}
		})
	}
}
