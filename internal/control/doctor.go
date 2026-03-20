package control

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/openclio/openclio/internal/config"
)

type DoctorStatus string

const (
	DoctorOK   DoctorStatus = "ok"
	DoctorWarn DoctorStatus = "warn"
	DoctorFail DoctorStatus = "fail"
)

type DoctorCheck struct {
	ID          string       `json:"id"`
	Category    string       `json:"category"`
	Label       string       `json:"label"`
	Status      DoctorStatus `json:"status"`
	Message     string       `json:"message"`
	Details     []string     `json:"details,omitempty"`
	ActionHint  string       `json:"action_hint,omitempty"`
}

type DoctorReport struct {
	Status    DoctorStatus  `json:"status"`
	Time      string        `json:"time"`
	Summary   string        `json:"summary"`
	Checks    []DoctorCheck `json:"checks"`
	Healthy   int           `json:"healthy"`
	Warnings  int           `json:"warnings"`
	Failures  int           `json:"failures"`
}

type MCPRuntimeStatus struct {
	Name    string `json:"name"`
	Healthy bool   `json:"healthy"`
	Status  string `json:"status,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ChannelDoctorStatus struct {
	Name       string `json:"name"`
	Configured bool   `json:"configured"`
	Running    bool   `json:"running"`
	Healthy    bool   `json:"healthy"`
	Message    string `json:"message,omitempty"`
}

type DoctorInput struct {
	Config       *config.Config
	DataDir      string
	ConfigPath   string
	MCPStatuses  []MCPRuntimeStatus
	ChannelStats []ChannelDoctorStatus
}

func BuildDoctorReport(in DoctorInput) DoctorReport {
	now := time.Now().UTC().Format(time.RFC3339)
	report := DoctorReport{
		Status:  DoctorOK,
		Time:    now,
		Checks:  make([]DoctorCheck, 0, 8),
		Summary: "All core checks passed.",
	}

	add := func(ch DoctorCheck) {
		report.Checks = append(report.Checks, ch)
		switch ch.Status {
		case DoctorOK:
			report.Healthy++
		case DoctorWarn:
			report.Warnings++
			if report.Status != DoctorFail {
				report.Status = DoctorWarn
			}
		case DoctorFail:
			report.Failures++
			report.Status = DoctorFail
		}
	}

	cfg := in.Config
	if cfg == nil {
		add(DoctorCheck{
			ID:         "config.missing",
			Category:   "config",
			Label:      "Configuration",
			Status:     DoctorFail,
			Message:    "Config is not loaded.",
			ActionHint: "Run openclio init or provide a valid config.yaml.",
		})
		report.Summary = "Configuration is unavailable."
		return report
	}

	configPath := in.ConfigPath
	if strings.TrimSpace(configPath) == "" && strings.TrimSpace(in.DataDir) != "" {
		configPath = filepath.Join(in.DataDir, "config.yaml")
	}
	if strings.TrimSpace(configPath) == "" {
		add(DoctorCheck{
			ID:       "config.path",
			Category: "config",
			Label:    "Config file path",
			Status:   DoctorWarn,
			Message:  "Config path is unknown in this runtime.",
		})
	} else if _, err := os.Stat(configPath); err != nil {
		add(DoctorCheck{
			ID:         "config.file",
			Category:   "config",
			Label:      "Config file",
			Status:     DoctorFail,
			Message:    fmt.Sprintf("Config file is missing at %s.", configPath),
			ActionHint: "Run openclio init to generate config.yaml.",
		})
	} else {
		add(DoctorCheck{
			ID:       "config.file",
			Category: "config",
			Label:    "Config file",
			Status:   DoctorOK,
			Message:  fmt.Sprintf("Config file found at %s.", configPath),
		})
	}

	if strings.TrimSpace(in.DataDir) == "" {
		add(DoctorCheck{
			ID:       "data.dir",
			Category: "storage",
			Label:    "Data directory",
			Status:   DoctorWarn,
			Message:  "Data directory is not set.",
		})
	} else {
		if err := os.MkdirAll(in.DataDir, 0o755); err != nil {
			add(DoctorCheck{
				ID:         "data.dir",
				Category:   "storage",
				Label:      "Data directory",
				Status:     DoctorFail,
				Message:    fmt.Sprintf("Failed to create or access %s: %v", in.DataDir, err),
				ActionHint: "Ensure the data directory is writable.",
			})
		} else {
			testFile := filepath.Join(in.DataDir, ".doctor-write-test")
			if err := os.WriteFile(testFile, []byte("ok"), 0o600); err != nil {
				add(DoctorCheck{
					ID:         "data.writable",
					Category:   "storage",
					Label:      "Data directory writable",
					Status:     DoctorFail,
					Message:    fmt.Sprintf("Cannot write to %s: %v", in.DataDir, err),
					ActionHint: "Fix filesystem permissions for the OpenClio data directory.",
				})
			} else {
				_ = os.Remove(testFile)
				add(DoctorCheck{
					ID:       "data.writable",
					Category: "storage",
					Label:    "Data directory writable",
					Status:   DoctorOK,
					Message:  fmt.Sprintf("Data directory %s is writable.", in.DataDir),
				})
			}
		}
	}

	provider := strings.TrimSpace(cfg.Model.Provider)
	model := strings.TrimSpace(cfg.Model.Model)
	switch {
	case provider == "":
		add(DoctorCheck{
			ID:         "model.provider",
			Category:   "provider",
			Label:      "Provider",
			Status:     DoctorFail,
			Message:    "No model provider configured.",
			ActionHint: "Choose a provider in init or the Control UI.",
		})
	default:
		add(DoctorCheck{
			ID:       "model.provider",
			Category: "provider",
			Label:    "Provider",
			Status:   DoctorOK,
			Message:  fmt.Sprintf("Provider %s is configured.", provider),
			Details:  []string{fmt.Sprintf("model=%s", model)},
		})
	}

	if providerNeedsAPIKey(provider) {
		envName := strings.TrimSpace(cfg.Model.APIKeyEnv)
		if envName == "" {
			add(DoctorCheck{
				ID:         "provider.apikey.env",
				Category:   "provider",
				Label:      "Provider credentials",
				Status:     DoctorFail,
				Message:    fmt.Sprintf("%s requires an API key env var, but none is configured.", provider),
				ActionHint: "Set the provider API key env in config or rerun init.",
			})
		} else if strings.TrimSpace(os.Getenv(envName)) == "" {
			add(DoctorCheck{
				ID:         "provider.apikey.value",
				Category:   "provider",
				Label:      "Provider credentials",
				Status:     DoctorWarn,
				Message:    fmt.Sprintf("%s is configured but %s is not set in the environment.", provider, envName),
				ActionHint: "Export the API key or add it to ~/.openclio/.env.",
			})
		} else {
			add(DoctorCheck{
				ID:       "provider.apikey.value",
				Category: "provider",
				Label:    "Provider credentials",
				Status:   DoctorOK,
				Message:  fmt.Sprintf("%s credentials are available via %s.", provider, envName),
			})
		}
	}

	if cfg.Tools.Browser.Enabled {
		bin, ok := firstExecutable(cfg.Tools.Browser.ChromePath, cfg.Tools.Browser.ChromiumPath, "google-chrome", "chromium", "chromium-browser")
		if !ok {
			add(DoctorCheck{
				ID:         "browser.binary",
				Category:   "tools",
				Label:      "Browser tool",
				Status:     DoctorWarn,
				Message:    "Browser automation is enabled but no Chrome/Chromium binary was found.",
				ActionHint: "Install Chrome/Chromium or configure chrome_path/chromium_path.",
			})
		} else {
			add(DoctorCheck{
				ID:       "browser.binary",
				Category: "tools",
				Label:    "Browser tool",
				Status:   DoctorOK,
				Message:  fmt.Sprintf("Browser automation is enabled and using %s.", bin),
			})
		}
	}

	add(execProfileCheck(cfg))
	add(basicBinaryCheck("tool.git", "tools", "Git", "git", "--version"))

	mcpCheck := buildMCPCheck(cfg, in.MCPStatuses)
	add(mcpCheck)

	channelChecks := buildChannelChecks(cfg, in.ChannelStats)
	for _, ch := range channelChecks {
		add(ch)
	}

	sort.Slice(report.Checks, func(i, j int) bool {
		if report.Checks[i].Category == report.Checks[j].Category {
			return report.Checks[i].Label < report.Checks[j].Label
		}
		return report.Checks[i].Category < report.Checks[j].Category
	})

	switch report.Status {
	case DoctorOK:
		report.Summary = "All core checks passed."
	case DoctorWarn:
		report.Summary = fmt.Sprintf("%d warning(s) detected. OpenClio is usable but not fully healthy.", report.Warnings)
	case DoctorFail:
		report.Summary = fmt.Sprintf("%d failure(s) detected. Fix the failing checks before broad beta use.", report.Failures)
	}

	return report
}

func providerNeedsAPIKey(provider string) bool {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "", "ollama", "lmstudio":
		return false
	default:
		return true
	}
}

func execProfileCheck(cfg *config.Config) DoctorCheck {
	profile := strings.TrimSpace(cfg.Tools.Exec.Profile)
	if profile == "" {
		profile = "safe"
	}
	msg := fmt.Sprintf("Exec profile is %s.", profile)
	status := DoctorOK
	if profile == "power-user" {
		status = DoctorWarn
		msg = "Exec profile is power-user."
	}
	return DoctorCheck{
		ID:       "tools.exec_profile",
		Category: "tools",
		Label:    "Exec profile",
		Status:   status,
		Message:  msg,
		Details: []string{
			fmt.Sprintf("approval_on_block=%t", cfg.Tools.Exec.ApprovalOnBlock),
			fmt.Sprintf("allowed_commands=%d", len(cfg.Tools.Exec.AllowedCommands)),
		},
	}
}

func basicBinaryCheck(id, category, label, bin string, args ...string) DoctorCheck {
	path, err := exec.LookPath(bin)
	if err != nil {
		return DoctorCheck{
			ID:         id,
			Category:   category,
			Label:      label,
			Status:     DoctorWarn,
			Message:    fmt.Sprintf("%s is not available on PATH.", bin),
			ActionHint: fmt.Sprintf("Install %s if you want to use related local tooling.", label),
		}
	}
	return DoctorCheck{
		ID:       id,
		Category: category,
		Label:    label,
		Status:   DoctorOK,
		Message:  fmt.Sprintf("%s is available at %s.", label, path),
	}
}

func firstExecutable(candidates ...string) (string, bool) {
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		if filepath.IsAbs(candidate) {
			if _, err := os.Stat(candidate); err == nil {
				return candidate, true
			}
			continue
		}
		if p, err := exec.LookPath(candidate); err == nil {
			return p, true
		}
	}
	return "", false
}

func buildMCPCheck(cfg *config.Config, statuses []MCPRuntimeStatus) DoctorCheck {
	if len(cfg.MCPServers) == 0 {
		return DoctorCheck{
			ID:       "mcp.none",
			Category: "mcp",
			Label:    "MCP servers",
			Status:   DoctorOK,
			Message:  "No MCP servers configured.",
		}
	}

	statusByName := make(map[string]MCPRuntimeStatus, len(statuses))
	for _, st := range statuses {
		statusByName[st.Name] = st
	}

	details := make([]string, 0, len(cfg.MCPServers))
	failures := 0
	warnings := 0
	for _, srv := range cfg.MCPServers {
		enabled := srv.Enabled == nil || *srv.Enabled
		if !enabled {
			details = append(details, fmt.Sprintf("%s: disabled", srv.Name))
			continue
		}
		if p, err := exec.LookPath(strings.TrimSpace(srv.Command)); err != nil {
			failures++
			details = append(details, fmt.Sprintf("%s: missing executable %s", srv.Name, srv.Command))
			continue
		} else {
			details = append(details, fmt.Sprintf("%s: command=%s", srv.Name, p))
		}
		if st, ok := statusByName[srv.Name]; ok {
			if !st.Healthy {
				warnings++
				msg := st.Status
				if strings.TrimSpace(msg) == "" {
					msg = "unhealthy"
				}
				if strings.TrimSpace(st.Error) != "" {
					msg += " (" + st.Error + ")"
				}
				details = append(details, fmt.Sprintf("%s runtime: %s", srv.Name, msg))
			}
		}
	}

	status := DoctorOK
	message := fmt.Sprintf("%d MCP server(s) configured.", len(cfg.MCPServers))
	if failures > 0 {
		status = DoctorFail
		message = fmt.Sprintf("%d MCP server(s) have missing executables.", failures)
	} else if warnings > 0 {
		status = DoctorWarn
		message = fmt.Sprintf("%d MCP server(s) are configured but unhealthy.", warnings)
	}

	return DoctorCheck{
		ID:        "mcp.servers",
		Category:  "mcp",
		Label:     "MCP servers",
		Status:    status,
		Message:   message,
		Details:   details,
		ActionHint: "Install missing MCP executables or disable unused servers in config.",
	}
}

func buildChannelChecks(cfg *config.Config, runtime []ChannelDoctorStatus) []DoctorCheck {
	checks := make([]DoctorCheck, 0, 4)
	runtimeByName := make(map[string]ChannelDoctorStatus, len(runtime))
	for _, st := range runtime {
		runtimeByName[st.Name] = st
	}
	maybeAdd := func(name, envName string, configured bool) {
		if !configured {
			return
		}
		label := titleName(name)
		ch := DoctorCheck{
			ID:       "channel." + name,
			Category: "channels",
			Label:    label,
			Status:   DoctorOK,
			Message:  fmt.Sprintf("%s is configured.", label),
		}
		if envName != "" && strings.TrimSpace(os.Getenv(envName)) == "" {
			ch.Status = DoctorWarn
			ch.Message = fmt.Sprintf("%s is configured but %s is not set.", label, envName)
			ch.ActionHint = "Set the channel credential env var or update the channel configuration."
		}
		if st, ok := runtimeByName[name]; ok {
			ch.Details = append(ch.Details,
				fmt.Sprintf("running=%t", st.Running),
				fmt.Sprintf("healthy=%t", st.Healthy),
			)
			if st.Configured && !st.Healthy {
				if ch.Status == DoctorOK {
					ch.Status = DoctorWarn
				}
				if strings.TrimSpace(st.Message) != "" {
					ch.Details = append(ch.Details, st.Message)
				}
			}
		}
		checks = append(checks, ch)
	}
	telegramEnv := ""
	if cfg.Channels.Telegram != nil {
		telegramEnv = cfg.Channels.Telegram.TokenEnv
	}
	discordEnv := ""
	if cfg.Channels.Discord != nil {
		discordEnv = cfg.Channels.Discord.TokenEnv
	}
	slackEnv := ""
	if cfg.Channels.Slack != nil {
		slackEnv = cfg.Channels.Slack.TokenEnv
	}
	maybeAdd("telegram", telegramEnv, cfg.Channels.Telegram != nil)
	maybeAdd("discord", discordEnv, cfg.Channels.Discord != nil)
	maybeAdd("slack", slackEnv, cfg.Channels.Slack != nil)
	if cfg.Channels.WhatsApp != nil && cfg.Channels.WhatsApp.Enabled {
		ch := DoctorCheck{
			ID:       "channel.whatsapp",
			Category: "channels",
			Label:    "WhatsApp",
			Status:   DoctorOK,
			Message:  "WhatsApp is enabled.",
		}
		if st, ok := runtimeByName["whatsapp"]; ok {
			ch.Details = append(ch.Details,
				fmt.Sprintf("running=%t", st.Running),
				fmt.Sprintf("healthy=%t", st.Healthy),
			)
			if !st.Healthy {
				ch.Status = DoctorWarn
			}
			if strings.TrimSpace(st.Message) != "" {
				ch.Details = append(ch.Details, st.Message)
			}
		}
		checks = append(checks, ch)
	}
	return checks
}

func titleName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_' || r == ' '
	})
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + strings.ToLower(parts[i][1:])
	}
	return strings.Join(parts, " ")
}

func FormatDoctorReportText(report DoctorReport) string {
	var b strings.Builder
	b.WriteString("\nDoctor Summary\n")
	b.WriteString("--------------\n")
	b.WriteString(fmt.Sprintf("Status: %s\n", strings.ToUpper(string(report.Status))))
	b.WriteString(fmt.Sprintf("Summary: %s\n", report.Summary))
	b.WriteString(fmt.Sprintf("Healthy: %d  Warnings: %d  Failures: %d\n\n", report.Healthy, report.Warnings, report.Failures))
	for _, ch := range report.Checks {
		icon := "✓"
		switch ch.Status {
		case DoctorWarn:
			icon = "!"
		case DoctorFail:
			icon = "✗"
		}
		b.WriteString(fmt.Sprintf("%s [%s] %s — %s\n", icon, ch.Category, ch.Label, ch.Message))
		for _, detail := range ch.Details {
			b.WriteString(fmt.Sprintf("    - %s\n", detail))
		}
		if strings.TrimSpace(ch.ActionHint) != "" {
			b.WriteString(fmt.Sprintf("    -> %s\n", ch.ActionHint))
		}
	}
	b.WriteString("\n")
	return b.String()
}
