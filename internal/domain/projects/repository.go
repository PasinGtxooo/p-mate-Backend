//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../../mocks/project_repository.go -package=mocks
package projects

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
)

type ProjectRepository interface {
	CreateProject(ctx context.Context, proj *entity.Project) error
	GetProjectByID(ctx context.Context, projectID uuid.UUID, nodeID uuid.UUID) (*entity.Project, error)
	ListProjectByAccountID(ctx context.Context, accountID uuid.UUID, nodeID uuid.UUID, limit, offset int) ([]*entity.Project, int, error)
	UpdateProject(ctx context.Context, proj *entity.Project, nodeID uuid.UUID) error
	DeleteProject(ctx context.Context, projectID uuid.UUID, nodeID uuid.UUID) error
	AddMember(ctx context.Context, member *entity.ProjectMember) error
	RemoveMember(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID uuid.UUID) error
	GetProjectMember(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID uuid.UUID) (*entity.ProjectMember, error)
	ListProjectMembers(ctx context.Context, projectID uuid.UUID, nodeID uuid.UUID) ([]*entity.ProjectMember, error)
	UpdateMemberRole(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID, nodeID uuid.UUID, role entity.ProjectRole) error
}
