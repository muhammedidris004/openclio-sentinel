package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// ConfigModelPersister persists model changes to the config file.
// Implement this in main.go and pass it to NewSwitchModelTool so that
// switching models survives a restart.
type ConfigModelPersister interface {
	PersistModel(provider, model string) error
}

// SwitchModelTool lets users change the active LLM provider and model
// mid-conversation via an injected runtime ProviderSwitcher.
type SwitchModelTool struct {
	switcher  ProviderSwitcher
	persister ConfigModelPersister // optional — persists change to config.yaml
	providers map[string]struct{}
}

// NewSwitchModelTool creates a switch_model tool.
func NewSwitchModelTool(switcher ProviderSwitcher) *SwitchModelTool {
	return &SwitchModelTool{
		switcher: switcher,
		providers: map[string]struct{}{
			"anthropic": {},
			"openai":    {},
			"gemini":    {},
			"ollama":    {},
			"groq":      {},
			"deepseek":  {},
		},
	}
}

// SetConfigPersister wires config persistence so model changes survive restart.
func (t *SwitchModelTool) SetConfigPersister(p ConfigModelPersister) {
	t.persister = p
}

func (t *SwitchModelTool) Name() string { return "switch_model" }

func (t *SwitchModelTool) Description() string {
	return "Switch the active AI provider and model. Call this when the user asks to change the model, use a different AI, switch to GPT, Claude, Gemini, or a local model."
}

func (t *SwitchModelTool) Schema() json.RawMessage {
	return json.RawMessage(`{
		"type": "object",
		"properties": {
			"provider": {
				"type": "string",
				"enum": ["anthropic", "openai", "gemini", "ollama", "groq", "deepseek"],
				"description": "Provider name"
			},
			"model": {
				"type": "string",
				"description": "Exact model name. Anthropic: claude-opus-4-5, claude-sonnet-4-6, claude-haiku-4-5-20251001, claude-3-5-sonnet-20241022, claude-3-5-haiku-20241022. OpenAI: gpt-4o, gpt-4o-mini, o3-mini. Gemini: gemini-2.0-flash, gemini-1.5-pro. Ollama: llama3.2, mistral. Groq: llama-3.3-70b-versatile. DeepSeek: deepseek-chat, deepseek-reasoner."
			}
		},
		"required": ["provider", "model"]
	}`)
}

type switchModelParams struct {
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

func (t *SwitchModelTool) Execute(ctx context.Context, params json.RawMessage) (string, error) {
	_ = ctx
	if t.switcher == nil {
		return "", fmt.Errorf("switch_model is unavailable")
	}

	var p switchModelParams
	if err := json.Unmarshal(params, &p); err != nil {
		return "", fmt.Errorf("invalid params: %w", err)
	}

	provider := strings.ToLower(strings.TrimSpace(p.Provider))
	model := strings.TrimSpace(p.Model)
	if provider == "" {
		return "", fmt.Errorf("provider is required")
	}
	if model == "" {
		return "", fmt.Errorf("model is required")
	}
	if _, ok := t.providers[provider]; !ok {
		return "", fmt.Errorf("unsupported provider %q (supported: anthropic, openai, gemini, ollama, groq, deepseek)", provider)
	}

	if err := t.switcher.SwitchProvider(provider, model); err != nil {
		return "", fmt.Errorf("switching model failed: %w", err)
	}

	// Persist to config so the model survives a restart.
	if t.persister != nil {
		if err := t.persister.PersistModel(provider, model); err != nil {
			// Non-fatal — model is live for this session even if save fails.
			return fmt.Sprintf("Switched to %s/%s (warning: could not save to config: %v)", provider, model, err), nil
		}
		return fmt.Sprintf("Switched to %s/%s and saved to config.", provider, model), nil
	}

	return fmt.Sprintf("Switched active model to %s/%s (session only — not saved to config).", provider, model), nil
}
