package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/profile"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/FrostBitzX/smart-task-ai/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
)

type ProfileService struct {
	repo        profiles.ProfileRepository
	accountRepo accounts.AccountRepository
}

func NewProfileService(repo profiles.ProfileRepository, accountRepo accounts.AccountRepository) *ProfileService {
	return &ProfileService{
		repo:        repo,
		accountRepo: accountRepo,
	}
}

func (s *ProfileService) CheckAndGetProfile(ctx context.Context, accountID string, nodeID string) (*entity.Profile, error) {
	prof, err := s.repo.GetProfile(ctx, accountID, nodeID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, apperror.NewInternalServerError("failed to check and get profile", "CHECK_AND_GET_PROFILE_ERROR", nil)
	}

	return prof, nil
}

func (s *ProfileService) CreateProfile(ctx context.Context, req *profile.CreateProfileRequest, nodeID string) (*entity.Profile, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", nil)
	}

	// Check if account exists
	_, err = s.accountRepo.GetAccount(ctx, req.AccountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewBadRequestError(
				"account ID does not exist",
				"ACCOUNT_NOT_FOUND",
				nil,
			)
		}
		return nil, apperror.NewInternalServerError("failed to check account existence", "CHECK_ACCOUNT_ERROR", nil)
	}

	// Check if profile already exists
	exists, err := s.CheckAndGetProfile(ctx, req.AccountID, nodeID)
	if err != nil {
		return nil, err
	}
	if exists != nil {
		return nil, apperror.NewBadRequestError("profile already exists", "PROFILE_ALREADY_EXISTS", nil)
	}

	// create domain entity
	now := time.Now()
	prof := &entity.Profile{
		ID:         uuid.New(),
		NodeID:     nodeUUID,
		AccountID:  uuid.MustParse(req.AccountID),
		FirstName:  req.FirstName,
		LastName:   req.LastName,
		Nickname:   req.Nickname,
		AvatarPath: req.AvatarPath,
		Skills:     req.Skills,
		State:      "active",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// persist profile to database
	err = s.repo.CreateProfile(ctx, prof)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to create profile", "CREATE_PROFILE_ERROR", nil)
	}

	return prof, nil
}

func (s *ProfileService) UpdateProfile(ctx context.Context, req *profile.UpdateProfileRequest, nodeID string) (*entity.Profile, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", nil)
	}

	// Check if account exists
	_, err = s.accountRepo.GetAccount(ctx, req.AccountID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperror.NewBadRequestError(
				fmt.Sprintf("account with ID %s does not exist", req.AccountID),
				"ACCOUNT_NOT_FOUND",
				nil,
			)
		}
		return nil, apperror.NewInternalServerError("failed to check account existence", "CHECK_ACCOUNT_ERROR", nil)
	}

	// Check if profile exists
	exists, err := s.CheckAndGetProfile(ctx, req.AccountID, nodeID)
	if err != nil {
		return nil, err
	}
	if exists == nil {
		return nil, apperror.NewBadRequestError("profile not found", "PROFILE_NOT_FOUND", nil)
	}

	// create domain entity
	now := time.Now()
	prof := &entity.Profile{
		ID:        exists.ID,
		NodeID:    nodeUUID,
		AccountID: uuid.MustParse(req.AccountID),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		State:     "active",
		CreatedAt: exists.CreatedAt,
		UpdatedAt: now,
	}

	// Nullable fields
	utils.UpdateNullableString(&prof.Nickname, req.Nickname)
	utils.UpdateNullableString(&prof.AvatarPath, req.AvatarPath)
	if req.Skills != nil {
		prof.Skills = req.Skills
	}

	// persist account to database
	err = s.repo.UpdateProfile(ctx, prof, nodeID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to update profile", "UPDATE_PROFILE_ERROR", nil)
	}

	return prof, nil
}
