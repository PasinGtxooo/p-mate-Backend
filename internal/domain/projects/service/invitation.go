package service

import (
	"context"
	"errors"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/accounts"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects"
	"github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
)

type InvitationService struct {
	invitationRepo projects.InvitationRepository
	accountRepo    accounts.AccountRepository
	memberRepo     projects.ProjectMemberRepository
	projectRepo    projects.ProjectRepository
}

func NewInvitationService(invitationRepo projects.InvitationRepository, accountRepo accounts.AccountRepository, memberRepo projects.ProjectMemberRepository, projectRepo projects.ProjectRepository) *InvitationService {
	return &InvitationService{
		invitationRepo: invitationRepo,
		accountRepo:    accountRepo,
		memberRepo:     memberRepo,
		projectRepo:    projectRepo,
	}
}

// CreateInvitation represents creates an invitation
func (s *InvitationService) CreateInvitation(ctx context.Context, projectID uuid.UUID, inviterAccountID uuid.UUID, inviteeShortID string, _ string, nodeID uuid.UUID) (*entity.ProjectInvitation, error) {
	isMember, err := s.memberRepo.IsMember(ctx, projectID, inviterAccountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	role := string(entity.RoleMember) // Member

	// Parse short ID to UUID
	inviteeID, err := utils.ParseID(inviteeShortID, "acc")
	if err != nil {
		return nil, apperror.NewBadRequestError("invalid account short ID format", "INVALID_SHORT_ID", err)
	}

	invitee, err := s.accountRepo.GetAccount(ctx, inviteeID.String())
	if err != nil {
		return nil, apperror.NewNotFoundError("account not found", "ACCOUNT_NOT_FOUND", err)
	}

	isAlreadyMember, err := s.memberRepo.IsMember(ctx, projectID, invitee.ID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if isAlreadyMember {
		return nil, apperror.NewConflictError("user is already a member of this project", "ALREADY_MEMBER", nil)
	}

	hasPendingInvitation, err := s.invitationRepo.ExistsPendingInvitation(ctx, projectID, invitee.ID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to check pending invitation", "CHECK_INVITATION_ERROR", err)
	}
	if hasPendingInvitation {
		return nil, apperror.NewConflictError("pending invitation already exists", "PENDING_INVITATION_EXISTS", nil)
	}

	now := time.Now()
	expiresAt := now.AddDate(0, 0, 30) // 30 days

	invitation := &entity.ProjectInvitation{
		ID:               uuid.New(),
		NodeID:           nodeID,
		ProjectID:        projectID,
		InviterAccountID: inviterAccountID,
		InviteeAccountID: invitee.ID,
		Role:             role,
		Status:           "pending",
		CreatedAt:        now,
		ExpiresAt:        expiresAt,
		UpdatedAt:        now,
	}

	err = s.invitationRepo.Create(ctx, invitation)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to create invitation", "CREATE_INVITATION_ERROR", err)
	}

	project, err := s.projectRepo.GetProjectByID(ctx, projectID, nodeID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return nil, apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return nil, apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	inviter, err := s.accountRepo.GetAccount(ctx, inviterAccountID.String())
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to get inviter details", "GET_INVITER_ERROR", err)
	}

	invitation.Project = *project
	invitation.Inviter = *inviter
	invitation.Invitee = *invitee

	return invitation, nil
}

// AcceptInvitation represent accepts an invitation
func (s *InvitationService) AcceptInvitation(ctx context.Context, projectID uuid.UUID, inviteeAccountID uuid.UUID, nodeID uuid.UUID) error {
	inv, err := s.invitationRepo.FindByProjectAndInvitee(ctx, projectID, inviteeAccountID)
	if err != nil {
		return apperror.NewNotFoundError("invitation not found", "INVITATION_NOT_FOUND", err)
	}

	if inv.InviteeAccountID != inviteeAccountID {
		return apperror.NewForbiddenError("you are not authorized to accept this invitation", "UNAUTHORIZED", nil)
	}

	if inv.Status != "pending" {
		return apperror.NewConflictError("invitation has already been processed", "INVITATION_ALREADY_PROCESSED", nil)
	}

	now := time.Now()
	if now.After(inv.ExpiresAt) {
		return apperror.NewConflictError("invitation has expired", "INVITATION_EXPIRED", nil)
	}

	// Get project to use its nodeID (not the invitee's nodeID from JWT)
	project, err := s.projectRepo.GetProjectByID(ctx, projectID, inv.NodeID)
	if err != nil {
		if errors.Is(err, apperror.ErrRecordNotFound) {
			return apperror.NewNotFoundError("project not found", "PROJECT_NOT_FOUND", err)
		}
		return apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
	}

	inv.Status = "accepted"
	inv.RespondedAt = &now
	inv.UpdatedAt = now

	err = s.invitationRepo.UpdateStatus(ctx, inv)
	if err != nil {
		return apperror.NewInternalServerError("failed to accept invitation", "UPDATE_INVITATION_ERROR", err)
	}

	newMember := &entity.ProjectMember{
		ProjectID: projectID,
		AccountID: inviteeAccountID,
		NodeID:    project.NodeID,
		Role:      entity.ProjectRole(inv.Role),
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = s.memberRepo.Create(ctx, newMember)
	if err != nil {
		return apperror.NewInternalServerError("failed to add member to project", "CREATE_MEMBER_ERROR", err)
	}

	return nil
}

// RejectInvitation represent rejects an invitation
func (s *InvitationService) RejectInvitation(ctx context.Context, projectID uuid.UUID, inviteeAccountID uuid.UUID) error {
	inv, err := s.invitationRepo.FindByProjectAndInvitee(ctx, projectID, inviteeAccountID)
	if err != nil {
		return apperror.NewNotFoundError("invitation not found", "INVITATION_NOT_FOUND", err)
	}

	if inv.InviteeAccountID != inviteeAccountID {
		return apperror.NewForbiddenError("you are not authorized to reject this invitation", "UNAUTHORIZED", nil)
	}

	if inv.Status != "pending" {
		return apperror.NewConflictError("invitation has already been processed", "INVITATION_ALREADY_PROCESSED", nil)
	}

	now := time.Now()
	if now.After(inv.ExpiresAt) {
		return apperror.NewConflictError("invitation has expired", "INVITATION_EXPIRED", nil)
	}

	inv.Status = "rejected"
	inv.RespondedAt = &now
	inv.UpdatedAt = now

	err = s.invitationRepo.UpdateStatus(ctx, inv)
	if err != nil {
		return apperror.NewInternalServerError("failed to reject invitation", "UPDATE_INVITATION_ERROR", err)
	}

	return nil
}

// CancelInvitation represents cancels an invitation
func (s *InvitationService) CancelInvitation(ctx context.Context, projectID uuid.UUID, inviteeAccountID uuid.UUID, cancellerID uuid.UUID) error {
	isMember, err := s.memberRepo.IsMember(ctx, projectID, cancellerID)
	if err != nil {
		return apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	inv, err := s.invitationRepo.FindByProjectAndInvitee(ctx, projectID, inviteeAccountID)
	if err != nil {
		return apperror.NewNotFoundError("invitation not found", "INVITATION_NOT_FOUND", err)
	}

	if inv.Status != "pending" {
		return apperror.NewConflictError("invitation has already been processed", "INVITATION_ALREADY_PROCESSED", nil)
	}

	err = s.invitationRepo.Delete(ctx, projectID, inviteeAccountID)
	if err != nil {
		return apperror.NewInternalServerError("failed to cancel invitation", "CANCEL_INVITATION_ERROR", err)
	}

	return nil
}

// ListMyInvitations represents lists all invitations for a user
func (s *InvitationService) ListMyInvitations(ctx context.Context, inviteeAccountID uuid.UUID) ([]*entity.ProjectInvitation, error) {
	invitations, err := s.invitationRepo.ListByInvitee(ctx, inviteeAccountID, "pending")
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list invitations", "LIST_INVITATIONS_ERROR", err)
	}

	return invitations, nil
}

// ListProjectInvitations represents lists all invitations for a project
func (s *InvitationService) ListProjectInvitations(ctx context.Context, projectID uuid.UUID, accountID uuid.UUID) ([]*entity.ProjectInvitation, error) {
	isMember, err := s.memberRepo.IsMember(ctx, projectID, accountID)
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to verify membership", "CHECK_MEMBER_ERROR", err)
	}
	if !isMember {
		return nil, apperror.NewForbiddenError("you are not a member of this project", "NOT_PROJECT_MEMBER", nil)
	}

	invitations, err := s.invitationRepo.ListByProject(ctx, projectID, "pending")
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to list project invitations", "LIST_INVITATIONS_ERROR", err)
	}

	return invitations, nil
}
