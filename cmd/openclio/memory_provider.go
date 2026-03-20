package main

import (
	"fmt"
	"strings"

	"github.com/openclio/openclio/internal/config"
	agentctx "github.com/openclio/openclio/internal/context"
	internlog "github.com/openclio/openclio/internal/logger"
	memoryserving "github.com/openclio/openclio/internal/memory/serving"
)

func buildMemoryProvider(
	cfg *config.Config,
	dataDir string,
	log *internlog.Logger,
) (agentctx.MemoryProvider, func(), error) {
	providerName := "workspace"
	if cfg != nil {
		v := strings.ToLower(strings.TrimSpace(cfg.Memory.Provider))
		if v != "" {
			providerName = v
		}
	}

	switch providerName {
	case "workspace":
		if log != nil {
			log.Info("memory provider configured", "provider", "workspace")
		}
		return memoryserving.NewWorkspaceFileProvider(dataDir), nil, nil
	case "mem0style":
		mem0Cfg := memoryserving.Mem0WorkspaceConfig{}
		if cfg != nil {
			mem0Cfg.DBPath = strings.TrimSpace(cfg.Memory.Mem0Style.DBPath)
			mem0Cfg.MinSalience = cfg.Memory.Mem0Style.MinSalience
		}
		p, err := memoryserving.NewMem0WorkspaceProvider(dataDir, mem0Cfg)
		if err != nil {
			return nil, nil, fmt.Errorf("new mem0style workspace provider: %w", err)
		}
		if log != nil {
			log.Info("memory provider configured", "provider", "mem0style")
		}
		return p, func() { _ = p.Close() }, nil
	default:
		return nil, nil, fmt.Errorf("unsupported memory provider %q", providerName)
	}
}
