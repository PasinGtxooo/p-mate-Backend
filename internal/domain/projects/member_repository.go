//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../../mocks/member_repository.go -package=mocks
package projects

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
)

// ProjectMemberRepository defines the interface for managing project members
type ProjectMemberRepository interface {
	Create(ctx context.Context, member *entity.ProjectMember) error
	IsMember(ctx context.Context, projectID, accountID uuid.UUID) (bool, error)
}
