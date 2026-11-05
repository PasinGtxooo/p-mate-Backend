package service

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/FrostBitzX/smart-task-ai/internal/domain/chats"
	taskEntity "github.com/FrostBitzX/smart-task-ai/internal/domain/tasks/entity"
	"github.com/FrostBitzX/smart-task-ai/internal/utils"
)

//go:embed instruction.txt
var instructionPrompt string

const (
	maxTasksInPrompt = 10
	MaxHistoryMsg    = 8
)

type PromptBuilder interface {
	BuildSystemPrompt(config *chats.AIConfig, tasks []*taskEntity.Task) string
}

type promptBuilder struct{}

func NewPromptBuilder() PromptBuilder {
	return &promptBuilder{}
}

// BuildSystemPrompt builds a system prompt for the AI assistant
// It includes the AI config settings and current task list
func (p *promptBuilder) BuildSystemPrompt(config *chats.AIConfig, tasks []*taskEntity.Task) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Style: %s | ภาษา: %s | ความเชี่ยวชาญ: %s\n\n",
		config.ChatStyle,
		config.Language,
		strings.Join(config.DomainKnowledge, ", "),
	))

	// จำกัดแค่ N tasks ล่าสุด ไม่ยัดทั้งหมด
	limited := tasks
	if len(tasks) > maxTasksInPrompt {
		limited = tasks[len(tasks)-maxTasksInPrompt:]
	}

	if len(limited) == 0 {
		sb.WriteString("Tasks: ไม่มี\n")
	} else {
		sb.WriteString("Tasks ล่าสุด:\n")
		for _, task := range limited {
			taskID := utils.ShortUUIDWithPrefix(task.ID, taskEntity.TaskIDPrefix)
			sb.WriteString(fmt.Sprintf("- [%s] %s (%s/%s)", taskID, task.Name, task.Status, task.Priority))
			if task.StartDateTime != nil {
				sb.WriteString(fmt.Sprintf(" %s", *task.StartDateTime))
			}
			if task.EndDateTime != nil {
				sb.WriteString(fmt.Sprintf("→%s", *task.EndDateTime))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString("\n")
	sb.WriteString(instructionPrompt)

	return sb.String()
}
