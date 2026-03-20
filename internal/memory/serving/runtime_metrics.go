package serving

import (
	"sync"
	"sync/atomic"
)

// EAMRuntimeMetricsSnapshot is a point-in-time view of EAM serving counters.
type EAMRuntimeMetricsSnapshot struct {
	AddendumCalls             uint64  `json:"addendum_calls"`
	BeliefBlockInjected       uint64  `json:"belief_block_injected"`
	KnowledgeGapBlockInjected uint64  `json:"knowledge_gap_block_injected"`
	PolicyBlockInjected       uint64  `json:"policy_block_injected"`
	StaleBeliefSignals        uint64  `json:"stale_belief_signals"`
	LowConfidenceSignals      uint64  `json:"low_confidence_signals"`
	GapDetections             uint64  `json:"gap_detections"`
	ContradictedGapSignals    uint64  `json:"contradicted_gap_signals"`
	ExpiredGapSignals         uint64  `json:"expired_gap_signals"`
	BeliefAccessUpdates       uint64  `json:"belief_access_updates"`
	LoadFromStaged            uint64  `json:"load_from_staged"`
	LoadFromSearch            uint64  `json:"load_from_search"`
	LoadFromActive            uint64  `json:"load_from_active"`
	PreloadHitRate            float64 `json:"preload_hit_rate"`
}

// EAMRuntimeMetrics tracks runtime behavior of EAM serving logic.
type EAMRuntimeMetrics struct {
	addendumCalls             atomic.Uint64
	beliefBlockInjected       atomic.Uint64
	knowledgeGapBlockInjected atomic.Uint64
	policyBlockInjected       atomic.Uint64
	staleBeliefSignals        atomic.Uint64
	lowConfidenceSignals      atomic.Uint64
	gapDetections             atomic.Uint64
	contradictedGapSignals    atomic.Uint64
	expiredGapSignals         atomic.Uint64
	beliefAccessUpdates       atomic.Uint64
	loadFromStaged            atomic.Uint64
	loadFromSearch            atomic.Uint64
	loadFromActive            atomic.Uint64
}

func (m *EAMRuntimeMetrics) IncAddendumCalls() {
	if m == nil {
		return
	}
	m.addendumCalls.Add(1)
}

func (m *EAMRuntimeMetrics) IncBeliefBlockInjected() {
	if m == nil {
		return
	}
	m.beliefBlockInjected.Add(1)
}

func (m *EAMRuntimeMetrics) IncKnowledgeGapBlockInjected() {
	if m == nil {
		return
	}
	m.knowledgeGapBlockInjected.Add(1)
}

func (m *EAMRuntimeMetrics) IncPolicyBlockInjected() {
	if m == nil {
		return
	}
	m.policyBlockInjected.Add(1)
}

func (m *EAMRuntimeMetrics) AddStaleBeliefSignals(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.staleBeliefSignals.Add(uint64(n))
}

func (m *EAMRuntimeMetrics) AddLowConfidenceSignals(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.lowConfidenceSignals.Add(uint64(n))
}

func (m *EAMRuntimeMetrics) AddGapDetections(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.gapDetections.Add(uint64(n))
}

func (m *EAMRuntimeMetrics) AddContradictedGapSignals(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.contradictedGapSignals.Add(uint64(n))
}

func (m *EAMRuntimeMetrics) AddExpiredGapSignals(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.expiredGapSignals.Add(uint64(n))
}

func (m *EAMRuntimeMetrics) AddBeliefAccessUpdates(n int) {
	if m == nil || n <= 0 {
		return
	}
	m.beliefAccessUpdates.Add(uint64(n))
}

func (m *EAMRuntimeMetrics) IncLoadFromStaged() {
	if m == nil {
		return
	}
	m.loadFromStaged.Add(1)
}

func (m *EAMRuntimeMetrics) IncLoadFromSearch() {
	if m == nil {
		return
	}
	m.loadFromSearch.Add(1)
}

func (m *EAMRuntimeMetrics) IncLoadFromActive() {
	if m == nil {
		return
	}
	m.loadFromActive.Add(1)
}

func (m *EAMRuntimeMetrics) Snapshot() EAMRuntimeMetricsSnapshot {
	if m == nil {
		return EAMRuntimeMetricsSnapshot{}
	}
	staged := m.loadFromStaged.Load()
	search := m.loadFromSearch.Load()
	active := m.loadFromActive.Load()
	totalLoads := staged + search + active
	preloadHitRate := 0.0
	if totalLoads > 0 {
		preloadHitRate = float64(staged) / float64(totalLoads)
	}
	return EAMRuntimeMetricsSnapshot{
		AddendumCalls:             m.addendumCalls.Load(),
		BeliefBlockInjected:       m.beliefBlockInjected.Load(),
		KnowledgeGapBlockInjected: m.knowledgeGapBlockInjected.Load(),
		PolicyBlockInjected:       m.policyBlockInjected.Load(),
		StaleBeliefSignals:        m.staleBeliefSignals.Load(),
		LowConfidenceSignals:      m.lowConfidenceSignals.Load(),
		GapDetections:             m.gapDetections.Load(),
		ContradictedGapSignals:    m.contradictedGapSignals.Load(),
		ExpiredGapSignals:         m.expiredGapSignals.Load(),
		BeliefAccessUpdates:       m.beliefAccessUpdates.Load(),
		LoadFromStaged:            staged,
		LoadFromSearch:            search,
		LoadFromActive:            active,
		PreloadHitRate:            preloadHitRate,
	}
}

func (m *EAMRuntimeMetrics) Reset() {
	if m == nil {
		return
	}
	m.addendumCalls.Store(0)
	m.beliefBlockInjected.Store(0)
	m.knowledgeGapBlockInjected.Store(0)
	m.policyBlockInjected.Store(0)
	m.staleBeliefSignals.Store(0)
	m.lowConfidenceSignals.Store(0)
	m.gapDetections.Store(0)
	m.contradictedGapSignals.Store(0)
	m.expiredGapSignals.Store(0)
	m.beliefAccessUpdates.Store(0)
	m.loadFromStaged.Store(0)
	m.loadFromSearch.Store(0)
	m.loadFromActive.Store(0)
}

var (
	defaultEAMRuntimeMetricsMu sync.Mutex
	defaultEAMRuntimeMetrics   = &EAMRuntimeMetrics{}
)

// DefaultEAMRuntimeMetrics returns the process-global EAM metrics collector.
func DefaultEAMRuntimeMetrics() *EAMRuntimeMetrics {
	defaultEAMRuntimeMetricsMu.Lock()
	defer defaultEAMRuntimeMetricsMu.Unlock()
	if defaultEAMRuntimeMetrics == nil {
		defaultEAMRuntimeMetrics = &EAMRuntimeMetrics{}
	}
	return defaultEAMRuntimeMetrics
}

// ResetDefaultEAMRuntimeMetrics clears process-global counters.
func ResetDefaultEAMRuntimeMetrics() {
	DefaultEAMRuntimeMetrics().Reset()
}
