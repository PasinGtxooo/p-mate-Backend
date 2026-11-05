package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/profile"
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type UpdateProfileUseCase struct {
	profileService *service.ProfileService
	logger         logger.Logger
}

func NewUpdateProfileUseCase(svc *service.ProfileService, l logger.Logger) *UpdateProfileUseCase {
	return &UpdateProfileUseCase{
		profileService: svc,
		logger:         l,
	}
}

func (uc *UpdateProfileUseCase) Execute(ctx context.Context, req *profile.UpdateProfileRequest, nodeID string) (*profile.UpdateProfileResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	prof, err := uc.profileService.UpdateProfile(ctx, req, nodeID)
	if err != nil {
		return nil, err
	}

	// Convert UUID to string with prefix
	accountID := utils.ShortUUIDWithPrefix(prof.AccountID, accountEntity.AccountIDPrefix)
	profileID := utils.ShortUUIDWithPrefix(prof.ID, entity.ProfileIDPrefix)

	res := &profile.UpdateProfileResponse{
		AccountID:  accountID,
		ProfileID:  profileID,
		FirstName:  prof.FirstName,
		LastName:   prof.LastName,
		Nickname:   prof.Nickname,
		AvatarPath: prof.AvatarPath,
	}
	return res, nil
}
