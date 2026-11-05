package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/project"
	"github.com/FrostBitzX/smart-task-ai/internal/application/project/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ProjectHandler struct {
	CreateProjectUC        *usecase.CreateProjectUseCase
	ListProjectByAccountUC *usecase.ListProjectByAccountUseCase
	GetProjectByIDUC       *usecase.GetProjectByIDUseCase
	UpdateProjectUC        *usecase.UpdateProjectUseCase
	DeleteProjectUC        *usecase.DeleteProjectUseCase
	AddMemberUC            *usecase.AddMemberUseCase
	RemoveMemberUC         *usecase.RemoveMemberUseCase
	ListMembersUC          *usecase.ListProjectMembersUseCase
	logger                 logger.Logger
}

func NewProjectHandler(
	create *usecase.CreateProjectUseCase,
	list *usecase.ListProjectByAccountUseCase,
	get *usecase.GetProjectByIDUseCase,
	update *usecase.UpdateProjectUseCase,
	delete *usecase.DeleteProjectUseCase,
	addMember *usecase.AddMemberUseCase,
	removeMember *usecase.RemoveMemberUseCase,
	listMembers *usecase.ListProjectMembersUseCase,
	l logger.Logger,
) *ProjectHandler {
	return &ProjectHandler{
		CreateProjectUC:        create,
		ListProjectByAccountUC: list,
		GetProjectByIDUC:       get,
		UpdateProjectUC:        update,
		DeleteProjectUC:        delete,
		AddMemberUC:            addMember,
		RemoveMemberUC:         removeMember,
		ListMembersUC:          listMembers,
		logger:                 l,
	}
}

func (h *ProjectHandler) CreateProject(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[project.CreateProjectRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	// Get AccountID and NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	// Set AccountID from JWT
	req.AccountID = accountID

	data, err := h.CreateProjectUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Project created successfully")
}

func (h *ProjectHandler) ListProject(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidateQuery[project.ListProjectRequest](c)
	if err != nil {
		h.logger.Warn("Invalid query parameters", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	// Get AccountID and NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	req.AccountID = accountID

	data, err := h.ListProjectByAccountUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "List Projects successfully")
}

func (h *ProjectHandler) GetProject(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	// Get NodeID and AccountID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	data, err := h.GetProjectByIDUC.Execute(c.Context(), projectID, nodeID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Project retrieved successfully")
}

func (h *ProjectHandler) UpdateProject(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	req, err := requests.ParseAndValidate[project.UpdateProjectRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	// Get NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	req.ProjectID = projectID

	data, err := h.UpdateProjectUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Project updated successfully")
}

func (h *ProjectHandler) DeleteProject(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	// Get NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	data, err := h.DeleteProjectUC.Execute(c.Context(), projectID, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Project deleted successfully")
}

func (h *ProjectHandler) AddMember(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	req, err := requests.ParseAndValidate[project.AddMemberRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	// Get NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	req.ProjectID = projectID

	data, err := h.AddMemberUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Member added successfully")
}

func (h *ProjectHandler) RemoveMember(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	accountID := c.Params("accountId")
	if accountID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing accountId", "MISSING_ACCOUNT_ID", nil))
	}

	// Get NodeID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	req := &project.RemoveMemberRequest{
		ProjectID: projectID,
		AccountID: accountID,
	}

	data, err := h.RemoveMemberUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Member removed successfully")
}

func (h *ProjectHandler) ListMembers(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("missing projectId", "MISSING_PROJECT_ID", nil))
	}

	// Get NodeID and AccountID from JWT claims
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	nodeID, ok := jwtClaims["NodeId"].(string)
	if !ok || nodeID == "" {
		h.logger.Error("Missing NodeId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", nil)
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	data, err := h.ListMembersUC.Execute(c.Context(), projectID, nodeID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Members retrieved successfully")
}
