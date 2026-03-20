package anticipation

import (
	"context"
	"database/sql"
	"time"

	"github.com/openclio/openclio/internal/memory/eam"
)

type Engine struct{}

type EngineConfig struct {
	PreLoadTopK       int
	MinRelevanceScore float64
}

type GapConfig struct {
	Enabled             bool
	SparseThreshold     int
	LowConfidenceCutoff float64
	MaxGapsToSurface    int
	GapAskCooldownHours int
}

type GapType string

const (
	GapContradicted GapType = "contradicted"
	GapExpired      GapType = "expired"
)

type KnowledgeGap struct {
	GapType  GapType `json:"gap_type,omitempty"`
	IsActive bool    `json:"is_active,omitempty"`
}

type SQLiteGapStore struct{}
type GapDetector struct{}
type BeliefScorer struct{}
type StagingCache struct{}

func DefaultEngineConfig() EngineConfig { return EngineConfig{} }
func DefaultGapConfig() GapConfig       { return GapConfig{MaxGapsToSurface: 3} }
func NewBeliefScorer() *BeliefScorer    { return &BeliefScorer{} }
func NewStagingCache(time.Duration) *StagingCache {
	return &StagingCache{}
}
func NewEngine(eam.BeliefStore, *BeliefScorer, *StagingCache, EngineConfig) *Engine {
	return &Engine{}
}
func NewSQLiteGapStore(*sql.DB) (*SQLiteGapStore, error) { return &SQLiteGapStore{}, nil }
func NewGapDetector(*SQLiteGapStore, GapConfig) *GapDetector {
	return &GapDetector{}
}
func (e *Engine) SetGapDetector(*GapDetector) {}
func (e *Engine) PreLoad(context.Context, string, any) ([]*eam.Belief, error) {
	return nil, nil
}
func (e *Engine) GetStaged(context.Context, string) ([]*eam.Belief, error) {
	return nil, nil
}
