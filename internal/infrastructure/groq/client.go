package groq

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	DefaultModel               = "openai/gpt-oss-120b"
	DefaultTemperature         = 1.0
	DefaultMaxCompletionTokens = 4096
	DefaultTopP                = 1.0
	DefaultReasoningEffort     = "medium"
	DefaultTimeout             = 30 * time.Second
	DefaultAPIURL              = "https://api.groq.com/openai/v1/chat/completions"
)

// GroqClient defines the interface for interacting with Groq API
type GroqClient interface {
	SendChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error)
	SendChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan StreamChunk, error)
}

// ChatCompletionRequest represents a request to the Groq chat completion API
type ChatCompletionRequest struct {
	Messages            []ChatMessage `json:"messages"`
	Model               string        `json:"model"`
	Temperature         float64       `json:"temperature"`
	MaxCompletionTokens int           `json:"max_completion_tokens"`
	TopP                float64       `json:"top_p"`
	ReasoningEffort     string        `json:"reasoning_effort"`
	Stream              bool          `json:"stream"`
	Stop                *string       `json:"stop,omitempty"`
}

// ChatMessage represents a message in the chat conversation
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionResponse represents the response from Groq API
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice represents a completion choice in the response
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage represents token usage information
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamChunk represents a chunk of streamed response
type StreamChunk struct {
	Content string
	Done    bool
	Error   error
}

// StreamDelta represents the delta content in a stream chunk
type StreamDelta struct {
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
}

// StreamChoice represents a choice in a stream response
type StreamChoice struct {
	Index        int         `json:"index"`
	Delta        StreamDelta `json:"delta"`
	FinishReason *string     `json:"finish_reason,omitempty"`
}

// StreamResponse represents a streaming response chunk from Groq API
type StreamResponse struct {
	ID      string         `json:"id"`
	Object  string         `json:"object"`
	Created int64          `json:"created"`
	Model   string         `json:"model"`
	Choices []StreamChoice `json:"choices"`
}

// APIError represents an error response from Groq API
type APIError struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

// groqClient implements the GroqClient interface
type groqClient struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

// ClientOption is a function that configures the groqClient
type ClientOption func(*groqClient)

// WithAPIURL sets a custom API URL
func WithAPIURL(url string) ClientOption {
	return func(c *groqClient) {
		c.apiURL = url
	}
}

// WithHTTPClient sets a custom HTTP client
func WithHTTPClient(client *http.Client) ClientOption {
	return func(c *groqClient) {
		c.httpClient = client
	}
}

// NewGroqClient creates a new Groq client
func NewGroqClient(opts ...ClientOption) (GroqClient, error) {
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GROQ_API_KEY environment variable is not set")
	}

	apiURL := os.Getenv("GROQ_API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	client := &groqClient{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// NewGroqClientWithKey creates a new Groq client with explicit API key
func NewGroqClientWithKey(apiKey string, opts ...ClientOption) (GroqClient, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	apiURL := os.Getenv("GROQ_API_URL")
	if apiURL == "" {
		apiURL = DefaultAPIURL
	}

	client := &groqClient{
		apiKey: apiKey,
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
		},
	}

	for _, opt := range opts {
		opt(client)
	}

	return client, nil
}

// NewDefaultRequest creates a ChatCompletionRequest with default values
func NewDefaultRequest(messages []ChatMessage) *ChatCompletionRequest {
	return &ChatCompletionRequest{
		Messages:            messages,
		Model:               DefaultModel,
		Temperature:         DefaultTemperature,
		MaxCompletionTokens: DefaultMaxCompletionTokens,
		TopP:                DefaultTopP,
		ReasoningEffort:     DefaultReasoningEffort,
		Stream:              false,
	}
}

// SendChatCompletion sends a chat completion request to Groq API
func (c *groqClient) SendChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	// Ensure model is set
	if req.Model == "" {
		req.Model = DefaultModel
	}

	// Ensure stream is false for non-streaming requests
	req.Stream = false

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout after %v", DefaultTimeout)
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatCompletionResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &chatResp, nil
}

// SendChatCompletionStream sends a streaming chat completion request to Groq API
func (c *groqClient) SendChatCompletionStream(ctx context.Context, req *ChatCompletionRequest) (<-chan StreamChunk, error) {
	// Ensure model is set
	if req.Model == "" {
		req.Model = DefaultModel
	}

	// Ensure stream is true for streaming requests
	req.Stream = true

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	// Create a client without timeout for streaming
	streamClient := &http.Client{}
	resp, err := streamClient.Do(httpReq)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			return nil, fmt.Errorf("request timeout")
		}
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		respBody, _ := io.ReadAll(resp.Body)
		var apiErr APIError
		if err := json.Unmarshal(respBody, &apiErr); err == nil && apiErr.Error.Message != "" {
			return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, apiErr.Error.Message)
		}
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	chunkChan := make(chan StreamChunk)

	go func() {
		defer close(chunkChan)
		defer resp.Body.Close()

		reader := bufio.NewReader(resp.Body)

		for {
			select {
			case <-ctx.Done():
				chunkChan <- StreamChunk{Error: ctx.Err(), Done: true}
				return
			default:
			}

			line, err := reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					chunkChan <- StreamChunk{Done: true}
					return
				}
				chunkChan <- StreamChunk{Error: fmt.Errorf("failed to read stream: %w", err), Done: true}
				return
			}

			line = strings.TrimSpace(line)

			// Skip empty lines
			if line == "" {
				continue
			}

			// Check for SSE data prefix
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			data := strings.TrimPrefix(line, "data: ")

			// Check for stream end marker
			if data == "[DONE]" {
				chunkChan <- StreamChunk{Done: true}
				return
			}

			var streamResp StreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				chunkChan <- StreamChunk{Error: fmt.Errorf("failed to parse stream chunk: %w", err), Done: true}
				return
			}

			// Extract content from the first choice
			if len(streamResp.Choices) > 0 {
				choice := streamResp.Choices[0]
				if choice.Delta.Content != "" {
					chunkChan <- StreamChunk{Content: choice.Delta.Content}
				}
				if choice.FinishReason != nil && *choice.FinishReason != "" {
					chunkChan <- StreamChunk{Done: true}
					return
				}
			}
		}
	}()

	return chunkChan, nil
}
