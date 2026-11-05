package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/task"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type GetTaskByIDUseCase struct {
	taskService *service.TaskService
	logger      logger.Logger
}

func NewGetTaskByIDUseCase(svc *service.TaskService, l logger.Logger) *GetTaskByIDUseCase {
	return &GetTaskByIDUseCase{
		taskService: svc,
		logger:      l,
	}
}

func (uc *GetTaskByIDUseCase) Execute(ctx context.Context, taskID string, nodeID string) (*task.GetTaskByIDResponse, error) {
	parsedTaskID, err := utils.ParseID(taskID, entity.TaskIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid task ID format", "INVALID_TASK_ID", err)
	}

	tsk, err := uc.taskService.GetTaskByID(ctx, parsedTaskID, nodeID)
	if err != nil {
		return nil, err
	}

	res := &task.GetTaskByIDResponse{
		ID:             utils.ShortUUIDWithPrefix(tsk.ID, entity.TaskIDPrefix),
		Status:         tsk.Status,
		Name:           tsk.Name,
		Description:    tsk.Description,
		Priority:       tsk.Priority,
		StartDateTime:  tsk.StartDateTime,
		EndDateTime:    tsk.EndDateTime,
		Location:       tsk.Location,
		RecurringDays:  tsk.RecurringDays,
		RecurringUntil: tsk.RecurringUntil,
		EstimateManday: tsk.EstimateManday,
		RequiredSkill:  tsk.RequiredSkill,
		AssigneeID:     mapAssignee(tsk.AssigneeID),
		CreatedAt:      tsk.CreatedAt,
		UpdatedAt:      tsk.UpdatedAt,
	}

	return res, nil
}
