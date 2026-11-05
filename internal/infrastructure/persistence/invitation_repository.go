package persistence

import (
	"context"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type invitationRepository struct {
	db *gorm.DB
}

func NewInvitationRepository(db *gorm.DB) projects.InvitationRepository {
	return &invitationRepository{db: db}
}

// Create creates a new project invitation
func (r *invitationRepository) Create(ctx context.Context, invitation *entity.ProjectInvitation) error {
	return r.db.WithContext(ctx).Create(invitation).Error
}

// FindByProjectAndInvitee finds an invitation by project and invitee IDs
func (r *invitationRepository) FindByProjectAndInvitee(ctx context.Context, projectID, inviteeID uuid.UUID) (*entity.ProjectInvitation, error) {
	var invitation entity.ProjectInvitation
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Inviter").
		Preload("Invitee").
		Where("project_id = ? AND invitee_account_id = ? AND status = ?", projectID, inviteeID, "pending").
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

// UpdateStatus updates the status and responded_at of an invitation
func (r *invitationRepository) UpdateStatus(ctx context.Context, invitation *entity.ProjectInvitation) error {
	return r.db.WithContext(ctx).
		Model(&entity.ProjectInvitation{}).
		Where("project_id = ? AND invitee_account_id = ?", invitation.ProjectID, invitation.InviteeAccountID).
		Updates(map[string]interface{}{
			"status":       invitation.Status,
			"responded_at": invitation.RespondedAt,
		}).Error
}

// Delete performs GORM soft delete by setting deleted_at
func (r *invitationRepository) Delete(ctx context.Context, projectID, inviteeID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND invitee_account_id = ?", projectID, inviteeID).
		Delete(&entity.ProjectInvitation{}).Error
}

// ListByInvitee lists invitations for an invitee with status and expiration filtering
func (r *invitationRepository) ListByInvitee(ctx context.Context, inviteeID uuid.UUID, status string) ([]*entity.ProjectInvitation, error) {
	var invitations []*entity.ProjectInvitation
	query := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Inviter").
		Where("invitee_account_id = ?", inviteeID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Filter out expired invitations for pending status
	if status == "pending" {
		query = query.Where("expires_at > ?", time.Now())
	}

	err := query.Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

// ListByProject lists invitations for a project with status filtering
func (r *invitationRepository) ListByProject(ctx context.Context, projectID uuid.UUID, status string) ([]*entity.ProjectInvitation, error) {
	var invitations []*entity.ProjectInvitation
	query := r.db.WithContext(ctx).
		Preload("Invitee").
		Preload("Inviter").
		Where("project_id = ?", projectID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Order("created_at DESC").Find(&invitations).Error
	return invitations, err
}

// ExistsPendingInvitation checks if a pending invitation exists
func (r *invitationRepository) ExistsPendingInvitation(ctx context.Context, projectID, inviteeID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.ProjectInvitation{}).
		Where("project_id = ? AND invitee_account_id = ? AND status = ?", projectID, inviteeID, "pending").
		Count(&count).Error
	return count > 0, err
}
