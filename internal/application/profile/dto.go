package profile

import (
	"time"

	profileEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
)

type CreateProfileRequest struct {
	AccountID  string  `json:"account_id"`
	FirstName  string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName   string  `json:"last_name" validate:"required,min=1,max=100"`
	Nickname   *string `json:"nickname,omitempty" validate:"omitempty,max=50"`
	AvatarPath *string              `json:"avatar_path,omitempty" validate:"omitempty,max=500"`
	Skills     []profileEntity.ProfileSkill `json:"skills,omitempty"`
}

type CreateProfileResponse struct {
	AccountID string `json:"account_id"`
	ProfileID string `json:"profile_id"`
}

type GetProfileByAccountIDRequest struct {
	AccountID string `json:"account_id"`
}

type GetProfileByAccountIDResponse struct {
	AccountID  string    `json:"account_id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Nickname   *string   `json:"nickname,omitempty"`
	AvatarPath *string   `json:"avatar_path,omitempty"`
	State      string                       `json:"state"`
	Skills     []profileEntity.ProfileSkill `json:"skills"`
	CreatedAt  time.Time                    `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type UpdateProfileRequest struct {
	AccountID  string  `json:"account_id"`
	FirstName  string  `json:"first_name" validate:"required,min=1,max=100"`
	LastName   string  `json:"last_name" validate:"required,min=1,max=100"`
	Nickname   *string `json:"nickname,omitempty" validate:"omitempty,max=50"`
	AvatarPath *string              `json:"avatar_path,omitempty" validate:"omitempty,max=500"`
	Skills     []profileEntity.ProfileSkill `json:"skills,omitempty"`
}

type UpdateProfileResponse struct {
	AccountID  string  `json:"account_id"`
	ProfileID  string  `json:"profile_id"`
	FirstName  string  `json:"first_name"`
	LastName   string  `json:"last_name"`
	Nickname   *string `json:"nickname,omitempty"`
	AvatarPath *string `json:"avatar_path,omitempty"`
}
