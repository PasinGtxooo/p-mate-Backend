package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/task"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type CreateTaskUseCase struct {
	taskService *service.TaskService
	logger      logger.Logger
}

func NewCreateTaskUseCase(svc *service.TaskService, l logger.Logger) *CreateTaskUseCase {
	return &CreateTaskUseCase{
		taskService: svc,
		logger:      l,
	}
}

func (uc *CreateTaskUseCase) Execute(ctx context.Context, projectID string, req *task.CreateTaskRequest, nodeID string) (*task.CreateTaskResponse, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	parsedProjectID, err := utils.ParseID(projectID, projectEntity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	tsk, err := uc.taskService.CreateTask(ctx, parsedProjectID, req, nodeID)
	if err != nil {
		return nil, err
	}

	// Convert UUID to string with prefix
	taskID := utils.ShortUUIDWithPrefix(tsk.ID, entity.TaskIDPrefix)

	res := &task.CreateTaskResponse{
		ID:             taskID,
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
	}
	return res, nil
}
