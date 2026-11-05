package rest

import (
	"github.com/FrostBitzX/smart-task-ai/internal/application/file"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/logger"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/requests"
	"github.com/FrostBitzX/smart-task-ai/internal/interfaces/http/responses"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// FileHandler handles file upload operations (presigned URL generation and upload completion)
type FileHandler struct {
	fileService *file.FileService
	logger      logger.Logger
}

// NewFileHandler creates a new FileHandler instance
func NewFileHandler(fileService *file.FileService, l logger.Logger) *FileHandler {
	return &FileHandler{
		fileService: fileService,
		logger:      l,
	}
}

// GeneratePresignedURL handles POST /api/files/avatar/presign endpoint
func (h *FileHandler) GeneratePresignedURL(c *fiber.Ctx) error {
	// Extract file metadata from request body
	req, err := requests.ParseAndValidate[file.PresignRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, apperror.ErrInvalidData)
	}

	// Extract account_id and node_id from JWT token
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

	// Set account_id and node_id from JWT
	req.AccountID = accountID
	req.NodeID = nodeID

	// Call FileService.GeneratePresignedURL
	response, err := h.fileService.GeneratePresignedURL(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL", map[string]interface{}{
			"error":      err.Error(),
			"account_id": accountID,
			"node_id":    nodeID,
			"file_name":  req.FileName,
			"file_size":  req.FileSize,
		})
		return responses.Error(c, err)
	}

	// Log successful presigned URL generation
	h.logger.Info("Presigned URL generated successfully", map[string]interface{}{
		"account_id": accountID,
		"node_id":    nodeID,
		"key":        response.Key,
		"file_size":  req.FileSize,
	})

	// Return JSON response with presigned URL and key
	return responses.Success(c, response, "Presigned URL generated successfully")
}

// UploadFileS3 handles POST /api/files/avatar/upload endpoint
func (h *FileHandler) UploadFileS3(c *fiber.Ctx) error {
	// Extract key from request body
	req, err := requests.ParseAndValidate[file.UploadRequest](c)
	if err != nil {
		h.logger.Warn("Invalid request data", map[string]interface{}{
			"error": err.Error(),
		})
		return responses.Error(c, apperror.ErrInvalidData)
	}

	// Extract account_id and node_id from JWT token
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

	// Set account_id and node_id from JWT
	req.AccountID = accountID
	req.NodeID = nodeID

	// Call FileService.UploadFileS3
	response, err := h.fileService.UploadFileS3(c.Context(), req)
	if err != nil {
		h.logger.Error("Failed to complete upload", map[string]interface{}{
			"error":      err.Error(),
			"account_id": accountID,
			"node_id":    nodeID,
			"key":        req.Key,
		})
		return responses.Error(c, err)
	}

	// Log successful upload completion
	h.logger.Info("Upload completed successfully", map[string]interface{}{
		"account_id": accountID,
		"node_id":    nodeID,
		"key":        req.Key,
		"url":        response.URL,
	})

	// Return JSON response with public URL
	return responses.Success(c, response, "Upload completed successfully")
}
