package control

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/openclio/openclio/internal/config"
)

type ModelSummary struct {
	Provider          string            `json:"provider"`
	Model             string            `json:"model"`
	BaseURL           string            `json:"base_url,omitempty"`
	APIKeyEnv         string            `json:"api_key_env,omitempty"`
	FallbackProviders []string          `json:"fallback_providers,omitempty"`
	FallbackModels    map[string]string `json:"fallback_models,omitempty"`
	DelegationEnabled bool              `json:"delegation_enabled"`
	SubAgentModel     string            `json:"sub_agent_model,omitempty"`
	SynthesisModel    string            `json:"synthesis_model,omitempty"`
}

type ChannelSummary struct {
	AllowAll bool                 `json:"allow_all"`
	Channels []ChannelSummaryItem `json:"channels"`
}

type ChannelSummaryItem struct {
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
	TokenEnv   string `json:"token_env,omitempty"`
	Notes      string `json:"notes,omitempty"`
}

type ToolingSummary struct {
	Packs           []string             `json:"packs,omitempty"`
	AllowedTools    []string             `json:"allowed_tools,omitempty"`
	MCPPresets      []string             `json:"mcp_presets,omitempty"`
	MCPServers      []ToolingMCPServer   `json:"mcp_servers,omitempty"`
	ExecProfile     string               `json:"exec_profile"`
	AllowedCommands []string             `json:"allowed_commands,omitempty"`
	ApprovalOnBlock bool                 `json:"approval_on_block"`
	Browser         ToolingBrowserStatus `json:"browser"`
}

type ToolingMCPServer struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	Command string `json:"command,omitempty"`
}

type ToolingBrowserStatus struct {
	Enabled  bool   `json:"enabled"`
	Headless bool   `json:"headless"`
	Path     string `json:"path,omitempty"`
}

type BrowserSummary struct {
	Enabled      bool   `json:"enabled"`
	Headless     bool   `json:"headless"`
	ConfigPath   string `json:"config_path,omitempty"`
	ResolvedPath string `json:"resolved_path,omitempty"`
	Available    bool   `json:"available"`
}

type SessionSummary struct {
	Total    int                  `json:"total"`
	Sessions []SessionSummaryItem `json:"sessions,omitempty"`
}

type SessionSummaryItem struct {
	ID         string `json:"id"`
	Channel    string `json:"channel"`
	LastActive string `json:"last_active,omitempty"`
}

type CronSummary struct {
	Total          int               `json:"total"`
	Enabled        int               `json:"enabled"`
	RecentFailures int               `json:"recent_failures"`
	Jobs           []CronSummaryItem `json:"jobs,omitempty"`
}

type CronSummaryItem struct {
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
	NextRun string `json:"next_run,omitempty"`
	Source  string `json:"source,omitempty"`
}

type ApprovalsSummary struct {
	AllowAll         bool     `json:"allow_all"`
	ApprovedCount    int      `json:"approved_count"`
	ApprovedSenders  []string `json:"approved_senders,omitempty"`
	ApprovalOnBlock  bool     `json:"approval_on_block"`
	HasExecApprovals bool     `json:"has_exec_approvals"`
}

type LogsSummary struct {
	Output        string `json:"output"`
	FileBacked    bool   `json:"file_backed"`
	FileExists    bool   `json:"file_exists"`
	FileSizeBytes int64  `json:"file_size_bytes,omitempty"`
	ModifiedAt    string `json:"modified_at,omitempty"`
}

type StatusSummary struct {
	Status        string `json:"status"`
	SetupRequired bool   `json:"setup_required"`
	SetupReason   string `json:"setup_reason,omitempty"`
	Provider      string `json:"provider,omitempty"`
	Model         string `json:"model,omitempty"`
	UptimeSeconds int64  `json:"uptime_seconds"`
	SessionCount  int    `json:"session_count"`
	ChannelCount  int    `json:"channel_count"`
	CronJobCount  int    `json:"cron_job_count"`
}

type AuthSummary struct {
	Provider   string `json:"provider"`
	Configured bool   `json:"configured"`
	SignedIn   bool   `json:"signed_in"`
	ExpiresAt  string `json:"expires_at,omitempty"`
	Message    string `json:"message,omitempty"`
}

type PluginSummary struct {
	Registered int                 `json:"registered"`
	Running    int                 `json:"running"`
	Healthy    int                 `json:"healthy"`
	Plugins    []PluginSummaryItem `json:"plugins,omitempty"`
}

type PluginSummaryItem struct {
	Name      string `json:"name"`
	Running   bool   `json:"running"`
	Healthy   bool   `json:"healthy"`
	Restarted int    `json:"restarted"`
	Message   string `json:"message,omitempty"`
}

func BuildModelSummary(cfg *config.Config) ModelSummary {
	if cfg == nil {
		return ModelSummary{}
	}
	fallbackProviders := append([]string(nil), cfg.Model.FallbackProviders...)
	fallbackModels := make(map[string]string, len(cfg.Model.FallbackModels))
	for k, v := range cfg.Model.FallbackModels {
		fallbackModels[k] = v
	}
	return ModelSummary{
		Provider:          cfg.Model.Provider,
		Model:             cfg.Model.Model,
		BaseURL:           cfg.Model.BaseURL,
		APIKeyEnv:         cfg.Model.APIKeyEnv,
		FallbackProviders: fallbackProviders,
		FallbackModels:    fallbackModels,
		DelegationEnabled: cfg.Agent.Delegation.Enabled,
		SubAgentModel:     cfg.Agent.Delegation.SubAgentModel,
		SynthesisModel:    cfg.Agent.Delegation.SynthesisModel,
	}
}

func BuildChannelSummary(cfg *config.Config) ChannelSummary {
	if cfg == nil {
		return ChannelSummary{}
	}
	var rows []ChannelSummaryItem
	rows = append(rows, ChannelSummaryItem{
		Name:       "webchat",
		Configured: true,
		Notes:      "Built-in local web UI channel",
	})
	if cfg.Channels.Telegram != nil {
		rows = append(rows, ChannelSummaryItem{
			Name:       "telegram",
			Configured: true,
			TokenEnv:   cfg.Channels.Telegram.TokenEnv,
		})
	}
	if cfg.Channels.Discord != nil {
		rows = append(rows, ChannelSummaryItem{
			Name:       "discord",
			Configured: true,
			TokenEnv:   cfg.Channels.Discord.TokenEnv,
			Notes:      "App ID optional for slash commands",
		})
	}
	if cfg.Channels.Slack != nil {
		rows = append(rows, ChannelSummaryItem{
			Name:       "slack",
			Configured: true,
			TokenEnv:   cfg.Channels.Slack.TokenEnv,
		})
	}
	if cfg.Channels.WhatsApp != nil {
		note := "QR login flow"
		if cfg.Channels.WhatsApp.Enabled {
			note = "Enabled • QR login flow"
		}
		rows = append(rows, ChannelSummaryItem{
			Name:       "whatsapp",
			Configured: cfg.Channels.WhatsApp.Enabled,
			TokenEnv:   cfg.Channels.WhatsApp.TokenEnv,
			Notes:      note,
		})
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return ChannelSummary{
		AllowAll: cfg.Channels.AllowAll,
		Channels: rows,
	}
}

func BuildToolingSummary(cfg *config.Config) ToolingSummary {
	if cfg == nil {
		return ToolingSummary{}
	}
	var servers []ToolingMCPServer
	for _, server := range cfg.MCPServers {
		enabled := true
		if server.Enabled != nil {
			enabled = *server.Enabled
		}
		servers = append(servers, ToolingMCPServer{
			Name:    server.Name,
			Enabled: enabled,
			Command: server.Command,
		})
	}
	sort.Slice(servers, func(i, j int) bool { return servers[i].Name < servers[j].Name })
	browserPath := cfg.Tools.Browser.ChromePath
	if browserPath == "" {
		browserPath = cfg.Tools.Browser.ChromiumPath
	}
	return ToolingSummary{
		Packs:           append([]string(nil), cfg.Tools.Packs...),
		AllowedTools:    append([]string(nil), cfg.Tools.AllowedTools...),
		MCPPresets:      append([]string(nil), cfg.Tools.MCPPresets...),
		MCPServers:      servers,
		ExecProfile:     cfg.Tools.Exec.Profile,
		AllowedCommands: append([]string(nil), cfg.Tools.Exec.AllowedCommands...),
		ApprovalOnBlock: cfg.Tools.Exec.ApprovalOnBlock,
		Browser: ToolingBrowserStatus{
			Enabled:  cfg.Tools.Browser.Enabled,
			Headless: cfg.Tools.Browser.Headless,
			Path:     browserPath,
		},
	}
}

func BuildApprovalsSummary(cfg *config.Config, allowAll bool, approved []string) ApprovalsSummary {
	senders := append([]string(nil), approved...)
	sort.Strings(senders)
	summary := ApprovalsSummary{
		AllowAll:        allowAll,
		ApprovedCount:   len(senders),
		ApprovedSenders: senders,
	}
	if cfg != nil {
		summary.ApprovalOnBlock = cfg.Tools.Exec.ApprovalOnBlock
		summary.HasExecApprovals = cfg.Tools.Exec.ApprovalOnBlock
	}
	return summary
}

func BuildLogsSummary(cfg *config.Config) LogsSummary {
	if cfg == nil {
		return LogsSummary{}
	}
	output := expandHome(strings.TrimSpace(cfg.Logging.Output))
	summary := LogsSummary{
		Output:     output,
		FileBacked: output != "" && output != "stderr" && output != "stdout",
	}
	if !summary.FileBacked {
		return summary
	}
	info, err := os.Stat(output)
	if err != nil {
		return summary
	}
	summary.FileExists = true
	summary.FileSizeBytes = info.Size()
	summary.ModifiedAt = info.ModTime().UTC().Format(time.RFC3339)
	return summary
}

func BuildBrowserSummary(cfg *config.Config) BrowserSummary {
	if cfg == nil {
		return BrowserSummary{}
	}
	configPath := strings.TrimSpace(cfg.Tools.Browser.ChromePath)
	if configPath == "" {
		configPath = strings.TrimSpace(cfg.Tools.Browser.ChromiumPath)
	}
	resolved, ok := firstExecutable(
		cfg.Tools.Browser.ChromePath,
		cfg.Tools.Browser.ChromiumPath,
		"google-chrome",
		"chromium",
		"chromium-browser",
	)
	return BrowserSummary{
		Enabled:      cfg.Tools.Browser.Enabled,
		Headless:     cfg.Tools.Browser.Headless,
		ConfigPath:   configPath,
		ResolvedPath: resolved,
		Available:    ok,
	}
}

func BuildSessionSummary(total int, sessions []SessionSummaryItem) SessionSummary {
	rows := append([]SessionSummaryItem(nil), sessions...)
	return SessionSummary{
		Total:    total,
		Sessions: rows,
	}
}

func BuildCronSummary(total, enabled, recentFailures int, jobs []CronSummaryItem) CronSummary {
	rows := append([]CronSummaryItem(nil), jobs...)
	return CronSummary{
		Total:          total,
		Enabled:        enabled,
		RecentFailures: recentFailures,
		Jobs:           rows,
	}
}

func BuildStatusSummary(status string, cfg *config.Config, setupRequired bool, setupReason string, uptimeSeconds int64, sessionCount, channelCount, cronJobCount int) StatusSummary {
	provider := ""
	model := ""
	if cfg != nil {
		provider = cfg.Model.Provider
		model = cfg.Model.Model
	}
	return StatusSummary{
		Status:        nonEmpty(strings.TrimSpace(status), "ok"),
		SetupRequired: setupRequired,
		SetupReason:   strings.TrimSpace(setupReason),
		Provider:      provider,
		Model:         model,
		UptimeSeconds: uptimeSeconds,
		SessionCount:  sessionCount,
		ChannelCount:  channelCount,
		CronJobCount:  cronJobCount,
	}
}

func BuildAuthSummary(configured, signedIn bool, expiresAt time.Time, message string) AuthSummary {
	summary := AuthSummary{
		Provider:   "openai",
		Configured: configured,
		SignedIn:   signedIn,
		Message:    strings.TrimSpace(message),
	}
	if !expiresAt.IsZero() {
		summary.ExpiresAt = expiresAt.UTC().Format(time.RFC3339)
	}
	return summary
}

func BuildPluginSummary(plugins []PluginSummaryItem) PluginSummary {
	rows := append([]PluginSummaryItem(nil), plugins...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	summary := PluginSummary{
		Registered: len(rows),
		Plugins:    rows,
	}
	for _, item := range rows {
		if item.Running {
			summary.Running++
		}
		if item.Healthy {
			summary.Healthy++
		}
	}
	return summary
}

func FormatModelSummaryText(summary ModelSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "MODELS [%s]\n", nonEmpty(summary.Provider, "unknown"))
	fmt.Fprintf(&b, "- Provider: %s\n", nonEmpty(summary.Provider, "not configured"))
	fmt.Fprintf(&b, "- Model: %s\n", nonEmpty(summary.Model, "not configured"))
	if summary.BaseURL != "" {
		fmt.Fprintf(&b, "- Base URL: %s\n", summary.BaseURL)
	}
	if summary.APIKeyEnv != "" {
		state := "missing"
		if os.Getenv(summary.APIKeyEnv) != "" {
			state = "set"
		}
		fmt.Fprintf(&b, "- API key env: %s (%s)\n", summary.APIKeyEnv, state)
	}
	if len(summary.FallbackProviders) == 0 {
		b.WriteString("- Fallback providers: none\n")
	} else {
		fmt.Fprintf(&b, "- Fallback providers: %s\n", strings.Join(summary.FallbackProviders, ", "))
		if len(summary.FallbackModels) > 0 {
			keys := make([]string, 0, len(summary.FallbackModels))
			for key := range summary.FallbackModels {
				keys = append(keys, key)
			}
			sort.Strings(keys)
			for _, key := range keys {
				fmt.Fprintf(&b, "  · %s -> %s\n", key, summary.FallbackModels[key])
			}
		}
	}
	if summary.DelegationEnabled {
		fmt.Fprintf(&b, "- Delegation: enabled\n")
		fmt.Fprintf(&b, "  · Sub-agent model: %s\n", nonEmpty(summary.SubAgentModel, "inherits primary"))
		fmt.Fprintf(&b, "  · Synthesis model: %s\n", nonEmpty(summary.SynthesisModel, "inherits primary"))
	} else {
		b.WriteString("- Delegation: disabled\n")
	}
	return b.String()
}

func FormatChannelSummaryText(summary ChannelSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CHANNELS [%d configured]\n", len(summary.Channels))
	if summary.AllowAll {
		b.WriteString("- Allowlist mode: allow_all=true\n")
	} else {
		b.WriteString("- Allowlist mode: strict\n")
	}
	if len(summary.Channels) == 0 {
		b.WriteString("- No channels configured\n")
		return b.String()
	}
	for _, ch := range summary.Channels {
		state := "configured"
		if !ch.Configured {
			state = "disabled"
		}
		fmt.Fprintf(&b, "- %s: %s\n", ch.Name, state)
		if ch.TokenEnv != "" {
			tokenState := "missing"
			if os.Getenv(ch.TokenEnv) != "" {
				tokenState = "set"
			}
			fmt.Fprintf(&b, "  · token env: %s (%s)\n", ch.TokenEnv, tokenState)
		}
		if ch.Notes != "" {
			fmt.Fprintf(&b, "  · notes: %s\n", ch.Notes)
		}
	}
	return b.String()
}

func FormatToolingSummaryText(summary ToolingSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "TOOLS [%s]\n", nonEmpty(summary.ExecProfile, "safe"))
	if len(summary.Packs) == 0 {
		b.WriteString("- Packs: none\n")
	} else {
		fmt.Fprintf(&b, "- Packs: %s\n", strings.Join(summary.Packs, ", "))
	}
	if len(summary.AllowedTools) == 0 {
		b.WriteString("- Built-in tools: all/default\n")
	} else {
		fmt.Fprintf(&b, "- Built-in tools: %s\n", strings.Join(summary.AllowedTools, ", "))
	}
	fmt.Fprintf(&b, "- Exec profile: %s\n", nonEmpty(summary.ExecProfile, "safe"))
	if len(summary.AllowedCommands) > 0 {
		fmt.Fprintf(&b, "- Allowed commands: %s\n", strings.Join(summary.AllowedCommands, ", "))
	}
	if summary.ApprovalOnBlock {
		b.WriteString("- Blocked command approvals: enabled\n")
	} else {
		b.WriteString("- Blocked command approvals: disabled\n")
	}
	if summary.Browser.Enabled {
		fmt.Fprintf(&b, "- Browser: enabled (%s)\n", ternary(summary.Browser.Headless, "headless", "headed"))
		if summary.Browser.Path != "" {
			fmt.Fprintf(&b, "  · binary: %s\n", summary.Browser.Path)
		}
	} else {
		b.WriteString("- Browser: disabled\n")
	}
	if len(summary.MCPPresets) == 0 {
		b.WriteString("- MCP presets: none\n")
	} else {
		fmt.Fprintf(&b, "- MCP presets: %s\n", strings.Join(summary.MCPPresets, ", "))
	}
	if len(summary.MCPServers) == 0 {
		b.WriteString("- MCP servers: none\n")
	} else {
		for _, server := range summary.MCPServers {
			state := "disabled"
			if server.Enabled {
				state = "enabled"
			}
			fmt.Fprintf(&b, "  · %s: %s", server.Name, state)
			if server.Command != "" {
				fmt.Fprintf(&b, " (%s)", server.Command)
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}

func FormatApprovalsSummaryText(summary ApprovalsSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "APPROVALS [%d]\n", summary.ApprovedCount)
	if summary.AllowAll {
		b.WriteString("- Channel allowlist mode: allow_all=true\n")
	} else {
		b.WriteString("- Channel allowlist mode: strict\n")
	}
	if summary.ApprovedCount == 0 {
		b.WriteString("- Approved senders: none\n")
	} else {
		fmt.Fprintf(&b, "- Approved senders: %d\n", summary.ApprovedCount)
		limit := len(summary.ApprovedSenders)
		if limit > 5 {
			limit = 5
		}
		for _, sender := range summary.ApprovedSenders[:limit] {
			fmt.Fprintf(&b, "  · %s\n", sender)
		}
		if len(summary.ApprovedSenders) > limit {
			fmt.Fprintf(&b, "  · ... and %d more\n", len(summary.ApprovedSenders)-limit)
		}
	}
	if summary.ApprovalOnBlock {
		b.WriteString("- Exec approval-on-block: enabled\n")
	} else {
		b.WriteString("- Exec approval-on-block: disabled\n")
	}
	return b.String()
}

func FormatLogsSummaryText(summary LogsSummary) string {
	var b strings.Builder
	b.WriteString("LOGS\n")
	if strings.TrimSpace(summary.Output) == "" {
		b.WriteString("- Output: not configured\n")
		return b.String()
	}
	fmt.Fprintf(&b, "- Output: %s\n", summary.Output)
	if !summary.FileBacked {
		b.WriteString("- Mode: stream output (not file-backed)\n")
		return b.String()
	}
	if !summary.FileExists {
		b.WriteString("- File: missing\n")
		return b.String()
	}
	b.WriteString("- File: present\n")
	fmt.Fprintf(&b, "- Size: %d bytes\n", summary.FileSizeBytes)
	if summary.ModifiedAt != "" {
		fmt.Fprintf(&b, "- Modified: %s\n", summary.ModifiedAt)
	}
	b.WriteString("- Inspect recent lines via /api/v1/logs or the Logs panel\n")
	return b.String()
}

func FormatBrowserSummaryText(summary BrowserSummary) string {
	var b strings.Builder
	b.WriteString("BROWSER\n")
	if summary.Enabled {
		fmt.Fprintf(&b, "- Status: enabled (%s)\n", ternary(summary.Headless, "headless", "headed"))
	} else {
		b.WriteString("- Status: disabled\n")
	}
	if summary.ConfigPath != "" {
		fmt.Fprintf(&b, "- Configured path: %s\n", summary.ConfigPath)
	}
	if summary.Available {
		fmt.Fprintf(&b, "- Resolved binary: %s\n", summary.ResolvedPath)
	} else {
		b.WriteString("- Resolved binary: unavailable\n")
	}
	return b.String()
}

func FormatSessionSummaryText(summary SessionSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "SESSIONS [%d]\n", summary.Total)
	if len(summary.Sessions) == 0 {
		b.WriteString("- No recent sessions\n")
		return b.String()
	}
	limit := len(summary.Sessions)
	if limit > 5 {
		limit = 5
	}
	for _, s := range summary.Sessions[:limit] {
		label := nonEmpty(s.Channel, "chat")
		fmt.Fprintf(&b, "- %s / %s\n", label, trimSessionID(s.ID))
		if s.LastActive != "" {
			fmt.Fprintf(&b, "  · last active: %s\n", s.LastActive)
		}
	}
	if len(summary.Sessions) > limit {
		fmt.Fprintf(&b, "- ... and %d more\n", len(summary.Sessions)-limit)
	}
	return b.String()
}

func FormatCronSummaryText(summary CronSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "CRON [%d jobs]\n", summary.Total)
	fmt.Fprintf(&b, "- Enabled: %d\n", summary.Enabled)
	fmt.Fprintf(&b, "- Recent failures: %d\n", summary.RecentFailures)
	if len(summary.Jobs) == 0 {
		b.WriteString("- No cron jobs configured\n")
		return b.String()
	}
	limit := len(summary.Jobs)
	if limit > 5 {
		limit = 5
	}
	for _, j := range summary.Jobs[:limit] {
		state := "disabled"
		if j.Enabled {
			state = "enabled"
		}
		fmt.Fprintf(&b, "- %s: %s\n", nonEmpty(j.Name, "job"), state)
		if j.NextRun != "" {
			fmt.Fprintf(&b, "  · next run: %s\n", j.NextRun)
		}
		if j.Source != "" {
			fmt.Fprintf(&b, "  · source: %s\n", j.Source)
		}
	}
	if len(summary.Jobs) > limit {
		fmt.Fprintf(&b, "- ... and %d more\n", len(summary.Jobs)-limit)
	}
	return b.String()
}

func FormatStatusSummaryText(summary StatusSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "STATUS [%s]\n", strings.ToUpper(nonEmpty(summary.Status, "ok")))
	if summary.SetupRequired {
		b.WriteString("- Setup required: yes\n")
		if summary.SetupReason != "" {
			fmt.Fprintf(&b, "- Setup reason: %s\n", summary.SetupReason)
		}
	} else {
		b.WriteString("- Setup required: no\n")
	}
	fmt.Fprintf(&b, "- Provider: %s\n", nonEmpty(summary.Provider, "not configured"))
	fmt.Fprintf(&b, "- Model: %s\n", nonEmpty(summary.Model, "not configured"))
	fmt.Fprintf(&b, "- Uptime: %ds\n", summary.UptimeSeconds)
	fmt.Fprintf(&b, "- Sessions: %d\n", summary.SessionCount)
	fmt.Fprintf(&b, "- Channels: %d\n", summary.ChannelCount)
	fmt.Fprintf(&b, "- Cron jobs: %d\n", summary.CronJobCount)
	return b.String()
}

func FormatAuthSummaryText(summary AuthSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "AUTH [%s]\n", strings.ToUpper(nonEmpty(summary.Provider, "openai")))
	if summary.Configured {
		b.WriteString("- Configured: yes\n")
	} else {
		b.WriteString("- Configured: no\n")
	}
	if summary.SignedIn {
		b.WriteString("- Signed in: yes\n")
		if summary.ExpiresAt != "" {
			fmt.Fprintf(&b, "- Expires at: %s\n", summary.ExpiresAt)
		}
	} else {
		b.WriteString("- Signed in: no\n")
	}
	if summary.Message != "" {
		fmt.Fprintf(&b, "- Message: %s\n", summary.Message)
	}
	return b.String()
}

func FormatPluginSummaryText(summary PluginSummary) string {
	var b strings.Builder
	fmt.Fprintf(&b, "PLUGINS [%d]\n", summary.Registered)
	fmt.Fprintf(&b, "- Running: %d\n", summary.Running)
	fmt.Fprintf(&b, "- Healthy: %d\n", summary.Healthy)
	if len(summary.Plugins) == 0 {
		b.WriteString("- No runtime plugins registered\n")
		return b.String()
	}
	limit := len(summary.Plugins)
	if limit > 5 {
		limit = 5
	}
	for _, item := range summary.Plugins[:limit] {
		state := "stopped"
		if item.Running {
			state = "running"
		}
		health := "degraded"
		if item.Healthy {
			health = "healthy"
		}
		fmt.Fprintf(&b, "- %s: %s / %s\n", item.Name, state, health)
		if item.Restarted > 0 {
			fmt.Fprintf(&b, "  · restarts: %d\n", item.Restarted)
		}
		if item.Message != "" {
			fmt.Fprintf(&b, "  · message: %s\n", item.Message)
		}
	}
	if len(summary.Plugins) > limit {
		fmt.Fprintf(&b, "- ... and %d more\n", len(summary.Plugins)-limit)
	}
	return b.String()
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func ternary[T any](cond bool, a, b T) T {
	if cond {
		return a
	}
	return b
}

func expandHome(path string) string {
	if path == "" || !strings.HasPrefix(path, "~") {
		return path
	}
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, strings.TrimPrefix(path, "~/"))
	}
	return path
}

func trimSessionID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		return id
	}
	return id[:8]
}
