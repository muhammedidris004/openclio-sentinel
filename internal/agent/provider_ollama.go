package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// OllamaProvider calls the Ollama local API.
// Zero cost, full privacy — runs on your machine.
type OllamaProvider struct {
	model       string
	baseURL     string
	numCtx      int      // context window tokens; 0 = use Ollama default (4096)
	temperature *float64 // nil = provider default
	seed        *int     // nil = provider default
}

// ollamaBaseURL normalizes the base URL to IPv4 localhost, which is where Ollama commonly binds.
func ollamaBaseURL(raw string) string {
	raw = strings.TrimRight(strings.TrimSpace(raw), "/")
	if raw == "" {
		return "http://127.0.0.1:11434"
	}
	if strings.Contains(raw, "//localhost") {
		raw = strings.Replace(raw, "//localhost", "//127.0.0.1", 1)
	}
	return raw
}

// NewOllamaProvider creates an Ollama provider with default base URL.
func NewOllamaProvider(model string) *OllamaProvider {
	return NewOllamaProviderWithBaseURL(model, "")
}

// NewOllamaProviderWithBaseURL creates an Ollama provider with an explicit base URL.
func NewOllamaProviderWithBaseURL(model, baseURL string) *OllamaProvider {
	return &OllamaProvider{
		model:   model,
		baseURL: ollamaBaseURL(baseURL),
	}
}

// SetNumCtx overrides the Ollama context window size (num_ctx option).
// Call this after construction when the config specifies ollama_num_ctx.
func (p *OllamaProvider) SetNumCtx(n int) {
	p.numCtx = n
}

// SetTemperature overrides Ollama sampling temperature.
func (p *OllamaProvider) SetTemperature(v float64) {
	p.temperature = &v
}

// SetSeed pins Ollama sampling to a deterministic seed when supported.
func (p *OllamaProvider) SetSeed(v int) {
	p.seed = &v
}

func (p *OllamaProvider) Name() string { return "ollama" }

// Ollama API types
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Tools    []ollamaTool    `json:"tools,omitempty"`
	Stream   bool            `json:"stream"`
	Options  *ollamaOptions  `json:"options,omitempty"`
}

type ollamaOptions struct {
	NumCtx      int      `json:"num_ctx,omitempty"`
	Temperature *float64 `json:"temperature,omitempty"`
	Seed        *int     `json:"seed,omitempty"`
}

type ollamaMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaTool struct {
	Type     string         `json:"type"`
	Function ollamaFunction `json:"function"`
}

type ollamaFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type ollamaToolCall struct {
	Function ollamaFunctionCall `json:"function"`
}

type ollamaFunctionCall struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ollamaResponse struct {
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
	Error           string        `json:"error,omitempty"`
}

type ollamaTagsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

// stripThinkingTags removes <think>...</think> blocks from local reasoning
// model output (deepseek-r1, qwen3, phi4-reasoning, etc.).
// The user only sees the final answer, not the internal chain-of-thought.
func stripThinkingTags(s string) string {
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(s[start:], "</think>")
		if end == -1 {
			// unclosed tag — strip everything from <think> onwards
			s = strings.TrimSpace(s[:start])
			break
		}
		s = s[:start] + s[start+end+len("</think>"):]
	}
	return strings.TrimSpace(s)
}

func canonicalOllamaModel(name string) string {
	c := strings.ToLower(strings.TrimSpace(name))
	c = strings.TrimSuffix(c, ":latest")
	c = strings.ReplaceAll(c, ":", "-")
	return c
}

// resolveModelName maps friendly/local aliases to actual installed Ollama tags.
// Example: gpt-oss-20b -> gpt-oss:20b when that tag exists locally.
func (p *OllamaProvider) resolveModelName(ctx context.Context, requested string) string {
	model := strings.TrimSpace(requested)
	if model == "" {
		return model
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.baseURL+"/api/tags", nil)
	if err != nil {
		return model
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return model
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return model
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return model
	}
	var tags ollamaTagsResponse
	if err := json.Unmarshal(body, &tags); err != nil {
		return model
	}
	if len(tags.Models) == 0 {
		return model
	}

	for _, m := range tags.Models {
		if strings.EqualFold(strings.TrimSpace(m.Name), model) {
			return strings.TrimSpace(m.Name)
		}
	}

	target := canonicalOllamaModel(model)
	for _, m := range tags.Models {
		name := strings.TrimSpace(m.Name)
		if canonicalOllamaModel(name) == target {
			return name
		}
	}
	return model
}

func (p *OllamaProvider) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	// Convert messages
	msgs := make([]ollamaMessage, 0, len(req.Messages)+1)

	if req.SystemPrompt != "" {
		msgs = append(msgs, ollamaMessage{Role: "system", Content: req.SystemPrompt})
	}

	for _, m := range req.Messages {
		msg := ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		msgs = append(msgs, msg)
	}

	// Convert tools
	var tools []ollamaTool
	for _, t := range req.Tools {
		tools = append(tools, ollamaTool{
			Type: "function",
			Function: ollamaFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  t.Parameters,
			},
		})
	}

	model := req.Model
	if model == "" {
		model = p.model
	}
	model = p.resolveModelName(ctx, model)

	apiReq := ollamaRequest{
		Model:    model,
		Messages: msgs,
		Tools:    tools,
		Stream:   false,
	}
	if opts := p.buildOptions(); opts != nil {
		apiReq.Options = opts
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("marshalling request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Ollama API (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Ollama API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result ollamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	if result.Error != "" {
		return nil, fmt.Errorf("Ollama error: %s", result.Error)
	}

	chatResp := &ChatResponse{
		Content: stripThinkingTags(result.Message.Content),
		Usage: Usage{
			InputTokens:  result.PromptEvalCount,
			OutputTokens: result.EvalCount,
		},
		StopReason: "stop",
	}

	// Parse tool calls
	for i, tc := range result.Message.ToolCalls {
		chatResp.ToolCalls = append(chatResp.ToolCalls, ToolCall{
			ID:        fmt.Sprintf("ollama_%d", i),
			Name:      tc.Function.Name,
			Arguments: tc.Function.Arguments,
		})
	}

	return chatResp, nil
}

func (p *OllamaProvider) buildOptions() *ollamaOptions {
	if p == nil {
		return nil
	}
	if p.numCtx <= 0 && p.temperature == nil && p.seed == nil {
		return nil
	}
	return &ollamaOptions{
		NumCtx:      p.numCtx,
		Temperature: p.temperature,
		Seed:        p.seed,
	}
}
