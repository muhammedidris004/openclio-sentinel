# OpenClio Sentinel

OpenClio Sentinel is the public Synthesis hackathon submission layer for a
trusted local-first agent workflow.

This repo is intentionally narrow. It contains:

- the Synthesis submission docs
- trust and cooperation demo scripts
- environment templates
- export and verification helpers

It does not contain the private backend implementation details for OpenClio,
EAM, or internal research systems.

## What This Demonstrates

OpenClio Sentinel is built for:

- `Agents that trust`
- `Agents that cooperate`

The demo shows:

- bounded authority
- visible trust-boundary enforcement
- exportable evidence bundles
- scoped delegation between cooperating agents

## Repo Layout

- `integrations/synthesis/README.md`
  - detailed hackathon workflow
- `integrations/synthesis/PROJECT.md`
  - project framing
- `integrations/synthesis/SUBMISSION.md`
  - submission draft
- `integrations/synthesis/SUBMISSION_BREAKDOWN.md`
  - what happened and what is being submitted
- `integrations/synthesis/TRUST_MODEL.md`
  - trust model
- `integrations/synthesis/COOPERATION_MODEL.md`
  - cooperation model
- `integrations/synthesis/*.sh`
  - registration, demo, export, and verification helpers

## Backend Model

This public repo talks to a running OpenClio backend over HTTP.

Required environment:

- `OPENCLIO_BASE_URL`
- `OPENCLIO_TOKEN`

Optional local helper env file:

- `integrations/synthesis/openclio.example.env`

## Quick Start

1. Configure access to a running OpenClio backend.
2. Read `integrations/synthesis/README.md`.
3. Run:

```bash
./integrations/synthesis/status.sh
./integrations/synthesis/demo_runner.sh
./integrations/synthesis/demo_runner_cooperate.sh
```

## Public/Private Split

Public here:

- hackathon-facing docs
- wrapper scripts
- demo workflow

Private elsewhere:

- full OpenClio runtime internals
- EAM implementation
- research artifacts
- private deployment/configuration

## Submission Assets

The Synthesis submission uses:

- this public repo
- a short demo video
- screenshots
- generated demo artifacts

See:

- `integrations/synthesis/SUBMISSION.md`
- `integrations/synthesis/VIDEO_CHECKLIST.md`
