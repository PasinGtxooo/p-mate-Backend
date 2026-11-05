package persistence

import (
	"context"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type taskRepository struct {
	db *gorm.DB
}

func NewTaskRepository(db *gorm.DB) tasks.TaskRepository {
	return &taskRepository{db: db}
}

func (r *taskRepository) CreateTask(ctx context.Context, task *entity.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *taskRepository) GetTaskByID(ctx context.Context, taskID uuid.UUID) (*entity.Task, error) {
	var task entity.Task
	err := r.db.WithContext(ctx).Where("id = ?", taskID).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *taskRepository) ListTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	var tasks []*entity.Task
	err := r.db.WithContext(ctx).
		Where("project_id = ?", projectID).
		Find(&tasks).Error
	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) CountTasksByProject(ctx context.Context, projectID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.Task{}).
		Where("project_id = ?", projectID).
		Count(&count).Error
	return count, err
}

func (r *taskRepository) UpdateTask(ctx context.Context, task *entity.Task) error {
	return r.db.WithContext(ctx).
		Model(&entity.Task{}).
		Select("*").
		Where("id = ?", task.ID).
		Updates(task).Error
}

func (r *taskRepository) DeleteTask(ctx context.Context, taskID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", taskID).
		Delete(&entity.Task{}).Error
}

func (r *taskRepository) CountTasksByStatusAndProject(ctx context.Context, projectID uuid.UUID) ([]tasks.StatusCount, error) {
	var results []tasks.StatusCount
	err := r.db.WithContext(ctx).
		Model(&entity.Task{}).
		Select("status, COUNT(*) as count").
		Where("project_id = ?", projectID).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (r *taskRepository) ListUnscheduledTasksByProject(ctx context.Context, projectID uuid.UUID) ([]*entity.Task, error) {
	var tasks []*entity.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("project_id = ?", projectID).
		Where("status IN (?)", []string{"todo", "in_progress"}).
		Where("(start_datetime IS NULL OR end_datetime IS NULL)").
		Order("updated_at DESC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (r *taskRepository) ListTodayTasksByProject(ctx context.Context, projectID uuid.UUID, today time.Time) ([]*entity.Task, error) {
	// Get start of day (00:00:00)
	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	// Get end of day (23:59:59)
	endOfDay := time.Date(today.Year(), today.Month(), today.Day(), 23, 59, 59, 999999999, today.Location())

	var tasks []*entity.Task
	err := r.db.WithContext(ctx).
		Preload("Project").
		Where("project_id = ?", projectID).
		Where(
			r.db.Where("start_datetime >= ? AND start_datetime <= ?", startOfDay, endOfDay).
				Or("end_datetime >= ? AND end_datetime <= ?", startOfDay, endOfDay).
				Or("(start_datetime < ? AND end_datetime > ?)", startOfDay, endOfDay),
		).
		Order("start_datetime ASC").
		Find(&tasks).Error

	if err != nil {
		return nil, err
	}

	return tasks, nil
}
