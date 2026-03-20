# OpenClio Sentinel

OpenClio Sentinel is the public Synthesis hackathon submission repo for a
trusted local-first agent workflow.

This repo is intentionally curated. It contains:

- a runnable public OpenClio build for the hackathon demo path
- the Synthesis submission docs
- trust and cooperation demo scripts
- environment templates
- export and verification helpers

It intentionally excludes the private EAM implementation and research
artifacts. The public build keeps the trusted runtime, tools, gateway, and
hackathon workflow needed to run the demo, while omitting private research-only
modules.

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

## Public Build Model

This repo can be used in two ways:

1. build and run the included public OpenClio runtime
2. point the hackathon scripts at another running OpenClio backend over HTTP

Required environment:

- `OPENCLIO_BASE_URL`
- `OPENCLIO_TOKEN`

Optional local helper env file:

- `integrations/synthesis/openclio.example.env`

## Quick Start

1. Build the included runtime:

```bash
go build -o openclio ./cmd/openclio
```

2. Start the server:

```bash
./openclio serve
```

3. Read `integrations/synthesis/README.md`.
4. Run:

```bash
./integrations/synthesis/status.sh
./integrations/synthesis/demo_runner.sh
./integrations/synthesis/demo_runner_cooperate.sh
```

## Public/Private Split

Public here:

- runnable public runtime for the hackathon demo path
- hackathon-facing docs
- demo scripts
- demo workflow

Private elsewhere:

- private EAM implementation
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
