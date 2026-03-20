# OpenClio Sentinel: Trust + Cooperate + Pay Plan

## Goal

Extend the current `Agents that trust` submission so it can also credibly demonstrate:

- `Agents that cooperate`
- `Agents that pay`

without weakening the core trust story.

## Principle

Do not turn this into a broad AGI platform demo.

Keep the system:

- local-first
- bounded
- auditable
- role-scoped
- human-controlled at risk boundaries

The rule is:

- trust remains the primary story
- cooperation is the first extension
- payment is the second extension

## What We Already Have

Current strengths:

- local OpenClio runtime
- bounded workspace/tool behavior
- allowlist/approval model
- session export and artifact generation
- canonical trust demo with real blocked boundary enforcement

This is already enough for `Agents that trust`.

## What We Need To Add

### 1. Cooperation MVP

Add a visible multi-agent workflow with scoped roles.

Suggested roles:

- `Coordinator`
  - receives the user goal
  - creates the plan
  - decides what each subagent is allowed to do

- `Inspector`
  - read-only repo inspection
  - identifies risks, files, constraints

- `Planner`
  - converts findings into recommended actions
  - does not execute changes

Optional later:

- `Payer`
  - handles payment approval and settlement preparation

What the demo should show:

- the coordinator delegates a subtask
- each subagent has a narrower scope than the coordinator
- outputs are merged back into one final result
- the run exports a coordination trace

### 2. Payment MVP

Do not build a full autonomous finance layer.

Build a minimal real payment flow with explicit controls:

- payment intent
- budget cap
- payee
- human approval
- execution receipt

Recommended payment story:

- the agent identifies a paid action or service request
- it prepares a `payment intent`
- it cannot execute payment unless the user approves
- if approved, it records:
  - amount
  - asset
  - destination
  - transaction reference
  - reason

This is enough to make `Agents that pay` real.

## Recommended Hackathon Story

### Project framing

OpenClio Sentinel is a trusted local-first operator agent that can:

- work inside bounded authority
- coordinate with role-scoped subagents
- prepare and execute user-approved payments with a verifiable receipt

### Strong combined theme fit

- `Agents that trust`
  - bounded authority
  - blocked boundary check
  - local-first memory and execution

- `Agents that cooperate`
  - coordinator + scoped subagents
  - explicit handoff and shared result

- `Agents that pay`
  - real payment intent and approval flow
  - no hidden spending autonomy

## MVP Scope

### Must-have

1. Keep the existing trust demo intact
2. Add a cooperation demo
3. Add a minimal payment approval flow
4. Export all of it as one artifact bundle

### Nice-to-have

- onchain transaction execution
- wallet signing integration
- richer payment policy engine
- agent-to-agent payment transfer

Do not block the submission on the nice-to-have list.

## Product Shape

### User flow

1. User gives a task
2. Coordinator decides:
   - what to inspect
   - which subagent handles what
   - whether a payment intent is needed
3. Inspector gathers facts
4. Planner proposes action
5. Payer prepares payment intent
6. Human approves or denies
7. If approved, payment executes and receipt is stored
8. Final output + trust/cooperation/payment trace is exported

## Payment Design

### Minimal real payment object

Fields:

- `intent_id`
- `reason`
- `amount`
- `currency_or_token`
- `destination`
- `budget_limit`
- `status`
- `approved_by_user`
- `tx_hash`
- `created_at`
- `executed_at`

### Required statuses

- `draft`
- `awaiting_approval`
- `approved`
- `rejected`
- `executed`
- `failed`

### Safety rules

- no payment executes without explicit approval
- payment amount must be <= configured budget
- destination must be visible before approval
- every execution must produce a receipt

## Cooperation Design

### Minimal coordination object

Fields:

- `task_id`
- `coordinator_goal`
- `subagents`
- `delegations`
- `outputs`
- `final_summary`

### Required evidence

- which subagent did what
- what scope each subagent had
- what tool calls each subagent made
- how the coordinator combined results

## Deliverables

Add these hackathon-facing artifacts:

- `COOPERATION_MODEL.md`
- `PAYMENT_MODEL.md`
- `demo_runner_coop_pay.sh`
- exported artifact bundle containing:
  - trust evidence
  - subagent coordination trace
  - payment intent / approval / receipt

## Build Order

### Phase A

- lock current trust demo
- keep canonical artifact as fallback

### Phase B

- implement coordinator + inspector + planner flow
- generate cooperation trace

### Phase C

- implement payment intent model
- human approval step
- receipt/export

### Phase D

- merge all evidence into one polished demo run
- update submission docs and video plan

## Success Criteria

The extended submission is good enough when:

- trust still works
- at least one subagent delegation is visible and useful
- at least one payment intent is created and handled end-to-end
- the human approval boundary is explicit
- the exported bundle proves all of the above

## Recommendation

Yes, build payment too, but keep it minimal and real.

The winning version is not:

- “the agent can spend money freely”

The winning version is:

- “the agent can prepare and complete a real payment under human-approved bounded authority, and the whole process is auditable.”
