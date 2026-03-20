package main

import (
	"strings"

	"github.com/openclio/openclio/internal/config"
	memoryserving "github.com/openclio/openclio/internal/memory/serving"
)

const (
	memoryModeOff      = "off"
	memoryModeStandard = "standard"
	memoryModeEnhanced = "enhanced"
)

func effectiveMemoryMode(cfg *config.Config) string {
	if cfg == nil {
		return memoryModeEnhanced
	}

	switch strings.ToLower(strings.TrimSpace(cfg.Memory.Mode)) {
	case memoryModeOff, memoryModeStandard, memoryModeEnhanced:
		return strings.ToLower(strings.TrimSpace(cfg.Memory.Mode))
	}

	if !cfg.Epistemic.Enabled || !cfg.Memory.EAMServingEnabled {
		return memoryModeOff
	}
	return memoryModeEnhanced
}

func shouldEnableEAMServing(cfg *config.Config) bool {
	if cfg == nil {
		return true
	}
	if !cfg.Epistemic.Enabled {
		return false
	}
	switch effectiveMemoryMode(cfg) {
	case memoryModeOff:
		return false
	case memoryModeStandard, memoryModeEnhanced:
		return true
	default:
		return cfg.Memory.EAMServingEnabled
	}
}

func shouldEnableAnticipationRuntime(cfg *config.Config) bool {
	if cfg == nil {
		return true
	}
	if !cfg.Epistemic.Enabled || !cfg.Epistemic.Anticipation.Enabled {
		return false
	}
	return resolveEAMProviderOptions(cfg).EnableAnticipation
}

func resolveEAMProviderOptions(cfg *config.Config) memoryserving.EAMProviderOptions {
	mode := effectiveMemoryMode(cfg)
	switch mode {
	case memoryModeStandard:
		return memoryserving.EAMProviderOptions{
			EnableContradictionPolicy: false,
			EnableTemporalValidity:    true,
			EnableGapSurfacing:        false,
			EnableAnticipation:        false,
		}
	case memoryModeEnhanced:
		return memoryserving.EAMProviderOptions{
			EnableContradictionPolicy: true,
			EnableTemporalValidity:    true,
			EnableGapSurfacing:        false,
			EnableAnticipation:        true,
		}
	default:
		return memoryserving.EAMProviderOptions{}
	}
}
