package persistence

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
	"gorm.io/gorm"
)

type profileRepository struct {
	db *gorm.DB
}

func NewProfileRepository(db *gorm.DB) profiles.ProfileRepository {
	return &profileRepository{db: db}
}

func (r *profileRepository) CreateProfile(ctx context.Context, prof *entity.Profile) error {
	return r.db.WithContext(ctx).Create(prof).Error
}

func (r *profileRepository) GetProfile(ctx context.Context, accountID string, nodeID string) (*entity.Profile, error) {
	var profile entity.Profile
	err := r.db.WithContext(ctx).
		Where("account_id = ? AND node_id = ?", accountID, nodeID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) CheckAndGetProfile(ctx context.Context, accountID string) (*entity.Profile, error) {
	var profile entity.Profile
	err := r.db.WithContext(ctx).
		Select("id, account_id, first_name, last_name, nickname, avatar_path, skills, state, created_at, updated_at").
		Where("account_id = ?", accountID).
		First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, prof *entity.Profile, nodeID string) error {

	return r.db.WithContext(ctx).
		Model(&entity.Profile{}).
		Select("*").
		Where("account_id = ? AND node_id = ?", prof.AccountID, nodeID).
		Updates(prof).Error
}

func (r *profileRepository) ListProfilesByAccountIDs(ctx context.Context, accountIDs []string, nodeID string) ([]*entity.Profile, error) {
	var profiles []*entity.Profile
	if len(accountIDs) == 0 {
		return profiles, nil
	}
	err := r.db.WithContext(ctx).
		Where("account_id IN ? AND node_id = ?", accountIDs, nodeID).
		Find(&profiles).Error
	if err != nil {
		return nil, err
	}
	return profiles, nil
}

func (r *profileRepository) UpdateProfileSkills(ctx context.Context, accountID string, nodeID string, skills []entity.ProfileSkill) error {
	return r.db.WithContext(ctx).
		Model(&entity.Profile{}).
		Where("account_id = ? AND node_id = ?", accountID, nodeID).
		Update("skills", skills).Error
}

func (r *profileRepository) UpdateCapacity(ctx context.Context, accountID string, nodeID string, capacityMDPerYear float64) error {
	return r.db.WithContext(ctx).
		Model(&entity.Profile{}).
		Where("account_id = ? AND node_id = ?", accountID, nodeID).
		Update("capacity_md_per_year", capacityMDPerYear).Error
}
