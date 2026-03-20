package eam

import (
	"context"
	"time"
)

type Category string
type Provenance string

type Belief struct {
	ID          int64      `json:"id,omitempty"`
	Claim       string     `json:"claim,omitempty"`
	Confidence  float64    `json:"confidence,omitempty"`
	Category    Category   `json:"category,omitempty"`
	Provenance  Provenance `json:"provenance,omitempty"`
	ValidUntil  *time.Time `json:"valid_until,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at,omitempty"`
	AccessCount int        `json:"access_count,omitempty"`
	IsActive    bool       `json:"is_active,omitempty"`
}

type BeliefStore interface {
	GetActive(context.Context, int) ([]*Belief, error)
}

type SQLiteBeliefStore struct{}

func NewSQLiteBeliefStore(string) (*SQLiteBeliefStore, error) { return &SQLiteBeliefStore{}, nil }
func (s *SQLiteBeliefStore) Close() error                     { return nil }
func (s *SQLiteBeliefStore) GetActive(context.Context, int) ([]*Belief, error) {
	return nil, nil
}

type ExtractionConfig struct{}
type HybridExtractor struct{}

func DefaultExtractionConfig() ExtractionConfig { return ExtractionConfig{} }
func NewHybridExtractor(ExtractionConfig, any) (*HybridExtractor, error) {
	return &HybridExtractor{}, nil
}

func ExtractAndUpsert(context.Context, *HybridExtractor, BeliefStore, string, string, any) ([]*Belief, error) {
	return nil, nil
}
