//go:generate go run go.uber.org/mock/mockgen -source=$GOFILE -destination=../../mocks/invitation_repository.go -package=mocks
package projects

import (
	"context"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
)

// InvitationRepository defines the interface for managing project invitations
type InvitationRepository interface {
	Create(ctx context.Context, invitation *entity.ProjectInvitation) error
	FindByProjectAndInvitee(ctx context.Context, projectID, inviteeAccountID uuid.UUID) (*entity.ProjectInvitation, error)
	UpdateStatus(ctx context.Context, invitation *entity.ProjectInvitation) error
	Delete(ctx context.Context, projectID, inviteeAccountID uuid.UUID) error
	ListByInvitee(ctx context.Context, inviteeAccountID uuid.UUID, status string) ([]*entity.ProjectInvitation, error)
	ListByProject(ctx context.Context, projectID uuid.UUID, status string) ([]*entity.ProjectInvitation, error)
	ExistsPendingInvitation(ctx context.Context, projectID, inviteeAccountID uuid.UUID) (bool, error)
}
