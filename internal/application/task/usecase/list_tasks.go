package usecase

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/application/common"
	"github.com/FrostBitzX/smart-task-ai/internal/application/task"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	taskEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
)

type ListTasksByProjectUseCase struct {
	taskService *service.TaskService
	logger      logger.Logger
}

func NewListTasksByProjectUseCase(svc *service.TaskService, l logger.Logger) *ListTasksByProjectUseCase {
	return &ListTasksByProjectUseCase{
		taskService: svc,
		logger:      l,
	}
}

func (uc *ListTasksByProjectUseCase) Execute(ctx context.Context, projectID string, nodeID string) (*task.ListTasksByProjectResponse, error) {
	parsedProjectID, err := utils.ParseID(projectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	tsks, err := uc.taskService.ListTasksByProject(ctx, parsedProjectID, nodeID)
	if err != nil {
		return nil, err
	}

	items := make([]task.GetTaskByIDResponse, 0, len(tsks))
	for _, t := range tsks {
		items = append(items, task.GetTaskByIDResponse{
			ID:             utils.ShortUUIDWithPrefix(t.ID, taskEntity.TaskIDPrefix),
			Status:         t.Status,
			Name:           t.Name,
			Description:    t.Description,
			Priority:       t.Priority,
			StartDateTime:  t.StartDateTime,
			EndDateTime:    t.EndDateTime,
			Location:       t.Location,
			RecurringDays:  t.RecurringDays,
			RecurringUntil: t.RecurringUntil,
			EstimateManday: t.EstimateManday,
			RequiredSkill:  t.RequiredSkill,
			AssigneeID:     mapAssignee(t.AssigneeID),
			CreatedAt:      t.CreatedAt,
			UpdatedAt:      t.UpdatedAt,
		})
	}

	res := &task.ListTasksByProjectResponse{
		Items: items,
		Pagination: common.Pagination{
			Total:   len(items),
			Limit:   10, // Default limit for now, can be parameterized later
			Offset:  0,
			HasMore: false,
		},
	}

	return res, nil
}
