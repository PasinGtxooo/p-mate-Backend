package service

import (
	"context"
	"errors"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	pkgutils "github.com/FrostBitzX/smart-task-ai/pkg/utils"
	"github.com/google/uuid"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/tasks"
)

type ProjectService struct {
	repo       projects.ProjectRepository
	taskRepo   tasks.TaskRepository
	memberRepo projects.ProjectMemberRepository
}

func NewProjectService(repo projects.ProjectRepository, taskRepo tasks.TaskRepository, memberRepo projects.ProjectMemberRepository) *ProjectService {
	return &ProjectService{
		repo:       repo,
		taskRepo:   taskRepo,
		memberRepo: memberRepo,
	}
}

func (s *ProjectService) CreateProject(ctx context.Context, req *project.CreateProjectRequest, nodeID string) (*entity.Project, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	// create domain entity
	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account ID format", "INVALID_ACCOUNT_ID", err)
	}

	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	now := time.Now()
	proj := &entity.Project{
		ID:        uuid.New(),
		NodeID:    nodeUUID,
		OwnerID:   accountID,
		Name:      req.Name,
		Config:    req.Config,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// persist project to database
	err = s.repo.CreateProject(ctx, proj)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to create project", "CREATE_PROJECT_ERROR", err)
	}

	// Add owner to project_members table
	ownerMember := &entity.ProjectMember{
		ProjectID: proj.ID,
		AccountID: accountID,
		NodeID:    nodeUUID,
		Role:      entity.RoleOwner,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.memberRepo.Create(ctx, ownerMember)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to add owner to project members", "ADD_OWNER_ERROR", err)
	}

	return proj, nil
}

func (s *ProjectService) GetProjectByID(ctx context.Context, projectID uuid.UUID, nodeID string, accountID uuid.UUID) (*entity.Project, error) {
	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	proj, err := s.repo.GetProjectByID(ctx, projectID, nodeUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	// Check if user has access to this project
	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to check membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you don't have access to this project", "NO_ACCESS", nil)
	}

	return proj, nil
}

func (s *ProjectService) ListProjectByAccountID(ctx context.Context, accountID uuid.UUID, nodeID string, limit, offset int) ([]*entity.Project, int, error) {
	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, 0, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	projs, total, err := s.repo.ListProjectByAccountID(ctx, accountID, nodeUUID, limit, offset)
	if err != nil {
		return nil, 0, apperror.NewInternalServerError("failed to list projects", "LIST_PROJECT_ERROR", err)
	}

	return projs, total, nil
}

func (s *ProjectService) UpdateProject(ctx context.Context, req *project.UpdateProjectRequest, nodeID string) (*entity.Project, error) {
	if req == nil {
		return nil, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", nil)
	}

	projectID, err := utils.ParseID(req.ProjectID, entity.ProjectIDPrefix)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid project ID format", "INVALID_PROJECT_ID", err)
	}

	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Get existing project
	proj, err := s.repo.GetProjectByID(ctx, projectID, nodeUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	proj.Name = req.Name

	// Nullable field
	pkgutils.UpdateNullableJSON((*[]byte)(&proj.Config), req.Config)

	proj.UpdatedAt = time.Now()

	err = s.repo.UpdateProject(ctx, proj, nodeUUID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to update project", "UPDATE_PROJECT_ERROR", err)
	}

	return proj, nil
}

func (s *ProjectService) DeleteProject(ctx context.Context, projectID uuid.UUID, nodeID string) error {
	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Get existing project
	_, err = s.repo.GetProjectByID(ctx, projectID, nodeUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	if err := s.deleteProjectCheck(ctx, projectID); err != nil {
		return err
	}

	err = s.repo.DeleteProject(ctx, projectID, nodeUUID)
	if err != nil {
		return apperror.NewInternalServerError("failed to delete project", "DELETE_PROJECT_ERROR", err)
	}

	return nil
}

func (s *ProjectService) deleteProjectCheck(ctx context.Context, projectID uuid.UUID) error {
	count, err := s.taskRepo.CountTasksByProject(ctx, projectID)
	if err != nil {
		return apperror.NewInternalServerError("failed to check tasks in project", "CHECK_TASKS_ERROR", err)
	}

	if count > 0 {
		return apperror.NewBadRequestError("cannot delete project with existing tasks", "PROJECT_HAS_TASKS", nil)
	}

	return nil
}

func (s *ProjectService) AddMember(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, role string, nodeID string) (*entity.ProjectMember, error) {
	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Check if project exists
	_, err = s.repo.GetProjectByID(ctx, projectID, nodeUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	// Check if member already exists
	_, err = s.repo.GetProjectMember(ctx, projectID, accountID, nodeUUID)
	if err == nil {
		return nil, apperror.NewBadRequestError("member already exists in project", "MEMBER_ALREADY_EXISTS", nil)
	}
	if !errors.Is(err, apperror.ErrRecordNotFound) {
		return nil, apperror.NewInternalServerError("failed to check member", "CHECK_MEMBER_ERROR", err)
	}

	projectRole := entity.RoleMember

	now := time.Now()
	member := &entity.ProjectMember{
		ProjectID: projectID,
		AccountID: accountID,
		NodeID:    nodeUUID,
		Role:      projectRole,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.repo.AddMember(ctx, member)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to add member", "ADD_MEMBER_ERROR", err)
	}

	return member, nil
}

func (s *ProjectService) RemoveMember(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID string) error {
	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Check if member exists
	member, err := s.repo.GetProjectMember(ctx, projectID, accountID, nodeUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return apperror.NewNotFoundError("member not found in project", "MEMBER_NOT_FOUND", err)
		}
		return apperror.NewInternalServerError("failed to get member", "GET_MEMBER_ERROR", err)
	}

	// Prevent removing owner
	if member.Role == entity.RoleOwner {
		return apperror.NewBadRequestError("cannot remove project owner", "CANNOT_REMOVE_OWNER", nil)
	}

	err = s.repo.RemoveMember(ctx, projectID, accountID, nodeUUID)
	if err != nil {
		return apperror.NewInternalServerError("failed to remove member", "REMOVE_MEMBER_ERROR", err)
	}

	return nil
}

func (s *ProjectService) ListProjectMembers(ctx context.Context, projectID uuid.UUID, nodeID string, accountID uuid.UUID) ([]*entity.ProjectMember, error) {
	// Parse nodeID
	nodeUUID, err := uuid.Parse(nodeID)
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid node ID format", "INVALID_NODE_ID", err)
	}

	// Check if project exists
	_, err = s.repo.GetProjectByID(ctx, projectID, nodeUUID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	// Check if user has access to this project
	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to check membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you don't have access to this project", "NO_ACCESS", nil)
	}

	members, err := s.repo.ListProjectMembers(ctx, projectID, nodeUUID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list members", "LIST_MEMBERS_ERROR", err)
	}

	return members, nil
}
