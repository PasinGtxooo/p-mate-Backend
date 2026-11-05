package entity

import (
	"time"

	"github.com/google/uuid"
)

type ProjectMember struct {
	ProjectID uuid.UUID   `gorm:"type:char(36);primaryKey"`
	AccountID uuid.UUID   `gorm:"type:char(36);primaryKey"`
	NodeID    uuid.UUID   `gorm:"type:char(36);not null;index"`
	Role      ProjectRole `gorm:"type:project_role;not null;default:'member'"`
	CreatedAt time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time   `gorm:"not null;default:CURRENT_TIMESTAMP"`
}

func (ProjectMember) TableName() string {
	return "project_members"
}
