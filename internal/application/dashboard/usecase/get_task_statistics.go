package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/dashboard"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
)

type GetTaskStatisticsUseCase struct {
	taskService *service.TaskService
	logger      logger.Logger
}

func NewGetTaskStatisticsUseCase(svc *service.TaskService, l logger.Logger) *GetTaskStatisticsUseCase {
	return &GetTaskStatisticsUseCase{
		taskService: svc,
		logger:      l,
	}
}

func (uc *GetTaskStatisticsUseCase) Execute(ctx context.Context, nodeIDStr string) (*dashboard.TaskStatisticsResponse, error) {
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

	// Get task statistics from service
	statistics, err := uc.taskService.GetTaskStatistics(ctx, nodeID)
	if err != nil {
		uc.logger.Error("Failed to get task statistics", map[string]interface{}{
			"nodeID": nodeID.String(),
			"error":  err.Error(),
		})
		return nil, err
	}

	// Convert domain model to DTO
	response := &dashboard.TaskStatisticsResponse{
		Todo:       statistics.Todo,
		InProgress: statistics.InProgress,
		InReview:   statistics.InReview,
		Done:       statistics.Done,
	}

	uc.logger.Info("Task statistics retrieved successfully", map[string]interface{}{
		"nodeID": nodeID.String(),
		"total":  statistics.Todo + statistics.InProgress + statistics.InReview + statistics.Done,
	})

	return response, nil
}
