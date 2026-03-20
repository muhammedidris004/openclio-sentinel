# OpenClio Sentinel

## One-line Pitch

OpenClio Sentinel is a local-first trusted operator agent that acts on behalf of a user under explicit permissions, auditable actions, and exportable session evidence.

## Track

- Primary: `Agents that trust`
- Secondary build extension: `Agents that cooperate`
- Supporting property: `Agents that keep secrets`

## Problem

Most AI agents ask users to trust opaque cloud execution, broad permissions, and weak auditability. That makes them difficult to adopt for real work involving repositories, private notes, and sensitive local context.

## Solution

OpenClio Sentinel uses OpenClio as a real local-first agent runtime:

- memory stays local by default
- dangerous actions are bounded by approvals and allowlists
- tool usage is visible and reviewable
- every demo run can be exported as a clean evidence package

The hackathon submission is not a wrapper-only shell. It is a real OpenClio-powered agent workflow with a trust story:

- a user gives a concrete operator/developer task
- the agent inspects context and proposes safe actions
- risky capabilities stay visibly bounded
- the run is exported as a session trace plus human-readable summary

## Canonical Demo

### Trusted local builder/operator assistant

The user asks Sentinel to inspect a repository, summarize risks, and propose a safe next-step plan.

Expected behavior:

1. The agent analyzes the local project through OpenClio.
2. The agent uses built-in tools and controlled local CLI access where allowed.
3. The agent avoids destructive execution unless explicitly allowed.
4. The run is exported as:
   - raw session trace
   - tool/action summary
   - trust-policy summary
   - human-readable conversation/process log

## Cooperation Extension

The next hackathon extension uses OpenClio's real delegation runtime.

In the cooperation demo:

1. the coordinator receives the user goal
2. the coordinator delegates two scoped subagents
3. the `Inspector` performs read-only repo inspection
4. the `Planner` turns those findings into safe human actions
5. the coordinator synthesizes the final answer
6. the exported artifact includes a cooperation trace

This keeps trust as the core story while making inter-agent cooperation visible.

Canonical cooperation artifact:

- `integrations/synthesis/artifacts/cooperate_20260315_011837/` (generated locally; not committed by default)

## Why This Fits Synthesis

This project is aligned to `Agents that trust` because it demonstrates:

- user-controlled permissions instead of hidden authority
- local-first execution instead of mandatory cloud dependence
- auditable behavior instead of opaque black-box action
- exportable proof of what the agent did

## What Is Core OpenClio vs Hackathon-Specific

### Core OpenClio

- agent runtime
- local memory
- tool execution
- HTTP API
- approvals / allowlists
- session export

### Hackathon Layer

- Synthesis registration helpers
- hackathon-specific demo runner
- evidence packaging scripts
- submission-facing project narrative and artifacts

## Success Criteria

The hackathon branch is ready when:

1. A judge can run one command sequence and get a complete demo evidence bundle.
2. The trust model is obvious from the README and artifacts.
3. The exported evidence clearly shows the user request, agent behavior, and resulting recommendation.
4. No secret material is required inside tracked files.
