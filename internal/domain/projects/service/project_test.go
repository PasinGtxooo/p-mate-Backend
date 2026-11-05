package service

import (
	"context"
	"errors"
	"testing"

	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestProjectService_CreateProject(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProjectRepository(ctrl)
	mockTaskRepo := mocks.NewMockTaskRepository(ctrl)
	mockMemberRepo := mocks.NewMockProjectMemberRepository(ctrl)
	svc := NewProjectService(mockRepo, mockTaskRepo, mockMemberRepo)

	ctx := context.Background()

	validAccountID := "550e8400-e29b-41d4-a716-446655440000"

	tests := []struct {
		name          string
		request       *project.CreateProjectRequest
		setupMock     func()
		expectedError string
		expectNil     bool
		validate      func(t *testing.T, res *entity.Project)
	}{
		{
			name: "success - creates project with all fields",
			request: &project.CreateProjectRequest{
				AccountID: validAccountID,
				Name:      "Test Project",
				Config:    []byte(`{"color": "blue", "theme": "dark"}`),
			},
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProject(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, proj *entity.Project) error {
						assert.Equal(t, "Test Project", proj.Name)
						assert.NotEmpty(t, proj.ID)
						assert.NotEmpty(t, proj.OwnerID)
						return nil
					}).
					Times(1)
				mockMemberRepo.EXPECT().
					Create(ctx, gomock.Any()).
					DoAndReturn(func(_ context.Context, member *entity.ProjectMember) error {
						assert.Equal(t, entity.RoleOwner, member.Role)
						return nil
					}).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
			validate: func(t *testing.T, res *entity.Project) {
				assert.Equal(t, "Test Project", res.Name)
				assert.JSONEq(t, `{"color": "blue", "theme": "dark"}`, string(res.Config))
			},
		},
		{
			name: "success - creates project with minimal fields",
			request: &project.CreateProjectRequest{
				AccountID: validAccountID,
				Name:      "Minimal Project",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProject(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				mockMemberRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
			validate: func(t *testing.T, res *entity.Project) {
				assert.Equal(t, "Minimal Project", res.Name)
			},
		},
		{
			name:          "error - nil request",
			request:       nil,
			setupMock:     func() {},
			expectedError: "invalid request body",
			expectNil:     true,
			validate:      nil,
		},
		{
			name: "error - invalid account ID format (not UUID)",
			request: &project.CreateProjectRequest{
				AccountID: "invalid-uuid",
				Name:      "Test Project",
			},
			setupMock:     func() {},
			expectedError: "invalid account ID format",
			expectNil:     true,
			validate:      nil,
		},
		{
			name: "error - invalid account ID format (empty)",
			request: &project.CreateProjectRequest{
				AccountID: "",
				Name:      "Test Project",
			},
			setupMock:     func() {},
			expectedError: "invalid account ID format",
			expectNil:     true,
			validate:      nil,
		},
		{
			name: "error - invalid account ID format (partial UUID)",
			request: &project.CreateProjectRequest{
				AccountID: "550e8400-e29b-41d4",
				Name:      "Test Project",
			},
			setupMock:     func() {},
			expectedError: "invalid account ID format",
			expectNil:     true,
			validate:      nil,
		},
		{
			name: "error - repository create fails",
			request: &project.CreateProjectRequest{
				AccountID: validAccountID,
				Name:      "Test Project",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProject(ctx, gomock.Any()).
					Return(errors.New("database error")).
					Times(1)
			},
			expectedError: "failed to create project",
			expectNil:     true,
			validate:      nil,
		},
		{
			name: "error - add owner to members fails",
			request: &project.CreateProjectRequest{
				AccountID: validAccountID,
				Name:      "Test Project",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProject(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				mockMemberRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(errors.New("member creation error")).
					Times(1)
			},
			expectedError: "failed to add owner to project members",
			expectNil:     true,
			validate:      nil,
		},
		{
			name: "success - creates project with empty name",
			request: &project.CreateProjectRequest{
				AccountID: validAccountID,
				Name:      "",
			},
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProject(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				mockMemberRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
			validate: func(t *testing.T, res *entity.Project) {
				assert.Equal(t, "", res.Name)
			},
		},
		{
			name: "success - creates project with nil config",
			request: &project.CreateProjectRequest{
				AccountID: validAccountID,
				Name:      "No Config Project",
				Config:    nil,
			},
			setupMock: func() {
				mockRepo.EXPECT().
					CreateProject(ctx, gomock.Any()).
					Return(nil).
					Times(1)
				mockMemberRepo.EXPECT().
					Create(ctx, gomock.Any()).
					Return(nil).
					Times(1)
			},
			expectedError: "",
			expectNil:     false,
			validate: func(t *testing.T, res *entity.Project) {
				assert.Nil(t, res.Config)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			res, err := svc.CreateProject(ctx, tt.request, "550e8400-e29b-41d4-a716-446655440000")

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
