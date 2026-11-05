package chats

// AIConfig represents AI configuration within project config
// Used to customize the AI assistant's behavior per project
type AIConfig struct {
	ChatStyle       string   `json:"chat_style"`       // "formal", "casual", "friendly"
	DomainKnowledge []string `json:"domain_knowledge"` // Areas of expertise the AI should emphasize
	Language        string   `json:"language"`         // Preferred response language: "th", "en"
}

// DefaultAIConfig returns the default AI configuration
// Used when no config is specified in the project
var DefaultAIConfig = AIConfig{
	ChatStyle:       "casual",
	DomainKnowledge: []string{"task_management", "scheduling"},
	Language:        "th",
}
