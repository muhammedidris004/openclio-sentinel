package agent

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/openclio/openclio/internal/config"
	agentctx "github.com/openclio/openclio/internal/context"
)

const (
	defaultDelegationParallelism = 5
	defaultDelegationTimeout     = 90 * time.Second
)

type delegationResult struct {
	task   string
	answer string
	err    error
}

// Delegate runs sub-tasks in parallel as real sub-agent instances, then
// synthesizes a final answer from all sub-agent outputs.
//
// Each sub-agent is a full Agent with access to the parent's tool registry
// (exec, web_fetch, web_search, memory_read, etc.) and runs its own
// tool-iteration loop. Sub-agents are isolated — they start with a fresh
// session and do not share conversation history with each other or the
// parent agent.
func (a *Agent) Delegate(ctx context.Context, objective string, tasks []string, cfg config.AgentDelegationConfig) (string, error) {
	return a.DelegateWithEvents(ctx, objective, tasks, cfg, nil)
}

// DelegateWithEvents runs delegated sub-tasks and emits subagent telemetry
// through onEvent when provided.
func (a *Agent) DelegateWithEvents(ctx context.Context, objective string, tasks []string, cfg config.AgentDelegationConfig, onEvent func(AgentEvent)) (string, error) {
	a.emitterMu.Lock()
	a.eventEmitter = onEvent
	a.emitterMu.Unlock()
	defer func() {
		a.emitterMu.Lock()
		a.eventEmitter = nil
		a.emitterMu.Unlock()
	}()

	provider, activeModel := a.providerSnapshot()
	if provider == nil {
		return "", fmt.Errorf("setup required: model provider is not configured")
	}

	objective = strings.TrimSpace(objective)
	if objective == "" {
		return "", fmt.Errorf("objective is required")
	}

	cleanTasks := make([]string, 0, len(tasks))
	for _, task := range tasks {
		task = strings.TrimSpace(task)
		if task == "" {
			continue
		}
		cleanTasks = append(cleanTasks, task)
	}
	if len(cleanTasks) == 0 {
		return "", fmt.Errorf("at least one non-empty task is required")
	}

	maxParallel := cfg.MaxParallelSubAgents
	if maxParallel <= 0 {
		maxParallel = defaultDelegationParallelism
	}
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultDelegationTimeout
	}

	subModel := strings.TrimSpace(cfg.SubAgentModel)
	if subModel == "" {
		subModel = activeModel
	}
	synthesisModel := strings.TrimSpace(cfg.SynthesisModel)
	if synthesisModel == "" {
		synthesisModel = activeModel
	}

	// Re-use the parent's context engine if available; otherwise create a
	// minimal noop one (covers test scenarios where no engine is wired).
	subCtxEngine := a.contextEngine
	if subCtxEngine == nil {
		subCtxEngine = agentctx.NewEngine(agentctx.NewNoOpEmbedder(), 8000, 5)
	}

	subMaxIter := a.maxIterations
	if subMaxIter <= 0 {
		// Keep delegated cooperation bounded even when the parent agent is in
		// unlimited mode; otherwise sub-agents can loop indefinitely and make
		// the hackathon cooperation path unreliable.
		subMaxIter = 6
	}

	// Collect tool names once — all sub-agents share the same executor.
	var toolNames []string
	if a.toolExecutor != nil {
		toolNames = a.toolExecutor.ListNames()
	}

	// Notify the caller that delegation is about to begin.
	a.emit(AgentEvent{
		Type:      EventDelegationStart,
		TaskCount: len(cleanTasks),
		Tasks:     cleanTasks,
	})

	runCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	results := make([]delegationResult, len(cleanTasks))
	sem := make(chan struct{}, maxParallel)
	var wg sync.WaitGroup

	for i, task := range cleanTasks {
		i := i
		task := task
		wg.Add(1)
		go func() {
			defer wg.Done()
			select {
			case <-runCtx.Done():
				results[i] = delegationResult{task: task, err: runCtx.Err()}
				return
			case sem <- struct{}{}:
			}
			defer func() { <-sem }()

			// Each sub-agent is a real Agent with the parent's tool registry
			// and context engine, but an isolated session (nil providers →
			// no shared history, no memory cross-contamination).
			subAgent := &Agent{
				provider:      provider,
				contextEngine: subCtxEngine,
				toolExecutor:  a.toolExecutor,
				cfg:           a.cfg,
				model:         subModel,
				maxIterations: subMaxIter,
			}

			// Unique session ID per sub-task — guarantees isolation.
			subSessionID := fmt.Sprintf("__sub_%d_%d__", time.Now().UnixNano(), i)

			// Notify that a new sub-agent is being spawned.
			a.emit(AgentEvent{
				Type:         EventSubagentSpawn,
				SubagentID:   subSessionID,
				Task:         task,
				Tools:        toolNames,
				MemoryAccess: "none",
			})

			taskMsg := fmt.Sprintf(
				"Overall objective: %s\n\nYour assigned sub-task: %s\n\nUse available tools if needed. Return concise findings.",
				objective, task,
			)

			capturedSubID := subSessionID
			resp, err := subAgent.runSubTask(runCtx, capturedSubID, taskMsg,
				func(toolName string) {
					a.emit(AgentEvent{
						Type:       EventSubagentTool,
						SubagentID: capturedSubID,
						ToolName:   toolName,
					})
				},
			)

			if err != nil {
				a.emit(AgentEvent{
					Type:       EventSubagentDone,
					SubagentID: capturedSubID,
					Error:      err.Error(),
				})
				results[i] = delegationResult{task: task, err: err}
				return
			}

			summary := strings.TrimSpace(resp.Text)
			if len(summary) > 200 {
				summary = summary[:200] + "..."
			}
			a.emit(AgentEvent{
				Type:          EventSubagentDone,
				SubagentID:    capturedSubID,
				ResultSummary: summary,
				Usage:         resp.Usage,
			})
			results[i] = delegationResult{
				task:   task,
				answer: strings.TrimSpace(resp.Text),
			}
		}()
	}

	wg.Wait()
	if err := runCtx.Err(); err != nil {
		return "", fmt.Errorf("delegation canceled: %w", err)
	}

	success := 0
	var synthesizedInput strings.Builder
	synthesizedInput.WriteString("Objective:\n")
	synthesizedInput.WriteString(objective)
	synthesizedInput.WriteString("\n\nSub-agent findings:\n")
	for i, res := range results {
		synthesizedInput.WriteString(fmt.Sprintf("%d. Task: %s\n", i+1, res.task))
		if res.err != nil {
			synthesizedInput.WriteString(fmt.Sprintf("Result: ERROR: %v\n\n", res.err))
			continue
		}
		success++
		if res.answer == "" {
			synthesizedInput.WriteString("Result: (empty)\n\n")
			continue
		}
		synthesizedInput.WriteString("Result: ")
		synthesizedInput.WriteString(res.answer)
		synthesizedInput.WriteString("\n\n")
	}
	if success == 0 {
		return "", fmt.Errorf("all delegated sub-tasks failed")
	}

	synthesisReq := ChatRequest{
		SystemPrompt: "You are a synthesis agent. Merge sub-agent findings into one coherent response. Resolve conflicts, call out uncertainty, and keep it concise.",
		Messages: []Message{
			{Role: "user", Content: synthesizedInput.String()},
		},
		MaxTokens: 1536,
		Model:     synthesisModel,
	}

	synthResp, err := provider.Chat(runCtx, synthesisReq)
	if err != nil {
		return "", fmt.Errorf("synthesizing delegated results: %w", err)
	}
	out := strings.TrimSpace(synthResp.Content)

	a.emit(AgentEvent{
		Type:          EventDelegationDone,
		SubagentCount: len(cleanTasks),
	})

	if out == "" {
		return synthesizedInput.String(), nil
	}
	return out, nil
}

// runSubTask executes a single sub-task for delegation.
//
// It runs the full tool-iteration loop (the sub-agent can call exec,
// web_fetch, web_search, etc.) but does NOT retry on transient provider
// errors — a failed sub-task is reported as an error in the aggregated
// results rather than blocking the whole delegation.
//
// Memory providers are intentionally nil so the sub-agent starts with a
// clean context — no history bleed between sub-agents or from the parent.
//
// onToolCall, if non-nil, is invoked each time the sub-agent calls a tool.
func (a *Agent) runSubTask(ctx context.Context, sessionID, userMessage string, onToolCall func(string)) (*Response, error) {
	provider, model := a.providerSnapshot()

	// Build tool definitions for the sub-agent.
	var toolDefs []ToolDef
	var toolNames []string
	var ctxToolDefs []agentctx.ToolDef
	if a.toolExecutor != nil {
		toolNames = a.toolExecutor.ListNames()
		if p, ok := a.toolExecutor.(ToolDefinitionProvider); ok {
			toolDefs = p.ToolDefinitions()
		}
		for _, name := range toolNames {
			ctxToolDefs = append(ctxToolDefs, agentctx.ToolDef{Name: name})
		}
	}

	systemPrompt := "You are a focused specialist sub-agent. Solve only the assigned sub-task using available tools. Return concise factual findings."

	// Assemble context. nil msgProvider and nil memProvider mean the sub-agent
	// starts with no history and no memory — fully isolated.
	assembled, err := a.contextEngine.Assemble(
		sessionID, userMessage, systemPrompt,
		ctxToolDefs, nil, nil,
	)
	if err != nil {
		return nil, fmt.Errorf("assembling sub-agent context: %w", err)
	}

	var messages []Message
	for _, cm := range assembled.Messages {
		messages = append(messages, Message{Role: cm.Role, Content: cm.Content})
	}

	response := &Response{}

	for iteration := 0; !a.hasIterationLimit() || iteration < a.maxIterations; iteration++ {
		response.Iterations = iteration + 1

		chatReq := ChatRequest{
			SystemPrompt: assembled.SystemPrompt,
			Messages:     messages,
			Tools:        toolDefs,
			MaxTokens:    1024,
			Model:        model,
		}

		// Direct call — no retry for sub-agents. Transient failures are
		// surfaced as task-level errors so other sub-tasks can still succeed.
		chatResp, err := provider.Chat(ctx, chatReq)
		if err != nil {
			return nil, err
		}

		response.Usage.InputTokens += chatResp.Usage.InputTokens
		response.Usage.OutputTokens += chatResp.Usage.OutputTokens
		response.Usage.LLMCalls++

		// No tool calls — sub-task is done.
		if len(chatResp.ToolCalls) == 0 {
			response.Text = chatResp.Content
			return response, nil
		}

		// No tool executor available — return whatever text we have.
		if a.toolExecutor == nil {
			response.Text = chatResp.Content
			return response, nil
		}

		// Add the assistant turn (with tool calls) to the conversation.
		messages = append(messages, Message{
			Role:      "assistant",
			Content:   chatResp.Content,
			ToolCalls: chatResp.ToolCalls,
		})

		// Execute each tool and append the result for the next iteration.
		for _, tc := range chatResp.ToolCalls {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			default:
			}

			if onToolCall != nil {
				onToolCall(tc.Name)
			}

			var result string
			if !a.toolExecutor.HasTool(tc.Name) {
				result = fmt.Sprintf("Error: the tool '%s' does not exist.", tc.Name)
			} else {
				var execErr error
				result, execErr = a.toolExecutor.Execute(ctx, tc.Name, tc.Arguments)
				if execErr != nil {
					result = fmt.Sprintf("Error: %v", execErr)
				}
			}

			messages = append(messages, Message{
				Role:       "tool",
				Content:    WrapToolResult(tc.Name, result),
				ToolCallID: tc.ID,
			})
		}
	}

	response.Text = "[Sub-agent reached maximum tool iterations.]"
	return response, nil
}
