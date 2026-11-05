package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/task"
	"github.com/FrostBitzX/smart-task-ai/internal/application/task/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type TaskHandler struct {
	CreateTaskUC         *usecase.CreateTaskUseCase
	GetTaskByIDUC        *usecase.GetTaskByIDUseCase
	ListTasksByProjectUC *usecase.ListTasksByProjectUseCase
	UpdateTaskUC         *usecase.UpdateTaskUseCase
	DeleteTaskUC         *usecase.DeleteTaskUseCase
	logger               logger.Logger
}

func NewTaskHandler(
	create *usecase.CreateTaskUseCase,
	getByID *usecase.GetTaskByIDUseCase,
	listByProject *usecase.ListTasksByProjectUseCase,
	update *usecase.UpdateTaskUseCase,
	delete *usecase.DeleteTaskUseCase,
	l logger.Logger,
) *TaskHandler {
	return &TaskHandler{
		CreateTaskUC:         create,
		GetTaskByIDUC:        getByID,
		ListTasksByProjectUC: listByProject,
		UpdateTaskUC:         update,
		DeleteTaskUC:         delete,
		logger:               l,
	}
}

func (h *TaskHandler) CreateTask(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[task.CreateTaskRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	// Parse project ID from URL
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID is required", "INVALID_PROJECT_ID", nil))
	}

	// Get AccountID from JWT claims (for membership check in soft multi-tenant)
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

	data, err := h.CreateTaskUC.Execute(c.Context(), projectID, req, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Task created successfully")
}

func (h *TaskHandler) GetTaskByID(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	if taskID == "" {
		return responses.Error(c, apperror.NewBadRequestError("task ID is required", "INVALID_TASK_ID", nil))
	}

	// Get AccountID from JWT claims (for membership check in soft multi-tenant)
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

	data, err := h.GetTaskByIDUC.Execute(c.Context(), taskID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Task retrieved successfully")
}

func (h *TaskHandler) ListTasksByProject(c *fiber.Ctx) error {
	projectID := c.Params("projectId")
	if projectID == "" {
		return responses.Error(c, apperror.NewBadRequestError("project ID is required", "INVALID_PROJECT_ID", nil))
	}

	// Get AccountID from JWT claims (for membership check in soft multi-tenant)
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

	data, err := h.ListTasksByProjectUC.Execute(c.Context(), projectID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Tasks retrieved successfully")
}

func (h *TaskHandler) UpdateTask(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[task.UpdateTaskRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, err)
	}

	taskID := c.Params("taskId")
	if taskID == "" {
		return responses.Error(c, apperror.NewBadRequestError("task ID is required", "INVALID_TASK_ID", nil))
	}

	// Get AccountID from JWT claims (for membership check in soft multi-tenant)
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

	data, err := h.UpdateTaskUC.Execute(c.Context(), taskID, req, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Task updated successfully")
}

func (h *TaskHandler) DeleteTask(c *fiber.Ctx) error {
	taskID := c.Params("taskId")
	if taskID == "" {
		return responses.Error(c, apperror.NewBadRequestError("task ID is required", "INVALID_TASK_ID", nil))
	}

	// Get AccountID from JWT claims (for membership check in soft multi-tenant)
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

	deletedID, err := h.DeleteTaskUC.Execute(c.Context(), taskID, accountID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, fiber.Map{"task_id": deletedID}, "Task deleted successfully")
}
