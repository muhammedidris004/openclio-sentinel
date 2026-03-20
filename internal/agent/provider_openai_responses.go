package agent

// Responses API support for OpenAI Codex models (gpt-5.x-codex, gpt-5.x).
// Codex models are NOT available on the Chat Completions endpoint (/v1/chat/completions).
// They require the Responses API endpoint (/v1/responses).

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// isCodexResponsesModel returns true for models that require the Responses API.
// Covers all gpt-5.x-codex variants and gpt-5.x base models served via Codex.
func isCodexResponsesModel(model string) bool {
	m := strings.ToLower(strings.TrimSpace(model))
	return strings.Contains(m, "codex") || strings.HasPrefix(m, "gpt-5.")
}

// responsesEndpoint returns the correct Responses API URL for this provider.
// ChatGPT OAuth tokens route through chatgpt.com/backend-api/codex/responses.
// Regular API keys use api.openai.com/v1/responses.
func (p *OpenAIProvider) responsesEndpoint() string {
	if strings.Contains(p.baseURL, "chatgpt.com") {
		return p.baseURL + "/responses"
	}
	return openAIEndpoint(p.baseURL, "/responses")
}

// ── Responses API request types ──────────────────────────────────────────────

type responsesAPIInput struct {
	// Regular message (role = "user" | "assistant" | "system" | "developer")
	Role    string `json:"role,omitempty"`
	Content string `json:"content,omitempty"`
	// Function call made by the model
	Type      string `json:"type,omitempty"`   // "function_call" or "function_call_output"
	CallID    string `json:"call_id,omitempty"` // used by both function_call and function_call_output
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"` // JSON string
	Output    string `json:"output,omitempty"`    // tool result text
}

type responsesAPITool struct {
	Type        string          `json:"type"` // "function"
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type responsesAPIReasoning struct {
	Effort string `json:"effort"` // "low" | "medium" | "high"
}

type responsesAPIRequest struct {
	Model           string                 `json:"model"`
	Input           []responsesAPIInput    `json:"input"`
	Instructions    string                 `json:"instructions,omitempty"` // chatgpt.com endpoint requires this
	Tools           []responsesAPITool     `json:"tools,omitempty"`
	MaxOutputTokens int                    `json:"max_output_tokens,omitempty"`
	Reasoning       *responsesAPIReasoning `json:"reasoning,omitempty"`
	Stream          bool                   `json:"stream,omitempty"`
}

// ── Responses API response types ─────────────────────────────────────────────

type responsesAPIResponse struct {
	ID     string                `json:"id"`
	Output []responsesOutputItem `json:"output"`
	Usage  responsesAPIUsage     `json:"usage"`
	Error  *openAIError          `json:"error,omitempty"`
}

type responsesOutputItem struct {
	Type      string              `json:"type"`    // "message", "reasoning", "function_call"
	Role      string              `json:"role,omitempty"`
	Content   []responsesContent  `json:"content,omitempty"`
	// Function call fields
	CallID    string `json:"call_id,omitempty"`
	Name      string `json:"name,omitempty"`
	Arguments string `json:"arguments,omitempty"`
	Status    string `json:"status,omitempty"`
}

type responsesContent struct {
	Type string `json:"type"` // "output_text"
	Text string `json:"text"`
}

type responsesAPIUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// ── Input conversion ─────────────────────────────────────────────────────────

// buildResponsesInput converts Chat Completions message format to Responses API input items.
// The system prompt is NOT included here — it is passed as Instructions in the top-level request.
func buildResponsesInput(messages []Message) []responsesAPIInput {
	var input []responsesAPIInput
	for _, m := range messages {
		switch m.Role {
		case "tool":
			// Tool result → function_call_output
			input = append(input, responsesAPIInput{
				Type:   "function_call_output",
				CallID: m.ToolCallID,
				Output: m.Content,
			})
		case "assistant":
			if len(m.ToolCalls) > 0 {
				// Assistant requested tool calls — emit one function_call item per call
				for _, tc := range m.ToolCalls {
					input = append(input, responsesAPIInput{
						Type:      "function_call",
						CallID:    tc.ID,
						Name:      tc.Name,
						Arguments: string(tc.Arguments),
					})
				}
				// Also emit any text content
				if m.Content != "" {
					input = append(input, responsesAPIInput{Role: "assistant", Content: m.Content})
				}
			} else {
				input = append(input, responsesAPIInput{Role: "assistant", Content: m.Content})
			}
		default:
			input = append(input, responsesAPIInput{Role: m.Role, Content: m.Content})
		}
	}
	return input
}

// buildResponsesTools converts the tool list to the Responses API flat format.
func buildResponsesTools(tools []ToolDef) []responsesAPITool {
	out := make([]responsesAPITool, 0, len(tools))
	for _, t := range tools {
		out = append(out, responsesAPITool{
			Type:        "function",
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}
	return out
}

// ── Non-streaming Responses API call ─────────────────────────────────────────

func (p *OpenAIProvider) chatViaResponsesAPI(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	apiReq := responsesAPIRequest{
		Model:           model,
		Input:           buildResponsesInput(req.Messages),
		Instructions:    req.SystemPrompt,
		Tools:           buildResponsesTools(req.Tools),
		MaxOutputTokens: req.MaxTokens,
	}
	if req.Thinking {
		budget := req.ThinkingBudget
		effort := "medium"
		switch {
		case budget >= 15000:
			effort = "high"
		case budget < 3000:
			effort = "low"
		}
		apiReq.Reasoning = &responsesAPIReasoning{Effort: effort}
	}

	body, _ := json.Marshal(apiReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.responsesEndpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating responses request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Responses API: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading Responses API response: %w", err)
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("Responses API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var result responsesAPIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parsing Responses API response: %w", err)
	}
	if result.Error != nil {
		return nil, fmt.Errorf("Responses API error: %s", result.Error.Message)
	}

	chatResp := &ChatResponse{
		Usage: Usage{
			InputTokens:  result.Usage.InputTokens,
			OutputTokens: result.Usage.OutputTokens,
		},
	}

	for _, item := range result.Output {
		switch item.Type {
		case "message":
			for _, c := range item.Content {
				if c.Type == "output_text" {
					chatResp.Content += c.Text
				}
			}
		case "function_call":
			args := json.RawMessage(item.Arguments)
			if len(args) == 0 {
				args = json.RawMessage("{}")
			}
			chatResp.ToolCalls = append(chatResp.ToolCalls, ToolCall{
				ID:        item.CallID,
				Name:      item.Name,
				Arguments: args,
			})
		}
	}

	if len(chatResp.ToolCalls) > 0 {
		chatResp.StopReason = "tool_use"
	} else {
		chatResp.StopReason = "end_turn"
	}
	return chatResp, nil
}

// ── Streaming Responses API call ─────────────────────────────────────────────

func (p *OpenAIProvider) streamViaResponsesAPI(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	model := req.Model
	if model == "" {
		model = p.model
	}

	apiReq := responsesAPIRequest{
		Model:           model,
		Input:           buildResponsesInput(req.Messages),
		Instructions:    req.SystemPrompt,
		Tools:           buildResponsesTools(req.Tools),
		MaxOutputTokens: req.MaxTokens,
		Stream:          true,
	}
	if req.Thinking {
		budget := req.ThinkingBudget
		effort := "medium"
		switch {
		case budget >= 15000:
			effort = "high"
		case budget < 3000:
			effort = "low"
		}
		apiReq.Reasoning = &responsesAPIReasoning{Effort: effort}
	}

	body, _ := json.Marshal(apiReq)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.responsesEndpoint(), bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating responses stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Responses API stream: %w", err)
	}
	if resp.StatusCode != 200 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 8*1024))
		resp.Body.Close()
		return nil, fmt.Errorf("Responses API stream error (HTTP %d): %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}

	ch := make(chan StreamChunk, 64)
	go func() {
		defer close(ch)
		defer resp.Body.Close()

		send := func(chunk StreamChunk) bool {
			select {
			case ch <- chunk:
				return true
			case <-ctx.Done():
				return false
			}
		}

		// Track in-progress function calls by item_id.
		type pendingCall struct {
			callID string
			name   string
			args   strings.Builder
		}
		pending := map[string]*pendingCall{}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "" {
				continue
			}

			var evt struct {
				Type     string `json:"type"`
				// output_text.delta
				Delta    string `json:"delta"`
				// output_item.added / output_item.done
				Item     *responsesOutputItem `json:"item"`
				// function_call_arguments.delta
				ItemID   string `json:"item_id"`
				// response.completed
				Response *responsesAPIResponse `json:"response"`
			}
			if err := json.Unmarshal([]byte(data), &evt); err != nil {
				continue
			}

			switch evt.Type {
			case "response.output_text.delta":
				if evt.Delta != "" {
					if !send(StreamChunk{Text: evt.Delta}) {
						return
					}
				}

			case "response.output_item.added":
				if evt.Item != nil && evt.Item.Type == "function_call" {
					// Start accumulating a new function call.
					pc := &pendingCall{callID: evt.Item.CallID, name: evt.Item.Name}
					pending[evt.ItemID] = pc
					if evt.Item.Arguments != "" {
						pc.args.WriteString(evt.Item.Arguments)
					}
				}

			case "response.function_call_arguments.delta":
				if pc, ok := pending[evt.ItemID]; ok {
					pc.args.WriteString(evt.Delta)
				}

			case "response.output_item.done":
				if evt.Item == nil {
					continue
				}
				if evt.Item.Type == "function_call" {
					// Resolve the call ID: prefer accumulated pending, fall back to item fields.
					callID := evt.Item.CallID
					name := evt.Item.Name
					argsStr := evt.Item.Arguments

					if pc, ok := pending[evt.ItemID]; ok {
						if pc.callID != "" {
							callID = pc.callID
						}
						if pc.name != "" {
							name = pc.name
						}
						fullArgs := pc.args.String()
						if fullArgs != "" {
							argsStr = fullArgs
						}
						delete(pending, evt.ItemID)
					}

					args := json.RawMessage(argsStr)
					if len(args) == 0 {
						args = json.RawMessage("{}")
					}
					tc := ToolCall{ID: callID, Name: name, Arguments: args}
					if !send(StreamChunk{ToolCalls: []ToolCall{tc}}) {
						return
					}
				}

			case "response.completed":
				// Flush any remaining pending tool calls.
				for _, pc := range pending {
					args := json.RawMessage(pc.args.String())
					if len(args) == 0 {
						args = json.RawMessage("{}")
					}
					if !send(StreamChunk{ToolCalls: []ToolCall{{ID: pc.callID, Name: pc.name, Arguments: args}}}) {
						return
					}
				}
				send(StreamChunk{Done: true})
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("reading Responses API stream: %w", err), Done: true}
		} else {
			ch <- StreamChunk{Done: true}
		}
	}()

	return ch, nil
}
