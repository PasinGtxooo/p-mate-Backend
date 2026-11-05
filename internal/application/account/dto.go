package account

import "github.com/FrostBitzX/smart-task-ai/internal/application/common"

type CreateAccountRequest struct {
	Username        string `json:"username" validate:"required,min=3,max=20"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required,min=4"`
	ConfirmPassword string `json:"confirm_password" validate:"required,min=4,eqfield=Password"`
}

type ListAccountsRequest struct {
	Limit  *int `query:"limit" validate:"omitempty,min=1,max=100"`
	Offset *int `query:"offset" validate:"omitempty,min=0"`
}

type CreateAccountResponse struct {
	AccountID string `json:"account_id"`
}

type AccountDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Status   string `json:"status"`
}

type ListAccountsResponse struct {
	Items      []AccountDTO      `json:"items"`
	Pagination common.Pagination `json:"pagination"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}
