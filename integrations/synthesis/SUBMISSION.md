# OpenClio Sentinel — Synthesis Submission Draft

## Title

OpenClio Sentinel

## Track

- Primary: `Agents that trust`
- Secondary extension: `Agents that cooperate`
- Supporting property: `Agents that keep secrets`

## One-line Pitch

OpenClio Sentinel is a local-first trusted operator agent that works under explicit permissions, stays inside bounded authority, and exports auditable evidence of what it did.

## Short Submission Description

OpenClio Sentinel is built on top of the OpenClio local agent runtime. It is designed to show that useful agents do not need opaque cloud execution or broad hidden permissions to be effective. Instead, Sentinel keeps execution local-first, limits authority through visible boundaries, and produces a clean trace of what happened during a run.

The project focuses on a trusted builder/operator workflow. A user gives Sentinel a real repository inspection task. Sentinel performs safe read-only analysis inside the allowed workspace, reports what it inspected, explains the reliability and trust risks it found, and recommends the safest next steps for the human operator.

The key trust demonstration is not just that the agent works, but that it is visibly constrained. In the canonical demo, Sentinel performs an explicit boundary test, attempts one harmless inspection outside the allowed workspace, gets blocked by the runtime, and then falls back to a safe in-workspace action. That blocked action is exported in the tool trace and summarized in the final artifact bundle. This makes the trust story concrete: the agent is useful, but it is not unconstrained.

## Problem

Most agents ask users to trust too much:

- cloud execution with weak visibility
- broad filesystem or tool authority
- little proof of what actually happened
- no clean way to distinguish allowed behavior from dangerous behavior

This is a serious problem for real operator and developer workflows, especially when repositories, local files, and sensitive context are involved.

## Solution

OpenClio Sentinel treats trust as a runtime property, not just a marketing claim.

It combines:

- local-first execution
- bounded authority
- visible tool usage
- exportable session evidence

The result is an agent that can do real work while staying inside constraints a human can understand and verify.

## What the Demo Shows

The canonical demo does two things:

### 1. Safe workspace inspection

Sentinel is asked to inspect the repository, summarize trust and reliability risks, and propose safe next actions.

The agent:

- stays inside the target workspace
- uses safe read-only inspection tools
- avoids destructive behavior
- produces a trust-focused report

### 2. Explicit trust-boundary check

Sentinel is then asked to perform one harmless inspection outside the workspace if policy allows it.

In the canonical run:

- it attempts a single `list_dir` call against the parent directory
- the runtime blocks the call because the path resolves outside the allowed workspace
- the agent explains the denial
- the agent falls back to a safe in-workspace action

This is the central trust proof for the project.

## Why This Fits "Agents that trust"

OpenClio Sentinel fits `Agents that trust` because it demonstrates:

- **bounded authority** instead of hidden unrestricted power
- **local-first execution** instead of mandatory cloud trust
- **auditable behavior** instead of opaque tool usage
- **visible policy enforcement** instead of informal promises

The most important part is that trust is enforced by the runtime and then exported as evidence.

## Cooperation Extension

OpenClio Sentinel also supports a cooperation demo built on the real delegation runtime.

In that extension:

- the coordinator uses the `delegate` tool
- an `Inspector` subagent identifies trust and reliability risks
- a `Planner` subagent proposes the safest human next steps
- the coordinator synthesizes the final answer
- the export includes a cooperation trace with subagent events

This means the project is not only about bounded authority, but also about role-scoped agent collaboration that remains auditable.

## Canonical Demo Evidence

Canonical trust artifact:

- `integrations/synthesis/artifacts/demo_20260314_004629/` (generated locally; not committed by default)

Canonical cooperation artifact:

- `integrations/synthesis/artifacts/cooperate_20260315_011837/` (generated locally; not committed by default)

Important files:

- `SUMMARY.md`
- `chat_response.json`
- `boundary_response.json`
- `tool_trace.json`
- `session_export.json`
- `status.json`

Key measurements from that run:

- provider: `ollama`
- model: `gpt-oss:20b`
- tool calls observed: `2`
- successful tool calls: `1`
- blocked/denied tool calls: `1`
- total messages: `4`
- total tokens: `981`

Most important blocked error:

```text
path <parent-of-workspace> resolves outside allowed directory <workspace>
```

Key cooperation measurements from the cooperation run:

- provider: `openai-compat`
- model: `NVIDIA NIM`
- delegation starts: `1`
- subagent spawns: `2`
- subagent tool events: `6`
- subagent completions: `2`
- delegation completions: `1`

## What Is Core OpenClio vs Hackathon-Specific

### Core OpenClio

- local agent runtime
- memory and session handling
- tool execution
- bounded workspace/tool behavior
- HTTP API
- exportable sessions and stats

### Hackathon-Specific Layer

- Synthesis registration helper
- demo runner
- session export bundling
- trust-focused submission docs
- artifact verification flow

## What We Are Submitting

We are submitting:

1. the public hackathon branch
2. the OpenClio Sentinel project docs
3. the canonical demo artifact bundle
4. the demo runner and export scripts
5. the final written Devfolio submission text
6. a short demo video / walkthrough

## What We Are Not Claiming

We are not claiming:

- unrestricted autonomy
- hidden elevated execution
- universal trust without evidence
- a fake wrapper around a non-participating agent

The project is specifically about trusted bounded execution, not unconstrained autonomy.

## Public Repo / Deliverables

- public hackathon branch
- `integrations/synthesis/README.md`
- `integrations/synthesis/PROJECT.md`
- `integrations/synthesis/TRUST_MODEL.md`
- `integrations/synthesis/ARCHITECTURE.md`
- `integrations/synthesis/SUBMISSION_BREAKDOWN.md`
- canonical artifact bundle generated by the demo runners
- short demo video

## Judge-Friendly Summary

OpenClio Sentinel is a trusted local-first agent for operator/developer workflows. In the canonical demo, it inspects a real repository safely, then performs an explicit trust-boundary test by attempting one harmless inspection outside the allowed workspace. The runtime blocks that action, the denial is captured in the tool trace, and the agent falls back to a safe in-workspace operation. This gives the judge concrete evidence that the system is useful, auditable, and bounded by real runtime policy rather than by a verbal promise alone.
