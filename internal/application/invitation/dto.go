package invitation

import (
	"time"
)

// CreateInvitationRequest represents the request to create a new project invitation
type CreateInvitationRequest struct {
	ProjectID        string `json:"-"`
	InviterAccountID string `json:"-"`
	InviteeShortID   string `json:"invitee_short_id" validate:"required,startswith=acc_"`
}

// CreateInvitationResponse represents the response after creating an invitation
type CreateInvitationResponse struct {
	InvitationID     string    `json:"invitation_id"`
	ProjectID        string    `json:"project_id"`
	ProjectName      string    `json:"project_name"`
	InviterAccountID string    `json:"inviter_account_id"`
	InviterName      string    `json:"inviter_name"`
	InviteeAccountID string    `json:"invitee_account_id"`
	InviteeShortID   string    `json:"invitee_short_id"`
	InviteeName      string    `json:"invitee_name"`
	Role             string    `json:"role"`
	Status           string    `json:"status"`
	CreatedAt        time.Time `json:"created_at"`
	ExpiresAt        time.Time `json:"expires_at"`
}

// AcceptInvitationRequest represents the request to accept an invitation
type AcceptInvitationRequest struct {
	ProjectID        string `json:"-"`
	InviteeAccountID string `json:"-"`
}

// AcceptInvitationResponse represents the response after accepting an invitation
type AcceptInvitationResponse struct {
	Message string `json:"message"`
}

// RejectInvitationRequest represents the request to reject an invitation
type RejectInvitationRequest struct {
	ProjectID        string `json:"-"`
	InviteeAccountID string `json:"-"`
}

// RejectInvitationResponse represents the response after rejecting an invitation
type RejectInvitationResponse struct {
	Message string `json:"message"`
}

// CancelInvitationRequest represents the request to cancel an invitation
type CancelInvitationRequest struct {
	ProjectID        string `json:"-"`
	InviteeAccountID string `json:"-"`
	CancellerID      string `json:"-"`
}

// CancelInvitationResponse represents the response after cancelling an invitation
type CancelInvitationResponse struct {
	Message string `json:"message"`
}

// InvitationResponse represents a single invitation in list responses
type InvitationResponse struct {
	InvitationID     string     `json:"invitation_id"`
	ProjectID        string     `json:"project_id"`
	ProjectName      string     `json:"project_name"`
	InviterAccountID string     `json:"inviter_account_id"`
	InviterName      string     `json:"inviter_name"`
	InviteeAccountID string     `json:"invitee_account_id"`
	InviteeShortID   string     `json:"invitee_short_id"`
	InviteeName      string     `json:"invitee_name"`
	Role             string     `json:"role"`
	Status           string     `json:"status"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RespondedAt      *time.Time `json:"responded_at,omitempty"`
}

// ListMyInvitationsRequest represents the request to list invitations for the current user
type ListMyInvitationsRequest struct {
	InviteeAccountID string `json:"-"`
}

// ListMyInvitationsResponse represents the response with list of invitations
type ListMyInvitationsResponse struct {
	Invitations []InvitationResponse `json:"invitations"`
	Total       int                  `json:"total"`
}

// ListProjectInvitationsRequest represents the request to list invitations for a project
type ListProjectInvitationsRequest struct {
	ProjectID string `json:"-"`
	AccountID string `json:"-"`
}

// ListProjectInvitationsResponse represents the response with list of project invitations
type ListProjectInvitationsResponse struct {
	Invitations []InvitationResponse `json:"invitations"`
	Total       int                  `json:"total"`
}
