package chat

// SendMessageRequestDTO represents the request to send a message to the AI assistant
type SendMessageRequestDTO struct {
	ProjectID      string       `json:"project_id,omitempty"` // Set from URL parameter
	Content        string       `json:"content" validate:"required"`
	SessionHistory []MessageDTO `json:"session_history,omitempty"`
}

// MessageDTO represents a message in the conversation history
type MessageDTO struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// SendMessageResponseDTO represents the response from the AI assistant
type SendMessageResponseDTO struct {
	Type    string    `json:"type"`    // "text" or "task_actions"
	Message string    `json:"message"` // AI response message
	Tasks   []TaskDTO `json:"tasks"`   // List of tasks (null when type is "text")
}

// TaskDTO represents a task in the chat response
type TaskDTO struct {
	Name           string   `json:"name"`
	Description    string   `json:"description,omitempty"`
	Priority       string   `json:"priority,omitempty"`
	StartDatetime  *string  `json:"start_datetime,omitempty"`
	EndDatetime    *string  `json:"end_datetime,omitempty"`
	Location       *string  `json:"location,omitempty"`
	RecurringDays  *int     `json:"recurring_days,omitempty"`
	RecurringUntil *string  `json:"recurring_until,omitempty"`
	RequiredSkill  *string  `json:"required_skill,omitempty"`
	EstimateManday *float64 `json:"estimate_manday,omitempty"`
}
