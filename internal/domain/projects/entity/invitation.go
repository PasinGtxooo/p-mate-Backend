package entity

import (
	"time"

	accountEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/accounts/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const InvitationIDPrefix = "inv"

type ProjectInvitation struct {
	ID               uuid.UUID             `json:"id" gorm:"type:uuid;primaryKey"`
	NodeID           uuid.UUID             `json:"nodeId" gorm:"type:uuid;not null;index"`
	ProjectID        uuid.UUID             `json:"projectId" gorm:"type:uuid;not null;index:idx_project_invitee,unique"`
	InviterAccountID uuid.UUID             `json:"inviterAccountId" gorm:"type:uuid;not null"`
	InviteeAccountID uuid.UUID             `json:"inviteeAccountId" gorm:"type:uuid;not null;index:idx_project_invitee,unique"`
	Role             string                `json:"role" gorm:"type:varchar(50);not null"`
	Status           string                `json:"status" gorm:"type:varchar(20);not null;default:'pending'"`
	ExpiresAt        time.Time             `json:"expiresAt" gorm:"not null"`
	RespondedAt      *time.Time            `json:"respondedAt,omitempty"`
	CreatedAt        time.Time             `json:"createdAt" gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt        time.Time             `json:"updatedAt" gorm:"not null;default:CURRENT_TIMESTAMP"`
	DeletedAt        gorm.DeletedAt        `json:"deletedAt,omitempty" gorm:"index"`
	Project          Project               `json:"project,omitempty" gorm:"foreignKey:ProjectID"`
	Inviter          accountEntity.Account `json:"inviter,omitempty" gorm:"foreignKey:InviterAccountID"`
	Invitee          accountEntity.Account `json:"invitee,omitempty" gorm:"foreignKey:InviteeAccountID"`
}

func (ProjectInvitation) TableName() string {
	return "project_invitations"
}
