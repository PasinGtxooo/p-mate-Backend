package service

import (
	"context"
	"errors"
	"testing"

	"github.com/FrostBitzX/smart-task-ai/internal/application/profile"
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	entity "github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"gorm.io/gorm"
)

func TestProfileService_CreateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProfileRepository(ctrl)
	mockAccountRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewProfileService(mockRepo, mockAccountRepo)
	ctx := context.Background()
	accountID := uuid.New().String()

	tests := []struct {
		name          string
		request       *profile.CreateProfileRequest
		setupMock     func()
		expectedError string
		expectNil     bool
	}{
		{
			name: "success - creates profile with all fields",
			request: &profile.CreateProfileRequest{
				AccountID:  accountID,
				FirstName:  "John",
				LastName:   "Doe",
				Nickname:   strPtr("johnd"),
				AvatarPath: strPtr("/avatars/john.png"),
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound).
					Times(1)
				mockRepo.EXPECT().
					CreateProfile(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, prof *entity.Profile) error {
						assert.Equal(t, "John", prof.FirstName)
						assert.Equal(t, "Doe", prof.LastName)
						assert.NotNil(t, prof.Nickname)
						assert.Equal(t, "johnd", *prof.Nickname)
						assert.NotNil(t, prof.AvatarPath)
						assert.Equal(t, "/avatars/john.png", *prof.AvatarPath)
						assert.Equal(t, "active", prof.State)
						return nil
					}).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name: "success - creates profile with minimal fields",
			request: &profile.CreateProfileRequest{
				AccountID: accountID,
				FirstName: "Jane",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound).
					Times(1)
				mockRepo.EXPECT().
					CreateProfile(ctx, gomock.Any()).
					Return(nil).
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
			name: "error - profile already exists",
			request: &profile.CreateProfileRequest{
				AccountID: accountID,
				FirstName: "John",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				existingProfile := &entity.Profile{
					ID:        uuid.New(),
					AccountID: uuid.MustParse(accountID),
				}
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(existingProfile, nil).
					Times(1)
			},
			expectedError: "profile already exists",
			expectNil:     true,
		},
		{
			name: "error - repository check fails",
			request: &profile.CreateProfileRequest{
				AccountID: accountID,
				FirstName: "John",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				// The implementation calls CheckAndGetProfile which fails
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to check and get profile",
			expectNil:     true,
		},
		{
			name: "error - repository create fails with duplicate key",
			request: &profile.CreateProfileRequest{
				AccountID: accountID,
				FirstName: "John",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound).
					Times(1)
				mockRepo.EXPECT().
					CreateProfile(ctx, gomock.Any()).
					Return(gorm.ErrDuplicatedKey).
					Times(1)
			},
			expectedError: "failed to create profile",
			expectNil:     true,
		},
		{
			name: "error - repository create fails with generic error",
			request: &profile.CreateProfileRequest{
				AccountID: accountID,
				FirstName: "John",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound).
					Times(1)
				mockRepo.EXPECT().
					CreateProfile(ctx, gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to create profile",
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.CreateProfile(ctx, tt.request, "550e8400-e29b-41d4-a716-446655440000")

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
					assert.Equal(t, tt.request.FirstName, res.FirstName)
				}
			}
		})
	}
}

func TestProfileService_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProfileRepository(ctrl)
	mockAccountRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewProfileService(mockRepo, mockAccountRepo)
	ctx := context.Background()
	accountID := uuid.New().String()
	profileID := uuid.New()

	tests := []struct {
		name          string
		request       *profile.UpdateProfileRequest
		setupMock     func()
		expectedError string
		expectNil     bool
	}{
		{
			name: "success - updates all fields",
			request: &profile.UpdateProfileRequest{
				AccountID:  accountID,
				FirstName:  "Updated",
				LastName:   "Name",
				Nickname:   strPtr("updated"),
				AvatarPath: strPtr("/new/avatar.png"),
			},
			setupMock: func() {
				existingProfile := &entity.Profile{
					ID:        profileID,
					AccountID: uuid.MustParse(accountID),
					FirstName: "Old",
					LastName:  "Name",
				}
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(existingProfile, nil).
					Times(1)
				mockRepo.EXPECT().
					UpdateProfile(ctx, gomock.Any(), gomock.Any()).
					DoAndReturn(func(_ context.Context, prof *entity.Profile, _ string) error {
						assert.Equal(t, profileID, prof.ID)
						assert.Equal(t, "Updated", prof.FirstName)
						assert.Equal(t, "Name", prof.LastName)
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
			name: "error - profile not found",
			request: &profile.UpdateProfileRequest{
				AccountID: accountID,
				FirstName: "Updated",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "profile not found",
			expectNil:     true,
		},
		{
			name: "error - repository get fails",
			request: &profile.UpdateProfileRequest{
				AccountID: accountID,
				FirstName: "Updated",
			},
			setupMock: func() {
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to check and get profile",
			expectNil:     true,
		},
		{
			name: "error - repository update fails",
			request: &profile.UpdateProfileRequest{
				AccountID: accountID,
				FirstName: "Updated",
			},
			setupMock: func() {
				existingProfile := &entity.Profile{
					ID:        profileID,
					AccountID: uuid.MustParse(accountID),
				}
				mockAccountRepo.EXPECT().
					GetAccount(ctx, accountID).
					Return(&accountEntity.Account{ID: uuid.MustParse(accountID)}, nil).
					Times(1)
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(existingProfile, nil).
					Times(1)
				mockRepo.EXPECT().
					UpdateProfile(ctx, gomock.Any(), gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to update profile",
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.UpdateProfile(ctx, tt.request, "550e8400-e29b-41d4-a716-446655440000")

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
					assert.Equal(t, tt.request.FirstName, res.FirstName)
				}
			}
		})
	}
}

func TestProfileService_CheckAndGetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProfileRepository(ctrl)
	mockAccountRepo := mocks.NewMockAccountRepository(ctrl)
	svc := NewProfileService(mockRepo, mockAccountRepo)
	ctx := context.Background()
	accountID := uuid.New().String()

	tests := []struct {
		name          string
		accountID     string
		setupMock     func()
		expectedError string
		expectNil     bool
	}{
		{
			name:      "success - returns profile",
			accountID: accountID,
			setupMock: func() {
				prof := &entity.Profile{
					ID:        uuid.New(),
					AccountID: uuid.MustParse(accountID),
					FirstName: "John",
					LastName:  "Doe",
				}
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(prof, nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name:      "success - returns nil when not found",
			accountID: accountID,
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, gorm.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "",
			expectNil:     true,
		},
		{
			name:      "error - repository fails",
			accountID: accountID,
			setupMock: func() {
				mockRepo.EXPECT().GetProfile(ctx, accountID, gomock.Any()).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to check and get profile",
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.CheckAndGetProfile(ctx, tt.accountID, "550e8400-e29b-41d4-a716-446655440000")

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
			}
		})
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}
