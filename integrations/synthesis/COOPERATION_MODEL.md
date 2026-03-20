# OpenClio Sentinel Cooperation Model

OpenClio Sentinel demonstrates `Agents that cooperate` by using OpenClio's
real delegation runtime instead of a fake multi-agent wrapper.

## Roles

- `Coordinator`
  - receives the human goal
  - decides whether delegation is useful
  - creates the sub-task split
  - synthesizes the final answer

- `Inspector`
  - read-only repository inspection
  - identifies trust, reliability, and boundary risks
  - does not modify files

- `Planner`
  - turns findings into safe next actions for the human operator
  - does not execute changes

## Cooperation Rules

- delegation must stay inside the repository workspace
- subagents are scoped to read-only analysis
- the coordinator must merge findings and call out uncertainty
- all cooperation evidence must be exported

## Cooperation Evidence

The cooperation demo is considered valid when the exported artifact contains:

- a final answer produced by the coordinator
- at least one `delegate` tool call
- agent events showing:
  - delegation start
  - subagent spawn
  - subagent tool use and/or completion
  - delegation done

## Why This Helps the Hackathon Story

This makes cooperation visible and bounded:

- the system does not claim hidden collaboration
- subagent work is observable
- each role has a narrower purpose than the coordinator
- trust and cooperation reinforce each other
