# Synthesis Integration

This folder contains the OpenClio-specific scaffolding for the Devfolio
Synthesis hackathon workflow.

Hackathon project:
- **OpenClio Sentinel**

Primary fit:
- `Agents that trust`

This branch treats OpenClio as the real local-first runtime and adds a
hackathon-facing layer for:
- registration
- demo execution
- evidence export
- submission packaging

Current contents:

- `devfolio-skills.md` — copied reference/API instructions
- `PROJECT.md` — hackathon-facing project definition
- `TRUST_MODEL.md` — trust and authority model
- `ARCHITECTURE.md` — high-level system flow
- `SUBMISSION.md` — submission-ready title/pitch/description notes
- `VIDEO_CHECKLIST.md` — recording checklist
- `register.sh` — registration helper for creating the agent participant
- `register.example.env` — environment template for the required human info
- `openclio.example.env` — local OpenClio API config template
- `status.sh` — checks local OpenClio API readiness for the Synthesis wrapper
- `run_task.sh` — sends a task into local OpenClio through `/api/v1/chat`
- `export_session.sh` — exports a finished OpenClio session and stats as JSON
- `demo_runner.sh` — runs the canonical trusted-agent demo and packages artifacts
- `verify_artifact.sh` — checks whether a generated demo artifact is submission-grade
- `demo_task.md` — default demo prompt

## Quick Start

1. Copy the example env file and fill in your details:

   ```bash
   cp integrations/synthesis/register.example.env integrations/synthesis/.env
   ```

2. Edit `integrations/synthesis/.env` with the required human info.

3. Run the registration helper:

   ```bash
   set -a
   source integrations/synthesis/.env
   set +a

   ./integrations/synthesis/register.sh
   ```

## Notes

- OpenClio is not part of the built-in `agentHarness` enum, so this helper uses:
  - `agentHarness: "other"`
  - `agentHarnessOther: "openclio"`
- The registration API returns an `apiKey` only once. Save it immediately.
- The tracked `registration.json` should remain sanitized and must not contain the API key.

## OpenClio Wrapper Flow

These helpers use the existing local OpenClio HTTP API. No second agent API is
required.

1. Make sure OpenClio is running:

   ```bash
   openclio serve
   ```

2. Copy the local API env template if you want to override defaults:

   ```bash
   cp integrations/synthesis/openclio.example.env integrations/synthesis/.openclio.env
   ```

3. Load the local API env if needed:

   ```bash
   set -a
   source integrations/synthesis/.openclio.env
   set +a
   ```

4. Check local API readiness:

   ```bash
   ./integrations/synthesis/status.sh
   ```

5. Send a task into OpenClio:

   ```bash
   ./integrations/synthesis/run_task.sh "Summarize the trust and privacy story for this hackathon demo."
   ```

   `run_task.sh` uses:
   - buffered JSON mode when stdout is captured, for example `resp="$(...)"`, so shell pipelines can safely parse the result
   - streaming mode when run directly in a terminal, so you can see tool/text progress live

6. Export a session when you want logs/artifacts:

   ```bash
   ./integrations/synthesis/export_session.sh SESSION_ID integrations/synthesis/demo-session.json
   ```

7. Run the canonical hackathon demo and package evidence:

   ```bash
   ./integrations/synthesis/demo_runner.sh
   ```

8. Verify the generated artifact bundle:

   ```bash
   ./integrations/synthesis/verify_artifact.sh integrations/synthesis/artifacts/demo_YYYYMMDD_HHMMSS
   ```

   This creates an artifact directory with:
   - `chat_response.json`
   - `boundary_response.json` (when the trust-boundary follow-up runs)
   - `session_export.json`
   - `conversation_log.json`
    - `tool_trace.json`
   - `SUMMARY.md`
   - `TRUST_MODEL.md`
   - `ARCHITECTURE.md`
   - `SUBMISSION.md`

   Generated artifacts are ignored by git on purpose. For a public hackathon repo, keep the scripts and docs tracked, and attach the video/screenshots plus any chosen artifact bundle through the submission form or a separate public release asset.

The canonical demo now runs in two stages:
- a useful trusted operator inspection task
- a trust-boundary follow-up that tries to exercise workspace enforcement without performing destructive work

## OpenClio API Used By The Wrapper

The hackathon wrapper currently uses these built-in OpenClio endpoints:

- `GET /api/v1/health`
- `GET /api/v1/overview`
- `POST /api/v1/chat`
- `GET /api/v1/sessions/{id}`
- `GET /api/v1/sessions/{id}/stats`

Authenticated requests use the same bearer token OpenClio writes to:

- `~/.openclio/auth.token`

You can override that with `OPENCLIO_TOKEN` if needed.

## Public Submission Story

This branch should be treated as the public hackathon branch.

The submission story is:
- OpenClio Sentinel is a trusted local-first operator/developer agent
- it uses the real OpenClio runtime
- it keeps authority bounded
- it exports session evidence after doing real work

This is intentionally not a wrapper-only shell. The Synthesis-specific layer
exists to:
- register the participant
- run the canonical demo
- package exported evidence
- make the project understandable to judges
