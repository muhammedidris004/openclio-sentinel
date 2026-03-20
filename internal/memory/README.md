# Memory Module

This folder groups memory-specific components so memory behavior is easy to inspect in one place.

## What Lives Here

- `store.go`: generic local memory CRUD/search store (SQLite-backed utility).
- `serving/provider.go`: runtime memory providers used by the context engine.
  - `WorkspaceFileProvider`: reads `memory.md` from the configured data directory.
  - `StaticProvider`: serves parsed memory entries from in-memory text.
- `serving/mem0_provider.go`: runtime Mem0-style provider.
  - `Mem0WorkspaceProvider`: syncs `memory.md` into a local Mem0-style SQLite fact store.
- `serving/eam_provider.go`: Phase 5 integration provider.
  - wraps base semantic memory with EAM prompt addenda
  - injects staged beliefs (`[Epistemic beliefs]`) and active knowledge gaps (`[Knowledge gaps]`)
  - injects `[Memory response policy]` guidance for contradiction/staleness/uncertainty handling
  - enforces semantic token budget for prompt augmentation
- `eam/math/`: Phase 0 EAM mathematical foundation (pure functions + tests).
  - Bayesian update formulas (supporting/contradicting evidence)
  - Confidence decay (Ebbinghaus curve)
  - Contradiction scoring
  - Pre-load relevance scoring
- `eam/belief/`: Phase 1 + 2 belief layer.
  - `types.go`: belief domain types/constants.
  - `store.go`: `BeliefStore` interface contract.
  - `store_sqlite.go`: SQLite `BeliefStore` implementation (beliefs + evidence + contradictions + causal edges).
  - `extractor.go`: hybrid belief extraction (Layer A rule-based + Layer B trigger support).
  - `injector.go`: system-prompt context block builder with token/quality thresholds.
  - `causal.go`: causal-edge extraction parsing + persistence helpers.
  - `config.go`: extraction and injection defaults.
- `eam/belief_alias.go`: compatibility re-exports from `eam/belief` for existing imports.
- `mem0style/`: Go-native Mem0-style fact-memory baseline.
  - `store.go`: in-memory consolidated fact store (upsert, list/search, expiration counting).
  - `store_sqlite.go`: SQLite-backed consolidated fact store for durable baseline/runtime comparison.
  - intentionally excludes epistemic features (Bayesian updates, contradiction resolution, provenance weighting).
- `eam/ambient/`: Phase 3 ambient observation layer.
  - `calendar.go`: `.ics`/CalDAV-compatible calendar signal collector.
  - `filesystem.go`: `fsnotify`-based filesystem signal collector with rate limiting.
  - `temporal.go`: session pattern tracker that emits time-pattern signals.
  - `processor.go`: ambient signal to EAM belief update pipeline.
  - `store_sqlite.go`: ephemeral ambient signal persistence + cleanup support.
  - `privacy.go`: per-source opt-in and payload minimization.
- `eam/anticipation/`: Phase 4 anticipatory pre-loader and gap detection.
  - `scorer.go`: composite relevance scoring for belief staging.
  - `staging.go`: per-session staging cache with TTL.
  - `engine.go`: orchestrates pre-loading + optional gap detection.
  - `gaps.go`: knowledge gap registry and detector over belief confidence/sparsity.
- `eam/benchmark/`: Phase 6 benchmark harness.
  - `harness.go`: runs the full benchmark suite and emits structured results.
  - `external_adapter.go`: adapter contract for comparing staleness metrics with out-of-repo systems.
  - `datasets/`: curated replay datasets for each benchmark track.
  - `datasets.go` + `dataset_validation.go`: embedded dataset loader and quality gates (shape, balance, coverage checks).
  - `staleness.go`: contradiction/staleness metrics plus Mem0-style fact-memory baseline comparison.
  - `anticipation.go`: pre-load hit-rate and precision metrics.
  - `gaps.go`: knowledge-gap detection and false-positive metrics.
  - `causal.go`: causal-edge persistence and graph-reasoning proxy metrics.

## Runtime Wiring

- Main runtime creates one shared memory provider in `cmd/openclio/main.go`.
  - `memory.provider: workspace` -> `WorkspaceFileProvider`
  - `memory.provider: mem0style` -> `Mem0WorkspaceProvider`
  - `memory.eam_serving_enabled: true` (default) wraps the base provider with `EAMProvider`
  - set `memory.eam_serving_enabled: false` for fast fallback to base Tier-3 provider
- The provider is injected into all run surfaces:
  - CLI chat
  - HTTP/WS gateway
  - Plugin router
  - Cron scheduler
  - gRPC server
- The agent passes this provider into context assembly as Tier 3 (semantic memory).
- Runtime observability surfaces:
  - `GET /api/v1/memory/runtime`: process-local EAM serving counters.
  - `GET /metrics`: Prometheus counters/gauges for EAM addendum/gap/staleness signals.
  - `openclio memory report [--in <report.json>]`: summarizes latest Phase 6 benchmark run and target checks.
  - `openclio memory benchmark-e2e [--quick]`: runs end-to-end tier ablations (`none`, `tier2_only`, `tier3_only`, `full_eam`) and writes JSON + casebook.

## Tier Responsibilities

- Tier 1: working memory (recent turns)
- Tier 2: episodic retrieval (embedding similarity)
- Tier 3: semantic memory (this module's providers)
- Tier 4: knowledge graph entities/relations (via message provider extension)
