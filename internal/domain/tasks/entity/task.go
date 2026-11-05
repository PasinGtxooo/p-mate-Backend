package entity

import (
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const TaskIDPrefix = "tsk"

type Task struct {
	ID             uuid.UUID       `json:"id" gorm:"column:id"`
	NodeID         uuid.UUID       `json:"nodeId" gorm:"column:node_id;type:char(36);not null;index"`
	ProjectID      uuid.UUID       `json:"projectId" gorm:"column:project_id"`
	Project        *entity.Project `json:"project,omitempty" gorm:"foreignKey:ProjectID;references:ID"`
	Name           string          `json:"name" gorm:"column:name"`
	Description    *string         `json:"description" gorm:"column:description"`
	Priority       string          `json:"priority" gorm:"column:priority"`
	StartDateTime  *string         `json:"startDateTime" gorm:"column:start_datetime"`
	EndDateTime    *string         `json:"endDateTime" gorm:"column:end_datetime"`
	Location       *string         `json:"location" gorm:"column:location"`
	RecurringDays  *int            `json:"recurringDays" gorm:"column:recurring_days"`
	RecurringUntil *string         `json:"recurringUntil" gorm:"column:recurring_until"`
	EstimateManday *float64        `json:"estimateManday" gorm:"column:estimate_manday"`
	RequiredSkill  *string         `json:"requiredSkill" gorm:"column:required_skill"`
	AssigneeID     *uuid.UUID      `json:"assigneeId" gorm:"column:assignee_id;type:char(36);index"`
	Status         string          `json:"status" gorm:"column:status"`
	CreatedAt      time.Time       `json:"createdAt" gorm:"column:created_at"`
	UpdatedAt      time.Time       `json:"updatedAt" gorm:"column:updated_at"`
	DeletedAt      gorm.DeletedAt  `json:"deletedAt" gorm:"column:deleted_at;index"`
}

func (Task) TableName() string {
	return "tasks"
}
