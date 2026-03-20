package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	cfgpkg "github.com/openclio/openclio/internal/config"
	"github.com/openclio/openclio/internal/control"
	"github.com/openclio/openclio/internal/cost"
)

// HandleCommand processes a slash command. Returns true if it was handled.
func (c *CLI) HandleCommand(input string) bool {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return false
	}

	cmd := strings.ToLower(parts[0])
	switch cmd {
	case "/":
		c.cmdSlashPalette()
		return true
	case "/help":
		c.cmdHelp()
		return true
	case "/status":
		c.cmdStatus()
		return true
	case "/auth":
		c.cmdAuth()
		return true
	case "/plugins":
		c.cmdPlugins()
		return true
	case "/sessions":
		c.cmdSessions(parts)
		return true
	case "/new":
		c.cmdNew()
		return true
	case "/clear":
		c.cmdReset()
		return true
	case "/reset":
		c.cmdReset()
		return true
	case "/compact":
		c.cmdCompact()
		return true
	case "/history":
		c.cmdHistory()
		return true
	case "/usage":
		c.cmdUsage()
		return true
	case "/cost":
		c.cmdCost()
		return true
	case "/cron":
		c.cmdCron()
		return true
	case "/model":
		c.cmdModel()
		return true
	case "/models":
		c.cmdModels(parts)
		return true
	case "/channels":
		c.cmdChannels(parts)
		return true
	case "/tools":
		c.cmdTools(parts)
		return true
	case "/browser":
		c.cmdBrowser(parts)
		return true
	case "/approvals":
		c.cmdApprovals(parts)
		return true
	case "/logs":
		c.cmdLogs(parts)
		return true
	case "/memory":
		c.cmdMemory()
		return true
	case "/doctor":
		c.cmdDoctor(parts)
		return true
	case "/beliefs", "/belief":
		c.cmdBeliefs()
		return true
	case "/switch":
		c.cmdSwitch(parts)
		return true
	case "/skill", "/skills":
		c.cmdSkill(parts)
		return true
	case "/debug":
		if len(parts) > 1 {
			if parts[1] == "context" {
				c.cmdDebugContext()
				return true
			}
			if parts[1] == "tokens" {
				c.cmdDebugTokens()
				return true
			}
		}
		PrintError("Usage: /debug context | /debug tokens")
		return true
	default:
		if strings.HasPrefix(cmd, "/") {
			PrintError(fmt.Sprintf("Unknown command: %s (type /help)", cmd))
			return true
		}
		return false
	}
}

func (c *CLI) cmdHelp() {
	fmt.Println()
	fmt.Printf("%sAvailable commands:%s\n\n", colorBold(), colorReset())
	fmt.Printf("  %s/%s          Open the grouped slash-command palette\n", colorCyan(), colorReset())
	fmt.Printf("  %s/help%s       Show this help\n", colorCyan(), colorReset())
	fmt.Printf("  %s/status%s     Show overall runtime/model/session summary\n", colorCyan(), colorReset())
	fmt.Printf("  %s/auth%s       Show OpenAI OAuth configuration and sign-in state\n", colorCyan(), colorReset())
	fmt.Printf("  %s/plugins%s    Show runtime plugin/adapter summary\n", colorCyan(), colorReset())
	fmt.Printf("  %s/sessions%s   List recent sessions\n", colorCyan(), colorReset())
	fmt.Printf("  %s/sessions delete <id|current>%s  Delete a session\n", colorCyan(), colorReset())
	fmt.Printf("  %s/new%s        Start a new session (keep old history)\n", colorCyan(), colorReset())
	fmt.Printf("  %s/clear%s      Clear current session and start fresh\n", colorCyan(), colorReset())
	fmt.Printf("  %s/compact%s    Compact old history into a summary\n", colorCyan(), colorReset())
	fmt.Printf("  %s/reset%s      Clear current session and start fresh\n", colorCyan(), colorReset())
	fmt.Printf("  %s/history%s    Show messages in current session\n", colorCyan(), colorReset())
	fmt.Printf("  %s/model%s      Show the active model and provider\n", colorCyan(), colorReset())
	fmt.Printf("  %s/models%s     Show provider, fallback, and delegation model settings\n", colorCyan(), colorReset())
	fmt.Printf("  %s/models switch%s  Switch active provider/model pair\n", colorCyan(), colorReset())
	fmt.Printf("  %s/channels%s   Show configured channels and allowlist mode\n", colorCyan(), colorReset())
	fmt.Printf("  %s/channels open|strict%s  Set channel allowlist mode\n", colorCyan(), colorReset())
	fmt.Printf("  %s/tools%s      Show tool packs, exec profile, browser, and MCP setup\n", colorCyan(), colorReset())
	fmt.Printf("  %s/tools profile%s  Set exec profile (safe|developer|builder|power-user)\n", colorCyan(), colorReset())
	fmt.Printf("  %s/tools mcp enable|disable%s  Toggle configured MCP server stubs\n", colorCyan(), colorReset())
	fmt.Printf("  %s/browser%s    Show browser automation status and resolved binary\n", colorCyan(), colorReset())
	fmt.Printf("  %s/browser on|off%s  Toggle browser automation\n", colorCyan(), colorReset())
	fmt.Printf("  %s/approvals%s  Show channel allowlist and approval workflow status\n", colorCyan(), colorReset())
	fmt.Printf("  %s/approvals open|strict%s  Set approval/allowlist mode\n", colorCyan(), colorReset())
	fmt.Printf("  %s/logs%s       Show logging output and log file readiness\n", colorCyan(), colorReset())
	fmt.Printf("  %s/doctor%s     Run readiness and health checks\n", colorCyan(), colorReset())
	fmt.Printf("  %s/skill%s      Show workspace and configured jobs\n", colorCyan(), colorReset())
	fmt.Printf("  %s/usage%s      Show cumulative token usage\n", colorCyan(), colorReset())
	fmt.Printf("  %s/cost%s       Show token/cost summary\n", colorCyan(), colorReset())
	fmt.Printf("  %s/cron%s       Show configured cron jobs\n", colorCyan(), colorReset())
	fmt.Printf("  %s/memory%s     Show persistent memory notes\n", colorCyan(), colorReset())
	fmt.Printf("  %s/beliefs%s    Show what the agent currently believes about you\n", colorCyan(), colorReset())
	fmt.Printf("  %s/switch%s     Switch to a previous session (/switch <id-prefix>)\n", colorCyan(), colorReset())
	fmt.Printf("  %sexit%s        Quit the agent\n", colorCyan(), colorReset())
	fmt.Println()
}

func (c *CLI) cmdSlashPalette() {
	items := slashPaletteItems()
	if len(items) == 0 {
		PrintInfo("No slash commands available.")
		return
	}
	selected := promptSelectSingleSearch("Slash Commands", items, 0)
	selected = strings.TrimSpace(selected)
	if selected == "" {
		return
	}
	if !c.HandleCommand(selected) {
		PrintError("Selected command could not be executed: " + selected)
	}
}

func (c *CLI) cmdDoctor(parts []string) {
	if c.cfg == nil {
		PrintError("Doctor is unavailable in this runtime.")
		return
	}

	report := control.BuildDoctorReport(control.DoctorInput{
		Config:     c.cfg,
		DataDir:    c.dataDir,
		ConfigPath: filepath.Join(c.dataDir, "config.yaml"),
	})

	if len(parts) > 1 {
		filter := strings.ToLower(strings.TrimSpace(parts[1]))
		if filter != "" {
			filtered := report
			filtered.Checks = nil
			filtered.Healthy = 0
			filtered.Warnings = 0
			filtered.Failures = 0
			filtered.Status = control.DoctorOK
			for _, ch := range report.Checks {
				if strings.EqualFold(ch.Category, filter) || strings.EqualFold(ch.ID, filter) {
					filtered.Checks = append(filtered.Checks, ch)
					switch ch.Status {
					case control.DoctorOK:
						filtered.Healthy++
					case control.DoctorWarn:
						filtered.Warnings++
						if filtered.Status != control.DoctorFail {
							filtered.Status = control.DoctorWarn
						}
					case control.DoctorFail:
						filtered.Failures++
						filtered.Status = control.DoctorFail
					}
				}
			}
			if len(filtered.Checks) == 0 {
				PrintError("Unknown doctor section. Try: provider, channels, tools, mcp")
				return
			}
			switch filtered.Status {
			case control.DoctorOK:
				filtered.Summary = "All selected checks passed."
			case control.DoctorWarn:
				filtered.Summary = fmt.Sprintf("%d warning(s) in selected checks.", filtered.Warnings)
			case control.DoctorFail:
				filtered.Summary = fmt.Sprintf("%d failure(s) in selected checks.", filtered.Failures)
			}
			report = filtered
		}
	}

	fmt.Print(control.FormatDoctorReportText(report))
}

func (c *CLI) cmdStatus() {
	sessionCount := 0
	if c.sessions != nil {
		if n, err := c.sessions.Count(); err == nil {
			sessionCount = n
		}
	}
	summary := control.BuildStatusSummary("ok", c.cfg, false, "", 0, sessionCount, 0, len(c.cronJobs))
	fmt.Print(control.FormatStatusSummaryText(summary))
}

func (c *CLI) cmdAuth() {
	configured := false
	if c.cfg != nil {
		oc := c.cfg.Auth.OpenAIOAuth
		configured = oc.Enabled &&
			strings.TrimSpace(oc.ClientID) != "" &&
			strings.TrimSpace(oc.AuthorizationURL) != "" &&
			strings.TrimSpace(oc.TokenURL) != ""
	}
	message := "runtime token status available in gateway/UI"
	if !configured {
		message = "oauth not configured"
	}
	summary := control.BuildAuthSummary(configured, false, time.Time{}, message)
	fmt.Print(control.FormatAuthSummaryText(summary))
}

func (c *CLI) cmdPlugins() {
	items := []control.PluginSummaryItem{
		{Name: "cli", Running: true, Healthy: c.agent != nil},
	}
	summary := control.BuildPluginSummary(items)
	fmt.Print(control.FormatPluginSummaryText(summary))
}

func (c *CLI) cmdModels(parts []string) {
	if c.cfg == nil {
		PrintError("Models summary is unavailable in this runtime.")
		return
	}
	if len(parts) > 1 && strings.EqualFold(strings.TrimSpace(parts[1]), "switch") {
		provider := ""
		model := ""
		if len(parts) >= 3 {
			provider = strings.TrimSpace(parts[2])
		}
		if len(parts) >= 4 {
			model = strings.TrimSpace(parts[3])
		}
		if provider == "" || model == "" {
			if !isInteractiveTTY() {
				PrintError("Usage: /models switch <provider> <model>")
				return
			}
			provider, model = c.promptModelSwitch(provider, model)
			if provider == "" || model == "" {
				return
			}
		}
		c.applyModelSelection(provider, model, "")
		return
	}
	summary := control.BuildModelSummary(c.cfg)
	text := control.FormatModelSummaryText(summary)
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "fallback":
			var b strings.Builder
			fmt.Fprintf(&b, "MODELS [fallback]\n")
			if len(summary.FallbackProviders) == 0 {
				b.WriteString("- No fallback providers configured\n")
			} else {
				fmt.Fprintf(&b, "- Providers: %s\n", strings.Join(summary.FallbackProviders, ", "))
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
			text = b.String()
		case "delegation":
			var b strings.Builder
			fmt.Fprintf(&b, "MODELS [delegation]\n")
			if summary.DelegationEnabled {
				b.WriteString("- Delegation: enabled\n")
				fmt.Fprintf(&b, "  · Sub-agent model: %s\n", nonEmptyCLI(summary.SubAgentModel, "inherits primary"))
				fmt.Fprintf(&b, "  · Synthesis model: %s\n", nonEmptyCLI(summary.SynthesisModel, "inherits primary"))
			} else {
				b.WriteString("- Delegation: disabled\n")
			}
			text = b.String()
		default:
			PrintError("Unknown models section. Try: fallback, delegation")
			return
		}
	}
	fmt.Print(text)
}

func (c *CLI) cmdChannels(parts []string) {
	if c.cfg == nil {
		PrintError("Channels summary is unavailable in this runtime.")
		return
	}
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "open":
			c.applyAllowAllMode(true)
			return
		case "strict":
			c.applyAllowAllMode(false)
			return
		}
	}
	summary := control.BuildChannelSummary(c.cfg)
	text := control.FormatChannelSummaryText(summary)
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "allowlist":
			if summary.AllowAll {
				text = "CHANNELS [allowlist]\n- Mode: allow_all=true\n"
			} else {
				text = "CHANNELS [allowlist]\n- Mode: strict\n"
			}
		default:
			PrintError("Unknown channels section. Try: allowlist")
			return
		}
	}
	fmt.Print(text)
}

func (c *CLI) cmdTools(parts []string) {
	if c.cfg == nil {
		PrintError("Tools summary is unavailable in this runtime.")
		return
	}
	if len(parts) > 2 && strings.EqualFold(strings.TrimSpace(parts[1]), "mcp") {
		action := strings.ToLower(strings.TrimSpace(parts[2]))
		name := ""
		if len(parts) > 3 {
			name = strings.TrimSpace(parts[3])
		}
		if name == "" && isInteractiveTTY() && (action == "enable" || action == "disable") {
			name = c.promptMCPServerName(action)
		}
		switch action {
		case "enable":
			if name == "" {
				PrintError("Usage: /tools mcp enable <name>")
				return
			}
			c.applyMCPServerEnabled(name, true)
			return
		case "disable":
			if name == "" {
				PrintError("Usage: /tools mcp disable <name>")
				return
			}
			c.applyMCPServerEnabled(name, false)
			return
		}
	}
	if len(parts) > 1 && strings.EqualFold(strings.TrimSpace(parts[1]), "profile") {
		profile := ""
		if len(parts) >= 3 {
			profile = strings.TrimSpace(parts[2])
		}
		if profile == "" {
			if !isInteractiveTTY() {
				PrintError("Usage: /tools profile <safe|developer|builder|power-user>")
				return
			}
			profile = c.promptExecProfile()
			if profile == "" {
				return
			}
		}
		c.applyExecProfile(profile)
		return
	}
	summary := control.BuildToolingSummary(c.cfg)
	text := control.FormatToolingSummaryText(summary)
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "exec":
			var b strings.Builder
			fmt.Fprintf(&b, "TOOLS [exec]\n")
			fmt.Fprintf(&b, "- Profile: %s\n", nonEmptyCLI(summary.ExecProfile, "safe"))
			if len(summary.AllowedCommands) == 0 {
				b.WriteString("- Allowed commands: none configured\n")
			} else {
				fmt.Fprintf(&b, "- Allowed commands: %s\n", strings.Join(summary.AllowedCommands, ", "))
			}
			if summary.ApprovalOnBlock {
				b.WriteString("- Approval on block: enabled\n")
			} else {
				b.WriteString("- Approval on block: disabled\n")
			}
			text = b.String()
		case "browser":
			var b strings.Builder
			fmt.Fprintf(&b, "TOOLS [browser]\n")
			if summary.Browser.Enabled {
				fmt.Fprintf(&b, "- Browser: enabled (%s)\n", ternaryCLI(summary.Browser.Headless, "headless", "headed"))
				if summary.Browser.Path != "" {
					fmt.Fprintf(&b, "- Binary path: %s\n", summary.Browser.Path)
				}
			} else {
				b.WriteString("- Browser: disabled\n")
			}
			text = b.String()
		case "mcp":
			var b strings.Builder
			fmt.Fprintf(&b, "TOOLS [mcp]\n")
			if len(summary.MCPPresets) == 0 {
				b.WriteString("- MCP presets: none\n")
			} else {
				fmt.Fprintf(&b, "- Presets: %s\n", strings.Join(summary.MCPPresets, ", "))
			}
			if len(summary.MCPServers) == 0 {
				b.WriteString("- Servers: none\n")
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
			b.WriteString("- Mutations: /tools mcp enable <name> | /tools mcp disable <name>\n")
			text = b.String()
		default:
			PrintError("Unknown tools section. Try: exec, browser, mcp")
			return
		}
	}
	fmt.Print(text)
}

func (c *CLI) cmdApprovals(parts []string) {
	if c.cfg == nil {
		PrintError("Approvals summary is unavailable in this runtime.")
		return
	}
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "open":
			c.applyAllowAllMode(true)
			return
		case "strict":
			c.applyAllowAllMode(false)
			return
		}
	}
	summary := control.BuildApprovalsSummary(c.cfg, c.cfg.Channels.AllowAll, readApprovedSenders(c.dataDir))
	text := control.FormatApprovalsSummaryText(summary)
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "allowlist":
			var b strings.Builder
			fmt.Fprintf(&b, "APPROVALS [allowlist]\n")
			if summary.AllowAll {
				b.WriteString("- Mode: allow_all=true\n")
			} else {
				b.WriteString("- Mode: strict\n")
			}
			fmt.Fprintf(&b, "- Approved senders: %d\n", summary.ApprovedCount)
			text = b.String()
		default:
			PrintError("Unknown approvals section. Try: allowlist")
			return
		}
	}
	fmt.Print(text)
}

func (c *CLI) cmdBrowser(parts []string) {
	if c.cfg == nil {
		PrintError("Browser summary is unavailable in this runtime.")
		return
	}
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "on":
			c.applyBrowserEnabled(true)
			return
		case "off":
			c.applyBrowserEnabled(false)
			return
		}
	}
	summary := control.BuildBrowserSummary(c.cfg)
	text := control.FormatBrowserSummaryText(summary)
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "binary":
			var b strings.Builder
			fmt.Fprintf(&b, "BROWSER [binary]\n")
			if summary.ConfigPath != "" {
				fmt.Fprintf(&b, "- Configured path: %s\n", summary.ConfigPath)
			}
			if summary.Available {
				fmt.Fprintf(&b, "- Resolved binary: %s\n", summary.ResolvedPath)
			} else {
				b.WriteString("- Resolved binary: unavailable\n")
			}
			text = b.String()
		default:
			PrintError("Unknown browser section. Try: binary")
			return
		}
	}
	fmt.Print(text)
}

func (c *CLI) cmdLogs(parts []string) {
	if c.cfg == nil {
		PrintError("Logs summary is unavailable in this runtime.")
		return
	}
	summary := control.BuildLogsSummary(c.cfg)
	text := control.FormatLogsSummaryText(summary)
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "error":
			var b strings.Builder
			fmt.Fprintf(&b, "LOGS [error]\n")
			if strings.TrimSpace(summary.Output) == "" {
				b.WriteString("- Output: not configured\n")
			} else {
				fmt.Fprintf(&b, "- Output: %s\n", summary.Output)
			}
			b.WriteString("- Use /api/v1/logs?level=error or the Logs panel for recent error lines\n")
			text = b.String()
		default:
			PrintError("Unknown logs section. Try: error")
			return
		}
	}
	fmt.Print(text)
}

func nonEmptyCLI(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func ternaryCLI(cond bool, a, b string) string {
	if cond {
		return a
	}
	return b
}

func (c *CLI) applyBrowserEnabled(enabled bool) {
	result, err := control.SetBrowserEnabled(control.ActionEnv{
		Config:     c.cfg,
		DataDir:    c.dataDir,
		WriteTools: true,
	}, enabled)
	if err != nil {
		PrintError("Failed to update browser setting: " + err.Error())
		return
	}
	PrintInfo(result.Note + " (restart running services if needed)")
}

func (c *CLI) applyAllowAllMode(allowAll bool) {
	result, err := control.SetAllowAllMode(control.ActionEnv{
		Config:     c.cfg,
		DataDir:    c.dataDir,
		WriteTools: true,
	}, allowAll)
	if err != nil {
		PrintError("Failed to update channel allowlist mode: " + err.Error())
		return
	}
	PrintInfo(result.Note)
}

func (c *CLI) applyExecProfile(profile string) {
	result, err := control.SetExecProfile(control.ActionEnv{
		Config:     c.cfg,
		DataDir:    c.dataDir,
		WriteTools: true,
	}, profile)
	if err != nil {
		PrintError("Failed to update exec profile: " + err.Error())
		return
	}
	PrintInfo(result.Note + " (restart running services if needed)")
}

func (c *CLI) applyModelSelection(provider, model, baseURL string) {
	result, err := control.SetActiveModelConfig(control.ActionEnv{
		Config:     c.cfg,
		DataDir:    c.dataDir,
		WriteTools: true,
	}, provider, model, baseURL)
	if err != nil {
		PrintError("Failed to update model selection: " + err.Error())
		return
	}
	c.provider = c.cfg.Model.Provider
	c.model = c.cfg.Model.Model
	PrintInfo(result.Note + " (restart running services if needed)")
}

func (c *CLI) applyMCPServerEnabled(name string, enabled bool) {
	result, err := control.SetMCPServerEnabled(control.ActionEnv{
		Config:     c.cfg,
		DataDir:    c.dataDir,
		WriteTools: true,
	}, name, enabled)
	if err != nil {
		PrintError("Failed to update MCP server state: " + err.Error())
		return
	}
	PrintInfo(result.Note + " (restart running services if needed)")
}

func readApprovedSenders(dataDir string) []string {
	if strings.TrimSpace(dataDir) == "" {
		return nil
	}
	path := filepath.Join(dataDir, "allowed_senders.txt")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var out []string
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, line)
	}
	sort.Strings(out)
	return out
}

func (c *CLI) cmdSessions(parts []string) {
	if len(parts) > 1 {
		switch strings.ToLower(strings.TrimSpace(parts[1])) {
		case "delete", "rm", "remove":
			c.cmdSessionsDelete(parts[2:])
			return
		default:
			PrintError("Usage: /sessions | /sessions delete <id|current>")
			return
		}
	}

	sessions, err := c.sessions.List(10)
	if err != nil {
		PrintError("Failed to list sessions: " + err.Error())
		return
	}

	infos := make([]SessionInfo, len(sessions))
	for i, s := range sessions {
		infos[i] = SessionInfo{
			ID:         s.ID,
			Channel:    s.Channel,
			LastActive: s.LastActive.Format("2006-01-02 15:04"),
		}
	}
	PrintSessionList(infos)
}

func (c *CLI) cmdSessionsDelete(args []string) {
	if len(args) == 0 {
		if !isInteractiveTTY() {
			PrintError("Usage: /sessions delete <id|current>")
			return
		}
		selected := c.promptSessionDeleteTarget()
		if strings.TrimSpace(selected) == "" {
			return
		}
		args = []string{selected}
	}
	target := strings.TrimSpace(args[0])
	if strings.EqualFold(target, "current") {
		if strings.TrimSpace(c.sessionID) == "" {
			PrintError("No active session.")
			return
		}
		target = c.sessionID
	}

	result, err := control.DeleteSession(control.ActionEnv{
		DeleteSession: c.sessions.Delete,
	}, target)
	if err != nil {
		PrintError("Failed to delete session: " + err.Error())
		return
	}

	if target == c.sessionID {
		session, err := c.sessions.Create("cli", "local")
		if err != nil {
			PrintError("Session deleted, but failed to create replacement session: " + err.Error())
			c.sessionID = ""
			return
		}
		c.sessionID = session.ID
		c.totalUsage = totalUsage{}
		PrintInfo(result.Note + ". New session: " + session.ID[:8] + "...")
		return
	}

	PrintInfo(result.Note)
}

func (c *CLI) promptExecProfile() string {
	profiles := cfgpkg.ExecProfileCatalog()
	items := make([]promptSelectItem, 0, len(profiles))
	defaultIndex := 0
	current := strings.TrimSpace(c.cfg.Tools.Exec.Profile)
	for i, profile := range profiles {
		items = append(items, promptSelectItem{
			Label:       profile.Name,
			Hint:        fmt.Sprintf("%d commands allowed", len(profile.Commands)),
			Value:       profile.Name,
			Group:       "Tools",
			Description: profile.Description,
		})
		if strings.EqualFold(profile.Name, current) {
			defaultIndex = i
		}
	}
	return promptSelectSingleSearch("Choose exec profile", items, defaultIndex)
}

func (c *CLI) promptMCPServerName(action string) string {
	if len(c.cfg.MCPServers) == 0 {
		PrintInfo("No MCP servers are configured.")
		return ""
	}
	items := make([]promptSelectItem, 0, len(c.cfg.MCPServers))
	defaultIndex := 0
	wantEnable := strings.EqualFold(action, "enable")
	for i, server := range c.cfg.MCPServers {
		state := "disabled"
		enabled := server.Enabled != nil && *server.Enabled
		if enabled {
			state = "enabled"
		}
		items = append(items, promptSelectItem{
			Label:       server.Name,
			Hint:        state,
			Value:       server.Name,
			Group:       "MCP",
			Description: fmt.Sprintf("%s via %s", server.Name, nonEmptyCLI(server.Command, "configured command")),
		})
		if wantEnable && !enabled {
			defaultIndex = i
		}
		if !wantEnable && enabled {
			defaultIndex = i
		}
	}
	return promptSelectSingleSearch("Choose MCP server", items, defaultIndex)
}

func (c *CLI) promptSessionDeleteTarget() string {
	if c.sessions == nil {
		PrintInfo("Session store unavailable.")
		return ""
	}
	list, err := c.sessions.List(12)
	if err != nil {
		PrintError("Failed to list sessions: " + err.Error())
		return ""
	}
	if len(list) == 0 {
		PrintInfo("No sessions available to delete.")
		return ""
	}
	items := make([]promptSelectItem, 0, len(list)+1)
	defaultIndex := 0
	if strings.TrimSpace(c.sessionID) != "" {
		items = append(items, promptSelectItem{
			Label:       "current",
			Hint:        shortSessionID(c.sessionID),
			Value:       "current",
			Group:       "Sessions",
			Description: "Delete the active session and immediately create a replacement.",
		})
	}
	for i, s := range list {
		items = append(items, promptSelectItem{
			Label:       shortSessionID(s.ID),
			Hint:        fmt.Sprintf("%s • %s", s.Channel, s.LastActive.Format("2006-01-02 15:04")),
			Value:       s.ID,
			Group:       "Sessions",
			Description: "Delete this saved session from local storage.",
		})
		if s.ID == c.sessionID && defaultIndex == 0 {
			defaultIndex = i + 1
		}
	}
	return promptSelectSingleSearch("Choose session to delete", items, defaultIndex)
}

func shortSessionID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= 8 {
		return id
	}
	return id[:8] + "..."
}

func (c *CLI) promptModelSwitch(provider, model string) (string, string) {
	provider = strings.TrimSpace(provider)
	model = strings.TrimSpace(model)
	if provider == "" {
		provider = c.promptProviderName()
		if provider == "" {
			return "", ""
		}
	}
	if model == "" {
		model = c.promptProviderModel(provider)
	}
	return strings.TrimSpace(provider), strings.TrimSpace(model)
}

func (c *CLI) promptProviderName() string {
	items := make([]promptSelectItem, 0, 4)
	defaultIndex := 0
	for i, provider := range c.providerOptions() {
		items = append(items, promptSelectItem{
			Label:       provider,
			Hint:        c.defaultModelSuggestion(provider),
			Value:       provider,
			Group:       "Models",
			Description: "Switch the active provider.",
		})
		if strings.EqualFold(provider, c.cfg.Model.Provider) {
			defaultIndex = i
		}
	}
	return promptSelectSingleSearch("Choose provider", items, defaultIndex)
}

func (c *CLI) promptProviderModel(provider string) string {
	options := c.providerModelOptions(provider)
	if len(options) == 0 {
		PrintError("No model suggestions available for provider " + provider)
		return ""
	}
	defaultIndex := 0
	current := c.defaultModelSuggestion(provider)
	for i, option := range options {
		if option.Value == current {
			defaultIndex = i
			break
		}
	}
	return promptSelectSingleSearch("Choose model", options, defaultIndex)
}

func (c *CLI) providerOptions() []string {
	seen := map[string]struct{}{}
	var providers []string
	add := func(value string) {
		value = strings.TrimSpace(strings.ToLower(value))
		if value == "" {
			return
		}
		if _, ok := seen[value]; ok {
			return
		}
		seen[value] = struct{}{}
		providers = append(providers, value)
	}
	add(c.cfg.Model.Provider)
	for _, provider := range c.cfg.Model.FallbackProviders {
		add(provider)
	}
	for _, provider := range []string{"ollama", "openai", "anthropic", "gemini"} {
		add(provider)
	}
	return providers
}

func (c *CLI) defaultModelSuggestion(provider string) string {
	provider = strings.TrimSpace(strings.ToLower(provider))
	if provider == strings.TrimSpace(strings.ToLower(c.cfg.Model.Provider)) && strings.TrimSpace(c.cfg.Model.Model) != "" {
		return strings.TrimSpace(c.cfg.Model.Model)
	}
	if model, ok := c.cfg.Model.FallbackModels[provider]; ok && strings.TrimSpace(model) != "" {
		return strings.TrimSpace(model)
	}
	switch provider {
	case "ollama":
		return "gpt-oss:20b"
	case "openai":
		return "gpt-4.1-mini"
	case "anthropic":
		return "claude-sonnet-4-6"
	case "gemini":
		return "gemini-3.1-flash"
	default:
		return ""
	}
}

func (c *CLI) providerModelOptions(provider string) []promptSelectItem {
	provider = strings.TrimSpace(strings.ToLower(provider))
	seen := map[string]struct{}{}
	var items []promptSelectItem
	add := func(model, desc string) {
		model = strings.TrimSpace(model)
		if model == "" {
			return
		}
		if _, ok := seen[model]; ok {
			return
		}
		seen[model] = struct{}{}
		items = append(items, promptSelectItem{
			Label:       model,
			Hint:        provider,
			Value:       model,
			Group:       "Models",
			Description: desc,
		})
	}
	add(c.defaultModelSuggestion(provider), "Current or recommended model for this provider.")
	switch provider {
	case "ollama":
		add("gpt-oss:120b", "Higher-capability local model.")
		add("qwen3-coder", "Coding-focused local model.")
		add("llama3.1", "Stable general-purpose local family.")
	case "openai":
		add("gpt-4o-mini", "Fast multimodal OpenAI model.")
		add("gpt-4.1", "Higher-capability OpenAI model.")
		add("o4-mini", "Compact reasoning model.")
	case "anthropic":
		add("claude-opus-4-6", "Highest-capability Claude tier.")
		add("claude-haiku-4-5", "Fast Claude tier.")
		add("claude-sonnet-4-5", "Stable Claude Sonnet generation.")
	case "gemini":
		add("gemini-2.5-flash", "Fast Gemini tier.")
		add("gemini-3.1-pro", "Higher-capability Gemini tier.")
		add("gemini-2.5-pro", "Reasoning-focused Gemini tier.")
	}
	return items
}

func (c *CLI) cmdNew() {
	session, err := c.sessions.Create("cli", "local")
	if err != nil {
		PrintError("Failed to create session: " + err.Error())
		return
	}
	c.sessionID = session.ID
	c.totalUsage = totalUsage{}
	PrintInfo(fmt.Sprintf("New session: %s", session.ID[:8]+"..."))
}

func (c *CLI) cmdHistory() {
	if c.sessionID == "" {
		PrintError("No active session.")
		return
	}

	messages, err := c.messages.GetBySession(c.sessionID)
	if err != nil {
		PrintError("Failed to get history: " + err.Error())
		return
	}

	if len(messages) == 0 {
		PrintInfo("No messages yet.")
		return
	}

	fmt.Println()
	for _, m := range messages {
		roleColor := colorDim()
		switch m.Role {
		case "user":
			roleColor = colorBold()
		case "assistant":
			roleColor = colorGreen()
		}

		content := m.Content
		if len(content) > 200 {
			content = content[:200] + "..."
		}
		fmt.Printf("  %s[%s]%s %s\n", roleColor, m.Role, colorReset(), content)
	}
	fmt.Println()
}

func (c *CLI) cmdUsage() {
	fmt.Println()
	fmt.Printf("  %sTotal Tokens:%s %d in / %d out\n", colorDim(), colorReset(), c.totalUsage.inputTokens, c.totalUsage.outputTokens)
	fmt.Printf("  %sLLM Calls:%s    %d\n", colorDim(), colorReset(), c.totalUsage.llmCalls)

	if c.costTracker != nil {
		if summary, err := c.costTracker.GetSummaryBySession(c.sessionID); err == nil && summary.CallCount > 0 {
			fmt.Printf("  %sCost Check:%s   ~$%.4f\n", colorDim(), colorReset(), summary.TotalCost)
		} else {
			fmt.Printf("  %sCost Check:%s   ~$%.4f\n", colorDim(), colorReset(), 0.0)
		}
	} else {
		fmt.Printf("  %sCost Check:%s   not available (no tracker)\n", colorDim(), colorReset())
	}

	fmt.Println()
}

func (c *CLI) cmdCost() {
	if c.costTracker == nil {
		PrintError("Cost tracking is not enabled.")
		return
	}

	summaries := make(map[string]*cost.Summary)
	for _, period := range []string{"today", "week", "month", "all"} {
		if s, err := c.costTracker.GetSummary(period); err == nil {
			summaries[period] = s
		}
	}
	byProvider, _ := c.costTracker.ProviderBreakdown("all")
	currentSession, _ := c.costTracker.GetSummaryBySession(c.sessionID)

	fmt.Println()
	fmt.Print(cost.FormatSummary(summaries, byProvider, currentSession))
}

func (c *CLI) cmdMemory() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		PrintError("Failed to resolve home directory: " + err.Error())
		return
	}

	memoryPath := filepath.Join(homeDir, ".openclio", "memory.md")
	data, err := os.ReadFile(memoryPath)
	if err != nil {
		if os.IsNotExist(err) {
			PrintInfo("No memory saved yet. The agent will populate memory over time.")
			return
		}
		PrintError("Failed to read memory: " + err.Error())
		return
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		PrintInfo("Memory file is currently empty.")
		return
	}

	fmt.Println()
	fmt.Printf("%sPersistent memory (%s):%s\n", colorBold(), memoryPath, colorReset())
	fmt.Println(content)
	fmt.Println()
}

func (c *CLI) cmdCron() {
	fmt.Println()
	fmt.Printf("%sCRON%s\n", colorBold(), colorReset())
	if len(c.cronJobs) == 0 {
		fmt.Printf("  %sNo configured cron jobs.%s\n\n", colorDim(), colorReset())
		return
	}
	fmt.Printf("  %sConfigured jobs:%s %d\n", colorDim(), colorReset(), len(c.cronJobs))
	for _, job := range c.cronJobs {
		fmt.Printf("  · %s\n", strings.TrimSpace(job))
	}
	fmt.Println()
}

// cmdReset clears the current session and starts a fresh one.
func (c *CLI) cmdReset() {
	if c.sessionID != "" {
		if err := c.sessions.Delete(c.sessionID); err != nil {
			PrintError("Failed to delete session: " + err.Error())
		}
	}
	session, err := c.sessions.Create("cli", "local")
	if err != nil {
		PrintError("Failed to create session: " + err.Error())
		return
	}
	c.sessionID = session.ID
	c.totalUsage = totalUsage{}
	PrintInfo("🗑  Session cleared. New session: " + session.ID[:8] + "...")
}

func (c *CLI) cmdCompact() {
	if c.sessionID == "" {
		PrintError("No active session.")
		return
	}

	msgProvider := &cliMessageProvider{
		messages:  c.messages,
		sessionID: c.sessionID,
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	if err := c.agent.ForceCompaction(ctx, c.sessionID, msgProvider, nil); err != nil {
		PrintError("Compaction failed: " + err.Error())
		return
	}
	PrintInfo("Compaction complete.")
}

// cmdModel prints the active LLM provider and model.
func (c *CLI) cmdModel() {
	fmt.Println()
	fmt.Printf("  %sProvider:%s  %s\n", colorDim(), colorReset(), c.provider)
	fmt.Printf("  %sModel:%s     %s\n", colorDim(), colorReset(), c.model)
	fmt.Printf("  %sHint:%s      Change via model.provider/model.model in config.yaml\n", colorDim(), colorReset())
	fmt.Println()
}

// cmdSkill loads a skill into the current session, or prints workspace/cron status.
func (c *CLI) cmdSkill(parts []string) {
	if len(parts) == 1 {
		// Old behaviour: just print workspace and cron jobs
		fmt.Println()
		if c.workspaceName != "" {
			fmt.Printf("  %sWorkspace:%s  %s\n", colorDim(), colorReset(), c.workspaceName)
		} else {
			fmt.Printf("  %sWorkspace:%s  (not configured — add ~/.openclio/workspace.yaml)\n", colorDim(), colorReset())
		}
		if len(c.cronJobs) > 0 {
			fmt.Printf("  %sScheduled jobs:%s\n", colorDim(), colorReset())
			for _, j := range c.cronJobs {
				fmt.Printf("    • %s%s%s\n", colorCyan(), j, colorReset())
			}
		} else {
			fmt.Printf("  %sScheduled jobs:%s  none (add cron: entries in config.yaml)\n", colorDim(), colorReset())
		}
		fmt.Println()
		return
	}

	skillName := parts[1]
	homeDir, err := os.UserHomeDir()
	if err != nil {
		PrintError("Failed to get home dir: " + err.Error())
		return
	}

	skillPath := filepath.Join(homeDir, ".openclio", "skills", skillName+".md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		if os.IsNotExist(err) {
			PrintError(fmt.Sprintf("Skill '%s' not found in ~/.openclio/skills/", skillName))
			return
		}
		PrintError("Failed to read skill: " + err.Error())
		return
	}

	content := strings.TrimSpace(string(data))
	if content == "" {
		PrintError("Skill file is empty.")
		return
	}

	// Insert as a system message so the LLM internalizes the instruction
	// Using 0 for tokens as it's a dynamic system injection
	_, err = c.messages.Insert(c.sessionID, "system", fmt.Sprintf("[Skill Instruction: %s]\n%s", skillName, content), 0)
	if err != nil {
		PrintError("Failed to inject skill into session: " + err.Error())
		return
	}

	PrintInfo(fmt.Sprintf("✅ Loaded skill '%s' into the current session.", skillName))
}

func (c *CLI) cmdDebugContext() {
	if c.lastContext == nil {
		PrintError("No context assembled yet in this session.")
		return
	}
	fmt.Println()
	fmt.Printf("%s--- Last Assembled Context ---%s\n", colorBold(), colorReset())
	fmt.Printf("%sSystem Prompt:%s\n%s\n\n", colorCyan(), colorReset(), c.lastContext.SystemPrompt)
	fmt.Printf("%sMessages (%d):%s\n", colorCyan(), len(c.lastContext.Messages), colorReset())
	for i, m := range c.lastContext.Messages {
		content := m.Content
		if len(content) > 300 {
			content = content[:300] + "... (truncated)"
		}
		fmt.Printf("  [%d] %s%s%s: %s\n", i+1, colorBold(), m.Role, colorReset(), content)
	}
	fmt.Println()
}

func (c *CLI) cmdDebugTokens() {
	if c.lastContext == nil {
		PrintError("No context assembled yet in this session.")
		return
	}
	st := c.lastContext.Stats
	fmt.Println()
	fmt.Printf("%s--- Last Context Token Breakdown ---%s\n", colorBold(), colorReset())
	fmt.Printf("  System Prompt:      %d\n", st.SystemPromptTokens)
	fmt.Printf("  User Message:       %d\n", st.UserMessageTokens)
	fmt.Printf("  Recent Turns:       %d\n", st.RecentTurnTokens)
	fmt.Printf("  Retrieved History:  %d (from %d messages)\n", st.RetrievedHistoryTokens, st.RetrievedMessagesCount)
	fmt.Printf("  Semantic Memory:    %d\n", st.SemanticMemoryTokens)
	fmt.Printf("  Knowledge Graph:    %d (from %d nodes)\n", st.KnowledgeGraphTokens, st.KnowledgeNodesCount)
	fmt.Printf("  Tool Definitions:   %d\n", st.ToolDefTokens)
	fmt.Printf("  --------------------------\n")
	fmt.Printf("  Total Context:      %d / %d budget limit\n", st.TotalTokens, st.BudgetTotal)
	fmt.Printf("  Remaining Budget:   %d\n", st.BudgetRemaining)
	fmt.Println()
}

// cmdSwitch switches the active session by ID prefix.
// With no arg it lists sessions. With a prefix it switches to the matching session.
func (c *CLI) cmdSwitch(parts []string) {
	if len(parts) < 2 {
		// No prefix — list sessions with numbers for easy picking
		sessions, err := c.sessions.List(10)
		if err != nil {
			PrintError("Failed to list sessions: " + err.Error())
			return
		}
		if len(sessions) == 0 {
			PrintInfo("No sessions found.")
			return
		}
		fmt.Println()
		fmt.Printf("%sSessions (use /switch <id-prefix> to jump to one):%s\n\n", colorBold(), colorReset())
		for i, s := range sessions {
			marker := "  "
			if s.ID == c.sessionID {
				marker = colorGreen() + "▶ " + colorReset()
			}
			fmt.Printf("%s%d. %s%s  %s%s%s  %s\n",
				marker, i+1,
				colorBold(), s.ID[:8]+"...", colorReset(),
				colorDim(), s.LastActive.Format("2006-01-02 15:04"), colorReset(),
			)
		}
		fmt.Println()
		return
	}

	prefix := strings.ToLower(strings.TrimSpace(parts[1]))
	sessions, err := c.sessions.List(50)
	if err != nil {
		PrintError("Failed to list sessions: " + err.Error())
		return
	}

	var matched []string
	for _, s := range sessions {
		if strings.HasPrefix(strings.ToLower(s.ID), prefix) {
			matched = append(matched, s.ID)
		}
	}

	switch len(matched) {
	case 0:
		PrintError(fmt.Sprintf("No session found with prefix %q.", prefix))
	case 1:
		c.sessionID = matched[0]
		c.totalUsage = totalUsage{}
		PrintInfo(fmt.Sprintf("Switched to session %s", matched[0][:8]+"..."))
	default:
		PrintError(fmt.Sprintf("Ambiguous prefix %q — %d sessions match. Use more characters.", prefix, len(matched)))
	}
}
