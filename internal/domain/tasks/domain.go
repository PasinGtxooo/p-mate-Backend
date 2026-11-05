package tasks

import (
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/samber/lo"
)

type Status struct {
	Todo       string
	InProgress string
	InReview   string
	Done       string
}

// TaskStatistics represents the count of tasks grouped by status
type TaskStatistics struct {
	Todo       int64 `json:"todo"`
	InProgress int64 `json:"in_progress"`
	InReview   int64 `json:"in_review"`
	Done       int64 `json:"done"`
}

// StatusCount represents a single status count result from database
type StatusCount struct {
	Status string
	Count  int64
}

// Task represents the task data exposed via the HTTP API.
// It is mapped from the domain/entity Task model.
type Task struct {
	ID             string    `json:"id"`
	NodeID         string    `json:"nodeId"`
	ProjectID      string    `json:"projectId"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Priority       string    `json:"priority"`
	StartDateTime  string    `json:"startDateTime"`
	EndDateTime    string    `json:"endDateTime"`
	Location       string    `json:"location"`
	RecurringDays  int       `json:"recurringDays"`
	RecurringUntil string    `json:"recurringUntil"`
	EstimateManday float64   `json:"estimateManday"`
	Status         Status    `json:"status"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
	DeletedAt      time.Time `json:"deletedAt"`
}

// FromTaskModel converts a domain/entity Task model to the HTTP Task DTO.
func FromTaskModel(p *entity.Task) *Task {
	if p == nil {
		return nil
	}

	task := &Task{
		ID:             p.ID.String(),
		NodeID:         p.NodeID.String(),
		ProjectID:      p.ProjectID.String(),
		Name:           p.Name,
		Description:    lo.FromPtr(p.Description),
		Priority:       p.Priority,
		StartDateTime:  lo.FromPtr(p.StartDateTime),
		EndDateTime:    lo.FromPtr(p.EndDateTime),
		Location:       lo.FromPtr(p.Location),
		RecurringDays:  lo.FromPtr(p.RecurringDays),
		RecurringUntil: lo.FromPtr(p.RecurringUntil),
		EstimateManday: lo.FromPtr(p.EstimateManday),
		Status:         Status{Todo: "todo", InProgress: "in_progress", InReview: "in_review", Done: "done"},
		CreatedAt:      p.CreatedAt,
		UpdatedAt:      p.UpdatedAt,
	}

	return task
}
