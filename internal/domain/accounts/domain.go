package accounts

import (
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
)

type State struct {
	Active   string
	Inactive string
}

// Account represents the account data exposed via the HTTP API.
// It is mapped from the domain/entity Account model.
type Account struct {
	ID        string    `json:"id"`
	NodeID    string    `json:"nodeId"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	State     State     `json:"state"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// FromAccountModel converts a domain/entity Account model to the HTTP Account DTO.
func FromAccountModel(m *entity.Account) *Account {
	if m == nil {
		return nil
	}

	acc := &Account{
		ID:        m.ID.String(),
		NodeID:    m.NodeID.String(),
		Username:  m.Username,
		Email:     m.Email,
		State:     State{Active: m.State, Inactive: m.State},
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
	}

	return acc
}
