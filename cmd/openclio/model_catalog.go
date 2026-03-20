package main

import (
	"fmt"
	"strings"
)

type providerModelOption struct {
	ID    string
	Label string
}

type providerModelCatalog struct {
	ProviderLabel string
	OfficialURL   string
	Options       []providerModelOption
}

func catalogForProvider(provider string) providerModelCatalog {
	switch strings.ToLower(strings.TrimSpace(provider)) {
	case "ollama":
		return providerModelCatalog{
			ProviderLabel: "Ollama",
			OfficialURL:   "https://ollama.com/library",
			Options: []providerModelOption{
				{ID: "gpt-oss:20b", Label: "GPT OSS 20B — strong local default"},
				{ID: "gpt-oss:120b", Label: "GPT OSS 120B — highest-quality open-weight"},
				{ID: "llama3.1", Label: "Llama 3.1 — versatile general-purpose family"},
				{ID: "llama3.2-vision", Label: "Llama 3.2 Vision — multimodal local option"},
				{ID: "qwen3", Label: "Qwen3 — strong general reasoning"},
				{ID: "qwen3-coder", Label: "Qwen3 Coder — code and agent workflows"},
				{ID: "qwen2.5", Label: "Qwen2.5 — broad multilingual baseline"},
				{ID: "qwen2.5-coder", Label: "Qwen2.5 Coder — older stable coding family"},
				{ID: "deepseek-r1", Label: "DeepSeek R1 — reasoning-focused"},
				{ID: "deepseek-v3.2", Label: "DeepSeek V3.2 — strong agentic performance"},
				{ID: "gemma3", Label: "Gemma 3 — lightweight local option"},
				{ID: "mistral", Label: "Mistral — compact general-purpose family"},
				{ID: "mistral-nemo", Label: "Mistral Nemo — efficient long-context option"},
				{ID: "phi4", Label: "Phi-4 — small but capable local model"},
				{ID: "olmo2", Label: "OLMo 2 — open research-oriented model"},
				{ID: "llava", Label: "LLaVA — multimodal vision-language model"},
			},
		}
	case "openai":
		return providerModelCatalog{
			ProviderLabel: "OpenAI",
			OfficialURL:   "https://platform.openai.com/docs/models",
			Options: []providerModelOption{
				{ID: "gpt-5.4", Label: "GPT-5.4 — latest flagship model"},
				{ID: "gpt-5", Label: "GPT-5 — flagship general model"},
				{ID: "gpt-5-pro", Label: "GPT-5 Pro — highest reasoning tier"},
				{ID: "gpt-5-mini", Label: "GPT-5 mini — faster, cheaper GPT-5"},
				{ID: "gpt-5-nano", Label: "GPT-5 nano — smallest GPT-5 tier"},
				{ID: "gpt-4.1", Label: "GPT-4.1 — smartest non-reasoning model"},
				{ID: "gpt-4.1-mini", Label: "GPT-4.1 mini — fast balanced default"},
				{ID: "gpt-4.1-nano", Label: "GPT-4.1 nano — lowest-cost chat tier"},
				{ID: "gpt-4o", Label: "GPT-4o — multimodal flagship"},
				{ID: "gpt-4o-mini", Label: "GPT-4o mini — affordable multimodal"},
				{ID: "o3", Label: "o3 — complex reasoning model"},
				{ID: "o3-pro", Label: "o3-pro — higher-reliability reasoning"},
				{ID: "o4-mini", Label: "o4-mini — compact reasoning model"},
			},
		}
	case "anthropic":
		return providerModelCatalog{
			ProviderLabel: "Anthropic",
			OfficialURL:   "https://docs.anthropic.com/en/docs/about-claude/models",
			Options: []providerModelOption{
				{ID: "claude-opus-4-6", Label: "Claude Opus 4.6 — latest premium tier"},
				{ID: "claude-sonnet-4-6", Label: "Claude Sonnet 4.6 — latest balanced tier"},
				{ID: "claude-opus-4-1", Label: "Claude Opus 4.1 — current top-end capability"},
				{ID: "claude-sonnet-4-5", Label: "Claude Sonnet 4.5 — strongest everyday default"},
				{ID: "claude-haiku-4-5", Label: "Claude Haiku 4.5 — latest fast alias"},
				{ID: "claude-haiku-4-5-20251001", Label: "Claude Haiku 4.5 (2025-10-01) — pinned version"},
				{ID: "claude-opus-4-20250514", Label: "Claude Opus 4 — earlier pinned snapshot"},
				{ID: "claude-sonnet-4-20250514", Label: "Claude Sonnet 4 — earlier pinned snapshot"},
				{ID: "claude-3-7-sonnet-20250219", Label: "Claude Sonnet 3.7 — strong reasoning"},
				{ID: "claude-3-5-sonnet-20241022", Label: "Claude Sonnet 3.5 — older stable tier"},
				{ID: "claude-3-5-haiku-20241022", Label: "Claude Haiku 3.5 — fast and cheap"},
				{ID: "claude-3-haiku-20240307", Label: "Claude Haiku 3 — legacy fastest tier"},
			},
		}
	case "gemini":
		return providerModelCatalog{
			ProviderLabel: "Google Gemini",
			OfficialURL:   "https://ai.google.dev/gemini-api/docs/models",
			Options: []providerModelOption{
				{ID: "gemini-3.1-pro", Label: "Gemini 3.1 Pro — newer high-end tier"},
				{ID: "gemini-3.1-flash", Label: "Gemini 3.1 Flash — newer fast tier"},
				{ID: "gemini-3.1-flash-lite", Label: "Gemini 3.1 Flash-Lite — newer low-cost tier"},
				{ID: "gemini-3-pro-preview", Label: "Gemini 3 Pro Preview — latest top-end reasoning"},
				{ID: "gemini-3-flash-preview", Label: "Gemini 3 Flash Preview — fast new generation"},
				{ID: "gemini-2.5-pro", Label: "Gemini 2.5 Pro — strongest reasoning/modeling"},
				{ID: "gemini-2.5-flash", Label: "Gemini 2.5 Flash — balanced default"},
				{ID: "gemini-2.5-flash-lite", Label: "Gemini 2.5 Flash-Lite — cheaper tier"},
				{ID: "gemini-1.5-pro", Label: "Gemini 1.5 Pro — older long-context tier"},
				{ID: "gemini-1.5-flash", Label: "Gemini 1.5 Flash — older fast tier"},
				{ID: "gemini-flash-latest", Label: "Gemini Flash latest alias"},
				{ID: "gemini-pro-latest", Label: "Gemini Pro latest alias"},
			},
		}
	default:
		return providerModelCatalog{}
	}
}

type providerModelSelection struct {
	Primary    string
	Additional []string
}

func selectModelsForProvider(provider, fallback string) providerModelSelection {
	catalog := catalogForProvider(provider)
	if len(catalog.Options) == 0 {
		model := promptString(fmt.Sprintf("Model name? [%s]", fallback), fallback)
		return providerModelSelection{Primary: model}
	}

	fmt.Println()
	fmt.Printf("Available %s models:\n", catalog.ProviderLabel)
	fmt.Println("Use ↑/↓ to move, Space to toggle, Enter to confirm.")
	fmt.Printf("Official models page: %s\n", catalog.OfficialURL)
	fmt.Println()

	items := make([]checkboxItem, 0, len(catalog.Options)+1)
	defaultIndex := 0
	for i, option := range catalog.Options {
		items = append(items, checkboxItem{
			Label:   option.Label,
			Hint:    option.ID,
			Value:   option.ID,
			Checked: option.ID == fallback,
		})
		if option.ID == fallback {
			defaultIndex = i
		}
	}
	items = append(items, checkboxItem{
		Label: "Custom model name",
		Hint:  "Enter any provider-supported model id manually",
		Value: "__custom__",
	})

	selected := promptCheckboxMulti("Choose one or more models", items, defaultIndex)
	if len(selected) == 0 {
		return providerModelSelection{Primary: fallback}
	}

	customChosen := false
	models := make([]string, 0, len(selected))
	for _, value := range selected {
		if value == "__custom__" {
			customChosen = true
			continue
		}
		models = append(models, value)
	}
	if customChosen {
		customModel := promptString(fmt.Sprintf("Custom model name? [%s]", fallback), fallback)
		if strings.TrimSpace(customModel) != "" {
			models = append(models, strings.TrimSpace(customModel))
		}
	}
	if len(models) == 0 {
		models = append(models, fallback)
	}

	return providerModelSelection{
		Primary:    models[0],
		Additional: models[1:],
	}
}
