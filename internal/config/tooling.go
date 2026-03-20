package config

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

type ToolPackDef struct {
	Name               string   `json:"name" yaml:"name"`
	Description        string   `json:"description" yaml:"description"`
	Tools              []string `json:"tools,omitempty" yaml:"tools,omitempty"`
	RecommendedPresets []string `json:"recommended_presets,omitempty" yaml:"recommended_presets,omitempty"`
	DefaultExecProfile string   `json:"default_exec_profile,omitempty" yaml:"default_exec_profile,omitempty"`
}

type MCPPresetDef struct {
	Name        string            `json:"name" yaml:"name"`
	Description string            `json:"description" yaml:"description"`
	Command     string            `json:"command,omitempty" yaml:"command,omitempty"`
	Args        []string          `json:"args,omitempty" yaml:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
}

type ExecProfileDef struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Commands    []string `json:"commands,omitempty" yaml:"commands,omitempty"`
}

func ToolPackCatalog() []ToolPackDef {
	return []ToolPackDef{
		{
			Name:               "developer",
			Description:        "Coding, git, tests, repo search, delegation",
			Tools:              []string{"exec", "read_file", "write_file", "edit_file", "list_dir", "web_fetch", "web_search", "delegate", "switch_model", "memory_write", "channel_status"},
			RecommendedPresets: []string{"github", "linear"},
			DefaultExecProfile: "developer",
		},
		{
			Name:               "research",
			Description:        "Web search, browser, docs retrieval, note workflows",
			Tools:              []string{"read_file", "write_file", "edit_file", "list_dir", "web_fetch", "web_search", "browser", "memory_write", "image_analyze"},
			RecommendedPresets: []string{"notion", "database"},
			DefaultExecProfile: "safe",
		},
		{
			Name:               "builder",
			Description:        "Docker, package managers, builds, deployment CLIs",
			Tools:              []string{"exec", "read_file", "write_file", "edit_file", "list_dir", "web_fetch", "delegate", "memory_write"},
			RecommendedPresets: []string{"github", "database"},
			DefaultExecProfile: "builder",
		},
		{
			Name:               "writing",
			Description:        "Research, drafting, editing, export helpers",
			Tools:              []string{"read_file", "write_file", "edit_file", "list_dir", "web_fetch", "web_search", "browser", "memory_write", "image_generate", "image_analyze"},
			RecommendedPresets: []string{"notion"},
			DefaultExecProfile: "safe",
		},
	}
}

func MCPPresetCatalog() []MCPPresetDef {
	return []MCPPresetDef{
		{Name: "github", Description: "GitHub issues, PRs, repos", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-github"}, Env: map[string]string{"GITHUB_TOKEN": "GITHUB_TOKEN"}},
		{Name: "linear", Description: "Linear issues and projects", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-linear"}, Env: map[string]string{"LINEAR_API_KEY": "LINEAR_API_KEY"}},
		{Name: "notion", Description: "Notion docs and workspace content", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-notion"}, Env: map[string]string{"NOTION_TOKEN": "NOTION_TOKEN"}},
		{Name: "browser", Description: "Playwright MCP server (advanced; separate from the built-in browser tool)", Command: "npx", Args: []string{"-y", "@playwright/mcp@latest"}},
		{Name: "database", Description: "Database and SQL access", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-postgres"}, Env: map[string]string{"POSTGRES_URL": "POSTGRES_URL"}},
		{Name: "figma", Description: "Figma design context and assets", Command: "npx", Args: []string{"-y", "@modelcontextprotocol/server-figma"}, Env: map[string]string{"FIGMA_API_KEY": "FIGMA_API_KEY"}},
	}
}

func ExecProfileCatalog() []ExecProfileDef {
	return []ExecProfileDef{
		{Name: "safe", Description: "Read-only and low-risk local commands", Commands: []string{"cat", "pwd", "ls", "find", "rg", "grep", "sed", "awk", "head", "tail", "wc", "echo", "git"}},
		{Name: "developer", Description: "Developer workflow commands and CLIs", Commands: []string{"cat", "pwd", "ls", "find", "rg", "grep", "sed", "awk", "head", "tail", "wc", "echo", "git", "gh", "go", "npm", "pnpm", "yarn", "node", "python", "python3", "uv", "pip", "pip3", "pytest", "make"}},
		{Name: "builder", Description: "Build and packaging tools including containers", Commands: []string{"cat", "pwd", "ls", "find", "rg", "grep", "sed", "awk", "head", "tail", "wc", "echo", "git", "gh", "go", "npm", "pnpm", "yarn", "node", "python", "python3", "uv", "pip", "pip3", "pytest", "make", "docker", "docker-compose", "terraform", "kubectl"}},
		{Name: "power-user", Description: "Extended local CLI access with explicit allowlist", Commands: []string{"cat", "pwd", "ls", "find", "rg", "grep", "sed", "awk", "head", "tail", "wc", "echo", "git", "gh", "go", "npm", "pnpm", "yarn", "node", "python", "python3", "uv", "pip", "pip3", "pytest", "make", "docker", "docker-compose", "terraform", "kubectl", "ffmpeg", "sqlite3", "psql"}},
	}
}

func ResolveToolingConfig(cfg *Config) {
	if cfg == nil {
		return
	}
	packMap := mapToolPacks()
	presetMap := mapMCPPresets()
	profileMap := mapExecProfiles()

	cfg.Tools.Packs = dedupeStrings(cfg.Tools.Packs)
	cfg.Tools.MCPPresets = dedupeStrings(cfg.Tools.MCPPresets)
	cfg.Tools.AllowedTools = dedupeStrings(cfg.Tools.AllowedTools)
	cfg.Tools.Exec.AllowedCommands = dedupeStrings(cfg.Tools.Exec.AllowedCommands)

	toolSet := make(map[string]struct{})
	presetSet := make(map[string]struct{})
	profileChoice := strings.ToLower(strings.TrimSpace(cfg.Tools.Exec.Profile))
	profileExplicit := profileChoice != "" && profileChoice != "safe"

	for _, packName := range cfg.Tools.Packs {
		pack, ok := packMap[packName]
		if !ok {
			continue
		}
		for _, toolName := range pack.Tools {
			toolSet[toolName] = struct{}{}
		}
		for _, presetName := range pack.RecommendedPresets {
			presetSet[presetName] = struct{}{}
		}
		if (!profileExplicit && (profileChoice == "" || profileChoice == "safe")) && pack.DefaultExecProfile != "" {
			profileChoice = pack.DefaultExecProfile
		}
	}
	for _, toolName := range cfg.Tools.AllowedTools {
		toolSet[toolName] = struct{}{}
	}
	if len(toolSet) > 0 {
		cfg.Tools.AllowedTools = sortedSet(toolSet)
	}

	for _, presetName := range cfg.Tools.MCPPresets {
		if _, ok := presetMap[presetName]; ok {
			presetSet[presetName] = struct{}{}
		}
	}
	if len(presetSet) > 0 {
		cfg.Tools.MCPPresets = sortedSet(presetSet)
	}

	if profileChoice == "" {
		profileChoice = "safe"
	}
	cfg.Tools.Exec.Profile = profileChoice
	if profile, ok := profileMap[profileChoice]; ok {
		cmdSet := make(map[string]struct{})
		for _, cmd := range profile.Commands {
			cmdSet[cmd] = struct{}{}
		}
		for _, cmd := range cfg.Tools.Exec.AllowedCommands {
			cmdSet[cmd] = struct{}{}
		}
		cfg.Tools.Exec.AllowedCommands = sortedSet(cmdSet)
	}
	if (profileChoice == "developer" || profileChoice == "builder") && !cfg.Tools.Exec.ApprovalOnBlock {
		cfg.Tools.Exec.ApprovalOnBlock = true
	}
	mergePresetServers(cfg)
}

func WriteToolsReference(dataDir string, cfg *Config) error {
	if dataDir == "" || cfg == nil {
		return nil
	}
	path := filepath.Join(dataDir, "TOOLS_REFERENCE.md")
	return os.WriteFile(path, []byte(buildToolsReferenceMarkdown(cfg)), 0600)
}

func mapToolPacks() map[string]ToolPackDef {
	out := make(map[string]ToolPackDef)
	for _, p := range ToolPackCatalog() {
		out[p.Name] = p
	}
	return out
}

func mapMCPPresets() map[string]MCPPresetDef {
	out := make(map[string]MCPPresetDef)
	for _, p := range MCPPresetCatalog() {
		out[p.Name] = p
	}
	return out
}

func mapExecProfiles() map[string]ExecProfileDef {
	out := make(map[string]ExecProfileDef)
	for _, p := range ExecProfileCatalog() {
		out[p.Name] = p
	}
	return out
}

func mergePresetServers(cfg *Config) {
	if cfg == nil || len(cfg.Tools.MCPPresets) == 0 {
		return
	}
	presetMap := mapMCPPresets()
	existing := make(map[string]struct{}, len(cfg.MCPServers))
	for _, server := range cfg.MCPServers {
		existing[strings.ToLower(strings.TrimSpace(server.Name))] = struct{}{}
	}
	for _, presetName := range cfg.Tools.MCPPresets {
		if _, ok := existing[presetName]; ok {
			continue
		}
		preset, ok := presetMap[presetName]
		if !ok {
			continue
		}
		cfg.MCPServers = append(cfg.MCPServers, MCPServerConfig{
			Enabled: boolPtr(false),
			Name:    preset.Name,
			Command: preset.Command,
			Args:    slices.Clone(preset.Args),
			Env:     mapsClone(preset.Env),
		})
	}
}

func buildToolsReferenceMarkdown(cfg *Config) string {
	if cfg == nil {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("# OpenClio Tools Reference\n\n")
	sb.WriteString("This file summarizes the product-facing tool capabilities enabled for this workspace.\n\n")
	sb.WriteString("## Active Packs\n\n")
	if len(cfg.Tools.Packs) == 0 {
		sb.WriteString("- none\n")
	} else {
		for _, pack := range cfg.Tools.Packs {
			sb.WriteString(fmt.Sprintf("- %s\n", pack))
		}
	}
	sb.WriteString("\n## Built-in Tools\n\n")
	if len(cfg.Tools.AllowedTools) == 0 {
		sb.WriteString("- all built-in tools allowed\n")
	} else {
		for _, toolName := range cfg.Tools.AllowedTools {
			sb.WriteString(fmt.Sprintf("- %s\n", toolName))
		}
	}
	sb.WriteString("\n## MCP Presets\n\n")
	if len(cfg.Tools.MCPPresets) == 0 {
		sb.WriteString("- none\n")
	} else {
		for _, preset := range cfg.Tools.MCPPresets {
			sb.WriteString(fmt.Sprintf("- %s\n", preset))
		}
	}
	sb.WriteString("\n## Local CLI Profile\n\n")
	sb.WriteString(fmt.Sprintf("- profile: %s\n", cfg.Tools.Exec.Profile))
	if len(cfg.Tools.Exec.AllowedCommands) > 0 {
		sb.WriteString("- allowed commands:\n")
		for _, cmd := range cfg.Tools.Exec.AllowedCommands {
			sb.WriteString(fmt.Sprintf("  - %s\n", cmd))
		}
	}
	if cfg.Tools.Exec.ApprovalOnBlock {
		sb.WriteString("- blocked commands can request one-shot approval in interactive sessions\n")
	}
	return sb.String()
}

func dedupeStrings(values []string) []string {
	if len(values) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func sortedSet(values map[string]struct{}) []string {
	out := make([]string, 0, len(values))
	for value := range values {
		out = append(out, value)
	}
	sort.Strings(out)
	return out
}

func mapsClone(in map[string]string) map[string]string {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string]string, len(in))
	for k, v := range in {
		out[k] = v
	}
	return out
}

func boolPtr(v bool) *bool {
	return &v
}
