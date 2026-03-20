package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func openAIEndpoint(baseURL, path string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		return path
	}
	if strings.HasSuffix(base, "/v1") {
		return base + path
	}
	return base + "/v1" + path
}

// OpenAIProvider calls the OpenAI Chat Completions API.
// Also works with any OpenAI-compatible API (Groq, Together, etc.).
type OpenAIProvider struct {
	apiKey  string
	model   string
	baseURL string
}

// NewOpenAIProvider creates an OpenAI provider from an API key environment variable.
func NewOpenAIProvider(apiKeyEnv, model string) (*OpenAIProvider, error) {
	key := os.Getenv(apiKeyEnv)
	if key == "" {
		return nil, fmt.Errorf("environment variable %s is not set", apiKeyEnv)
	}
	return newOpenAIProviderWithKey(key, model), nil
}

// NewOpenAIProviderWithToken creates an OpenAI provider from a ChatGPT OAuth bearer token.
// ChatGPT OAuth tokens must call chatgpt.com/backend-api/codex, not api.openai.com.
func NewOpenAIProviderWithToken(accessToken, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey:  accessToken,
		model:   model,
		baseURL: "https://chatgpt.com/backend-api/codex",
	}
}

func newOpenAIProviderWithKey(apiKey, model string) *OpenAIProvider {
	baseURL := os.Getenv("OPENAI_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com"
	}
	return &OpenAIProvider{
		apiKey:  apiKey,
		model:   model,
		baseURL: baseURL,
	}
}

func (p *OpenAIProvider) Name() string { return "openai" }

// isReasoningModel returns true for OpenAI models that use internal reasoning
// (o1, o3, o4 family) and require max_completion_tokens instead of max_tokens.
func isReasoningModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.HasPrefix(m, "o1") || strings.HasPrefix(m, "o3") || strings.HasPrefix(m, "o4")
}

// OpenAI API types
type openAIRequest struct {
	Model               string          `json:"model"`
	Messages            []openAIMessage `json:"messages"`
	Tools               []openAITool    `json:"tools,omitempty"`
	MaxTokens           int             `json:"max_tokens,omitempty"`
	MaxCompletionTokens int             `json:"max_completion_tokens,omitempty"` // o1/o3/o4 models
	ReasoningEffort     string          `json:"reasoning_effort,omitempty"`      // "low" | "medium" | "high"
}

type openAIMessage struct {
	Role       string           `json:"role"`
	Content    string           `json:"content,omitempty"`
	ToolCalls  []openAIToolCall `json:"tool_calls,omitempty"`
	ToolCallID string           `json:"tool_call_id,omitempty"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openAIToolCall struct {
	ID       string             `json:"id"`
	Type     string             `json:"type"`
	Function openAIFunctionCall `json:"function"`
}

type openAIFunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openAIResponse struct {
	Choices []openAIChoice `json:"choices"`
	Usage   openAIUsage    `json:"usage"`
	Error   *openAIError   `json:"error,omitempty"`
}

type openAIChoice struct {
	Message      openAIMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

type openAIUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
}

type openAIError struct {
	Message string `json:"message"`
}

func (p *OpenAIProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}
	// Codex models (gpt-5.x-codex, gpt-5.x) require the Responses API.
	if isCodexResponsesModel(model) {
		return p.chatViaResponsesAPI(ctx, req)
	}

	// Convert messages
	msgs := make([]openAIMessage, 0, len(req.Messages)+1)

	// System prompt as first message
	if req.SystemPrompt != "" {
		msgs = append(msgs, openAIMessage{Role: "system", Content: req.SystemPrompt})
	}

	for _, m := range req.Messages {
		msg := openAIMessage{
			Role:       m.Role,
			Content:    m.Content,
			ToolCallID: m.ToolCallID,
		}
		// Convert tool calls
		for _, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, openAIToolCall{
				ID:   tc.ID,
				Type: "function",
				Function: openAIFunctionCall{
					Name:      tc.Name,
					Arguments: string(tc.Arguments),
				},
			})
		}
		msgs = append(msgs, msg)
	}

	// Convert tools
	var tools []openAITool
	for _, t := range req.Tools {
		tools = append(tools, openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	apiReq := openAIRequest{
		Model:    model,
		Messages: msgs,
		Tools:    tools,
	}

	if isReasoningModel(model) {
		// o1/o3/o4 models: use max_completion_tokens and reasoning_effort
		apiReq.MaxCompletionTokens = req.MaxTokens
		if req.Thinking {
			budget := req.ThinkingBudget
			switch {
			case budget >= 15000:
				apiReq.ReasoningEffort = "high"
			case budget >= 5000:
				apiReq.ReasoningEffort = "medium"
			default:
				apiReq.ReasoningEffort = "low"
			}
		}
	} else {
		apiReq.MaxTokens = req.MaxTokens
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", openAIEndpoint(p.baseURL, "/chat/completions"), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling OpenAI API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("OpenAI API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("OpenAI API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("OpenAI API returned no choices")
	}

	choice := result.Choices[0]
	chatResp := &ChatResponse{
		Content: choice.Message.Content,
		Usage: Usage{
			InputTokens:  result.Usage.PromptTokens,
			OutputTokens: result.Usage.CompletionTokens,
		},
		StopReason: choice.FinishReason,
	}

	// Parse tool calls
	for _, tc := range choice.Message.ToolCalls {
		chatResp.ToolCalls = append(chatResp.ToolCalls, ToolCall{
			ID:        tc.ID,
			Name:      tc.Function.Name,
			Arguments: json.RawMessage(tc.Function.Arguments),
		})
	}

	return chatResp, nil
}
