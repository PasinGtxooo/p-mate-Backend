package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/dashboard"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
)

type GetUnscheduledTasksUseCase struct {
	taskService *service.TaskService
	logger      logger.Logger
}

func NewGetUnscheduledTasksUseCase(svc *service.TaskService, l logger.Logger) *GetUnscheduledTasksUseCase {
	return &GetUnscheduledTasksUseCase{
		taskService: svc,
		logger:      l,
	}
}

func (uc *GetUnscheduledTasksUseCase) Execute(ctx context.Context, nodeIDStr string) (*dashboard.UnscheduledTasksResponse, error) {
	// Validate input
	if nodeIDStr == "" {
		uc.logger.Error("Empty node ID provided", nil)
		return nil, apperror.NewBadRequestError("node ID is required", "MISSING_NODE_ID", nil)
	}

	// Parse nodeID
	nodeID, err := uuid.Parse(nodeIDStr)
	if err != nil {
		uc.logger.Error("Invalid node ID format", map[string]interface{}{
			"nodeID": nodeIDStr,
			"error":  err.Error(),
		})
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Get unscheduled tasks from service
	tasks, err := uc.taskService.GetUnscheduledTasks(ctx, nodeID)
	if err != nil {
		uc.logger.Error("Failed to get unscheduled tasks", map[string]interface{}{
			"nodeID": nodeID.String(),
			"error":  err.Error(),
		})
		return nil, err
	}

	// Convert domain entities to DTOs
	items := make([]dashboard.TaskWithProjectResponse, 0, len(tasks))
	for _, task := range tasks {
		item := dashboard.TaskWithProjectResponse{
			ID:            task.ID.String(),
			Name:          task.Name,
			Description:   task.Description,
			Priority:      task.Priority,
			Status:        task.Status,
			StartDateTime: task.StartDateTime,
			EndDateTime:   task.EndDateTime,
		}

		// Add project information if available
		if task.Project != nil {
			item.Project = dashboard.ProjectInfo{
				ID:   task.Project.ID.String(),
				Name: task.Project.Name,
			}
		} else {
			uc.logger.Warn("Task has no project information", map[string]interface{}{
				"taskID": task.ID.String(),
			})
		}

		items = append(items, item)
	}

	response := &dashboard.UnscheduledTasksResponse{
		Items: items,
		Total: len(items),
	}

	uc.logger.Info("Unscheduled tasks retrieved successfully", map[string]interface{}{
		"nodeID": nodeID.String(),
		"count":  len(items),
	})

	return response, nil
}
