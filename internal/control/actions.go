package control

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openclio/openclio/internal/config"
	"github.com/openclio/openclio/internal/plugin"
)

type RuntimeSync interface {
	SyncBrowserTool()
	SyncExecTool()
}

type ChannelConnector interface {
	ConnectChannel(channelType string, credentials map[string]string) error
}

type ChannelLifecycle interface {
	DisconnectChannel(channelType string) error
}

type ActionEnv struct {
	Config           *config.Config
	DataDir          string
	Allowlist        *plugin.Allowlist
	Runtime          RuntimeSync
	ChannelConnector ChannelConnector
	ChannelLifecycle ChannelLifecycle
	DeleteSession    func(id string) error
	RunCron          func(name string) error
	SetCronEnabled   func(name string, enabled bool) error
	DeleteCron       func(name string) error
	WriteTools       bool
}

type ActionResult struct {
	Updated []string `json:"updated,omitempty"`
	Note    string   `json:"note,omitempty"`
}

func SetActiveModelConfig(env ActionEnv, provider, model, baseURL string) (ActionResult, error) {
	if env.Config == nil {
		return ActionResult{}, fmt.Errorf("config not available")
	}
	previousProvider := strings.ToLower(strings.TrimSpace(env.Config.Model.Provider))
	previousAPIKeyEnv := strings.TrimSpace(env.Config.Model.APIKeyEnv)
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider == "" {
		return ActionResult{}, fmt.Errorf("provider is required")
	}
	switch provider {
	case "anthropic", "openai", "gemini", "ollama", "cohere",
		"groq", "deepseek", "mistral", "xai", "cerebras",
		"together", "fireworks", "perplexity", "openrouter",
		"kimi", "sambanova", "lambda", "lmstudio", "openai-compat":
	default:
		return ActionResult{}, fmt.Errorf("unknown provider: %s", provider)
	}
	baseURL = strings.TrimSpace(baseURL)
	if provider == "openai-compat" && baseURL == "" {
		return ActionResult{}, fmt.Errorf("base_url is required for openai-compat provider")
	}

	model = strings.TrimSpace(model)
	if model == "" {
		model = defaultModelForProvider(provider)
	}
	if provider == "openai-compat" && model == "" {
		return ActionResult{}, fmt.Errorf("model is required for openai-compat provider")
	}

	env.Config.Model.Provider = provider
	env.Config.Model.Model = model
	if providerRequiresAPIKey(provider) {
		if previousProvider == provider && previousAPIKeyEnv != "" {
			env.Config.Model.APIKeyEnv = previousAPIKeyEnv
		} else {
			env.Config.Model.APIKeyEnv = defaultAPIKeyEnvForProvider(provider)
		}
	} else {
		env.Config.Model.APIKeyEnv = ""
	}
	env.Config.Model.BaseURL = baseURL
	if provider != "openai-compat" {
		env.Config.Model.Name = ""
	}

	if err := saveActionConfig(env); err != nil {
		return ActionResult{}, err
	}

	updated := []string{"model.provider", "model.model", "model.api_key_env"}
	if baseURL != "" || provider == "openai-compat" {
		updated = append(updated, "model.base_url")
	}
	return ActionResult{
		Updated: updated,
		Note:    "model selection saved",
	}, nil
}

func SetMCPServerEnabled(env ActionEnv, name string, enabled bool) (ActionResult, error) {
	if env.Config == nil {
		return ActionResult{}, fmt.Errorf("config not available")
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return ActionResult{}, fmt.Errorf("mcp server name is required")
	}
	found := false
	for i := range env.Config.MCPServers {
		if strings.EqualFold(env.Config.MCPServers[i].Name, name) {
			val := enabled
			env.Config.MCPServers[i].Enabled = &val
			name = env.Config.MCPServers[i].Name
			found = true
			break
		}
	}
	if !found {
		return ActionResult{}, fmt.Errorf("mcp server %q not found", name)
	}
	if err := saveActionConfig(env); err != nil {
		return ActionResult{}, err
	}
	state := "disabled"
	if enabled {
		state = "enabled"
	}
	return ActionResult{
		Updated: []string{"mcp_servers." + name + ".enabled"},
		Note:    fmt.Sprintf("mcp server %s %s", name, state),
	}, nil
}

func SetBrowserEnabled(env ActionEnv, enabled bool) (ActionResult, error) {
	if env.Config == nil {
		return ActionResult{}, fmt.Errorf("config not available")
	}
	env.Config.Tools.Browser.Enabled = enabled
	if err := saveActionConfig(env); err != nil {
		return ActionResult{}, err
	}
	if env.Runtime != nil {
		env.Runtime.SyncBrowserTool()
	}
	return ActionResult{
		Updated: []string{"tools.browser.enabled", "runtime.tools.browser"},
		Note:    "browser setting saved",
	}, nil
}

func SetAllowAllMode(env ActionEnv, allowAll bool) (ActionResult, error) {
	if env.Config == nil {
		return ActionResult{}, fmt.Errorf("config not available")
	}
	env.Config.Channels.AllowAll = allowAll
	if env.Allowlist != nil {
		env.Allowlist.SetAllowAll(allowAll)
	}
	if err := saveActionConfig(env); err != nil {
		return ActionResult{}, err
	}
	return ActionResult{
		Updated: []string{"channels.allow_all"},
		Note:    "channel allowlist mode saved",
	}, nil
}

func SetExecProfile(env ActionEnv, profile string) (ActionResult, error) {
	if env.Config == nil {
		return ActionResult{}, fmt.Errorf("config not available")
	}
	profile = strings.TrimSpace(profile)
	if profile == "" {
		return ActionResult{}, fmt.Errorf("exec profile is required")
	}
	switch profile {
	case "safe", "developer", "builder", "power-user":
	default:
		return ActionResult{}, fmt.Errorf("tools.exec.profile must be one of: safe, developer, builder, power-user")
	}
	env.Config.Tools.Exec.Profile = profile
	config.ResolveToolingConfig(env.Config)
	if err := saveActionConfig(env); err != nil {
		return ActionResult{}, err
	}
	if env.Runtime != nil {
		env.Runtime.SyncExecTool()
	}
	return ActionResult{
		Updated: []string{"tools.exec.profile", "tools.tooling_expanded", "runtime.tools.exec"},
		Note:    "exec profile saved",
	}, nil
}

func DeleteSession(env ActionEnv, id string) (ActionResult, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		return ActionResult{}, fmt.Errorf("session id is required")
	}
	if env.DeleteSession == nil {
		return ActionResult{}, fmt.Errorf("session deletion is not available")
	}
	if err := env.DeleteSession(id); err != nil {
		return ActionResult{}, fmt.Errorf("delete session: %w", err)
	}
	return ActionResult{
		Updated: []string{"runtime.sessions", "storage.sessions"},
		Note:    "session deleted",
	}, nil
}

func RunCronMutation(env ActionEnv, name, action string, enabled *bool) (ActionResult, error) {
	name = strings.TrimSpace(name)
	action = strings.TrimSpace(strings.ToLower(action))
	if name == "" {
		return ActionResult{}, fmt.Errorf("cron job name is required")
	}
	switch action {
	case "run":
		if env.RunCron == nil {
			return ActionResult{}, fmt.Errorf("cron run is not available")
		}
		if err := env.RunCron(name); err != nil {
			return ActionResult{}, fmt.Errorf("run cron: %w", err)
		}
		return ActionResult{
			Updated: []string{"runtime.cron." + name},
			Note:    "cron job triggered",
		}, nil
	case "enable", "disable":
		if env.SetCronEnabled == nil {
			return ActionResult{}, fmt.Errorf("cron enable/disable is not available")
		}
		target := action == "enable"
		if enabled != nil {
			target = *enabled
		}
		if err := env.SetCronEnabled(name, target); err != nil {
			return ActionResult{}, fmt.Errorf("set cron enabled: %w", err)
		}
		state := "disabled"
		if target {
			state = "enabled"
		}
		return ActionResult{
			Updated: []string{"runtime.cron." + name, "config.cron." + name},
			Note:    fmt.Sprintf("cron job %s %s", name, state),
		}, nil
	case "delete":
		if env.DeleteCron == nil {
			return ActionResult{}, fmt.Errorf("cron delete is not available")
		}
		if err := env.DeleteCron(name); err != nil {
			return ActionResult{}, fmt.Errorf("delete cron: %w", err)
		}
		return ActionResult{
			Updated: []string{"runtime.cron", "config.cron"},
			Note:    "cron job deleted",
		}, nil
	default:
		return ActionResult{}, fmt.Errorf("action must be one of: run|enable|disable|delete")
	}
}

func RunChannelAction(env ActionEnv, name, action string, forceReconnect bool) (ActionResult, error) {
	name = strings.TrimSpace(strings.ToLower(name))
	action = strings.TrimSpace(strings.ToLower(action))
	if name == "" || action == "" {
		return ActionResult{}, fmt.Errorf("channel name and action are required")
	}
	switch name {
	case "webchat", "telegram", "discord", "whatsapp", "slack":
	default:
		return ActionResult{}, fmt.Errorf("unknown channel: %s", name)
	}

	switch action {
	case "connect":
		if env.ChannelConnector == nil {
			return ActionResult{}, fmt.Errorf("runtime channel connect is not available")
		}
		credentials := map[string]string{}
		if name == "whatsapp" && forceReconnect {
			credentials["force_reconnect"] = "true"
		}
		if err := env.ChannelConnector.ConnectChannel(name, credentials); err != nil {
			return ActionResult{}, fmt.Errorf("connect failed: %w", err)
		}
		return ActionResult{
			Updated: []string{"runtime.channels." + name},
			Note:    "connect requested",
		}, nil
	case "disconnect":
		if env.ChannelLifecycle == nil {
			return ActionResult{}, fmt.Errorf("runtime channel disconnect is not available")
		}
		if err := env.ChannelLifecycle.DisconnectChannel(name); err != nil {
			return ActionResult{}, fmt.Errorf("disconnect failed: %w", err)
		}
		return ActionResult{
			Updated: []string{"runtime.channels." + name},
			Note:    "disconnect requested",
		}, nil
	case "restart":
		if env.ChannelLifecycle == nil || env.ChannelConnector == nil {
			return ActionResult{}, fmt.Errorf("runtime channel restart is not available")
		}
		if err := env.ChannelLifecycle.DisconnectChannel(name); err != nil {
			lower := strings.ToLower(err.Error())
			if !strings.Contains(lower, "not connected") {
				return ActionResult{}, fmt.Errorf("restart failed: %w", err)
			}
		}
		credentials := map[string]string{}
		if name == "whatsapp" {
			credentials["force_reconnect"] = "true"
		}
		if err := env.ChannelConnector.ConnectChannel(name, credentials); err != nil {
			return ActionResult{}, fmt.Errorf("restart failed: %w", err)
		}
		return ActionResult{
			Updated: []string{"runtime.channels." + name},
			Note:    "restart requested",
		}, nil
	case "ping", "refresh":
		return ActionResult{
			Updated: []string{"runtime.channels." + name},
			Note:    "status refreshed",
		}, nil
	default:
		return ActionResult{}, fmt.Errorf("action must be one of: ping|refresh|restart|connect|disconnect")
	}
}

func saveActionConfig(env ActionEnv) error {
	if strings.TrimSpace(env.DataDir) == "" {
		return nil
	}
	if err := config.Save(filepath.Join(env.DataDir, "config.yaml"), env.Config); err != nil {
		return err
	}
	if env.WriteTools {
		_ = config.WriteToolsReference(env.DataDir, env.Config)
	}
	return nil
}

func providerRequiresAPIKey(provider string) bool {
	switch provider {
	case "ollama", "lmstudio":
		return false
	default:
		return true
	}
}

func defaultAPIKeyEnvForProvider(provider string) string {
	switch provider {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "gemini":
		return "GEMINI_API_KEY"
	case "cohere":
		return "COHERE_API_KEY"
	case "groq":
		return "GROQ_API_KEY"
	case "deepseek":
		return "DEEPSEEK_API_KEY"
	case "mistral":
		return "MISTRAL_API_KEY"
	case "xai":
		return "XAI_API_KEY"
	case "cerebras":
		return "CEREBRAS_API_KEY"
	case "together":
		return "TOGETHER_API_KEY"
	case "fireworks":
		return "FIREWORKS_API_KEY"
	case "perplexity":
		return "PERPLEXITY_API_KEY"
	case "openrouter":
		return "OPENROUTER_API_KEY"
	case "kimi":
		return "KIMI_API_KEY"
	case "sambanova":
		return "SAMBANOVA_API_KEY"
	case "lambda":
		return "LAMBDA_API_KEY"
	case "openai-compat":
		return "OPENAI_API_KEY"
	default:
		return ""
	}
}

func defaultModelForProvider(provider string) string {
	switch provider {
	case "anthropic":
		return "claude-sonnet-4-5"
	case "openai":
		return "gpt-4.1-mini"
	case "gemini":
		return "gemini-2.5-flash"
	case "ollama":
		return "gpt-oss:20b"
	case "cohere":
		return "command-r-plus"
	case "groq":
		return "llama-3.3-70b-versatile"
	case "deepseek":
		return "deepseek-chat"
	case "mistral":
		return "mistral-large-latest"
	case "xai":
		return "grok-4"
	case "cerebras":
		return "qwen-3-coder-480b"
	case "together":
		return "meta-llama/Meta-Llama-3.1-70B-Instruct-Turbo"
	case "fireworks":
		return "accounts/fireworks/models/llama-v3p1-70b-instruct"
	case "perplexity":
		return "sonar"
	case "openrouter":
		return "openai/gpt-4.1-mini"
	case "kimi":
		return "kimi-k2-0711-preview"
	case "sambanova":
		return "Meta-Llama-3.3-70B-Instruct"
	case "lambda":
		return "hermes-3-llama-3.1-405b-fp8"
	default:
		return ""
	}
}
