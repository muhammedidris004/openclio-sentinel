package externaleval

import (
	"context"
	"encoding/json"
	"os"

	"github.com/openclio/openclio/internal/agent"
	agentctx "github.com/openclio/openclio/internal/context"
)

type EvalMode string

const (
	ModeEAMOff             EvalMode = "eam_off"
	ModeFullEAM            EvalMode = "full_eam"
	ModeFullEAMV2          EvalMode = "full_eam_v2"
	ModeNoContradictions   EvalMode = "no_contradictions"
	ModeNoTemporalValidity EvalMode = "no_temporal_validity"
	ModeNoGapSurfacing     EvalMode = "no_gap_surfacing"
	ModeNoAnticipation     EvalMode = "no_anticipation"

	BenchmarkLongMemEval = "longmemeval"
	BenchmarkLoCoMo      = "locomo"
)

type RunConfig struct {
	Provider              agent.Provider
	Model                 string
	Embedder              agentctx.Embedder
	MaxTokens             int
	RetrievalK            int
	RecentTurns           int
	Repeats               int
	Modes                 []EvalMode
	MemoryProviderFactory any
}

type Summary struct {
	Mode               EvalMode `json:"mode"`
	Cases              int      `json:"cases"`
	Accuracy           float64  `json:"accuracy"`
	AbstainRate        float64  `json:"abstain_rate"`
	ConfidentWrongRate float64  `json:"confident_wrong_rate"`
	AvgInputTokens     float64  `json:"avg_input_tokens"`
	AvgOutputTokens    float64  `json:"avg_output_tokens"`
	AvgLatencyMS       float64  `json:"avg_latency_ms"`
}

type RunResult struct {
	Summaries []Summary `json:"summaries,omitempty"`
}

type EvidenceBundleOptions struct {
	Tool        string
	Benchmark   string
	Result      *RunResult
	DatasetPath string
	ReportPath  string
	SummaryPath string
	ConfigPath  string
}

func NewIsolatedEAMFactory(any) (any, error)                                { return struct{}{}, nil }
func RunLongMemEval(context.Context, string, RunConfig) (*RunResult, error) { return &RunResult{}, nil }
func RunLoCoMo(context.Context, string, RunConfig) (*RunResult, error)      { return &RunResult{}, nil }
func SaveRunResult(path string, result *RunResult) error {
	data, _ := json.MarshalIndent(result, "", "  ")
	return os.WriteFile(path, data, 0644)
}
func WriteSummaryMarkdown(path string, _ *RunResult) error {
	return os.WriteFile(path, []byte("# Public build\nExternal EAM evaluation is unavailable in the public hackathon build.\n"), 0644)
}
func WriteEvidenceBundle(opts EvidenceBundleOptions) (string, error) {
	manifestPath := opts.ReportPath + ".manifest.json"
	data, _ := json.MarshalIndent(map[string]any{
		"tool":        opts.Tool,
		"benchmark":   opts.Benchmark,
		"datasetPath": opts.DatasetPath,
		"reportPath":  opts.ReportPath,
		"summaryPath": opts.SummaryPath,
		"configPath":  opts.ConfigPath,
		"note":        "Private EAM evidence bundle omitted from the public hackathon build.",
	}, "", "  ")
	return manifestPath, os.WriteFile(manifestPath, data, 0644)
}
