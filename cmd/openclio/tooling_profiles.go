package main

import "github.com/openclio/openclio/internal/config"

func resolveToolingConfig(cfg *config.Config) {
	config.ResolveToolingConfig(cfg)
}

func writeToolsReference(dataDir string, cfg *config.Config) error {
	return config.WriteToolsReference(dataDir, cfg)
}
