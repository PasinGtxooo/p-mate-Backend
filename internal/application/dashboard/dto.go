package dashboard

// TaskStatisticsResponse represents the response for task statistics API
type TaskStatisticsResponse struct {
	Todo       int64 `json:"todo"`
	InProgress int64 `json:"in_progress"`
	InReview   int64 `json:"in_review"`
	Done       int64 `json:"done"`
}

// ProjectInfo represents project information in task responses
type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TaskWithProjectResponse represents a task with its project information
type TaskWithProjectResponse struct {
	ID            string      `json:"id"`
	Name          string      `json:"name"`
	Description   *string     `json:"description,omitempty"`
	Priority      string      `json:"priority"`
	Status        string      `json:"status"`
	StartDateTime *string     `json:"start_datetime,omitempty"`
	EndDateTime   *string     `json:"end_datetime,omitempty"`
	Project       ProjectInfo `json:"project"`
}

// UnscheduledTasksResponse represents the response for unscheduled tasks API
type UnscheduledTasksResponse struct {
	Items []TaskWithProjectResponse `json:"items"`
	Total int                       `json:"total"`
}

// TodayTasksResponse represents the response for today's tasks API
type TodayTasksResponse struct {
	Items []TaskWithProjectResponse `json:"items"`
	Total int                       `json:"total"`
}
