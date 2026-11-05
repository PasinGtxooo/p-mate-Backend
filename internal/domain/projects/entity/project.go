package entity

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const ProjectIDPrefix = "proj"

type Project struct {
	ID        uuid.UUID       `json:"id" gorm:"type:char(36);primaryKey"`
	NodeID    uuid.UUID       `json:"nodeId" gorm:"type:char(36);not null;index"`
	OwnerID   uuid.UUID       `json:"ownerId" gorm:"type:char(36);not null;index"`
	Name      string          `json:"name" gorm:"type:varchar(255);not null"`
	Config    json.RawMessage `json:"config" gorm:"type:jsonb"`
	CreatedAt time.Time       `json:"createdAt" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time       `json:"updatedAt" gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt  `json:"deletedAt" gorm:"column:deleted_at;index"`
}

func (Project) TableName() string {
	return "projects"
}
