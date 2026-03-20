package cli

import (
	"sort"
	"strings"

	"github.com/chzyer/readline"
	"github.com/openclio/openclio/internal/control"
)

type slashCommandDef struct {
	Command     string
	Group       string
	Label       string
	Description string
}

func slashCommandCatalog() []slashCommandDef {
	seen := make(map[string]struct{})
	var defs []slashCommandDef

	add := func(cmd, group, label, description string) {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			return
		}
		if _, ok := seen[cmd]; ok {
			return
		}
		seen[cmd] = struct{}{}
		defs = append(defs, slashCommandDef{
			Command:     cmd,
			Group:       group,
			Label:       label,
			Description: description,
		})
	}

	for _, group := range control.Catalog() {
		for _, cmd := range group.Commands {
			if !hasSurface(cmd.Surfaces, control.SurfaceCLI) {
				continue
			}
			add(cmd.Slash, group.Label, cmd.Label, cmd.Description)
		}
	}

	base := []slashCommandDef{
		{Command: "/", Group: "CLI", Label: "Open command palette", Description: "Open the interactive slash-command palette."},
		{Command: "/help", Group: "CLI", Label: "Show help", Description: "Show the built-in CLI command reference."},
		{Command: "/sessions", Group: "CLI", Label: "List recent sessions", Description: "List recent chat sessions from local storage."},
		{Command: "/new", Group: "CLI", Label: "Start a new session", Description: "Create a fresh session without clearing the database."},
		{Command: "/clear", Group: "CLI", Label: "Clear current session", Description: "Delete the active session and start fresh."},
		{Command: "/reset", Group: "CLI", Label: "Clear current session", Description: "Delete the active session and start fresh."},
		{Command: "/compact", Group: "CLI", Label: "Compact old history", Description: "Run context compaction over the current session."},
		{Command: "/history", Group: "CLI", Label: "Show session history", Description: "Print recent messages from the active session."},
		{Command: "/usage", Group: "CLI", Label: "Show token usage", Description: "Show current session token usage totals."},
		{Command: "/cost", Group: "CLI", Label: "Show cost summary", Description: "Show cost-tracker summaries when enabled."},
		{Command: "/model", Group: "CLI", Label: "Show active model", Description: "Show the current provider and active model."},
		{Command: "/models switch", Group: "CLI", Label: "Switch active model", Description: "Interactively choose a provider and model, then switch to it."},
		{Command: "/memory", Group: "CLI", Label: "Show workspace memory", Description: "Show persistent workspace memory content."},
		{Command: "/beliefs", Group: "CLI", Label: "Show active beliefs", Description: "Show current EAM beliefs when available."},
		{Command: "/switch", Group: "CLI", Label: "Switch session", Description: "Switch the active CLI session."},
		{Command: "/skill", Group: "CLI", Label: "Show skills", Description: "List loaded skills and workspace context."},
		{Command: "/skills", Group: "CLI", Label: "Show skills", Description: "List loaded skills and workspace context."},
		{Command: "/browser on", Group: "CLI", Label: "Enable browser automation", Description: "Turn browser automation on for the current config."},
		{Command: "/browser off", Group: "CLI", Label: "Disable browser automation", Description: "Turn browser automation off for the current config."},
		{Command: "/approvals open", Group: "CLI", Label: "Open approvals mode", Description: "Allow all channels without approval gating."},
		{Command: "/approvals strict", Group: "CLI", Label: "Strict approvals mode", Description: "Require explicit approvals / allowlist mode."},
		{Command: "/channels open", Group: "CLI", Label: "Open channel mode", Description: "Set channels to allow-all mode."},
		{Command: "/channels strict", Group: "CLI", Label: "Strict channel mode", Description: "Set channels to strict allowlist mode."},
		{Command: "/tools profile", Group: "CLI", Label: "Set exec profile", Description: "Interactively choose the local command execution profile."},
		{Command: "/tools mcp enable", Group: "CLI", Label: "Enable MCP server", Description: "Interactively enable a configured MCP server."},
		{Command: "/tools mcp disable", Group: "CLI", Label: "Disable MCP server", Description: "Interactively disable a configured MCP server."},
		{Command: "/sessions delete", Group: "CLI", Label: "Delete session", Description: "Interactively choose a saved session to delete."},
		{Command: "/debug context", Group: "CLI", Label: "Show assembled context", Description: "Print the last assembled context tiers."},
		{Command: "/debug tokens", Group: "CLI", Label: "Show token breakdown", Description: "Print token-budget and prompt breakdown details."},
	}
	for _, cmd := range base {
		add(cmd.Command, cmd.Group, cmd.Label, cmd.Description)
	}

	sort.Slice(defs, func(i, j int) bool {
		if defs[i].Group == defs[j].Group {
			return defs[i].Command < defs[j].Command
		}
		return defs[i].Group < defs[j].Group
	})
	return defs
}

func hasSurface(surfaces []control.Surface, target control.Surface) bool {
	for _, surface := range surfaces {
		if surface == target {
			return true
		}
	}
	return false
}

func slashCompleter() *readline.PrefixCompleter {
	items := slashCommandCatalog()
	pcItems := make([]readline.PrefixCompleterInterface, 0, len(items))
	for _, item := range items {
		pcItems = append(pcItems, readline.PcItem(item.Command))
	}
	return readline.NewPrefixCompleter(pcItems...)
}

func slashPaletteItems() []promptSelectItem {
	defs := slashCommandCatalog()
	items := make([]promptSelectItem, 0, len(defs))
	for _, def := range defs {
		if def.Command == "/" {
			continue
		}
		hint := def.Group
		if strings.TrimSpace(def.Label) != "" {
			hint = def.Group + " • " + def.Label
		}
		items = append(items, promptSelectItem{
			Label:       def.Command,
			Hint:        hint,
			Value:       def.Command,
			Group:       def.Group,
			Description: def.Description,
		})
	}
	return items
}
