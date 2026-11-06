//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../../mocks/profile_repository.go -package=mocks
package profiles

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
)

type ProfileRepository interface {
	CreateProfile(ctx context.Context, prof *entity.Profile) error
	GetProfile(ctx context.Context, accountID string, nodeID string) (*entity.Profile, error)
	UpdateProfile(ctx context.Context, prof *entity.Profile, nodeID string) error
	ListProfilesByAccountIDs(ctx context.Context, accountIDs []string, nodeID string) ([]*entity.Profile, error)
	UpdateProfileSkills(ctx context.Context, accountID string, nodeID string, skills []entity.ProfileSkill) error
	UpdateCapacity(ctx context.Context, accountID string, nodeID string, capacityMDPerYear float64) error
}
