// Package cli provides the interactive terminal chat interface.
package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chzyer/readline"
	"github.com/openclio/openclio/internal/agent"
	"github.com/openclio/openclio/internal/config"
	agentctx "github.com/openclio/openclio/internal/context"
	"github.com/openclio/openclio/internal/cost"
	eam "github.com/openclio/openclio/internal/memory/eam"
	"github.com/openclio/openclio/internal/storage"
)

// totalUsage tracks cumulative token usage across the session.
type totalUsage struct {
	inputTokens  int
	outputTokens int
	llmCalls     int
}

// CLI is the interactive terminal chat interface.
type CLI struct {
	agent         *agent.Agent
	sessions      *storage.SessionStore
	messages      *storage.MessageStore
	memory        agentctx.MemoryProvider
	beliefStore   eam.BeliefStore
	contextEngine *agentctx.Engine
	costTracker   *cost.Tracker
	cfg           *config.Config
	dataDir       string
	sessionID     string
	provider      string
	model         string
	totalUsage    totalUsage
	scannerBuf    int
	// For /skill display
	workspaceName string
	cronJobs      []string

	// For /debug display
	lastContext *agentctx.AssembledContext
}

// NewCLI creates a new CLI instance.
func NewCLI(
	agentInstance *agent.Agent,
	sessions *storage.SessionStore,
	messages *storage.MessageStore,
	contextEngine *agentctx.Engine,
	costTracker *cost.Tracker,
	fullConfig *config.Config,
	dataDir string,
	cfg config.CLIConfig,
	provider, model string,
	workspaceName string,
	cronJobs []string,
) *CLI {
	buf := cfg.ScannerBuffer
	if buf == 0 {
		buf = 64 * 1024
	}
	return &CLI{
		agent:         agentInstance,
		sessions:      sessions,
		messages:      messages,
		contextEngine: contextEngine,
		costTracker:   costTracker,
		cfg:           fullConfig,
		dataDir:       dataDir,
		provider:      provider,
		model:         model,
		scannerBuf:    buf,
		workspaceName: workspaceName,
		cronJobs:      cronJobs,
	}
}

// SetMemoryProvider wires an optional tier-3 memory provider for chat runs.
func (c *CLI) SetMemoryProvider(provider agentctx.MemoryProvider) {
	c.memory = provider
}

// SetBeliefStore wires the EAM belief store for /beliefs display.
func (c *CLI) SetBeliefStore(store eam.BeliefStore) {
	c.beliefStore = store
}

func (c *CLI) cmdBeliefs() {
	if c.beliefStore == nil {
		PrintInfo("Belief store not available (EAM disabled or not wired).")
		return
	}
	ctx := context.Background()
	beliefs, err := c.beliefStore.GetActive(ctx, 30)
	if err != nil {
		PrintError("Failed to load beliefs: " + err.Error())
		return
	}
	if len(beliefs) == 0 {
		PrintInfo("No active beliefs yet. Keep chatting — beliefs are extracted automatically.")
		return
	}
	fmt.Println()
	fmt.Printf("%sEpistemic beliefs (%d active):%s\n", colorBold(), len(beliefs), colorReset())
	now := time.Now()
	for _, b := range beliefs {
		stale := ""
		if b.ValidUntil != nil && now.After(*b.ValidUntil) {
			stale = " [EXPIRED]"
		}
		age := now.Sub(b.UpdatedAt).Round(time.Hour)
		fmt.Printf("  [%.2f] (%s / %s)%s %s\n", b.Confidence, b.Category, b.Provenance, stale, b.Claim)
		fmt.Printf("         updated %s ago · accessed %d×\n", age, b.AccessCount)
	}
	fmt.Println()
}

// Run starts the interactive REPL loop with readline support
// (arrow-key history, Ctrl+C to cancel, multi-line paste).
func (c *CLI) Run() error {
	// Create a new session
	session, err := c.sessions.Create("cli", "local")
	if err != nil {
		return fmt.Errorf("creating session: %w", err)
	}
	c.sessionID = session.ID

	PrintWelcome(c.sessionID, c.provider, c.model)

	// Determine history file path (~/.openclio/history).
	historyFile := ""
	if home, err := os.UserHomeDir(); err == nil {
		historyFile = filepath.Join(home, ".openclio", "history")
	}

	rl, err := readline.NewEx(&readline.Config{
		Prompt:          colorBold() + "> " + colorReset(),
		HistoryFile:     historyFile,
		HistoryLimit:    500,
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		AutoComplete:    slashCompleter(),
	})
	if err != nil {
		// Fallback: readline unavailable (e.g. non-TTY pipe) — not fatal
		return c.runScanner()
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err == readline.ErrInterrupt {
			if strings.TrimSpace(line) == "" {
				break // Ctrl+C on empty line → exit
			}
			continue // Ctrl+C mid-input → clear and retry
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("input error: %w", err)
		}

		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			break
		}

		if c.HandleCommand(input) {
			continue
		}

		c.chat(input)
	}

	fmt.Println()
	PrintInfo("Goodbye!")
	return nil
}

// runScanner is the non-readline fallback for non-TTY environments (pipes, tests).
func (c *CLI) runScanner() error {
	buf := make([]byte, c.scannerBuf)
	scanner := bufScanner{buf: buf}
	for {
		fmt.Printf("%s> %s", colorBold(), colorReset())
		line, ok := scanner.readLine()
		if !ok {
			break
		}
		input := strings.TrimSpace(line)
		if input == "" {
			continue
		}
		if input == "exit" || input == "quit" {
			break
		}
		if !c.HandleCommand(input) {
			c.chat(input)
		}
	}
	fmt.Println()
	PrintInfo("Goodbye!")
	return nil
}

// bufScanner is a minimal line reader used when readline is unavailable.
type bufScanner struct {
	buf      []byte
	leftover string
}

func (s *bufScanner) readLine() (string, bool) {
	for {
		if idx := strings.IndexByte(s.leftover, '\n'); idx >= 0 {
			line := s.leftover[:idx]
			s.leftover = s.leftover[idx+1:]
			return line, true
		}
		n, err := os.Stdin.Read(s.buf)
		if n > 0 {
			s.leftover += string(s.buf[:n])
		}
		if err != nil {
			if s.leftover != "" {
				line := s.leftover
				s.leftover = ""
				return line, true
			}
			return "", false
		}
	}
}

func (c *CLI) chat(userMessage string) {
	// Store user message
	userTokens := agentctx.EstimateTokens(userMessage)
	if _, err := c.messages.Insert(c.sessionID, "user", userMessage, userTokens); err != nil {
		PrintError("Failed to store message: " + err.Error())
		return
	}

	// Update session activity
	c.sessions.UpdateLastActive(c.sessionID)

	// Create message provider
	msgProvider := &cliMessageProvider{
		messages:  c.messages,
		sessionID: c.sessionID,
	}

	// Run agent with a timeout — prevents hanging forever on slow LLM responses.
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Print the assistant header before tokens start arriving.
	firstToken := true
	onToken := func(token string) {
		if firstToken {
			fmt.Printf("\n%s%s⚡ Agent%s\n%s", colorBold(), colorGreen(), colorReset(), colorGreen())
			firstToken = false
		}
		fmt.Print(token)
	}

	onToolCall := func(toolName, result string) {
		// Ensure we're on a new line before printing a tool call.
		if !firstToken {
			fmt.Print(colorReset())
			firstToken = true // reset so next text gets the header again
		}
		PrintToolCall(toolName, "", result, "")
	}

	resp, err := c.agent.RunStream(ctx, c.sessionID, userMessage, msgProvider, c.memory, onToken, onToolCall, nil)
	if err != nil {
		PrintError("Agent error: " + err.Error())
		return
	}

	// Close the green color block and add newline after streamed text.
	if !firstToken {
		fmt.Printf("%s\n", colorReset())
	} else if resp.Text != "" {
		// Provider doesn't stream — RunStream fell back to Run and called onToken once.
		// The token callback already printed; just ensure the newline.
		fmt.Printf("%s\n", colorReset())
	}

	// Store for /debug commands
	c.lastContext = resp.AssembledContext

	// Display usage
	PrintUsage(resp.Usage.InputTokens, resp.Usage.OutputTokens, resp.Usage.LLMCalls, resp.Duration.Milliseconds())

	// Store assistant response
	assistantTokens := agentctx.EstimateTokens(resp.Text)
	c.messages.Insert(c.sessionID, "assistant", resp.Text, assistantTokens)

	// Track cumulative usage
	c.totalUsage.inputTokens += resp.Usage.InputTokens
	c.totalUsage.outputTokens += resp.Usage.OutputTokens
	c.totalUsage.llmCalls += resp.Usage.LLMCalls
}

// cliMessageProvider adapts storage to the context engine interface.
type cliMessageProvider struct {
	messages  *storage.MessageStore
	sessionID string
}

func (p *cliMessageProvider) GetRecentMessages(sessionID string, limit int) ([]agentctx.ContextMessage, error) {
	msgs, err := p.messages.GetRecent(sessionID, limit)
	if err != nil {
		return nil, err
	}
	var result []agentctx.ContextMessage
	for _, m := range msgs {
		result = append(result, agentctx.ContextMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return result, nil
}

func (p *cliMessageProvider) GetStoredEmbeddings(sessionID string) ([]agentctx.StoredEmbedding, error) {
	msgs, err := p.messages.GetEmbeddings(sessionID)
	if err != nil {
		return nil, err
	}
	var result []agentctx.StoredEmbedding
	for _, m := range msgs {
		result = append(result, agentctx.StoredEmbedding{
			MessageID: m.ID,
			SessionID: m.SessionID,
			Role:      m.Role,
			Content:   m.Content,
			Summary:   m.Summary,
			Tokens:    m.Tokens,
			Embedding: m.Embedding,
		})
	}
	return result, nil
}

func (p *cliMessageProvider) SearchKnowledge(query, nodeType string, limit int) ([]agentctx.KnowledgeNode, error) {
	nodes, err := p.messages.SearchKnowledge(query, nodeType, limit)
	if err != nil {
		return nil, err
	}
	out := make([]agentctx.KnowledgeNode, 0, len(nodes))
	for _, n := range nodes {
		out = append(out, agentctx.KnowledgeNode{
			ID:         n.ID,
			Type:       n.Type,
			Name:       n.Name,
			Confidence: n.Confidence,
		})
	}
	return out, nil
}

func (p *cliMessageProvider) GetOldMessages(sessionID string, keepRecentTurns int) ([]agent.CompactionMessage, error) {
	msgs, err := p.messages.GetOldMessages(sessionID, keepRecentTurns)
	if err != nil {
		return nil, err
	}
	result := make([]agent.CompactionMessage, 0, len(msgs))
	for _, m := range msgs {
		result = append(result, agent.CompactionMessage{
			ID:      m.ID,
			Role:    m.Role,
			Content: m.Content,
		})
	}
	return result, nil
}

func (p *cliMessageProvider) ArchiveMessages(sessionID string, olderThanID int64) (int64, error) {
	return p.messages.ArchiveMessages(sessionID, olderThanID)
}

func (p *cliMessageProvider) InsertCompactionSummary(sessionID, content string, tokens int) error {
	_, err := p.messages.Insert(sessionID, "system", content, tokens)
	return err
}
