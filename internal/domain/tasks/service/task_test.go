package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/task"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/mocks"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestTaskService_CreateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	ctx := context.Background()
	projectID := uuid.New()
	nodeIDStr := "550e8400-e29b-41d4-a716-446655440000"

	// Helper to create time strings
	now := time.Now()
	startTime := now.Add(time.Hour).Format(time.RFC3339)
	endTime := now.Add(2 * time.Hour).Format(time.RFC3339)
	sameTime := now.Format(time.RFC3339)

	tests := []struct {
		name          string
		projectID     uuid.UUID
		request       *task.CreateTaskRequest
		setupMock     func()
		expectedError string
		expectNil     bool
		validate      func(t *testing.T, res *entity.Task)
	}{
		{
			name:      "success - creates task with all fields",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:          "Test Task",
				Description:   strPtr("Task description"),
				Priority:      "1",
				StartDateTime: &startTime,
				EndDateTime:   &endTime,
				Location:      strPtr("Office"),
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
				mockRepo.EXPECT().
					CreateTask(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, tsk *entity.Task) error {
						assert.Equal(t, "Test Task", tsk.Name)
						assert.Equal(t, "todo", tsk.Status)
						assert.Equal(t, projectID, tsk.ProjectID)
						return nil
					}).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
			validate: func(t *testing.T, res *entity.Task) {
				assert.Equal(t, "Test Task", res.Name)
				assert.Equal(t, "todo", res.Status)
			},
		},
		{
			name:      "success - creates task with minimal fields",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:     "Minimal Task",
				Priority: "2",
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
				mockRepo.EXPECT().
					CreateTask(ctx, gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
			validate: func(t *testing.T, res *entity.Task) {
				assert.Equal(t, "Minimal Task", res.Name)
			},
		},
		{
			name:          "error - nil request",
			projectID:     projectID,
			request:       nil,
			setupMock:     func() {},
			expectedError: "invalid request body",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - project not found",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:     "Test Task",
				Priority: "1",
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(nil, apperror.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "project not found",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - project validation fails",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:     "Test Task",
				Priority: "1",
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to validate project",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - same start and end time",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:          "Test Task",
				StartDateTime: &sameTime,
				EndDateTime:   &sameTime,
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
			},
			expectedError: "cannot be the same",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - end time before start time",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:          "Test Task",
				StartDateTime: &endTime,   // Later time as start
				EndDateTime:   &startTime, // Earlier time as end
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
			},
			expectedError: "end_datetime must be greater than start_datetime",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - invalid start datetime format",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:          "Test Task",
				StartDateTime: strPtr("invalid-date"),
				EndDateTime:   &endTime,
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
			},
			expectedError: "invalid start_datetime format",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - invalid end datetime format",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name:          "Test Task",
				StartDateTime: &startTime,
				EndDateTime:   strPtr("invalid-date"),
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
			},
			expectedError: "invalid end_datetime format",
			expectNil:     true,
			validate:      nil,
		},
		{
			name:      "error - repository create fails",
			projectID: projectID,
			request: &task.CreateTaskRequest{
				Name: "Test Task",
			},
			setupMock: func() {
				mockProjectRepo.EXPECT().
					GetProjectByID(ctx, projectID, uuid.Nil).
					Return(&projectEntity.Project{ID: projectID}, nil).
					Times(1)
				mockRepo.EXPECT().
					CreateTask(ctx, gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to create task",
			expectNil:     true,
			validate:      nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.CreateTask(ctx, tt.projectID, tt.request, nodeIDStr)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
				if tt.validate != nil {
					tt.validate(t, res)
				}
			}
		})
	}
}

func TestTaskService_GetTaskByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	ctx := context.Background()
	taskID := uuid.New()
	nodeIDStr := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name          string
		taskID        uuid.UUID
		setupMock     func()
		expectedError string
		expectNil     bool
	}{
		{
			name:   "success - returns task",
			taskID: taskID,
			setupMock: func() {
				expectedTask := &entity.Task{
					ID:     taskID,
					Name:   "Test Task",
					Status: "todo",
				}
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(expectedTask, nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name:   "error - task not found",
			taskID: taskID,
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(nil, apperror.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "task not found",
			expectNil:     true,
		},
		{
			name:   "error - repository fails",
			taskID: taskID,
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to get task",
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.GetTaskByID(ctx, tt.taskID, nodeIDStr)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
				assert.Equal(t, taskID, res.ID)
			}
		})
	}
}

func TestTaskService_UpdateTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	ctx := context.Background()
	taskID := uuid.New()
	nodeIDStr := "550e8400-e29b-41d4-a716-446655440000"

	now := time.Now()
	newStartTime := now.Add(time.Hour).Format(time.RFC3339)
	todoStatus := "todo"
	projectID := uuid.New()
	accountID, _ := uuid.Parse(nodeIDStr)

	tests := []struct {
		name          string
		taskID        uuid.UUID
		request       *task.UpdateTaskRequest
		setupMock     func()
		expectedError string
		expectNil     bool
	}{
		{
			name:   "success - updates task name",
			taskID: taskID,
			request: &task.UpdateTaskRequest{
				Name: "Updated Name",
			},
			setupMock: func() {
				existingTask := &entity.Task{
					ID:        taskID,
					Status:    "todo",
					Name:      "Old Name",
					ProjectID: projectID,
				}
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(existingTask, nil).
					Times(1)
				mockMemberRepo.EXPECT().
					IsMember(ctx, projectID, accountID).
					Return(true, nil).
					Times(1)
				mockRepo.EXPECT().
					UpdateTask(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, tsk *entity.Task) error {
						assert.Equal(t, "Updated Name", tsk.Name)
						return nil
					}).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name:   "error - cannot update start_datetime when status is not todo",
			taskID: taskID,
			request: &task.UpdateTaskRequest{
				StartDateTime: &newStartTime,
			},
			setupMock: func() {
				existingTask := &entity.Task{
					ID:        taskID,
					Status:    "doing", // Not "todo"
					ProjectID: projectID,
				}
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(existingTask, nil).
					Times(1)
				mockMemberRepo.EXPECT().
					IsMember(ctx, projectID, accountID).
					Return(true, nil).
					Times(1)
			},
			expectedError: "cannot update start date/time when status is not todo",
			expectNil:     true,
		},
		{
			name:   "success - can update start_datetime when status is todo",
			taskID: taskID,
			request: &task.UpdateTaskRequest{
				StartDateTime: &newStartTime,
			},
			setupMock: func() {
				existingTask := &entity.Task{
					ID:        taskID,
					Status:    "todo",
					ProjectID: projectID,
				}
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(existingTask, nil).
					Times(1)
				mockMemberRepo.EXPECT().
					IsMember(ctx, projectID, accountID).
					Return(true, nil).
					Times(1)
				mockRepo.EXPECT().
					UpdateTask(ctx, gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name:   "success - can update start_datetime when changing status to todo",
			taskID: taskID,
			request: &task.UpdateTaskRequest{
				Status:        &todoStatus,
				StartDateTime: &newStartTime,
			},
			setupMock: func() {
				existingTask := &entity.Task{
					ID:        taskID,
					Status:    "in_progress", // Current status is not todo
					ProjectID: projectID,
				}
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(existingTask, nil).
					Times(1)
				mockMemberRepo.EXPECT().
					IsMember(ctx, projectID, accountID).
					Return(true, nil).
					Times(1)
				mockRepo.EXPECT().
					UpdateTask(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, tsk *entity.Task) error {
						assert.Equal(t, "todo", tsk.Status)
						assert.NotNil(t, tsk.StartDateTime)
						return nil
					}).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
		},
		{
			name:   "error - task not found",
			taskID: taskID,
			request: &task.UpdateTaskRequest{
				Name: "Updated",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(nil, apperror.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "task not found",
			expectNil:     true,
		},
		{
			name:   "error - repository update fails",
			taskID: taskID,
			request: &task.UpdateTaskRequest{
				Name: "Updated",
			},
			setupMock: func() {
				existingTask := &entity.Task{
					ID:        taskID,
					Status:    "todo",
					ProjectID: projectID,
				}
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(existingTask, nil).
					Times(1)
				mockMemberRepo.EXPECT().
					IsMember(ctx, projectID, accountID).
					Return(true, nil).
					Times(1)
				mockRepo.EXPECT().
					UpdateTask(ctx, gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to update task",
			expectNil:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.UpdateTask(ctx, tt.taskID, tt.request, nodeIDStr)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}

			if tt.expectNil {
				assert.Nil(t, res)
			} else {
				assert.NotNil(t, res)
			}
		})
	}
}

func TestTaskService_DeleteTask(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	ctx := context.Background()
	taskID := uuid.New()
	nodeIDStr := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name          string
		taskID        uuid.UUID
		setupMock     func()
		expectedError string
	}{
		{
			name:   "success - deletes task",
			taskID: taskID,
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(&entity.Task{ID: taskID}, nil).
					Times(1)
				mockRepo.EXPECT().
					DeleteTask(ctx, taskID).
					Return(nil).
					Times(1)
			},
			expectedError: "",
		},
		{
			name:   "error - task not found",
			taskID: taskID,
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(nil, apperror.ErrRecordNotFound).
					Times(1)
			},
			expectedError: "task not found",
		},
		{
			name:   "error - repository get fails",
			taskID: taskID,
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to get task",
		},
		{
			name:   "error - repository delete fails",
			taskID: taskID,
			setupMock: func() {
				mockRepo.EXPECT().
					GetTaskByID(ctx, taskID).
					Return(&entity.Task{ID: taskID}, nil).
					Times(1)
				mockRepo.EXPECT().
					DeleteTask(ctx, taskID).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to delete task",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := svc.DeleteTask(ctx, tt.taskID, nodeIDStr)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestTaskService_ListTasksByProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	ctx := context.Background()
	projectID := uuid.New()
	nodeIDStr := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name          string
		projectID     uuid.UUID
		setupMock     func()
		expectedCount int
		expectedError string
	}{
		{
			name:      "success - returns tasks",
			projectID: projectID,
			setupMock: func() {
				tasks := []*entity.Task{
					{ID: uuid.New(), Name: "Task 1", ProjectID: projectID},
					{ID: uuid.New(), Name: "Task 2", ProjectID: projectID},
				}
				mockRepo.EXPECT().
					ListTasksByProject(ctx, projectID).
					Return(tasks, nil).
					Times(1)
			},
			expectedCount: 2,
			expectedError: "",
		},
		{
			name:      "success - returns empty list",
			projectID: projectID,
			setupMock: func() {
				mockRepo.EXPECT().
					ListTasksByProject(ctx, projectID).
					Return([]*entity.Task{}, nil).
					Times(1)
			},
			expectedCount: 0,
			expectedError: "",
		},
		{
			name:      "error - repository fails",
			projectID: projectID,
			setupMock: func() {
				mockRepo.EXPECT().
					ListTasksByProject(ctx, projectID).
					Return(nil, errors.New("database error")).
					Times(1)
			},
			expectedCount: 0,
			expectedError: "failed to list tasks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.ListTasksByProject(ctx, tt.projectID, nodeIDStr)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Len(t, res, tt.expectedCount)
			}
		})
	}
}

// Helper function to create string pointers
func strPtr(s string) *string {
	return &s
}

func TestTaskService_GetTaskStatistics(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	ctx := context.Background()
	accountID := uuid.MustParse("550e8400-e29b-41d4-a716-446655440000")
	projectID1 := uuid.New()
	projectID2 := uuid.New()

	tests := []struct {
		name          string
		setupMock     func()
		expectedStats *tasks.TaskStatistics
		expectedError string
	}{
		{
			name: "success - returns statistics with all statuses from multiple projects",
			setupMock: func() {
				mockProjectRepo.EXPECT().
					ListProjectByAccountID(ctx, accountID, uuid.Nil, 1000, 0).
					Return([]*projectEntity.Project{
						{ID: projectID1},
						{ID: projectID2},
					}, 2, nil).
					Times(1)

				mockRepo.EXPECT().
					CountTasksByStatusAndProject(ctx, projectID1).
					Return([]tasks.StatusCount{
						{Status: "todo", Count: 3},
						{Status: "in_progress", Count: 2},
					}, nil).
					Times(1)

				mockRepo.EXPECT().
					CountTasksByStatusAndProject(ctx, projectID2).
					Return([]tasks.StatusCount{
						{Status: "todo", Count: 2},
						{Status: "in_review", Count: 2},
						{Status: "done", Count: 10},
					}, nil).
					Times(1)
			},
			expectedStats: &tasks.TaskStatistics{
				Todo:       5,
				InProgress: 2,
				InReview:   2,
				Done:       10,
			},
			expectedError: "",
		},
		{
			name: "success - returns zero statistics when no projects",
			setupMock: func() {
				mockProjectRepo.EXPECT().
					ListProjectByAccountID(ctx, accountID, uuid.Nil, 1000, 0).
					Return([]*projectEntity.Project{}, 0, nil).
					Times(1)
			},
			expectedStats: &tasks.TaskStatistics{
				Todo:       0,
				InProgress: 0,
				InReview:   0,
				Done:       0,
			},
			expectedError: "",
		},
		{
			name: "error - repository fails",
			setupMock: func() {
				mockProjectRepo.EXPECT().
					ListProjectByAccountID(ctx, accountID, uuid.Nil, 1000, 0).
					Return(nil, 0, errors.New("database error")).
					Times(1)
			},
			expectedStats: nil,
			expectedError: "failed to list user projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.GetTaskStatistics(ctx, accountID)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, res)
				assert.Equal(t, tt.expectedStats.Todo, res.Todo)
				assert.Equal(t, tt.expectedStats.InProgress, res.InProgress)
				assert.Equal(t, tt.expectedStats.InReview, res.InReview)
				assert.Equal(t, tt.expectedStats.Done, res.Done)
			}
		})
	}
}
