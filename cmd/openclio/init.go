package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/openclio/openclio/internal/brand"
	"github.com/openclio/openclio/internal/config"
	"github.com/openclio/openclio/internal/workspace"
)

const (
	initBannerBlue  = "\033[38;5;39m"
	initBannerDim   = "\033[2m"
	initBannerBold  = "\033[1m"
	initBannerReset = "\033[0m"
)

func initBannerAnsi(code string) string {
	if isInteractiveInitOutput() {
		return code
	}
	return ""
}

func initBannerBlueText() string  { return initBannerAnsi(initBannerBlue) }
func initBannerDimText() string   { return initBannerAnsi(initBannerDim) }
func initBannerBoldText() string  { return initBannerAnsi(initBannerBold) }
func initBannerResetText() string { return initBannerAnsi(initBannerReset) }

func isInteractiveInitOutput() bool {
	stat, err := os.Stdout.Stat()
	return err == nil && (stat.Mode()&os.ModeCharDevice != 0)
}

// runInit is the interactive first-time setup wizard.
// It runs BEFORE any config or database is loaded so it works on a fresh install.
func runInit(dataDir string) {
	configPath := filepath.Join(dataDir, "config.yaml")

	fmt.Println()
	printOpenClioBanner("Setup Wizard")
	fmt.Println()

	// Check for existing config
	if _, err := os.Stat(configPath); err == nil {
		fmt.Printf("⚠️  Config already exists at %s\n", configPath)
		if !promptConfirm("Overwrite and start fresh?", false) {
			fmt.Println("Setup cancelled. Your existing config is unchanged.")
			os.Exit(0)
		}
		fmt.Println()
	}

	// Ensure data directory exists
	if err := os.MkdirAll(dataDir, 0700); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating %s: %v\n", dataDir, err)
		os.Exit(1)
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 1: Create Your Assistant
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 1: Create Your Assistant                                  │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	assistantName := promptString("What should your assistant be called?", "Aria")
	if assistantName == "" {
		assistantName = "Aria"
	}

	fmt.Println()
	fmt.Println("Pick the default tone and working style for your assistant.")
	fmt.Println()
	personalityChoice := promptSelectSingle("Personality style", []selectItem{
		{Label: "Professional", Hint: "Direct, efficient, business-focused", Value: "1"},
		{Label: "Technical", Hint: "Precise, thorough, code-first mindset", Value: "2"},
		{Label: "Creative", Hint: "Exploratory, suggestive, brainstorming-oriented", Value: "3"},
		{Label: "Minimal", Hint: "Ultra-concise, no fluff, just facts", Value: "4"},
		{Label: "Balanced", Hint: "Friendly mix of all above", Value: "5"},
	}, 4)

	personalityTraits := map[string][]string{
		"1": {"Professional", "direct", "efficient", "business-focused", "clear communicator"},
		"2": {"Technical", "precise", "thorough", "code-first", "detail-oriented"},
		"3": {"Creative", "exploratory", "suggestive", "brainstorming", "idea-generator"},
		"4": {"Minimal", "ultra-concise", "no-fluff", "just-the-facts", "high-density"},
		"5": {"Balanced", "friendly", "adaptable", "clear", "helpful"},
	}
	traits := personalityTraits[personalityChoice]

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 2: About You
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 2: Tell Me About You                                      │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("This helps me understand your context and communicate better.")
	fmt.Println("(Press Enter to skip any question)")
	fmt.Println()

	userName := promptString("Your name?", "")
	userRole := promptString("Your profession/role?", "")
	userStack := promptString("Primary tech stack? (e.g., 'Go, React, PostgreSQL')", "")

	fmt.Println()
	fmt.Println("Pick the default level of detail for responses.")
	fmt.Println()
	responseStyle := promptSelectSingle("Response style", []selectItem{
		{Label: "Detailed", Hint: "Thorough explanations, examples, context", Value: "1"},
		{Label: "Balanced", Hint: "Good detail without overwhelming", Value: "2"},
		{Label: "Concise", Hint: "Bullet points, minimal text, just what you need", Value: "3"},
	}, 1)

	responseStyleLabels := map[string]string{
		"1": "detailed",
		"2": "balanced",
		"3": "concise",
	}

	fmt.Println()

	// Build rich user profile
	var userProfileParts []string
	if userName != "" {
		userProfileParts = append(userProfileParts, fmt.Sprintf("My name is %s.", userName))
	}
	if userRole != "" {
		userProfileParts = append(userProfileParts, fmt.Sprintf("I work as a %s.", userRole))
	}
	if userStack != "" {
		userProfileParts = append(userProfileParts, fmt.Sprintf("My primary tech stack includes: %s.", userStack))
	}
	userProfileParts = append(userProfileParts, fmt.Sprintf("I prefer %s responses.", responseStyleLabels[responseStyle]))

	if len(userProfileParts) == 0 {
		userProfileParts = append(userProfileParts, "I am a developer and prefer concise, practical answers.")
	}
	userProfile := strings.Join(userProfileParts, " ")

	// Install complete template set with the assistant name
	if err := workspace.InstallDefaults(dataDir, assistantName); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to install default templates: %v\n", err)
	}

	// Install user profile using the template
	if err := workspace.InstallUserProfile(dataDir, userProfile); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to install user profile: %v\n", err)
	}

	// Seed bundled default skills for backward compatibility
	if err := workspace.SeedDefaultSkills(dataDir); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: failed to seed default skills: %v\n", err)
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 3: Configure Memory
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 3: Configure Memory                                       │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("Pick the memory mode you want enabled by default.")
	fmt.Println()
	memoryModeChoice := promptSelectSingle("Memory mode", []selectItem{
		{Label: "Off", Hint: "Only the current conversation, no memory recall", Value: "1"},
		{Label: "Standard", Hint: "Stable long-term memory with conservative behavior", Value: "2"},
		{Label: "Enhanced", Hint: "Best memory experience, richer recall and anticipation", Value: "3"},
	}, 2)
	memoryModeMap := map[string]string{
		"1": memoryModeOff,
		"2": memoryModeStandard,
		"3": memoryModeEnhanced,
	}
	memoryMode := memoryModeMap[memoryModeChoice]

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 4: Configure Tools
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 4: Configure Tools                                        │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("Choose the capability packs you want available by default.")
	fmt.Println()
	selectedToolPacks := promptCheckboxMulti("Tool packs", []checkboxItem{
		{Label: "Developer", Hint: "Coding, git, tests, repo search, delegation", Value: "developer", Checked: true},
		{Label: "Research", Hint: "Web search, browser, docs retrieval, note workflows", Value: "research"},
		{Label: "Builder", Hint: "Docker, builds, package managers, deployment CLIs", Value: "builder"},
		{Label: "Writing", Hint: "Drafting, editing, export helpers", Value: "writing"},
	}, 0)
	if len(selectedToolPacks) == 0 {
		selectedToolPacks = []string{"developer"}
	}

	fmt.Println()
	advancedTooling := promptSelectSingle("Advanced tooling setup", []selectItem{
		{Label: "No", Hint: "Use recommended MCP presets and exec profile automatically", Value: "no"},
		{Label: "Yes", Hint: "Choose MCP presets and local CLI profile yourself", Value: "yes"},
	}, 0) == "yes"

	selectedMCPPresets := []string(nil)
	selectedExecProfile := ""
	if advancedTooling {
		fmt.Println()
		fmt.Println("Select MCP presets to scaffold into your config (disabled until you install them).")
		fmt.Println()
		selectedMCPPresets = promptCheckboxMulti("MCP presets", []checkboxItem{
			{Label: "GitHub", Hint: "Issues, PRs, repos", Value: "github"},
			{Label: "Linear", Hint: "Issues and projects", Value: "linear"},
			{Label: "Notion", Hint: "Docs and workspace content", Value: "notion"},
			{Label: "Playwright MCP", Hint: "Advanced browser MCP server; separate from the built-in browser tool", Value: "browser"},
			{Label: "Database", Hint: "Database and SQL access", Value: "database"},
			{Label: "Figma", Hint: "Design context and assets", Value: "figma"},
		}, 0)
		fmt.Println()
		selectedExecProfile = promptSelectSingle("Local CLI profile", []selectItem{
			{Label: "Safe", Hint: "Read-only and low-risk commands", Value: "safe"},
			{Label: "Developer", Hint: "Developer workflow commands", Value: "developer"},
			{Label: "Builder", Hint: "Build and container tools", Value: "builder"},
			{Label: "Power User", Hint: "Extended local CLI access", Value: "power-user"},
		}, 1)
	}

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 5: Configure Channels
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 5: Configure Channels                                     │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("Channels allow you to interact with your agent from different interfaces.")
	fmt.Println()
	fmt.Println("  [✓] CLI (always enabled)")
	fmt.Println()

	selectedChannels := promptCheckboxMulti("Enable channels", []checkboxItem{
		{Label: "Telegram", Hint: "Chat with your agent in Telegram", Value: "telegram"},
		{Label: "Discord", Hint: "Use your agent inside Discord", Value: "discord"},
		{Label: "WhatsApp", Hint: "Use QR-based WhatsApp linked devices", Value: "whatsapp"},
		{Label: "WebChat", Hint: "Browser UI on your local machine", Value: "webchat", Checked: true},
	}, 3)
	enableTelegram := containsString(selectedChannels, "telegram")
	enableDiscord := containsString(selectedChannels, "discord")
	enableWhatsApp := containsString(selectedChannels, "whatsapp")
	enableWebChat := containsString(selectedChannels, "webchat")

	fmt.Println()

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 6: Configure AI Provider
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 6: Configure AI Provider                                  │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()
	fmt.Println("Select one or more providers. The first selection becomes the default.")
	fmt.Println()
	selectedProviders := promptCheckboxMulti("Choose one or more providers", []checkboxItem{
		{Label: "Ollama", Hint: "Local, free, fully private", Value: "ollama"},
		{Label: "OpenAI", Hint: "GPT models, fast and reliable", Value: "openai", Checked: true},
		{Label: "Anthropic", Hint: "Claude models, excellent reasoning", Value: "anthropic"},
		{Label: "Google", Hint: "Gemini models, good balance", Value: "gemini"},
	}, 1)
	if len(selectedProviders) == 0 {
		selectedProviders = []string{"openai"}
	}

	provider := selectedProviders[0]
	fallbackProviders := append([]string(nil), selectedProviders[1:]...)
	providerModels := make(map[string]providerModelSelection, len(selectedProviders))
	providerAPIEnvs := make(map[string]string, len(selectedProviders))
	providerLabels := map[string]string{
		"ollama":    "Ollama",
		"openai":    "OpenAI",
		"anthropic": "Anthropic",
		"gemini":    "Google Gemini",
	}
	providerHints := map[string]string{
		"openai":    "https://platform.openai.com/api-keys",
		"anthropic": "https://console.anthropic.com/settings/keys",
		"gemini":    "https://aistudio.google.com/app/apikey",
	}
	for _, selectedProvider := range selectedProviders {
		fmt.Println()
		fmt.Printf("Configure models for %s:\n", providerLabels[selectedProvider])
		providerModels[selectedProvider] = selectModelsForProvider(selectedProvider, defaultModelForProvider(selectedProvider))
		if selectedProvider == "ollama" {
			providerAPIEnvs[selectedProvider] = ""
			continue
		}
		defaultEnv := defaultAPIKeyEnvForProvider(selectedProvider)
		fmt.Println()
		customEnv := promptString(fmt.Sprintf("%s API key environment variable? [%s]", providerLabels[selectedProvider], defaultEnv), defaultEnv)
		providerAPIEnvs[selectedProvider] = customEnv
		if hint := providerHints[selectedProvider]; hint != "" {
			fmt.Printf("   API keys: %s\n", hint)
		}
	}

	model := providerModels[provider].Primary
	additionalModels := providerModels[provider].Additional
	fmt.Println()
	portStr := promptString("HTTP port for Web UI? [18789]", "18789")
	port, err := strconv.Atoi(portStr)
	if err != nil || port < 1 || port > 65535 {
		fmt.Fprintf(os.Stderr, "Invalid port: %s\n", portStr)
		os.Exit(1)
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 7: Generate Configuration
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println("┌─────────────────────────────────────────────────────────────────┐")
	fmt.Println("│  STEP 7: Generating Configuration...                            │")
	fmt.Println("└─────────────────────────────────────────────────────────────────┘")
	fmt.Println()

	renderCfg := config.DefaultConfig()
	renderCfg.Tools.Packs = append([]string(nil), selectedToolPacks...)
	renderCfg.Tools.MCPPresets = append([]string(nil), selectedMCPPresets...)
	renderCfg.Tools.Exec.Profile = selectedExecProfile
	resolveToolingConfig(renderCfg)

	// Build YAML
	var sb strings.Builder

	sb.WriteString("# openclio Configuration\n")
	sb.WriteString("# Generated by 'openclio init'\n")
	sb.WriteString("# Edit freely or run 'openclio init' again to reconfigure.\n\n")

	sb.WriteString("# ── AI Model Configuration ─────────────────────────────────────────\n")
	sb.WriteString("model:\n")
	sb.WriteString(fmt.Sprintf("  provider:    %s          # AI provider\n", provider))
	sb.WriteString(fmt.Sprintf("  model:       %s  # Model to use\n", model))
	if apiKeyEnv := providerAPIEnvs[provider]; apiKeyEnv != "" {
		sb.WriteString(fmt.Sprintf("  api_key_env: %s     # Environment variable for API key\n", apiKeyEnv))
	}
	if len(additionalModels) > 0 {
		sb.WriteString("  # Additional selected models are used for delegation roles below.\n")
	}
	if len(fallbackProviders) > 0 {
		sb.WriteString(fmt.Sprintf("  fallback_providers: [%s]\n", strings.Join(fallbackProviders, ", ")))
		sb.WriteString("  fallback_models:\n")
		for _, fallbackProvider := range fallbackProviders {
			fallbackModel := providerModels[fallbackProvider].Primary
			if strings.TrimSpace(fallbackModel) == "" {
				fallbackModel = defaultModelForProvider(fallbackProvider)
			}
			sb.WriteString(fmt.Sprintf("    %s: %s\n", fallbackProvider, fallbackModel))
		}
		hasFallbackKeyEnv := false
		for _, fallbackProvider := range fallbackProviders {
			keyEnv := providerAPIEnvs[fallbackProvider]
			if keyEnv == "" {
				keyEnv = defaultAPIKeyEnvForProvider(fallbackProvider)
			}
			if keyEnv == "" {
				continue
			}
			if !hasFallbackKeyEnv {
				sb.WriteString("  fallback_api_key_env:\n")
				hasFallbackKeyEnv = true
			}
			sb.WriteString(fmt.Sprintf("    %s: %s\n", fallbackProvider, keyEnv))
		}
	}
	sb.WriteString("\n")

	sb.WriteString("# ── Gateway (HTTP + WebSocket) ─────────────────────────────────────\n")
	sb.WriteString("gateway:\n")
	sb.WriteString(fmt.Sprintf("  port: %d                    # Port for web UI\n", port))
	sb.WriteString("  bind: 127.0.0.1             # Interface to bind (127.0.0.1 = localhost only)\n")
	sb.WriteString("\n")

	sb.WriteString("# ── Memory ────────────────────────────────────────────────────────\n")
	sb.WriteString("memory:\n")
	sb.WriteString("  provider: workspace         # Local workspace memory files\n")
	sb.WriteString(fmt.Sprintf("  mode: %s              # off | standard | enhanced\n", memoryMode))
	sb.WriteString(fmt.Sprintf("  eam_serving_enabled: %t       # Backward-compatible EAM serving toggle\n", memoryMode != memoryModeOff))
	sb.WriteString("\n")

	sb.WriteString("# ── Context Engine ─────────────────────────────────────────────────\n")
	sb.WriteString("context:\n")
	sb.WriteString("  max_tokens_per_call:  8000  # Max tokens per LLM call\n")
	sb.WriteString("  history_retrieval_k:  10    # Number of relevant past messages to include\n")
	sb.WriteString("  proactive_compaction: 0.5   # Compact history when context is 50% full\n")
	sb.WriteString("\n")

	sb.WriteString("# ── Agent ──────────────────────────────────────────────────────────\n")
	sb.WriteString("agent:\n")
	sb.WriteString("  max_tool_iterations: 0\n")
	if len(additionalModels) > 0 {
		sb.WriteString("  delegation:\n")
		sb.WriteString(fmt.Sprintf("    enabled: %t\n", len(additionalModels) > 0))
		sb.WriteString("    max_parallel_sub_agents: 3\n")
		sb.WriteString(fmt.Sprintf("    sub_agent_model: %s\n", additionalModels[0]))
		synthesisModel := model
		if len(additionalModels) > 1 {
			synthesisModel = additionalModels[1]
		}
		sb.WriteString(fmt.Sprintf("    synthesis_model: %s\n", synthesisModel))
		sb.WriteString("    timeout: 90s\n")
	} else {
		sb.WriteString("  delegation:\n")
		sb.WriteString("    enabled: false\n")
		sb.WriteString("    max_parallel_sub_agents: 3\n")
		sb.WriteString("    timeout: 90s\n")
	}
	sb.WriteString("\n")

	sb.WriteString("# ── Tool Configuration ─────────────────────────────────────────────\n")
	sb.WriteString("tools:\n")
	sb.WriteString("  max_output_size: 102400     # Max bytes of tool output (100 KB)\n")
	sb.WriteString("  scrub_output: true          # Remove sensitive data from output\n")
	if len(renderCfg.Tools.Packs) > 0 {
		sb.WriteString(fmt.Sprintf("  packs: [%s]  # Product-facing tool packs\n", strings.Join(renderCfg.Tools.Packs, ", ")))
	}
	if len(renderCfg.Tools.MCPPresets) > 0 {
		sb.WriteString(fmt.Sprintf("  mcp_presets: [%s]  # Recommended MCP integrations\n", strings.Join(renderCfg.Tools.MCPPresets, ", ")))
	}
	if len(renderCfg.Tools.AllowedTools) > 0 {
		sb.WriteString(fmt.Sprintf("  allowed_tools: [%s]\n", strings.Join(renderCfg.Tools.AllowedTools, ", ")))
	}
	sb.WriteString("  exec:\n")
	sb.WriteString("    sandbox: none             # none | docker (sandbox mode for commands)\n")
	sb.WriteString("    timeout: 30s              # Max command execution time\n")
	sb.WriteString(fmt.Sprintf("    profile: %s             # safe | developer | builder | power-user\n", renderCfg.Tools.Exec.Profile))
	sb.WriteString(fmt.Sprintf("    approval_on_block: %t      # Ask before one-shot blocked CLI commands\n", renderCfg.Tools.Exec.ApprovalOnBlock))
	if len(renderCfg.Tools.Exec.AllowedCommands) > 0 {
		sb.WriteString(fmt.Sprintf("    allowed_commands: [%s]\n", strings.Join(renderCfg.Tools.Exec.AllowedCommands, ", ")))
	}
	sb.WriteString("  browser:\n")
	sb.WriteString("    enabled: true             # Enable web browser automation\n")
	sb.WriteString("    headless: true            # Run browser without visible window\n")
	sb.WriteString("    timeout: 30s              # Page load timeout\n")
	sb.WriteString("\n")

	if len(renderCfg.MCPServers) > 0 {
		sb.WriteString("# ── MCP Server Presets (disabled until you install them) ────────────\n")
		sb.WriteString("mcp_servers:\n")
		for _, srv := range renderCfg.MCPServers {
			sb.WriteString(fmt.Sprintf("  - name: %s\n", srv.Name))
			enabled := srv.Enabled == nil || *srv.Enabled
			sb.WriteString(fmt.Sprintf("    enabled: %t\n", enabled))
			sb.WriteString(fmt.Sprintf("    command: %s\n", srv.Command))
			if len(srv.Args) > 0 {
				sb.WriteString(fmt.Sprintf("    args: [%s]\n", strings.Join(srv.Args, ", ")))
			}
			if len(srv.Env) > 0 {
				sb.WriteString("    env:\n")
				keys := make([]string, 0, len(srv.Env))
				for key := range srv.Env {
					keys = append(keys, key)
				}
				sort.Strings(keys)
				for _, key := range keys {
					sb.WriteString(fmt.Sprintf("      %s: %s\n", key, srv.Env[key]))
				}
			}
		}
		sb.WriteString("\n")
	}

	sb.WriteString("# ── Channel Adapters ──────────────────────────────────────────────\n")
	sb.WriteString("channels:\n")
	sb.WriteString("  allow_all: true             # Allow all configured channels\n")
	if enableTelegram {
		sb.WriteString("  telegram:\n")
		sb.WriteString("    token_env: TELEGRAM_BOT_TOKEN\n")
	}
	if enableDiscord {
		sb.WriteString("  discord:\n")
		sb.WriteString("    token_env: DISCORD_BOT_TOKEN\n")
		sb.WriteString("    app_id_env: DISCORD_APP_ID\n")
	}
	if enableWhatsApp {
		sb.WriteString("  whatsapp:\n")
		sb.WriteString("    enabled: true\n")
		sb.WriteString("    # Uses QR login via WhatsApp Linked Devices\n")
	}
	if enableWebChat {
		sb.WriteString("  # WebChat UI enabled at http://127.0.0.1:" + portStr + "\n")
	}
	sb.WriteString("\n")

	sb.WriteString("# ── Logging ────────────────────────────────────────────────────────\n")
	sb.WriteString("logging:\n")
	sb.WriteString("  level: info                 # debug | info | warn | error\n")
	sb.WriteString("  output: ~/.openclio/openclio.log\n")

	// Write config file
	if err := os.WriteFile(configPath, []byte(sb.String()), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing config: %v\n", err)
		os.Exit(1)
	}

	_ = writeToolsReference(dataDir, renderCfg)

	// ═══════════════════════════════════════════════════════════════════════════
	// STEP 8: Customize Identity with Personality
	// ═══════════════════════════════════════════════════════════════════════════
	identityPath := filepath.Join(dataDir, "identity.md")
	if content, err := os.ReadFile(identityPath); err == nil {
		contentStr := string(content)

		// Add personality customization section at the end
		personalityNote := fmt.Sprintf("\n\n## 🎨 Personality Customization\n\n"+
			"<!-- Added during initialization -->\n\n"+
			"**Assistant Name:** %s\n\n"+
			"**Personality Style:** %s\n\n"+
			"**Key Traits:**\n", assistantName, traits[0])

		for i := 1; i < len(traits); i++ {
			personalityNote += fmt.Sprintf("- %s\n", traits[i])
		}

		personalityNote += fmt.Sprintf("\n**Response Style:** %s\n", responseStyleLabels[responseStyle])

		if userStack != "" {
			personalityNote += fmt.Sprintf("\n**User Stack:** %s\n", userStack)
		}

		// Append to the file
		contentStr = contentStr + personalityNote
		_ = os.WriteFile(identityPath, []byte(contentStr), 0600)
	}

	// ═══════════════════════════════════════════════════════════════════════════
	// Success Message
	// ═══════════════════════════════════════════════════════════════════════════
	fmt.Println()
	fmt.Println("╔══════════════════════════════════════════════════════════════════╗")
	fmt.Println("║  ✅ Setup Complete!                                              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Show what was created
	fmt.Printf("📁 Configuration directory: %s\n", dataDir)
	fmt.Println()
	fmt.Println("📄 Files created:")
	fmt.Println("   • config.yaml          — Your main configuration")
	fmt.Println("   • identity.md          — " + assistantName + "'s personality & values")
	fmt.Println("   • user.md              — Your profile and preferences")
	fmt.Println("   • memory.md            — Long-term memory structure")
	fmt.Println("   • PHILOSOPHY.md        — Project vision & principles")
	fmt.Println("   • AGENTS_REFERENCE.md  — Developer reference guide")
	fmt.Println("   • TOOLS_REFERENCE.md   — Enabled packs, MCP presets, and CLI profile")
	fmt.Println("   • skills/              — Default skill templates")
	fmt.Println()

	// Show API key instructions if needed
	if apiKeyEnv := providerAPIEnvs[provider]; apiKeyEnv != "" {
		fmt.Println("🔑 Next: Set your API key")
		fmt.Printf("   export %s=<your-api-key>\n", apiKeyEnv)
		if hint := providerHints[provider]; hint != "" {
			fmt.Printf("   Get one at: %s\n", hint)
		}
		fmt.Println()
	}
	for _, fallbackProvider := range fallbackProviders {
		keyEnv := providerAPIEnvs[fallbackProvider]
		if keyEnv == "" {
			continue
		}
		fmt.Printf("   export %s=<your-%s-api-key>\n", keyEnv, fallbackProvider)
	}
	if len(fallbackProviders) > 0 {
		fmt.Println()
	}

	// Show channel setup instructions
	if enableTelegram || enableDiscord || enableWhatsApp {
		fmt.Println("📡 Channel Setup:")
		if enableTelegram {
			fmt.Println("   Telegram: Set TELEGRAM_BOT_TOKEN=<token>")
			fmt.Println("             Get a token from @BotFather on Telegram")
		}
		if enableDiscord {
			fmt.Println("   Discord:  Set DISCORD_BOT_TOKEN=<token>")
			fmt.Println("             Get a token at https://discord.com/developers/applications")
		}
		if enableWhatsApp {
			fmt.Println("   WhatsApp: No token required (QR login)")
			fmt.Println("             Start `openclio serve` and scan the QR code in the web UI")
		}
		fmt.Println()
	}

	// Quick start guide
	fmt.Println("🚀 Quick Start:")
	fmt.Println("   openclio chat          — Start chatting in terminal")
	fmt.Println("   openclio serve         — Start web UI + all channels")
	fmt.Println()
	fmt.Printf("🧠 Primary model: %s\n", model)
	if len(additionalModels) > 0 {
		fmt.Println("🔀 Additional models:")
		for _, extra := range additionalModels {
			fmt.Printf("   • %s\n", extra)
		}
		fmt.Println("   These were assigned to delegation/subagent roles in config.yaml.")
		fmt.Println()
	}
	fmt.Printf("✨ %s is ready with full personality and memory!\n", assistantName)
	fmt.Println()
	fmt.Println("💡 Pro tip: Edit ~/.openclio/identity.md to customize your agent further.")
	fmt.Println()
}

// ── Prompt helpers ───────────────────────────────────────────────────────────

var initReader = bufio.NewReader(os.Stdin)

func promptString(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("📝 %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("📝 %s: ", label)
	}
	line, _ := initReader.ReadString('\n')
	line = strings.TrimSpace(line)
	if line == "" {
		return defaultVal
	}
	return line
}

func promptConfirm(label string, defaultYes bool) bool {
	hint := "y/N"
	if defaultYes {
		hint = "Y/n"
	}
	fmt.Printf("📝 %s [%s]: ", label, hint)
	line, _ := initReader.ReadString('\n')
	line = strings.ToLower(strings.TrimSpace(line))
	if line == "" {
		return defaultYes
	}
	return line == "y" || line == "yes"
}

func promptChoice(label string, choices []string, defaultChoice string) string {
	allowed := strings.Join(choices, "/")
	for {
		if defaultChoice != "" {
			fmt.Printf("📝 %s [%s]: ", label, defaultChoice)
		} else {
			fmt.Printf("📝 %s [%s]: ", label, allowed)
		}

		line, _ := initReader.ReadString('\n')
		line = strings.TrimSpace(line)
		if line == "" {
			if defaultChoice != "" {
				return defaultChoice
			}
			fmt.Printf("   Please choose one of: %s\n", allowed)
			continue
		}
		for _, c := range choices {
			if line == c {
				return c
			}
		}
		fmt.Printf("   Invalid choice '%s'. Enter one of: %s\n", line, allowed)
	}
}

func promptMultiline(prefix string) string {
	var lines []string
	for {
		fmt.Print(prefix)
		line, _ := initReader.ReadString('\n')
		line = strings.TrimRight(line, "\r\n")
		if strings.TrimSpace(line) == "" {
			break
		}
		lines = append(lines, line)
	}
	return strings.TrimSpace(strings.Join(lines, "\n"))
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func printOpenClioBanner(mode string) {
	lines := brand.TerminalBanner()
	width := 0
	for _, line := range lines {
		if len(line) > width {
			width = len(line)
		}
	}
	if len(mode) > width {
		width = len(mode)
	}

	fmt.Printf("%s╭%s╮%s\n", initBannerBlueText(), strings.Repeat("─", width+2), initBannerResetText())
	for _, line := range lines {
		fmt.Printf("%s│%s %s%-*s%s %s│%s\n",
			initBannerBlueText(),
			initBannerResetText(),
			initBannerBoldText()+initBannerBlueText(),
			width,
			line,
			initBannerResetText(),
			initBannerBlueText(),
			initBannerResetText(),
		)
	}
	if strings.TrimSpace(mode) != "" {
		fmt.Printf("%s│%s %-*s %s│%s\n",
			initBannerBlueText(),
			initBannerResetText(),
			width,
			"",
			initBannerBlueText(),
			initBannerResetText(),
		)
		fmt.Printf("%s│%s %s%-*s%s %s│%s\n",
			initBannerBlueText(),
			initBannerResetText(),
			initBannerBoldText(),
			width,
			centerText(mode, width),
			initBannerResetText(),
			initBannerBlueText(),
			initBannerResetText(),
		)
	}
	fmt.Printf("%s╰%s╯%s\n", initBannerBlueText(), strings.Repeat("─", width+2), initBannerResetText())
	fmt.Println()
}

func centerText(text string, width int) string {
	if len(text) >= width {
		return text
	}
	leftPad := (width - len(text)) / 2
	return strings.Repeat(" ", leftPad) + text
}
