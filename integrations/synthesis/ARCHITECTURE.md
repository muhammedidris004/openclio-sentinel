# Architecture

## High-Level Flow

1. Human gives a task to OpenClio Sentinel.
2. The hackathon wrapper sends the task to the local OpenClio HTTP API.
3. OpenClio executes the task using:
   - model provider
   - memory
   - tools
   - approval / allowlist policy
4. The wrapper exports:
   - session trace
   - session stats
   - conversation log
   - process summary
   - trust-policy summary

## Components

- **Synthesis Wrapper**
  - registration helper
  - status check
  - demo runner
  - export helpers

- **OpenClio Runtime**
  - chat/task execution API
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
OpenClio Sentinel wrapper
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
