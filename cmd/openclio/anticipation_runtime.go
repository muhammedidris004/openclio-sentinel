package main

import (
	"github.com/openclio/openclio/internal/config"
	internlog "github.com/openclio/openclio/internal/logger"
	"github.com/openclio/openclio/internal/memory/eam"
	"github.com/openclio/openclio/internal/memory/eam/anticipation"
	"github.com/openclio/openclio/internal/storage"
)

// Public hackathon build: anticipation runtime is intentionally disabled.
func setupAnticipationEngine(
	_ *config.Config,
	_ *storage.DB,
	_ eam.BeliefStore,
	_ *internlog.Logger,
) (*anticipation.Engine, error) {
	return nil, nil
}
