package benchmark

import (
	"context"
	"encoding/json"
	"os"
	"time"
)

type SuiteConfig struct {
	StalenessBeliefs         int
	StalenessContradictions  int
	StalenessExpiredBeliefs  int
	AnticipationTopics       int
	AnticipationSessions     int
	KnowledgeGapKnownTopics  int
	KnowledgeGapSparseTopics int
	KnowledgeGapQuestions    int
	CausalScenarios          int
}

type BenchmarkResult struct {
	Name      string             `json:"name"`
	Completed bool               `json:"completed"`
	Error     string             `json:"error,omitempty"`
	Metrics   map[string]float64 `json:"metrics,omitempty"`
	Notes     []string           `json:"notes,omitempty"`
}

type SuiteResult struct {
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt time.Time         `json:"finished_at"`
	Results    []BenchmarkResult `json:"results"`
}

type Harness struct{ cfg SuiteConfig }

type ExternalStalenessRequest struct{}
type ExternalStalenessResponse struct {
	Provider string             `json:"provider,omitempty"`
	Metrics  map[string]float64 `json:"metrics,omitempty"`
	Notes    []string           `json:"notes,omitempty"`
}

type E2EConfig struct {
	Cases     int
	Shuffle   bool
	Seed      int64
	Ablations []string
}

type E2EAblation struct {
	Name               string  `json:"name"`
	Cases              int     `json:"cases"`
	Accuracy           float64 `json:"accuracy"`
	AbstainRate        float64 `json:"abstain_rate"`
	ConfidentWrongRate float64 `json:"confident_wrong_rate"`
	AvgTotalTokens     float64 `json:"avg_total_tokens"`
	AvgTier2Retrieved  float64 `json:"avg_tier2_retrieved"`
}

type E2EResult struct {
	Ablations []E2EAblation `json:"ablations"`
}

const (
	AblationNone      = "none"
	AblationTier2Only = "tier2_only"
	AblationTier3Only = "tier3_only"
	AblationFullEAM   = "full_eam"
)

func DefaultSuiteConfig() SuiteConfig     { return SuiteConfig{} }
func NewHarness(cfg SuiteConfig) *Harness { return &Harness{cfg: cfg} }
func (h *Harness) Run(context.Context) (*SuiteResult, error) {
	return &SuiteResult{StartedAt: time.Now(), FinishedAt: time.Now(), Results: []BenchmarkResult{}}, nil
}
func BuildExternalStalenessRequest(SuiteConfig) (*ExternalStalenessRequest, error) {
	return &ExternalStalenessRequest{}, nil
}
func RunExternalStalenessAdapter(context.Context, string, *ExternalStalenessRequest) (*ExternalStalenessResponse, error) {
	return &ExternalStalenessResponse{Metrics: map[string]float64{}}, nil
}
func SaveReport(path string, result *SuiteResult) error {
	data, _ := json.MarshalIndent(result, "", "  ")
	return os.WriteFile(path, data, 0644)
}
func DefaultE2EConfig() E2EConfig                           { return E2EConfig{} }
func RunE2E(context.Context, E2EConfig) (*E2EResult, error) { return &E2EResult{}, nil }
func SaveE2EReport(path string, result *E2EResult) error {
	data, _ := json.MarshalIndent(result, "", "  ")
	return os.WriteFile(path, data, 0644)
}
func WriteE2ECasebook(path string, _ *E2EResult, _ string, _ int) error {
	return os.WriteFile(path, []byte("# Public build\nEAM benchmark casebook unavailable in the public hackathon build.\n"), 0644)
}
