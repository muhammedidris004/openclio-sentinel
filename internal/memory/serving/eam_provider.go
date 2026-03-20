package serving

import (
	agentctx "github.com/openclio/openclio/internal/context"
)

// EAMProviderOptions is kept in the public build for config compatibility.
// The public hackathon build does not ship the private EAM implementation.
type EAMProviderOptions struct {
	EnableContradictionPolicy bool
	EnableTemporalValidity    bool
	EnableGapSurfacing        bool
	EnableAnticipation        bool
}

// DefaultEAMProviderOptions returns the compatibility default values.
func DefaultEAMProviderOptions() EAMProviderOptions {
	return EAMProviderOptions{
		EnableContradictionPolicy: true,
		EnableTemporalValidity:    true,
		EnableGapSurfacing:        true,
		EnableAnticipation:        true,
	}
}

// NewEAMProviderWithOptions is a no-op in the public hackathon build.
// Private EAM augmentation remains in the private backend.
func NewEAMProviderWithOptions(
	base agentctx.MemoryProvider,
	_ any,
	_ any,
	_ EAMProviderOptions,
) agentctx.MemoryProvider {
	return base
}

// NewEAMProvider is a compatibility no-op in the public build.
func NewEAMProvider(base agentctx.MemoryProvider, _ any, _ any) agentctx.MemoryProvider {
	return base
}
