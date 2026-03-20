package agent

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

// Stream implements the Streamer interface for Ollama.
// Ollama returns newline-delimited JSON (NDJSON) when stream=true.
// Each line is a partial ollamaResponse; the final line has done=true and
// may carry tool_calls in the message field.
func (p *OllamaProvider) Stream(ctx context.Context, req ChatRequest) (<-chan StreamChunk, error) {
	msgs := make([]ollamaMessage, 0, len(req.Messages)+1)
	if req.SystemPrompt != "" {
		msgs = append(msgs, ollamaMessage{Role: "system", Content: req.SystemPrompt})
	}
	for _, m := range req.Messages {
		msg := ollamaMessage{
			Role:    m.Role,
			Content: m.Content,
		}
		// Carry tool calls on assistant messages.
		for i, tc := range m.ToolCalls {
			msg.ToolCalls = append(msg.ToolCalls, ollamaToolCall{
				Function: ollamaFunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
			_ = i
		}
		msgs = append(msgs, msg)
	}

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
		Stream:   true,
	}
	if opts := p.buildOptions(); opts != nil {
		apiReq.Options = opts
	}

	body, err := json.Marshal(apiReq)
	if err != nil {
		return nil, fmt.Errorf("marshalling stream request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating stream request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("calling Ollama stream API (is Ollama running?): %w", err)
	}
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		msg := strings.TrimSpace(string(body))
		if msg != "" {
			return nil, fmt.Errorf("Ollama stream API error (HTTP %d): %s", resp.StatusCode, msg)
		}
		return nil, fmt.Errorf("Ollama stream API error (HTTP %d)", resp.StatusCode)
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

		// inThinking suppresses <think>…</think> blocks token by token.
		// tagBuf accumulates partial tag text to detect boundary spans.
		var inThinking bool
		var tagBuf strings.Builder

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Bytes()
			if len(line) == 0 {
				continue
			}

			var evt ollamaResponse
			if err := json.Unmarshal(line, &evt); err != nil {
				continue
			}

			if evt.Error != "" {
				send(StreamChunk{Error: fmt.Errorf("Ollama error: %s", evt.Error), Done: true})
				return
			}

			// Filter <think>…</think> blocks from streaming output.
			if tok := evt.Message.Content; tok != "" {
				tagBuf.WriteString(tok)
				buf := tagBuf.String()

				// Process the buffer: emit safe text, suppress thinking spans.
				var safe strings.Builder
				for buf != "" {
					if inThinking {
						idx := strings.Index(buf, "</think>")
						if idx == -1 {
							buf = "" // still inside — consume everything
						} else {
							inThinking = false
							buf = buf[idx+len("</think>"):]
						}
					} else {
						idx := strings.Index(buf, "<think>")
						if idx == -1 {
							// No opening tag — check for partial match at end
							cutoff := len(buf)
							for cutoff > 0 && strings.HasPrefix("<think>", buf[cutoff-1:]) {
								cutoff--
							}
							safe.WriteString(buf[:cutoff])
							buf = buf[cutoff:] // hold potential partial tag
							break
						}
						safe.WriteString(buf[:idx])
						inThinking = true
						buf = buf[idx+len("<think>"):]
					}
				}
				tagBuf.Reset()
				tagBuf.WriteString(buf) // leftover partial tag

				if out := safe.String(); out != "" {
					if !send(StreamChunk{Text: out}) {
						return
					}
				}
				continue
			}

			if evt.Done {
				// Final chunk — collect any tool calls from the message.
				var toolCalls []ToolCall
				for i, tc := range evt.Message.ToolCalls {
					toolCalls = append(toolCalls, ToolCall{
						ID:        fmt.Sprintf("ollama_%d", i),
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					})
				}
				send(StreamChunk{ToolCalls: toolCalls, Done: true})
				return
			}
		}

		if err := scanner.Err(); err != nil {
			ch <- StreamChunk{Error: fmt.Errorf("reading Ollama stream: %w", err), Done: true}
		} else {
			ch <- StreamChunk{Done: true}
		}
	}()

	return ch, nil
}
