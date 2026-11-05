package service

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/chats"
	projectEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/entity"
	projectSvc "github.com/FrostBitzX/smart-task-ai/internal/domain/projects/service"
	taskSvc "github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/service"
	"github.com/FrostBitzX/smart-task-ai/internal/infrastructure/groq"
	"github.com/FrostBitzX/smart-task-ai/pkg/apperror"
	"github.com/google/uuid"
)

// Error codes for chat service
const (
	ErrCodeGroqUnavailable = "GROQ_UNAVAILABLE"
	ErrCodeGroqTimeout     = "GROQ_TIMEOUT"
	ErrCodeRateLimited     = "RATE_LIMITED"
	ErrCodeGroqAuthError   = "GROQ_AUTH_ERROR"
	ErrCodeProjectNotFound = "PROJECT_NOT_FOUND"
	ErrCodeInvalidMessage  = "INVALID_MESSAGE"
)

// Pre-compiled regex for extracting JSON from markdown code blocks
var codeBlockRegex = regexp.MustCompile("```(?:json)?\\s*([\\s\\S]*?)```")

// ChatService defines the interface for chat operations
type ChatService interface {
	SendMessage(ctx context.Context, req *SendMessageRequest, nodeID string, accountID uuid.UUID) (*SendMessageResponse, error)
}

// SendMessageRequest represents a request to send a message
type SendMessageRequest struct {
	ProjectID      uuid.UUID
	AccountID      uuid.UUID
	Content        string
	SessionHistory []groq.ChatMessage
}

// SendMessageResponse represents the response from sending a message
type SendMessageResponse struct {
	Type    string
	Message string
	Tasks   []TaskFromAI
}

// chatService implements the ChatService interface
type chatService struct {
	groqClient     groq.GroqClient
	taskService    *taskSvc.TaskService
	projectService *projectSvc.ProjectService
	promptBuilder  PromptBuilder
}

// NewChatService creates a new chat service
func NewChatService(
	groqClient groq.GroqClient,
	taskService *taskSvc.TaskService,
	projectService *projectSvc.ProjectService,
) ChatService {
	return &chatService{
		groqClient:     groqClient,
		taskService:    taskService,
		projectService: projectService,
		promptBuilder:  NewPromptBuilder(),
	}
}

// SendMessage sends a message to the AI and returns the response
func (s *chatService) SendMessage(ctx context.Context, req *SendMessageRequest, nodeID string, accountID uuid.UUID) (*SendMessageResponse, error) {
	if err := s.validateRequest(req); err != nil {
		return nil, err
	}

	project, err := s.projectService.GetProjectByID(ctx, req.ProjectID, nodeID, accountID)
	if err != nil {
		return nil, s.handleProjectError(err)
	}

	tasks, err := s.taskService.ListTasksByProject(ctx, req.ProjectID, accountID.String())
	if err != nil {
		return nil, apperror.NewInternalServerError("failed to get tasks", "GET_TASKS_ERROR", err)
	}

	aiConfig := s.getAIConfig(project)
	systemPrompt := s.promptBuilder.BuildSystemPrompt(aiConfig, tasks)
	messages := s.buildMessages(systemPrompt, req.SessionHistory, req.Content)
	groqReq := groq.NewDefaultRequest(messages)

	resp, err := s.groqClient.SendChatCompletion(ctx, groqReq)
	if err != nil {
		return nil, s.handleGroqError(err)
	}

	if len(resp.Choices) == 0 {
		return nil, apperror.NewInternalServerError("no response from AI", "EMPTY_RESPONSE", nil)
	}

	aiResponse := resp.Choices[0].Message.Content

	// Try to parse as structured JSON response from AI
	structuredResp := s.parseStructuredResponse(aiResponse)
	if structuredResp != nil {
		return structuredResp, nil
	}

	return &SendMessageResponse{
		Type:    "text",
		Message: aiResponse,
		Tasks:   nil,
	}, nil
}

// TaskListResponse represents the JSON response from AI with task_actions type
type TaskListResponse struct {
	Type    string       `json:"type"`
	Message string       `json:"message"`
	Tasks   []TaskFromAI `json:"tasks"`
}

// TaskFromAI represents a single task from AI response
type TaskFromAI struct {
	ID             string  `json:"id,omitempty"`
	Name           string  `json:"name"`
	Description    string  `json:"description,omitempty"`
	Priority       string  `json:"priority"`
	Status         string  `json:"status"`
	StartDateTime  string  `json:"start_datetime,omitempty"`
	EndDateTime    string  `json:"end_datetime,omitempty"`
	Location       string  `json:"location,omitempty"`
	RecurringDays  int     `json:"recurring_days,omitempty"`
	RecurringUntil string  `json:"recurring_until,omitempty"`
	RequiredSkill  string  `json:"required_skill,omitempty"`
	EstimateManday float64 `json:"estimate_manday,omitempty"`
}

// parseStructuredResponse tries to parse any structured JSON response from AI
func (s *chatService) parseStructuredResponse(response string) *SendMessageResponse {
	// Try to extract JSON from the response
	jsonStr := extractJSON(response)
	if jsonStr == "" {
		return nil
	}

	// Try to parse as TaskListResponse (supports "message" fields)
	var taskListResp TaskListResponse
	if err := json.Unmarshal([]byte(jsonStr), &taskListResp); err != nil {
		return nil
	}

	// Must have a type field
	if taskListResp.Type == "" {
		return nil
	}

	// Must have a message
	if taskListResp.Message == "" {
		return nil
	}

	return &SendMessageResponse{
		Type:    taskListResp.Type,
		Message: taskListResp.Message,
		Tasks:   taskListResp.Tasks,
	}
}

// ExtractHistoryText strips the tasks payload out of a prior structured AI
// response, keeping only the display message. Session history only needs the
// conversational text — the task list is redundant there since the system
// prompt already re-injects the project's current tasks on every turn — and
// echoing full task JSON back through history burns TPM budget for no benefit.
func ExtractHistoryText(content string) string {
	jsonStr := extractJSON(content)
	if jsonStr == "" {
		return content
	}

	var resp TaskListResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return content
	}

	if resp.Type == "" || resp.Message == "" {
		return content
	}

	return resp.Message
}

// extractJSON extracts JSON object from a string (handles markdown code blocks)
func extractJSON(s string) string {
	s = strings.TrimSpace(s)

	// Try to extract from markdown code block (using pre-compiled regex)
	matches := codeBlockRegex.FindStringSubmatch(s)
	if len(matches) > 1 {
		s = strings.TrimSpace(matches[1])
	}

	// Find JSON object boundaries
	start := strings.Index(s, "{")
	if start == -1 {
		return ""
	}

	// Find matching closing brace, accounting for strings
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(s); i++ {
		c := s[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch c {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return s[start : i+1]
			}
		}
	}

	return ""
}

func (s *chatService) validateRequest(req *SendMessageRequest) error {
	if req == nil {
		return apperror.NewBadRequestError("request is required", ErrCodeInvalidMessage, nil)
	}
	if strings.TrimSpace(req.Content) == "" {
		return apperror.NewBadRequestError("message content is required", ErrCodeInvalidMessage, nil)
	}
	if req.ProjectID == uuid.Nil {
		return apperror.NewBadRequestError("project ID is required", ErrCodeInvalidMessage, nil)
	}
	return nil
}

func (s *chatService) handleProjectError(err error) error {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		if appErr.Code == "PROJECT_NOT_FOUND" {
			return apperror.NewNotFoundError("project not found", ErrCodeProjectNotFound, err)
		}
	}
	return apperror.NewInternalServerError("failed to get project", "GET_PROJECT_ERROR", err)
}

func (s *chatService) handleGroqError(err error) error {
	errMsg := err.Error()

	if strings.Contains(errMsg, "timeout") || strings.Contains(errMsg, "deadline exceeded") {
		return apperror.NewTimeoutError("AI response timeout", ErrCodeGroqTimeout, err)
	}

	if strings.Contains(errMsg, "status 429") || strings.Contains(errMsg, "Too many requests") {
		return apperror.NewTooManyRequestsError("Too many requests", ErrCodeRateLimited, err)
	}

	if strings.Contains(errMsg, "status 401") || strings.Contains(errMsg, "Unauthorized") {
		return apperror.NewInternalServerError("AI service configuration error", ErrCodeGroqAuthError, err)
	}

	if strings.Contains(errMsg, "status 503") || strings.Contains(errMsg, "unavailable") {
		return &apperror.AppError{
			Status:   http.StatusServiceUnavailable,
			Code:     ErrCodeGroqUnavailable,
			Message:  "AI service temporarily unavailable",
			RawError: err,
		}
	}

	return &apperror.AppError{
		Status:   http.StatusServiceUnavailable,
		Code:     ErrCodeGroqUnavailable,
		Message:  "AI service temporarily unavailable",
		RawError: err,
	}
}

func (s *chatService) getAIConfig(project *projectEntity.Project) *chats.AIConfig {
	if len(project.Config) == 0 {
		return &chats.DefaultAIConfig
	}

	var configWrapper struct {
		AIConfig *chats.AIConfig `json:"ai_config"`
	}

	if err := json.Unmarshal(project.Config, &configWrapper); err != nil {
		return &chats.DefaultAIConfig
	}

	if configWrapper.AIConfig == nil {
		return &chats.DefaultAIConfig
	}

	config := configWrapper.AIConfig
	if config.ChatStyle == "" {
		config.ChatStyle = chats.DefaultAIConfig.ChatStyle
	}
	if len(config.DomainKnowledge) == 0 {
		config.DomainKnowledge = chats.DefaultAIConfig.DomainKnowledge
	}
	if config.Language == "" {
		config.Language = chats.DefaultAIConfig.Language
	}

	return config
}

func (s *chatService) buildMessages(systemPrompt string, sessionHistory []groq.ChatMessage, userContent string) []groq.ChatMessage {
	messages := make([]groq.ChatMessage, 0, len(sessionHistory)+2)

	messages = append(messages, groq.ChatMessage{
		Role:    "system",
		Content: systemPrompt,
	})

	messages = append(messages, sessionHistory...)

	messages = append(messages, groq.ChatMessage{
		Role:    "user",
		Content: userContent,
	})

	return messages
}
