package service

import (
	"context"
	"errors"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/task"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	pkgutils "github.com/FrostBitzX/smart-task-ai/pkg/utils"
	"github.com/google/uuid"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
)

type TaskService struct {
	repo        tasks.TaskRepository
	projectRepo projects.ProjectRepository
	memberRepo  projects.ProjectMemberRepository
}

func NewTaskService(repo tasks.TaskRepository, projectRepo projects.ProjectRepository, memberRepo projects.ProjectMemberRepository) *TaskService {
	return &TaskService{
		repo:        repo,
		projectRepo: projectRepo,
		memberRepo:  memberRepo,
	}
}

func (s *TaskService) CreateTask(ctx context.Context, projectID uuid.UUID, req *task.CreateTaskRequest, accountIDStr string) (*entity.Task, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// Parse accountID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Check if user is a member of the project
	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	// Get project to use its node_id for the task
	project, err := s.projectRepo.GetProjectByID(ctx, projectID, uuid.Nil)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	if err := s.validateTimeRange(req.StartDateTime, req.EndDateTime); err != nil {
		return nil, err
	}

	// create domain entity
	now := time.Now()
	task := &entity.Task{
		ID:            uuid.New(),
		NodeID:        project.NodeID,
		ProjectID:     projectID,
		Name:          req.Name,
		Priority:      req.Priority,
		StartDateTime: req.StartDateTime,
		EndDateTime:   req.EndDateTime,
		Status:        "todo",
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	if req.Description != nil {
		task.Description = req.Description
	}

	if req.Location != nil {
		task.Location = req.Location
	}

	if req.RecurringDays != nil {
		task.RecurringDays = req.RecurringDays
	}

	if req.RecurringUntil != nil {
		task.RecurringUntil = req.RecurringUntil
	}

	if req.EstimateManday != nil {
		task.EstimateManday = req.EstimateManday
	}

	if req.RequiredSkill != nil {
		task.RequiredSkill = req.RequiredSkill
	}

	if req.AssigneeID != nil && *req.AssigneeID != "" {
		assigneeID, perr := utils.ParseID(*req.AssigneeID, "acc")
		if perr != nil {
			return nil, apperror.NewBadRequestError("invalid assignee ID format", "INVALID_ASSIGNEE_ID", perr)
		}
		task.AssigneeID = &assigneeID
	}

	// persist task to database
	err = s.repo.CreateTask(ctx, task)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to create task", "CREATE_TASK_ERROR", err)
	}

	return task, nil
}

func (s *TaskService) GetTaskByID(ctx context.Context, taskID uuid.UUID, accountIDStr string) (*entity.Task, error) {
	// Parse accountID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Get task
	tsk, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("task not found", "TASK_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get task", "GET_TASK_ERROR", err)
	}

	// Check if user is a member of the project
	isMember, err := s.memberRepo.IsMember(ctx, tsk.ProjectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	return tsk, nil
}

func (s *TaskService) ListTasksByProject(ctx context.Context, projectID uuid.UUID, accountIDStr string) ([]*entity.Task, error) {
	// Parse accountID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Check if user is a member of the project
	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	tasks, err := s.repo.ListTasksByProject(ctx, projectID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list tasks", "LIST_TASKS_ERROR", err)
	}

	return tasks, nil
}

func (s *TaskService) UpdateTask(ctx context.Context, taskID uuid.UUID, req *task.UpdateTaskRequest, accountIDStr string) (*entity.Task, error) {
	// Parse accountID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Get task by ID for update
	tsk, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("task not found", "TASK_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get task", "GET_TASK_ERROR", err)
	}

	// Check if user is a member of the project
	isMember, err := s.memberRepo.IsMember(ctx, tsk.ProjectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	newStatus := tsk.Status
	if req.Status != nil {
		newStatus = *req.Status
	}

	if newStatus != "todo" && req.StartDateTime != nil {
		return nil, apperror.NewBadRequestError("cannot update start date/time when status is not todo", "INVALID_REQUEST", nil)
	}

	// Additional validation same as CreateTask
	if err := s.validateTimeRange(req.StartDateTime, req.EndDateTime); err != nil {
		return nil, err
	}

	// Update fields only if provided (PATCH semantics)
	if req.Name != "" {
		tsk.Name = req.Name
	}
	if req.Status != nil {
		tsk.Status = *req.Status
	}
	if req.Priority != "" {
		tsk.Priority = req.Priority
	}

	// Nullable fields
	pkgutils.UpdateNullableString(&tsk.Description, req.Description)
	pkgutils.UpdateNullableString(&tsk.Location, req.Location)
	pkgutils.UpdateNullableString(&tsk.StartDateTime, req.StartDateTime)
	pkgutils.UpdateNullableString(&tsk.EndDateTime, req.EndDateTime)
	pkgutils.UpdateNullableInt(&tsk.RecurringDays, req.RecurringDays)
	pkgutils.UpdateNullableString(&tsk.RecurringUntil, req.RecurringUntil)
	pkgutils.UpdateNullableFloat(&tsk.EstimateManday, req.EstimateManday)
	pkgutils.UpdateNullableString(&tsk.RequiredSkill, req.RequiredSkill)
	if req.AssigneeID != nil {
		if *req.AssigneeID == "" {
			tsk.AssigneeID = nil
		} else {
			assigneeID, perr := uptParseAssignee(*req.AssigneeID)
			if perr != nil {
				return nil, perr
			}
			tsk.AssigneeID = assigneeID
		}
	}

	tsk.UpdatedAt = time.Now()

	err = s.repo.UpdateTask(ctx, tsk)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to update task", "UPDATE_TASK_ERROR", err)
	}

	return tsk, nil
}

func (s *TaskService) DeleteTask(ctx context.Context, taskID uuid.UUID, accountIDStr string) error {
	// Parse accountID
	accountID, err := uuid.Parse(accountIDStr)
	if err != nil {
		return apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Get task
	tsk, err := s.repo.GetTaskByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return apperror.NewNotFoundError("task not found", "TASK_NOT_FOUND", err)
		}
		return apperror.NewInternalServerError("failed to get task", "GET_TASK_ERROR", err)
	}

	// Check if user is a member of the project
	isMember, err := s.memberRepo.IsMember(ctx, tsk.ProjectID, accountID)
	if err != nil {
		return apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	err = s.repo.DeleteTask(ctx, taskID)
	if err != nil {
		return apperror.NewInternalServerError("failed to delete task", "DELETE_TASK_ERROR", err)
	}

	return nil
}

func (s *TaskService) GetTaskStatistics(ctx context.Context, accountID uuid.UUID) (*tasks.TaskStatistics, error) {
	// Get all projects where user is a member
	projects, _, err := s.projectRepo.ListProjectByAccountID(ctx, accountID, uuid.Nil, 1000, 0)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list user projects", "LIST_PROJECTS_ERROR", err)
	}

	// Initialize statistics
	statistics := &tasks.TaskStatistics{
		Todo:       0,
		InProgress: 0,
		InReview:   0,
		Done:       0,
	}

	// Count tasks for each project
	for _, project := range projects {
		status, err := s.repo.CountTasksByStatusAndProject(ctx, project.ID)
		if err != nil {
			return nil, apperror.NewInternalServerError("failed to get task statistics", "GET_STATISTICS_ERROR", err)
		}

		for _, sc := range status {
			switch sc.Status {
			case "todo":
				statistics.Todo += sc.Count
			case "in_progress":
				statistics.InProgress += sc.Count
			case "in_review":
				statistics.InReview += sc.Count
			case "done":
				statistics.Done += sc.Count
			}
		}
	}

	return statistics, nil
}

func (s *TaskService) GetUnscheduledTasks(ctx context.Context, accountID uuid.UUID) ([]*entity.Task, error) {
	// Get all projects where user is a member
	projects, _, err := s.projectRepo.ListProjectByAccountID(ctx, accountID, uuid.Nil, 1000, 0)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list user projects", "LIST_PROJECTS_ERROR", err)
	}

	// Collect all unscheduled tasks from user's projects
	var allTasks []*entity.Task
	for _, project := range projects {
		tasks, err := s.repo.ListUnscheduledTasksByProject(ctx, project.ID)
		if err != nil {
			return nil, apperror.NewInternalServerError("failed to get unscheduled tasks", "GET_UNSCHEDULED_TASKS_ERROR", err)
		}
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

func (s *TaskService) ListTodayTasks(ctx context.Context, accountID uuid.UUID, today time.Time) ([]*entity.Task, error) {
	// Get all projects where user is a member
	projects, _, err := s.projectRepo.ListProjectByAccountID(ctx, accountID, uuid.Nil, 1000, 0)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list user projects", "LIST_PROJECTS_ERROR", err)
	}

	// Collect all today's tasks from user's projects
	var allTasks []*entity.Task
	for _, project := range projects {
		tasks, err := s.repo.ListTodayTasksByProject(ctx, project.ID, today)
		if err != nil {
			return nil, apperror.NewInternalServerError("failed to list today's tasks", "LIST_TODAY_TASKS_ERROR", err)
		}
		allTasks = append(allTasks, tasks...)
	}

	return allTasks, nil
}

func (s *TaskService) validateTimeRange(startStr, endStr *string) error {
	if startStr != nil && endStr != nil {
		start, err := time.Parse(time.RFC3339, *startStr)
		if err != nil {
			return apperror.NewBadRequestError("invalid start_datetime format", "INVALID_DATE_FORMAT", err)
		}
		end, err := time.Parse(time.RFC3339, *endStr)
		if err != nil {
			return apperror.NewBadRequestError("invalid end_datetime format", "INVALID_DATE_FORMAT", err)
		}

		if start.Equal(end) {
			return apperror.NewBadRequestError("start_datetime and end_datetime cannot be the same", "INVALID_REQUEST", nil)
		}
		if end.Before(start) {
			return apperror.NewBadRequestError("end_datetime must be greater than start_datetime", "INVALID_REQUEST", nil)
		}
	}
	return nil
}

// uptParseAssignee parses an acc_ prefixed account id into a *uuid.UUID.
func uptParseAssignee(raw string) (*uuid.UUID, error) {
	id, err := utils.ParseID(raw, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid assignee ID format", "INVALID_ASSIGNEE_ID", err)
	}
	return &id, nil
}
