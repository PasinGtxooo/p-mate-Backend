package profiles

import (
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/profiles/entity"
	"github.com/samber/lo"
)

type State struct {
	Active   string
	Inactive string
}

// Profile represents the profile data exposed via the HTTP API.
// It is mapped from the domain/entity Profile model.
type Profile struct {
	ID         string    `json:"id"`
	NodeID     string    `json:"nodeId"`
	AccountID  string    `json:"accountId"`
	FirstName  string    `json:"firstName"`
	LastName   string    `json:"lastName"`
	Nickname   string    `json:"nickname"`
	AvatarPath string                `json:"avatarPath"`
	Position   string                `json:"position"`
	Skills     []entity.ProfileSkill `json:"skills"`
	State      State     `json:"state"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// FromProfileModel converts a domain/entity Profile model to the HTTP Profile DTO.
func FromProfileModel(p *entity.Profile) *Profile {
	if p == nil {
		return nil
	}

	prof := &Profile{
		ID:         p.ID.String(),
		NodeID:     p.NodeID.String(),
		AccountID:  p.AccountID.String(),
		FirstName:  p.FirstName,
		LastName:   p.LastName,
		Nickname:   lo.FromPtr(p.Nickname),
		AvatarPath: lo.FromPtr(p.AvatarPath),
		Position:   lo.FromPtr(p.Position),
		Skills:     p.Skills,
		State:      State{Active: p.State, Inactive: p.State},
		CreatedAt:  p.CreatedAt,
		UpdatedAt:  p.UpdatedAt,
	}

	return prof
}
