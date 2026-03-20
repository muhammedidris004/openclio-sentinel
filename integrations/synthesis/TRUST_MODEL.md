# Trust Model

OpenClio Sentinel is designed to demonstrate trusted agent execution rather than unrestricted autonomy.

## Core Principles

- **Local-first by default**: runtime state and memory stay on the user machine unless the user explicitly connects external services.
- **Bounded authority**: tool packs, allowlists, and approval modes constrain what the agent can do.
- **Auditable execution**: sessions, stats, and tool usage can be exported after a run.
- **Human remains in control**: the demo intentionally avoids hidden destructive behavior.

## What the Demo Shows

- the user can give a real operator/developer task
- the agent can inspect and reason over the local project
- the agent works inside visible tool and permission boundaries
- the resulting session can be exported as evidence

## What the Demo Does Not Claim

- unrestricted shell autonomy
- hidden cloud trust assumptions
- autonomous self-directed destructive actions
- trust without auditability
