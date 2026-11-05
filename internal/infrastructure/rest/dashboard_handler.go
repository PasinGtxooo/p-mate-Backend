package rest

import (
	"time"

	"github.com/FrostBitzX/smart-task-ai/internal/application/dashboard/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	GetTaskStatisticsUC   *usecase.GetTaskStatisticsUseCase
	GetUnscheduledTasksUC *usecase.GetUnscheduledTasksUseCase
	ListTodayTasksUC      *usecase.ListTodayTasksUseCase
	logger                logger.Logger
}

func NewDashboardHandler(
	getTaskStatistics *usecase.GetTaskStatisticsUseCase,
	getUnscheduledTasks *usecase.GetUnscheduledTasksUseCase,
	listTodayTasks *usecase.ListTodayTasksUseCase,
	l logger.Logger,
) *DashboardHandler {
	return &DashboardHandler{
		GetTaskStatisticsUC:   getTaskStatistics,
		GetUnscheduledTasksUC: getUnscheduledTasks,
		ListTodayTasksUC:      listTodayTasks,
		logger:                l,
	}
}

func (h *DashboardHandler) GetTaskStatistics(c *fiber.Ctx) error {
	// Get AccountID from JWT claims (for soft multi-tenant)
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", map[string]interface{}{
			"path": c.Path(),
		})
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", map[string]interface{}{
			"path":   c.Path(),
			"claims": jwtClaims,
		})
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	data, err := h.GetTaskStatisticsUC.Execute(c.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get task statistics", map[string]interface{}{
			"accountID": accountID,
			"error":     err.Error(),
		})
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Task statistics retrieved successfully")
}

func (h *DashboardHandler) GetUnscheduledTasks(c *fiber.Ctx) error {
	// Get AccountID from JWT claims (for soft multi-tenant)
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", map[string]interface{}{
			"path": c.Path(),
		})
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", map[string]interface{}{
			"path":   c.Path(),
			"claims": jwtClaims,
		})
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	data, err := h.GetUnscheduledTasksUC.Execute(c.Context(), accountID)
	if err != nil {
		h.logger.Error("Failed to get unscheduled tasks", map[string]interface{}{
			"accountID": accountID,
			"error":     err.Error(),
		})
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Unscheduled tasks retrieved successfully")
}

func (h *DashboardHandler) ListTodayTasks(c *fiber.Ctx) error {
	// Get AccountID from JWT claims (for soft multi-tenant)
	jwtClaims, ok := c.Locals("jwt_claims").(map[string]interface{})
	if !ok {
		h.logger.Error("Invalid JWT claims", map[string]interface{}{
			"path": c.Path(),
		})
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	accountID, ok := jwtClaims["AccountId"].(string)
	if !ok || accountID == "" {
		h.logger.Error("Missing AccountId in JWT claims", map[string]interface{}{
			"path":   c.Path(),
			"claims": jwtClaims,
		})
		return responses.Error(c, apperror.ErrUnauthorized)
	}

	// Parse optional date query parameter (format: YYYY-MM-DD)
	var date *time.Time
	dateStr := c.Query("date")
	if dateStr != "" {
		parsedDate, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			h.logger.Error("Invalid date format", map[string]interface{}{
				"dateStr": dateStr,
				"error":   err.Error(),
			})
			return responses.Error(c, apperror.NewBadRequestError("invalid date format, expected YYYY-MM-DD", "INVALID_DATE_FORMAT", err))
		}
		date = &parsedDate
	}

	data, err := h.ListTodayTasksUC.Execute(c.Context(), accountID, date)
	if err != nil {
		h.logger.Error("Failed to list today's tasks", map[string]interface{}{
			"accountID": accountID,
			"date":      dateStr,
			"error":     err.Error(),
		})
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Today's tasks retrieved successfully")
}
