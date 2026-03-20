# Architecture

## High-Level Flow

1. Human gives a task to OpenClio Sentinel.
2. The Synthesis demo and export scripts invoke the local OpenClio runtime and HTTP API.
3. OpenClio executes the task using:
   - model provider
   - memory
   - tools
   - approval / allowlist policy
4. The Synthesis integration exports:
   - session trace
   - session stats
   - conversation log
   - process summary
   - trust-policy summary

## Components

- **OpenClio Sentinel Runtime**
  - runnable `openclio` binary
  - local HTTP API
  - trust and cooperation demo entrypoints
  - bounded tool and approval policy

- **Synthesis Integration Layer**
  - registration helper
  - status check
  - demo runners
  - export helpers

- **OpenClio Core Runtime**
  - chat and task execution API
  - session storage
  - memory
  - tools
  - approvals / allowlists

- **Evidence Bundle**
  - machine-readable run data
  - human-readable summary
  - trust model explanation

## Diagram

```text
Human
  |
  v
OpenClio Sentinel runtime
  |
  +--> Synthesis demo and export scripts
  |
  v
OpenClio local API
  |
  +--> Memory
  +--> Tools / controlled exec / approvals
  +--> Session + stats
  |
  v
Evidence bundle
```
