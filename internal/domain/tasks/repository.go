//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../../mocks/task_repository.go -package=mocks
package tasks

import (
	"context"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/google/uuid"
)

type TaskRepository interface {
	CreateTask(ctx context.Context, task *entity.Task) error
	GetTaskByID(ctx context.Context, taskID uuid.UUID) (*entity.Task, error)
	ListTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)
	CountTasksByProject(ctx context.Context, projectID uuid.UUID) (int64, error)
	UpdateTask(ctx context.Context, task *entity.Task) error
	DeleteTask(ctx context.Context, taskID uuid.UUID) error
	CountTasksByStatusAndProject(ctx context.Context, projectID uuid.UUID) ([]StatusCount, error)
	ListUnscheduledTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error)
	ListTodayTasksByProject(ctx context.Context, projectID uuid.UUID, today time.Time) ([]*entity.Task, error)
}
