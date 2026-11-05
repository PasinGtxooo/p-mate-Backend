package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/profile"
	"github.com/FrostBitzX/smart-task-ai/internal/application/profile/usecase"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

type ProfileHandler struct {
	CreateProfileUC *usecase.CreateProfileUseCase
	GetProfileUC    *usecase.GetProfileUseCase
	UpdateProfileUC *usecase.UpdateProfileUseCase
	logger          logger.Logger
}

func NewProfileHandler(
	create *usecase.CreateProfileUseCase,
	get *usecase.GetProfileUseCase,
	update *usecase.UpdateProfileUseCase,
	l logger.Logger,
) *ProfileHandler {
	return &ProfileHandler{
		CreateProfileUC: create,
		GetProfileUC:    get,
		UpdateProfileUC: update,
		logger:          l,
	}

}

func (h *ProfileHandler) CreateProfile(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[profile.CreateProfileRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, apperror.ErrInvalidData)
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

	data, err := h.CreateProfileUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Profile created successfully")
}

func (h *ProfileHandler) GetProfile(c *fiber.Ctx) error {
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

	req := &profile.GetProfileByAccountIDRequest{
		AccountID: accountID,
	}

	data, err := h.GetProfileUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Profile retrieved successfully")
}

func (h *ProfileHandler) UpdateProfile(c *fiber.Ctx) error {
	req, err := requests.ParseAndValidate[profile.UpdateProfileRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, apperror.ErrInvalidData)
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

	data, err := h.UpdateProfileUC.Execute(c.Context(), req, nodeID)
	if err != nil {
		return responses.Error(c, err)
	}

	return responses.Success(c, data, "Profile updated successfully")
}
