package usecase

import (
	"context"
	"errors"
	"testing"

	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/mocks"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// mockLogger is a simple logger implementation for testing
type mockLogger struct{}

func (m *mockLogger) Info(msg string, fields ...map[string]interface{})  {}
func (m *mockLogger) Error(msg string, fields ...map[string]interface{}) {}
func (m *mockLogger) Debug(msg string, fields ...map[string]interface{}) {}
func (m *mockLogger) Warn(msg string, fields ...map[string]interface{})  {}
func (m *mockLogger) With(fields map[string]interface{}) logger.Logger   { return m }

func TestGetTaskStatisticsUseCase_Execute(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockTaskRepository(ctrl)
	mockProjectRepo := mocks.NewMockProjectRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	taskService := service.NewTaskService(mockRepo, mockProjectRepo, mockMemberRepo)
	logger := &mockLogger{}
	uc := NewGetTaskStatisticsUseCase(taskService, logger)
	ctx := context.Background()
	nodeIDStr := "550e8400-e29b-41d4-a716-446655440000"
	nodeID := uuid.MustParse(nodeIDStr)
	projectID1 := uuid.New()
	projectID2 := uuid.New()

	tests := []struct {
		name          string
		nodeIDStr     string
		setupMock     func()
		expectedStats *struct {
			Todo       int64
			InProgress int64
			InReview   int64
			Done       int64
		}
		expectedError string
	}{
		{
			name:      "success - returns statistics with all statuses from multiple projects",
			nodeIDStr: nodeIDStr,
			setupMock: func() {
				// Mock ListProjectByAccountID
				mockProjectRepo.EXPECT().
					ListProjectByAccountID(ctx, nodeID, uuid.Nil, 1000, 0).
					Return([]*projectEntity.Project{
						{ID: projectID1, NodeID: nodeID},
						{ID: projectID2, NodeID: nodeID},
					}, 2, nil).
					Times(1)

				// Mock CountTasksByStatusAndProject for project 1
				mockRepo.EXPECT().
					CountTasksByStatusAndProject(ctx, projectID1).
					Return([]tasks.StatusCount{
						{Status: "todo", Count: 3},
						{Status: "in_progress", Count: 2},
					}, nil).
					Times(1)

				// Mock CountTasksByStatusAndProject for project 2
				mockRepo.EXPECT().
					CountTasksByStatusAndProject(ctx, projectID2).
					Return([]tasks.StatusCount{
						{Status: "todo", Count: 2},
						{Status: "in_review", Count: 2},
						{Status: "done", Count: 10},
					}, nil).
					Times(1)
			},
			expectedStats: &struct {
				Todo       int64
				InProgress int64
				InReview   int64
				Done       int64
			}{
				Todo:       5,
				InProgress: 2,
				InReview:   2,
				Done:       10,
			},
			expectedError: "",
		},
		{
			name:      "success - returns zero statistics when no projects",
			nodeIDStr: nodeIDStr,
			setupMock: func() {
				mockProjectRepo.EXPECT().
					ListProjectByAccountID(ctx, nodeID, uuid.Nil, 1000, 0).
					Return([]*projectEntity.Project{}, 0, nil).
					Times(1)
			},
			expectedStats: &struct {
				Todo       int64
				InProgress int64
				InReview   int64
				Done       int64
			}{
				Todo:       0,
				InProgress: 0,
				InReview:   0,
				Done:       0,
			},
			expectedError: "",
		},
		{
			name:      "error - invalid node ID format",
			nodeIDStr: "invalid-uuid",
			setupMock: func() {
				// No mock setup needed as validation happens before repository call
			},
			expectedStats: nil,
			expectedError: "invalid node ID format",
		},
		{
			name:      "error - repository fails",
			nodeIDStr: nodeIDStr,
			setupMock: func() {
				mockProjectRepo.EXPECT().
					ListProjectByAccountID(ctx, nodeID, uuid.Nil, 1000, 0).
					Return(nil, 0, apperror.NewInternalServerError("database error", "DB_ERROR", errors.New("connection failed"))).
					Times(1)
			},
			expectedStats: nil,
			expectedError: "failed to list user projects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := uc.Execute(ctx, tt.nodeIDStr)

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
