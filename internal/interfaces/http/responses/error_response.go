package responses

import (
	"net/http"
	"strings"

	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/gofiber/fiber/v2"
)

// ErrorResponse represents the standard error response
type ErrorResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   ErrorDetail `json:"error"`
}

// SuccessResponse represents the standard success response
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
	Error   interface{} `json:"error"`
}

type ErrorDetail struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

func Error(c *fiber.Ctx, err error) error {
	var status int
	var code string
	var message string

	if appErr, ok := apperror.IsAppError(err); ok {
		status = appErr.Status
		code = appErr.Code
		message = appErr.Message
	} else {
		status = apperror.StatusCode(err)
		code = "INTERNAL_SERVER_ERROR"
		message = err.Error()
	}

	errorResponse := ErrorResponse{
		Success: false,
		Message: message,
		Data:    nil,
		Error: ErrorDetail{
			Code:    status,
			Message: getStatusText(status),
			Details: code,
		},
	}

	return c.Status(status).JSON(errorResponse)
}

func getStatusText(status int) string {
	text := http.StatusText(status)
	if text == "" {
		return "INTERNAL_SERVER_ERROR"
	}
	return strings.ToUpper(strings.ReplaceAll(text, " ", "_"))
}

func Success(c *fiber.Ctx, data interface{}, message ...string) error {
	msg := "Success"
	if len(message) > 0 {
		msg = message[0]
	}

	successResponse := SuccessResponse{
		Success: true,
		Message: msg,
		Data:    data,
		Error:   nil,
	}

	return c.Status(fiber.StatusOK).JSON(successResponse)
}
