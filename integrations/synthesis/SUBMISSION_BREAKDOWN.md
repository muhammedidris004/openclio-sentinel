# OpenClio Sentinel Submission Breakdown

This file explains two things clearly:

1. what happened in the current canonical demo run
2. what we are actually going to submit for the Synthesis hackathon

## Project

- Project name: `OpenClio Sentinel`
- Track: `Agents that trust`
- Extension: `Agents that cooperate`
- Runtime: real local OpenClio API
- Agent harness registration: `other` -> `openclio`

## Canonical Demo Run

Canonical artifact folder:

- `integrations/synthesis/artifacts/demo_20260314_004629/` (generated locally; not committed by default)

Canonical cooperation artifact folder:

- `integrations/synthesis/artifacts/cooperate_20260315_011837/` (generated locally; not committed by default)

Primary artifact files:

- `SUMMARY.md`
- `chat_response.json`
- `boundary_response.json`
- `session_export.json`
- `conversation_log.json`
- `tool_trace.json`
- `status.json`

## What Happened

The demo was a two-stage trust-focused run.

### Stage 1: Safe in-workspace inspection

The agent was asked to:

- stay inside the repository workspace
- use only safe read/inspect actions
- avoid destructive behavior
- produce a trust-focused report

What the agent actually did:

- used `list_dir` on the repo root
- stayed inside the repository workspace
- did not modify files
- did not execute destructive actions

Result:

- 1 successful tool call
- safe read-only inspection
- trust-focused report produced

### Stage 2: Explicit trust-boundary test

The agent was then asked to perform one explicit boundary test.

What the agent actually did:

- attempted `list_dir` on the parent directory of the workspace
- attempted `list_dir` on the parent directory of the workspace
- the runtime blocked the request because it was outside the allowed workspace
- the agent then explained the denial and used the workspace-safe fallback

This is the most important trust proof in the demo.

The exact blocked error was:

```text
path <parent-of-workspace> resolves outside allowed directory <workspace>
```

## Measured Runtime Evidence

From the canonical artifact:

- Session ID: `57c7d883-f515-4c4e-8233-102eb677e462`
- OpenClio base URL: `http://127.0.0.1:18789`
- Provider: `ollama`
- Model: `gpt-oss:20b`
- Tool packs: `developer`, `research`
- Memory mode: `unset`
- Total messages: `4`
- Total tokens: `981`
- Tool calls observed: `2`
- Successful tool calls: `1`
- Blocked/denied tool calls: `1`

## Cooperation Evidence

From the cooperation artifact:

- Session ID: `1ca86c40-8a1c-4a28-8e09-be4e197d5917`
- Provider: `openai-compat`
- Model: `NVIDIA NIM`
- Delegation starts: `1`
- Subagent spawns: `2`
- Subagent tool events: `6`
- Subagent completions: `2`
- Delegation completions: `1`

This proves the cooperation story is backed by real runtime telemetry rather than a narrated multi-agent claim.

## Why This Fits "Agents that trust"

This demo shows the exact behavior the track wants:

- the agent is useful
- the agent is bounded
- the user can see what it did
- the runtime enforces policy
- the evidence bundle is exportable and auditable

This is not just a generic chat demo.

The key trust claim is:

- OpenClio Sentinel can do useful work
- while remaining inside explicit user-controlled authority boundaries

## What We Are Submitting

We are submitting:

### 1. The project

- `OpenClio Sentinel`
- a local-first trusted operator/developer agent built on OpenClio

### 2. The public code

The hackathon branch/repo code itself, including:

- hackathon integration layer in `integrations/synthesis/`
- trust model docs
- architecture docs
- demo runner
- export helpers

### 3. The canonical evidence bundle

Use this artifact as the primary submission evidence:

- `integrations/synthesis/artifacts/demo_20260314_004629/`

Use this artifact as the cooperation extension evidence:

- `integrations/synthesis/artifacts/cooperate_20260315_011837/`

### 4. The written submission content

Main files:

- `integrations/synthesis/SUBMISSION.md`
- `integrations/synthesis/PROJECT.md`
- `integrations/synthesis/TRUST_MODEL.md`
- `integrations/synthesis/ARCHITECTURE.md`
- `integrations/synthesis/VIDEO_CHECKLIST.md`

### 5. Demo video / screenshots

We should capture:

- one clean run of the demo
- one screenshot of the blocked boundary tool call
- one screenshot of the final trust summary

## What We Are Not Claiming

We are not claiming:

- unrestricted autonomous operation
- arbitrary filesystem access
- hidden privileged execution
- a fake wrapper around a non-participating agent

The trustworthy part of the project is exactly that it stays bounded.

## Submission Short Version

If someone asks "what happened?" in one paragraph:

OpenClio Sentinel was run as a real local agent through the OpenClio API. It first performed a safe read-only inspection inside the repository, then performed a deliberate trust-boundary test by attempting to inspect a parent directory outside the allowed workspace. The runtime blocked that out-of-scope tool call, the agent reported the denial, and then continued safely within the workspace. The final artifact bundle includes the conversation, exported session, tool trace, health snapshot, trust model, and architecture notes, providing direct evidence that the agent is useful but still constrained by explicit policy.

## Submission Decision

Current recommendation:

- use the canonical artifact from `demo_20260314_004629`
- use `OpenClio Sentinel` as the submission name
- present the project as a trusted local-first agent runtime for bounded operator workflows

This is now strong enough to package for the hackathon.
