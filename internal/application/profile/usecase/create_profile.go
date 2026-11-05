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

type CreateProfileUseCase struct {
	profileService *service.ProfileService
	logger         logger.Logger
}

func NewCreateProfileUseCase(svc *service.ProfileService, l logger.Logger) *CreateProfileUseCase {
	return &CreateProfileUseCase{
		profileService: svc,
		logger:         l,
	}
}

func (uc *CreateProfileUseCase) Execute(ctx context.Context, req *profile.CreateProfileRequest, nodeID string) (*profile.CreateProfileResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	prof, err := uc.profileService.CreateProfile(ctx, req, nodeID)
	if err != nil {
		return nil, err
	}

	// Convert UUID to string with prefix
	accountID := utils.ShortUUIDWithPrefix(prof.AccountID, accountEntity.AccountIDPrefix)
	profileID := utils.ShortUUIDWithPrefix(prof.ID, entity.ProfileIDPrefix)

	res := &profile.CreateProfileResponse{
		AccountID: accountID,
		ProfileID: profileID,
	}
	return res, nil
}
