package project

import (
	"encoding/json"
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/common"
)

type CreateProjectRequest struct {
	AccountID string          `json:"account_id"`
	Name      string          `json:"name" validate:"required,max=50"`
	Config    json.RawMessage `json:"config"`
}

type ListProjectRequest struct {
	AccountID string `query:"account_id" validate:"omitempty"`
	Limit     *int   `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset    *int   `query:"offset" validate:"omitempty,min=0"`
}

type CreateProjectResponse struct {
	ProjectID string `json:"project_id"`
}

type ProjectResponse struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Role      string          `json:"role"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type ListProjectResponse struct {
	Items      []ProjectResponse `json:"items"`
	Pagination common.Pagination `json:"pagination"`
}

type UpdateProjectRequest struct {
	ProjectID string          `json:"-"`
	Name      string          `json:"name" validate:"required,max=50"`
	Config    json.RawMessage `json:"config" validate:"omitempty"`
}

type UpdateProjectResponse struct {
	ID        string          `json:"id"`
	Name      string          `json:"name"`
	Config    json.RawMessage `json:"config"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

type DeleteProjectResponse struct {
	ProjectID string `json:"project_id"`
}

type AddMemberRequest struct {
	ProjectID string `json:"-"`
	Email     string `json:"email" validate:"required,email"`
}

type AddMemberResponse struct {
	ProjectID string `json:"project_id"`
	AccountID string `json:"account_id"`
	Role      string `json:"role"`
}

type RemoveMemberRequest struct {
	ProjectID string `json:"-"`
	AccountID string `json:"-"`
}

type RemoveMemberResponse struct {
	ProjectID string `json:"project_id"`
	AccountID string `json:"account_id"`
}

type ProjectMemberResponse struct {
	AccountID string    `json:"account_id"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}

type ListProjectMembersResponse struct {
	Members []ProjectMemberResponse `json:"members"`
}
