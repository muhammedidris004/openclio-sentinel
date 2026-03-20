package main

import (
	"context"

	"github.com/openclio/openclio/internal/config"
	internlog "github.com/openclio/openclio/internal/logger"
	"github.com/openclio/openclio/internal/memory/eam"
	"github.com/openclio/openclio/internal/storage"
)

type disabledAmbientRuntime struct{}

func (r *disabledAmbientRuntime) Stop() error                                 { return nil }
func (r *disabledAmbientRuntime) ProcessPending(context.Context) (int, error) { return 0, nil }

// Public hackathon build: ambient EAM runtime is intentionally disabled.
func setupAmbientRuntime(
	_ *config.Config,
	_ *storage.DB,
	_ string,
	_ string,
	_ *internlog.Logger,
) (*disabledAmbientRuntime, *eam.SQLiteBeliefStore, error) {
	return nil, nil, nil
}

func startAmbientRuntime(_ *disabledAmbientRuntime, _ *internlog.Logger) error {
	return nil
}
