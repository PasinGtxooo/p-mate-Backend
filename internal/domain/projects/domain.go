package projects

import (
	"encoding/json"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
)

// Project represents the project data exposed via the HTTP API.
// It is mapped from the domain/entity Project model.
type Project struct {
	ID        string          `json:"id"`
	NodeID    string          `json:"nodeId"`
	OwnerID   string          `json:"ownerId"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	DeletedAt time.Time       `json:"deletedAt"`
}

// FromProjectModel converts a domain/entity Project model to the HTTP Project DTO.
func FromProjectModel(p *entity.Project) *Project {
	if p == nil {
		return nil
	}

	proj := &Project{
		ID:        p.ID.String(),
		NodeID:    p.NodeID.String(),
		OwnerID:   p.OwnerID.String(),
		Name:      p.Name,
		Config:    p.Config,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}

	return proj
}
