package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/profile"
	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type GetProfileUseCase struct {
	profileService *service.ProfileService
	logger         logger.Logger
}

func NewGetProfileUseCase(svc *service.ProfileService, l logger.Logger) *GetProfileUseCase {
	return &GetProfileUseCase{
		profileService: svc,
		logger:         l,
	}
}

func (uc *GetProfileUseCase) Execute(ctx context.Context, req *profile.GetProfileByAccountIDRequest, nodeID string) (*profile.GetProfileByAccountIDResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	prof, err := uc.profileService.CheckAndGetProfile(ctx, req.AccountID, nodeID)
	if err != nil {
		return nil, err
	}
	if prof == nil {
		return nil, apperror.NewNotFoundError("profile not found", "PROFILE_NOT_FOUND", nil)
	}

	// Convert UUID to string with prefix
	accountID := utils.ShortUUIDWithPrefix(prof.AccountID, accountEntity.AccountIDPrefix)

	res := &profile.GetProfileByAccountIDResponse{
		AccountID:  accountID,
		FirstName:  prof.FirstName,
		LastName:   prof.LastName,
		Nickname:   prof.Nickname,
		AvatarPath: prof.AvatarPath,
		State:      prof.State,
		Skills:     prof.Skills,
		CreatedAt:  prof.CreatedAt,
		UpdatedAt:  prof.UpdatedAt,
	}
	return res, nil
}
