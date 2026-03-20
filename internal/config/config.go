// Package config handles loading and validating the agent configuration.
// Configuration is loaded from a YAML file with environment variable overlays.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// Config is the top-level configuration for the agent.
type Config struct {
	Gateway     GatewayConfig     `yaml:"gateway"`
	Model       ModelConfig       `yaml:"model"`
	ModelRouter ModelRouterConfig `yaml:"model_router"`
	Embeddings  EmbeddingsConfig  `yaml:"embeddings"`
	Context     ContextConfig     `yaml:"context"`
	Memory      MemoryConfig      `yaml:"memory"`
	Epistemic   EpistemicConfig   `yaml:"epistemic"`
	MCPServers  []MCPServerConfig `yaml:"mcp_servers,omitempty"`
	Channels    ChannelsConfig    `yaml:"channels"`
	Agent       AgentConfig       `yaml:"agent"`
	Tools       ToolsConfig       `yaml:"tools"`
	CLI         CLIConfig         `yaml:"cli"`
	Logging     LoggingConfig     `yaml:"logging"`
	Retention   RetentionConfig   `yaml:"retention"`
	Cron        []CronJob         `yaml:"cron"`
	Auth        AuthConfig        `yaml:"auth,omitempty"`

	// DataDir is runtime-only (not serialized) and points at ~/.openclio.
	DataDir string `yaml:"-"`
}

// AuthConfig holds optional OAuth and related auth settings.
type AuthConfig struct {
	OpenAIOAuth OpenAIOAuthConfig `yaml:"openai_oauth,omitempty"`
}

// OpenAIOAuthConfig configures "Sign in with OpenAI" (OAuth 2.0 + PKCE).
type OpenAIOAuthConfig struct {
	Enabled          bool   `yaml:"enabled"`
	ClientID         string `yaml:"client_id"`
	ClientSecret     string `yaml:"client_secret,omitempty"`
	AuthorizationURL string `yaml:"authorization_url"`
	TokenURL         string `yaml:"token_url"`
	Scope            string `yaml:"scope,omitempty"`
}

// ModelRouterConfig configures heuristic model routing.
type ModelRouterConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Strategy       string `yaml:"strategy"` // cost_optimized | quality_first | speed_first | privacy_first
	CheapModel     string `yaml:"cheap_model,omitempty"`
	MidModel       string `yaml:"mid_model,omitempty"`
	ExpensiveModel string `yaml:"expensive_model,omitempty"`
	PrivacyModel   string `yaml:"privacy_model,omitempty"`
}

// RetentionConfig controls automatic data pruning.
type RetentionConfig struct {
	// SessionsDays deletes sessions older than this many days.
	// Set to 0 (default) to keep sessions forever.
	SessionsDays int `yaml:"sessions_days"`
	// MessagesPerSession trims the oldest messages when a session exceeds this
	// count. Set to 0 (default) to disable trimming.
	MessagesPerSession int `yaml:"messages_per_session"`
}

// AgentConfig configures the core agent behavior.
type AgentConfig struct {
	// Name is the display name of the agent. Defaults to "openclio".
	// Set this in config.yaml to give your agent a custom name.
	Name                string                `yaml:"name,omitempty"`
	MaxToolIterations   int                   `yaml:"max_tool_iterations"`
	MaxTokensPerSession int                   `yaml:"max_tokens_per_session"`
	MaxTokensPerDay     int                   `yaml:"max_tokens_per_day"`
	Delegation          AgentDelegationConfig `yaml:"delegation"`

	// ReasoningEnabled enables extended thinking/reasoning.
	// Anthropic: injects a thinking block (requires claude-3-7-sonnet or newer).
	// OpenAI: sends reasoning_effort for o1/o3/o4-mini models.
	ReasoningEnabled bool `yaml:"reasoning_enabled"`
	// ReasoningBudget is the number of tokens allocated for internal reasoning.
	// Anthropic only. Minimum 1024. Defaults to 8000 when enabled.
	ReasoningBudget int `yaml:"reasoning_budget,omitempty"`
}

// AgentDelegationConfig controls optional parallel sub-agent delegation.
type AgentDelegationConfig struct {
	Enabled              bool          `yaml:"enabled"`
	MaxParallelSubAgents int           `yaml:"max_parallel_sub_agents"`
	SubAgentModel        string        `yaml:"sub_agent_model"`
	SynthesisModel       string        `yaml:"synthesis_model"`
	Timeout              time.Duration `yaml:"timeout"`
}

// ToolsConfig configures the tool system.
type ToolsConfig struct {
	MaxOutputSize int               `yaml:"max_output_size"`
	ScrubOutput   bool              `yaml:"scrub_output"` // redact passwords/secrets from tool results (default: true)
	Packs         []string          `yaml:"packs,omitempty"`
	MCPPresets    []string          `yaml:"mcp_presets,omitempty"`
	AllowedTools  []string          `yaml:"allowed_tools,omitempty"`
	Exec          ExecToolConfig    `yaml:"exec"`
	Browser       BrowserToolConfig `yaml:"browser"`
	WebSearch     *WebSearchConfig  `yaml:"web_search,omitempty"`
}

// CLIConfig configures the interactive terminal.
type CLIConfig struct {
	ScannerBuffer int `yaml:"scanner_buffer"`
}

// GatewayConfig configures the HTTP/WebSocket server.
type GatewayConfig struct {
	Port        int    `yaml:"port"`
	Bind        string `yaml:"bind"`
	TLSCertFile string `yaml:"tls_cert_file"` // PEM cert file; enables TLS when set together with tls_key_file
	TLSKeyFile  string `yaml:"tls_key_file"`  // PEM key file
	GRPCPort    int    `yaml:"grpc_port"`     // gRPC port for out-of-process adapters (0 = disabled)
}

// ModelConfig configures the LLM provider.
type ModelConfig struct {
	Provider  string `yaml:"provider"`
	Model     string `yaml:"model"`
	APIKeyEnv string `yaml:"api_key_env"`
	// BaseURL overrides the default API endpoint. Required for openai-compat;
	// optional for openai (overrides OPENAI_BASE_URL env var).
	BaseURL string `yaml:"base_url,omitempty"`
	// Name sets the display name reported by the provider (used by openai-compat).
	Name              string            `yaml:"name,omitempty"`
	FallbackProviders []string          `yaml:"fallback_providers,omitempty"`
	FallbackModels    map[string]string `yaml:"fallback_models,omitempty"`
	FallbackAPIKeyEnv map[string]string `yaml:"fallback_api_key_env,omitempty"`

	// OllamaNumCtx sets the context window size sent to Ollama (num_ctx option).
	// Ollama's default is 4096 which is far too small for a personal AI agent.
	// Recommended: 16384 for 7B models on 8GB RAM, 32768 for 13B+ on 16GB+.
	// Set this to at least max_tokens_per_call + 8000 to avoid silent truncation.
	OllamaNumCtx int `yaml:"ollama_num_ctx,omitempty"`
}

// APIKey reads the actual API key from the environment variable.
// The key is never stored in the config file.
func (m *ModelConfig) APIKey() string {
	if m.APIKeyEnv == "" {
		return ""
	}
	return os.Getenv(m.APIKeyEnv)
}

// ContextConfig configures the context engine.
type ContextConfig struct {
	MaxTokensPerCall     int     `yaml:"max_tokens_per_call"`
	HistoryRetrievalK    int     `yaml:"history_retrieval_k"`
	ProactiveCompaction  float64 `yaml:"proactive_compaction"`
	CompactionKeepRecent int     `yaml:"compaction_keep_recent"`
	CompactionModel      string  `yaml:"compaction_model"`
	ToolResultSummary    bool    `yaml:"tool_result_summary"`
}

// MemoryConfig configures tier-3 semantic memory provider behavior.
type MemoryConfig struct {
	Provider          string                `yaml:"provider"` // workspace | mem0style
	Mode              string                `yaml:"mode,omitempty"`
	EAMServingEnabled bool                  `yaml:"eam_serving_enabled"`
	Mem0Style         MemoryMem0StyleConfig `yaml:"mem0style"`
}

type MemoryMem0StyleConfig struct {
	DBPath      string  `yaml:"db_path"`
	MinSalience float64 `yaml:"min_salience"`
}

// EpistemicConfig controls the optional EAM pipeline.
type EpistemicConfig struct {
	Enabled      bool                        `yaml:"enabled"`
	Ambient      EpistemicAmbientConfig      `yaml:"ambient"`
	Anticipation EpistemicAnticipationConfig `yaml:"anticipation"`
}

// EpistemicAmbientConfig controls Phase 3 ambient observation.
type EpistemicAmbientConfig struct {
	Enabled           bool                             `yaml:"enabled"`
	SignalExpiryHours int                              `yaml:"signal_expiry_hours"`
	Calendar          EpistemicAmbientCalendarConfig   `yaml:"calendar"`
	Filesystem        EpistemicAmbientFilesystemConfig `yaml:"filesystem"`
	Temporal          EpistemicAmbientTemporalConfig   `yaml:"temporal"`
}

type EpistemicAmbientCalendarConfig struct {
	Enabled           bool     `yaml:"enabled"`
	ICSPaths          []string `yaml:"ics_paths"`
	CalDAVURL         string   `yaml:"caldav_url"`
	CalDAVUsername    string   `yaml:"caldav_username"`
	CalDAVPasswordEnv string   `yaml:"caldav_password_env"`
	LookaheadHours    int      `yaml:"lookahead_hours"`
	LookbehindHours   int      `yaml:"lookbehind_hours"`
}

type EpistemicAmbientFilesystemConfig struct {
	Enabled            bool     `yaml:"enabled"`
	WatchPaths         []string `yaml:"watch_paths"`
	ExcludePatterns    []string `yaml:"exclude_patterns"`
	MaxEventsPerMinute int      `yaml:"max_events_per_minute"`
}

type EpistemicAmbientTemporalConfig struct {
	Enabled            bool `yaml:"enabled"`
	PatternMinSessions int  `yaml:"pattern_min_sessions"`
}

type EpistemicAnticipationConfig struct {
	Enabled           bool                           `yaml:"enabled"`
	PreLoadTopK       int                            `yaml:"pre_load_top_k"`
	MinRelevanceScore float64                        `yaml:"min_relevance_score"`
	StagingTTLMinutes int                            `yaml:"staging_ttl_minutes"`
	GapDetection      EpistemicAnticipationGapConfig `yaml:"gap_detection"`
}

type EpistemicAnticipationGapConfig struct {
	Enabled             bool    `yaml:"enabled"`
	SparseThreshold     int     `yaml:"sparse_threshold"`
	LowConfidenceCutoff float64 `yaml:"low_confidence_threshold"`
	MaxGapsToSurface    int     `yaml:"max_gaps_to_surface"`
	GapAskCooldownHours int     `yaml:"gap_ask_cooldown_hours"`
}

// EmbeddingsConfig configures semantic embedding generation.
type EmbeddingsConfig struct {
	Provider  string `yaml:"provider"` // auto | openai | ollama
	Model     string `yaml:"model"`
	APIKeyEnv string `yaml:"api_key_env"`
	BaseURL   string `yaml:"base_url"`
}

// ChannelsConfig configures messaging channel adapters.
type ChannelsConfig struct {
	AllowAll bool            `yaml:"allow_all"` // default true; false = only approved senders
	Telegram *TelegramConfig `yaml:"telegram,omitempty"`
	WhatsApp *WhatsAppConfig `yaml:"whatsapp,omitempty"`
	Discord  *DiscordConfig  `yaml:"discord,omitempty"`
	Slack    *SlackConfig    `yaml:"slack,omitempty"`
}

// SlackConfig configures the Slack adapter.
type SlackConfig struct {
	TokenEnv string `yaml:"token_env"` // env var holding the Slack bot token (xoxb-...)
}

// TelegramConfig configures the Telegram adapter.
type TelegramConfig struct {
	TokenEnv string `yaml:"token_env"`
}

// WhatsAppConfig configures the WhatsApp adapter.
type WhatsAppConfig struct {
	Enabled bool `yaml:"enabled"`
	// TokenEnv and WebhookURL are reserved for a future Cloud API mode.
	// Currently the adapter uses whatsmeow QR-code login (no API key required).
	// Setting these fields has no effect — they are ignored by the current implementation.
	TokenEnv   string `yaml:"token_env,omitempty"`
	WebhookURL string `yaml:"webhook_url,omitempty"`
	DataDir    string `yaml:"data_dir"` // directory for whatsmeow session SQLite (defaults to ~/.openclio)
}

// DiscordConfig configures the Discord adapter.
type DiscordConfig struct {
	TokenEnv string `yaml:"token_env"`
	AppIDEnv string `yaml:"app_id_env"` // optional, for slash commands
}

// ExecToolConfig configures shell command execution.
type ExecToolConfig struct {
	Sandbox             string        `yaml:"sandbox"`
	Timeout             time.Duration `yaml:"timeout"`
	DockerImage         string        `yaml:"docker_image"`
	NetworkAccess       bool          `yaml:"network_access"`
	RequireConfirmation bool          `yaml:"require_confirmation"`
	Profile             string        `yaml:"profile,omitempty"`
	AllowedCommands     []string      `yaml:"allowed_commands,omitempty"`
	ApprovalOnBlock     bool          `yaml:"approval_on_block,omitempty"`
}

// BrowserToolConfig configures browser automation.
type BrowserToolConfig struct {
	Enabled      bool          `yaml:"enabled"`
	ChromePath   string        `yaml:"chrome_path"`
	ChromiumPath string        `yaml:"chromium_path,omitempty"` // alias supported for compatibility
	Headless     bool          `yaml:"headless"`
	Timeout      time.Duration `yaml:"timeout"`
}

// MCPServerConfig configures one MCP stdio server.
type MCPServerConfig struct {
	Enabled *bool             `yaml:"enabled,omitempty"`
	Name    string            `yaml:"name"`
	Command string            `yaml:"command"`
	Args    []string          `yaml:"args,omitempty"`
	Env     map[string]string `yaml:"env,omitempty"` // values may be literal or env var names
}

// WebSearchConfig configures web search.
type WebSearchConfig struct {
	Provider  string `yaml:"provider"`
	APIKeyEnv string `yaml:"api_key_env"`
}

// LoggingConfig configures structured logging.
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Output string `yaml:"output"` // "stderr", "stdout", or a file path
}

// CronJob defines a scheduled agent task.
type CronJob struct {
	Name        string `yaml:"name"`
	Schedule    string `yaml:"schedule"`          // standard cron expression (for time-based jobs)
	Trigger     string `yaml:"trigger,omitempty"` // event-driven watch trigger, e.g. "every 6 hours"
	Prompt      string `yaml:"prompt"`
	Channel     string `yaml:"channel,omitempty"`      // adapter to send result to
	SessionMode string `yaml:"session_mode,omitempty"` // isolated | shared (default: isolated)
	TimeoutSec  int    `yaml:"timeout_seconds,omitempty"`
}

// Load reads a YAML config file and returns a Config.
// Environment variables can override specific fields.
func Load(path string) (*Config, error) {
	// Load ~/.openclio/.env (or sibling .env) as a fallback source for API keys.
	// Existing process environment always takes precedence.
	_ = LoadDotEnv(filepath.Join(filepath.Dir(path), ".env"))

	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No config file — use defaults + env overrides.
			applyEnvOverrides(cfg)
			normalizeBrowserPath(cfg)
			if err := cfg.Validate(); err != nil {
				return nil, fmt.Errorf("invalid configuration: %w", err)
			}
			cfg.DataDir = filepath.Dir(path)
			return cfg, nil
		}
		return nil, fmt.Errorf("reading config %s: %w", path, err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config %s: %w", path, err)
	}

	// Apply environment variable overrides
	applyEnvOverrides(cfg)
	normalizeBrowserPath(cfg)

	// Validate configuration before returning
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Runtime-only field: the config data directory.
	cfg.DataDir = filepath.Dir(path)

	return cfg, nil
}

// Save writes the config to disk atomically.
// API keys are not persisted here (only env var names), matching Config design.
func Save(path string, cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config cannot be nil")
	}
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("creating config directory %s: %w", dir, err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshalling config: %w", err)
	}

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("writing config temp file %s: %w", tmpPath, err)
	}
	if err := os.Rename(tmpPath, path); err != nil {
		return fmt.Errorf("atomic rename %s -> %s failed: %w", tmpPath, path, err)
	}
	return nil
}

// applyEnvOverrides lets environment variables override config values.
func applyEnvOverrides(cfg *Config) {
	if v := os.Getenv("AGENT_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.Gateway.Port)
	}
	if v := os.Getenv("AGENT_BIND"); v != "" {
		cfg.Gateway.Bind = v
	}
	if v := os.Getenv("AGENT_MODEL_PROVIDER"); v != "" {
		cfg.Model.Provider = v
	}
	if v := os.Getenv("AGENT_MODEL"); v != "" {
		cfg.Model.Model = v
	}
	if v := os.Getenv("AGENT_MEMORY_PROVIDER"); v != "" {
		cfg.Memory.Provider = v
	}
	if v := os.Getenv("AGENT_MEMORY_MODE"); v != "" {
		cfg.Memory.Mode = v
	}
	if v := os.Getenv("AGENT_MEMORY_EAM_ENABLED"); v != "" {
		if parsed, err := strconv.ParseBool(strings.TrimSpace(v)); err == nil {
			cfg.Memory.EAMServingEnabled = parsed
		}
	}
	if v := os.Getenv("AGENT_LOG_LEVEL"); v != "" {
		cfg.Logging.Level = v
	}
}

func normalizeBrowserPath(cfg *Config) {
	if cfg == nil {
		return
	}
	if strings.TrimSpace(cfg.Tools.Browser.ChromePath) == "" && strings.TrimSpace(cfg.Tools.Browser.ChromiumPath) != "" {
		cfg.Tools.Browser.ChromePath = strings.TrimSpace(cfg.Tools.Browser.ChromiumPath)
	}
}

// Validate checks the configuration for obvious errors.
func (c *Config) Validate() error {
	if c.Gateway.Port < 1 || c.Gateway.Port > 65535 {
		return fmt.Errorf("gateway port must be between 1 and 65535, got %d", c.Gateway.Port)
	}
	if strings.TrimSpace(c.Model.Provider) == "" {
		// Allowed in setup mode; user picks provider during init or dashboard setup.
	}
	for _, p := range c.Model.FallbackProviders {
		switch p {
		case "anthropic", "openai", "gemini", "ollama", "cohere",
			"groq", "deepseek", "mistral", "xai", "cerebras",
			"together", "fireworks", "perplexity", "openrouter",
			"kimi", "sambanova", "lambda", "lmstudio", "openai-compat":
		default:
			return fmt.Errorf("model fallback provider %q is not a recognised provider", p)
		}
	}
	switch c.Embeddings.Provider {
	case "", "auto", "openai", "ollama":
	default:
		return fmt.Errorf("embeddings provider must be one of: auto, openai, ollama")
	}
	switch strings.ToLower(strings.TrimSpace(c.Memory.Provider)) {
	case "", "workspace", "mem0style":
	default:
		return fmt.Errorf("memory provider must be one of: workspace, mem0style")
	}
	switch strings.ToLower(strings.TrimSpace(c.Memory.Mode)) {
	case "", "off", "standard", "enhanced":
	default:
		return fmt.Errorf("memory mode must be one of: off, standard, enhanced")
	}
	if c.Memory.Mem0Style.MinSalience < 0 || c.Memory.Mem0Style.MinSalience > 1 {
		return fmt.Errorf("memory mem0style min_salience must be between 0 and 1")
	}
	if c.Context.MaxTokensPerCall <= 0 {
		return fmt.Errorf("context max tokens per call must be positive")
	}
	if c.Context.HistoryRetrievalK <= 0 {
		return fmt.Errorf("context history retrieval K must be positive")
	}
	if c.Context.ProactiveCompaction < 0 || c.Context.ProactiveCompaction > 1 {
		return fmt.Errorf("context proactive_compaction must be between 0 and 1")
	}
	if c.Context.CompactionKeepRecent < 0 {
		return fmt.Errorf("context compaction_keep_recent cannot be negative")
	}
	if c.Epistemic.Ambient.SignalExpiryHours < 0 {
		return fmt.Errorf("epistemic ambient signal_expiry_hours cannot be negative")
	}
	if c.Epistemic.Ambient.Calendar.LookaheadHours < 0 {
		return fmt.Errorf("epistemic ambient calendar lookahead_hours cannot be negative")
	}
	if c.Epistemic.Ambient.Calendar.LookbehindHours < 0 {
		return fmt.Errorf("epistemic ambient calendar lookbehind_hours cannot be negative")
	}
	if c.Epistemic.Ambient.Filesystem.MaxEventsPerMinute < 0 {
		return fmt.Errorf("epistemic ambient filesystem max_events_per_minute cannot be negative")
	}
	if c.Epistemic.Ambient.Temporal.PatternMinSessions < 0 {
		return fmt.Errorf("epistemic ambient temporal pattern_min_sessions cannot be negative")
	}
	if c.Epistemic.Anticipation.PreLoadTopK < 0 {
		return fmt.Errorf("epistemic anticipation pre_load_top_k cannot be negative")
	}
	if c.Epistemic.Anticipation.MinRelevanceScore < 0 || c.Epistemic.Anticipation.MinRelevanceScore > 1 {
		return fmt.Errorf("epistemic anticipation min_relevance_score must be between 0 and 1")
	}
	if c.Epistemic.Anticipation.StagingTTLMinutes < 0 {
		return fmt.Errorf("epistemic anticipation staging_ttl_minutes cannot be negative")
	}
	if c.Epistemic.Anticipation.GapDetection.SparseThreshold < 0 {
		return fmt.Errorf("epistemic anticipation gap_detection sparse_threshold cannot be negative")
	}
	if c.Epistemic.Anticipation.GapDetection.LowConfidenceCutoff < 0 || c.Epistemic.Anticipation.GapDetection.LowConfidenceCutoff > 1 {
		return fmt.Errorf("epistemic anticipation gap_detection low_confidence_threshold must be between 0 and 1")
	}
	if c.Epistemic.Anticipation.GapDetection.MaxGapsToSurface < 0 {
		return fmt.Errorf("epistemic anticipation gap_detection max_gaps_to_surface cannot be negative")
	}
	if c.Epistemic.Anticipation.GapDetection.GapAskCooldownHours < 0 {
		return fmt.Errorf("epistemic anticipation gap_detection gap_ask_cooldown_hours cannot be negative")
	}
	for _, s := range c.MCPServers {
		if s.Enabled != nil && !*s.Enabled {
			continue
		}
		if s.Name == "" {
			return fmt.Errorf("mcp_servers entry has empty name")
		}
		if s.Command == "" {
			return fmt.Errorf("mcp_servers[%s] has empty command", s.Name)
		}
	}
	if c.Agent.MaxToolIterations < 0 {
		return fmt.Errorf("agent max tool iterations cannot be negative, got %d", c.Agent.MaxToolIterations)
	}
	if c.Agent.MaxTokensPerSession < 0 {
		return fmt.Errorf("agent max tokens per session cannot be negative")
	}
	if c.Agent.MaxTokensPerDay < 0 {
		return fmt.Errorf("agent max tokens per day cannot be negative")
	}
	if c.Agent.Delegation.MaxParallelSubAgents < 0 {
		return fmt.Errorf("agent delegation max_parallel_sub_agents cannot be negative")
	}
	if c.Agent.Delegation.Enabled && c.Agent.Delegation.MaxParallelSubAgents == 0 {
		return fmt.Errorf("agent delegation max_parallel_sub_agents must be positive when enabled")
	}
	if c.Agent.Delegation.Timeout < 0 {
		return fmt.Errorf("agent delegation timeout cannot be negative")
	}
	switch c.Tools.Exec.Sandbox {
	case "", "none", "namespace", "docker":
	default:
		return fmt.Errorf("tools.exec.sandbox must be one of: none, namespace, docker")
	}
	switch c.Tools.Exec.Profile {
	case "", "safe", "developer", "builder", "power-user":
	default:
		return fmt.Errorf("tools.exec.profile must be one of: safe, developer, builder, power-user")
	}
	if c.Tools.Browser.Timeout < 0 {
		return fmt.Errorf("tools.browser.timeout cannot be negative")
	}
	for _, j := range c.Cron {
		if strings.TrimSpace(j.Schedule) == "" && strings.TrimSpace(j.Trigger) == "" {
			return fmt.Errorf("cron job %q must define either schedule or trigger", j.Name)
		}
		if strings.TrimSpace(j.Schedule) != "" && strings.TrimSpace(j.Trigger) != "" {
			return fmt.Errorf("cron job %q cannot define both schedule and trigger", j.Name)
		}
		switch j.SessionMode {
		case "", "isolated", "shared":
		default:
			return fmt.Errorf("cron job %q has invalid session_mode %q (valid: isolated, shared)", j.Name, j.SessionMode)
		}
		if j.TimeoutSec < 0 {
			return fmt.Errorf("cron job %q timeout_seconds cannot be negative", j.Name)
		}
	}
	switch c.ModelRouter.Strategy {
	case "", "cost_optimized", "quality_first", "speed_first", "privacy_first":
	default:
		return fmt.Errorf("model_router.strategy must be one of: cost_optimized, quality_first, speed_first, privacy_first")
	}
	return nil
}
