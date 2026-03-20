package control

// Surface identifies where a control is exposed.
type Surface string

const (
	SurfaceCLI Surface = "cli"
	SurfaceUI  Surface = "ui"
)

// Command describes one user-facing control command.
type Command struct {
	ID          string    `json:"id"`
	Group       string    `json:"group"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Slash       string    `json:"slash,omitempty"`
	Surfaces    []Surface `json:"surfaces,omitempty"`
}

// Group describes a command family.
type Group struct {
	ID          string    `json:"id"`
	Label       string    `json:"label"`
	Description string    `json:"description"`
	Surfaces    []Surface `json:"surfaces,omitempty"`
	Commands    []Command `json:"commands,omitempty"`
}

// Catalog returns the current shared control catalog used by CLI and UI.
func Catalog() []Group {
	return []Group{
		{
			ID:          "status",
			Label:       "Status",
			Description: "Overall runtime, setup, and service status.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "status.summary",
					Group:       "status",
					Label:       "Status summary",
					Description: "Show overall runtime status, provider/model, uptime, and service counts.",
					Slash:       "/status",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "auth",
			Label:       "Auth",
			Description: "Authentication and OAuth readiness state.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "auth.summary",
					Group:       "auth",
					Label:       "Auth summary",
					Description: "Show OpenAI OAuth configuration and sign-in status.",
					Slash:       "/auth",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "plugins",
			Label:       "Plugins",
			Description: "Runtime plugin and channel adapter status.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "plugins.summary",
					Group:       "plugins",
					Label:       "Plugin summary",
					Description: "Show registered plugins/adapters and their runtime health.",
					Slash:       "/plugins",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "sessions",
			Label:       "Sessions",
			Description: "Session listing and lifecycle controls.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "sessions.summary",
					Group:       "sessions",
					Label:       "Session summary",
					Description: "Show recent sessions and total session count.",
					Slash:       "/sessions",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "sessions.delete",
					Group:       "sessions",
					Label:       "Delete session",
					Description: "Delete one session by id or use current for the active session.",
					Slash:       "/sessions delete <id|current>",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "cron",
			Label:       "Cron",
			Description: "Cron job summaries and runtime actions.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "cron.summary",
					Group:       "cron",
					Label:       "Cron summary",
					Description: "Show current cron jobs and recent failures.",
					Slash:       "/cron",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "cron.run",
					Group:       "cron",
					Label:       "Run cron job",
					Description: "Trigger a cron job immediately.",
					Surfaces:    []Surface{SurfaceUI},
				},
				{
					ID:          "cron.toggle",
					Group:       "cron",
					Label:       "Enable or disable cron job",
					Description: "Toggle a persistent cron job.",
					Surfaces:    []Surface{SurfaceUI},
				},
				{
					ID:          "cron.delete",
					Group:       "cron",
					Label:       "Delete cron job",
					Description: "Delete a persistent cron job.",
					Surfaces:    []Surface{SurfaceUI},
				},
			},
		},
		{
			ID:          "doctor",
			Label:       "Doctor",
			Description: "Health and readiness checks for config, providers, channels, tools, and MCP.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "doctor.summary",
					Group:       "doctor",
					Label:       "Doctor summary",
					Description: "Show a consolidated health and readiness report.",
					Slash:       "/doctor",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "doctor.provider",
					Group:       "doctor",
					Label:       "Provider check",
					Description: "Validate provider configuration, model selection, and required credentials.",
					Slash:       "/doctor provider",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "doctor.channels",
					Group:       "doctor",
					Label:       "Channels check",
					Description: "Show configured channels and whether required credentials are present.",
					Slash:       "/doctor channels",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "doctor.tools",
					Group:       "doctor",
					Label:       "Tools check",
					Description: "Validate tool packs, browser, exec profile, and key local binaries.",
					Slash:       "/doctor tools",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "doctor.mcp",
					Group:       "doctor",
					Label:       "MCP check",
					Description: "Show MCP preset and server readiness, including missing executables.",
					Slash:       "/doctor mcp",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "models",
			Label:       "Models",
			Description: "Provider, model, fallback, and delegation model controls.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "models.summary",
					Group:       "models",
					Label:       "Model summary",
					Description: "Show the active provider, model, fallback providers, and delegation models.",
					Slash:       "/models",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "models.fallback",
					Group:       "models",
					Label:       "Fallback models",
					Description: "Show fallback providers and their configured fallback models.",
					Slash:       "/models fallback",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "models.delegation",
					Group:       "models",
					Label:       "Delegation models",
					Description: "Show sub-agent and synthesis model routing details.",
					Slash:       "/models delegation",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "models.switch",
					Group:       "models",
					Label:       "Switch model",
					Description: "Switch the active provider/model pair and persist it to config.",
					Slash:       "/models switch <provider> <model>",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "channels",
			Label:       "Channels",
			Description: "Channel connect, disconnect, approval, and runtime status controls.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "channels.summary",
					Group:       "channels",
					Label:       "Channel summary",
					Description: "Show configured channels and allowlist mode.",
					Slash:       "/channels",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "channels.allowlist",
					Group:       "channels",
					Label:       "Allowlist summary",
					Description: "Show whether channels are running in allow-all or strict mode.",
					Slash:       "/channels allowlist",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "channels.allowlist.open",
					Group:       "channels",
					Label:       "Set allowlist open",
					Description: "Switch channels into allow-all mode.",
					Slash:       "/approvals open",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "channels.allowlist.strict",
					Group:       "channels",
					Label:       "Set allowlist strict",
					Description: "Require approval/allowlist for incoming channel senders.",
					Slash:       "/approvals strict",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "channels.connect",
					Group:       "channels",
					Label:       "Connect channel",
					Description: "Connect a runtime channel adapter.",
					Surfaces:    []Surface{SurfaceUI},
				},
				{
					ID:          "channels.disconnect",
					Group:       "channels",
					Label:       "Disconnect channel",
					Description: "Disconnect a runtime channel adapter.",
					Surfaces:    []Surface{SurfaceUI},
				},
				{
					ID:          "channels.restart",
					Group:       "channels",
					Label:       "Restart channel",
					Description: "Restart a runtime channel adapter.",
					Surfaces:    []Surface{SurfaceUI},
				},
			},
		},
		{
			ID:          "tools",
			Label:       "Tools",
			Description: "Tool packs, exec profile, allowed tools, and browser controls.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "tools.summary",
					Group:       "tools",
					Label:       "Tooling summary",
					Description: "Show packs, browser state, exec profile, and configured MCP servers.",
					Slash:       "/tools",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "tools.exec",
					Group:       "tools",
					Label:       "Exec summary",
					Description: "Show exec profile, allowed commands, and approval-on-block behavior.",
					Slash:       "/tools exec",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "tools.exec.profile",
					Group:       "tools",
					Label:       "Set exec profile",
					Description: "Set the active local CLI execution profile.",
					Slash:       "/tools profile <safe|developer|builder|power-user>",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "tools.browser",
					Group:       "tools",
					Label:       "Browser summary",
					Description: "Show browser automation status and binary path selection.",
					Slash:       "/tools browser",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "tools.mcp",
					Group:       "tools",
					Label:       "MCP summary",
					Description: "Show MCP presets and configured server stubs.",
					Slash:       "/tools mcp",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "tools.mcp.enable",
					Group:       "tools",
					Label:       "Enable MCP server",
					Description: "Enable a configured MCP server stub in config.",
					Slash:       "/tools mcp enable <name>",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "tools.mcp.disable",
					Group:       "tools",
					Label:       "Disable MCP server",
					Description: "Disable a configured MCP server stub in config.",
					Slash:       "/tools mcp disable <name>",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "mcp",
			Label:       "MCP",
			Description: "MCP presets, configured servers, enabled state, and health.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
		},
		{
			ID:          "browser",
			Label:       "Browser",
			Description: "Browser automation status, binary selection, and headless mode.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "browser.summary",
					Group:       "browser",
					Label:       "Browser summary",
					Description: "Show browser automation status, selected binary, and headless mode.",
					Slash:       "/browser",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "browser.binary",
					Group:       "browser",
					Label:       "Browser binary",
					Description: "Show the resolved browser binary path and whether it is available.",
					Slash:       "/browser binary",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "browser.on",
					Group:       "browser",
					Label:       "Enable browser",
					Description: "Enable browser automation and browser-backed web fetching.",
					Slash:       "/browser on",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "browser.off",
					Group:       "browser",
					Label:       "Disable browser",
					Description: "Disable browser automation and browser-backed web fetching.",
					Slash:       "/browser off",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "logs",
			Label:       "Logs",
			Description: "Gateway and runtime log inspection.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "logs.summary",
					Group:       "logs",
					Label:       "Logs summary",
					Description: "Show configured logging output, file readiness, and recent file metadata.",
					Slash:       "/logs",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "logs.error",
					Group:       "logs",
					Label:       "Error log guidance",
					Description: "Show where to inspect recent error logs through the gateway log surface.",
					Slash:       "/logs error",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
		{
			ID:          "approvals",
			Label:       "Approvals",
			Description: "Approval and allowlist workflows for sensitive actions.",
			Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
			Commands: []Command{
				{
					ID:          "approvals.summary",
					Group:       "approvals",
					Label:       "Approvals summary",
					Description: "Show channel allowlist mode, approved sender count, and approval workflow status.",
					Slash:       "/approvals",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "approvals.allowlist",
					Group:       "approvals",
					Label:       "Allowlist summary",
					Description: "Show approved sender identities and whether channels are strict or open.",
					Slash:       "/approvals allowlist",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "approvals.open",
					Group:       "approvals",
					Label:       "Open approvals mode",
					Description: "Allow all senders without approval.",
					Slash:       "/approvals open",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
				{
					ID:          "approvals.strict",
					Group:       "approvals",
					Label:       "Strict approvals mode",
					Description: "Require approval/allowlist for channel senders.",
					Slash:       "/approvals strict",
					Surfaces:    []Surface{SurfaceCLI, SurfaceUI},
				},
			},
		},
	}
}
