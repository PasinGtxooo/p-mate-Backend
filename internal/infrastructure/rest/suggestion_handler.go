package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion"
	"github.com/FrostBitzX/smart-task-ai/internal/application/suggestion/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type SuggestionHandler struct {
	SuggestTeammatesUC *usecase.SuggestTeammatesUseCase
	GetManpowerUC      *usecase.GetManpowerUseCase
	SetMemberSkillsUC  *usecase.SetMemberSkillsUseCase
	ListMemberSkillsUC *usecase.ListMemberSkillsUseCase
	SetMemberRoleUC    *usecase.SetMemberRoleUseCase
	logger             logger.Logger
}

func NewSuggestionHandler(
	suggestTeammates *usecase.SuggestTeammatesUseCase,
	getManpower *usecase.GetManpowerUseCase,
	setMemberSkills *usecase.SetMemberSkillsUseCase,
	listMemberSkills *usecase.ListMemberSkillsUseCase,
	setMemberRole *usecase.SetMemberRoleUseCase,
	l logger.Logger,
) *SuggestionHandler {
	return &SuggestionHandler{
		SuggestTeammatesUC: suggestTeammates,
		GetManpowerUC:      getManpower,
		SetMemberSkillsUC:  setMemberSkills,
		ListMemberSkillsUC: listMemberSkills,
		SetMemberRoleUC:    setMemberRole,
		logger:             l,
	}
}

func (h *SuggestionHandler) claims(c *fiber.Ctx) (nodeID, accountID string, err error) {
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		return "", "", apperror.ErrUnauthorized
	}
	nodeID, ok = jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		return "", "", apperror.ErrUnauthorized
	}
	accountID, ok = jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		return "", "", apperror.ErrUnauthorized
	}
	return nodeID, accountID, nil
}

// SuggestTeammates ranks members for a task by skill and free capacity.
func (h *SuggestionHandler) SuggestTeammates(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	if taskID == "" {
		return responses.Error(c, apperror.NewBadRequestError("task ID is required", "INVALID_TASK_ID", nil))
	}
	nodeID, accountID, err := h.claims(c)
	if err != nil {
		return responses.Error(c, err)
	}
	data, err := h.SuggestTeammatesUC.Execute(c.Context(), taskID, nodeID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}
	return responses.Success(c, data, "Teammate suggestions retrieved successfully")
}

// GetManpower reports whether the project team can absorb all open work.
func (h *SuggestionHandler) GetManpower(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID is required", "INVALID_PROJECT_ID", nil))
	}
	nodeID, accountID, err := h.claims(c)
	if err != nil {
		return responses.Error(c, err)
	}
	data, err := h.GetManpowerUC.Execute(c.Context(), projectID, nodeID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}
	return responses.Success(c, data, "Manpower analysis retrieved successfully")
}

// SetMemberSkills lets a project owner set a member's skills.
func (h *SuggestionHandler) SetMemberSkills(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	targetAccountID := c.Params("accountId")
	if projectID == "" || targetAccountID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID and account ID are required", "INVALID_PARAMS", nil))
	}
	nodeID, callerAccountID, err := h.claims(c)
	if err != nil {
		return responses.Error(c, err)
	}

	var req suggestion.SetMemberSkillsRequest
	if err := c.BodyParser(&req); err != nil {
		return responses.Error(c, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", err))
	}

	if err := h.SetMemberSkillsUC.Execute(c.Context(), projectID, targetAccountID, nodeID, callerAccountID, &req); err != nil {
		return responses.Error(c, err)
	}
	return responses.Success(c, fiber.Map{"account_id": targetAccountID}, "Member skills updated successfully")
}

// ListMemberSkills returns every member of the project with their skills.
func (h *SuggestionHandler) ListMemberSkills(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID is required", "INVALID_PROJECT_ID", nil))
	}
	nodeID, accountID, err := h.claims(c)
	if err != nil {
		return responses.Error(c, err)
	}
	data, err := h.ListMemberSkillsUC.Execute(c.Context(), projectID, nodeID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}
	return responses.Success(c, data, "Member skills retrieved successfully")
}

// SetMemberRole lets the project owner promote a member to PM or back to member.
func (h *SuggestionHandler) SetMemberRole(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	targetAccountID := c.Params("accountId")
	if projectID == "" || targetAccountID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID and account ID are required", "INVALID_PARAMS", nil))
	}
	nodeID, callerAccountID, err := h.claims(c)
	if err != nil {
		return responses.Error(c, err)
	}

	var req suggestion.SetMemberRoleRequest
	if err := c.BodyParser(&req); err != nil {
		return responses.Error(c, apperror.NewBadRequestError("invalid request body", "INVALID_REQUEST", err))
	}

	if err := h.SetMemberRoleUC.Execute(c.Context(), projectID, targetAccountID, nodeID, callerAccountID, &req); err != nil {
		return responses.Error(c, err)
	}
	return responses.Success(c, fiber.Map{"account_id": targetAccountID, "role": req.Role}, "Member role updated successfully")
}
