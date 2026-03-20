package tools

import (
	"os"
	"strings"
	"sync"
)

var (
	runtimeToolPolicyMu sync.RWMutex
	runtimeAllowedTools map[string]struct{}
)

// SetAllowedTools configures a runtime allowlist for tool execution.
// An empty list means "no config-driven restriction". OPENCLIO_ALLOWED_TOOLS
// still takes precedence when set.
func SetAllowedTools(toolNames []string) {
	runtimeToolPolicyMu.Lock()
	defer runtimeToolPolicyMu.Unlock()
	if len(toolNames) == 0 {
		runtimeAllowedTools = nil
		return
	}
	runtimeAllowedTools = make(map[string]struct{}, len(toolNames))
	for _, toolName := range toolNames {
		toolName = strings.TrimSpace(toolName)
		if toolName == "" {
			continue
		}
		runtimeAllowedTools[toolName] = struct{}{}
	}
}

// IsToolAllowed checks runtime permission gates for tool execution.
// If OPENCLIO_ALLOWED_TOOLS is unset, allow all tools. Otherwise it's a
// comma-separated allowlist (tool names).
func IsToolAllowed(toolName string) bool {
	raw := os.Getenv("OPENCLIO_ALLOWED_TOOLS")
	if strings.TrimSpace(raw) == "" {
		return isRuntimeToolAllowed(toolName)
	}
	parts := strings.Split(raw, ",")
	for _, p := range parts {
		if strings.TrimSpace(p) == toolName {
			return true
		}
	}
	return false
}

func isRuntimeToolAllowed(toolName string) bool {
	runtimeToolPolicyMu.RLock()
	defer runtimeToolPolicyMu.RUnlock()
	if len(runtimeAllowedTools) == 0 {
		return true
	}
	_, ok := runtimeAllowedTools[toolName]
	return ok
}

// RuntimeAllowedTools returns the active config-driven tool allowlist.
func RuntimeAllowedTools() []string {
	runtimeToolPolicyMu.RLock()
	defer runtimeToolPolicyMu.RUnlock()
	if len(runtimeAllowedTools) == 0 {
		return nil
	}
	out := make([]string, 0, len(runtimeAllowedTools))
	for name := range runtimeAllowedTools {
		out = append(out, name)
	}
	return out
}
