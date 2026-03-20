#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=integrations/synthesis/lib.sh
source "${SCRIPT_DIR}/lib.sh"

require_json_tooling
require_openclio_token

repo_root="$(openclio_repo_root)"

prompt_file="${SCRIPT_DIR}/cooperation_task.md"
message="${1:-}"
if [[ -z "${message}" ]]; then
  message="$(python3 - <<'PY' "${prompt_file}"
from pathlib import Path
import sys
for line in Path(sys.argv[1]).read_text(encoding="utf-8").splitlines():
    line = line.strip()
    if line.startswith("> "):
        print(line[2:].strip())
        break
PY
)"
fi

if [[ -z "${message}" ]]; then
  echo "missing cooperation message" >&2
  exit 1
fi

objective="Produce a trusted read-only review of the repository at ${repo_root} using two scoped subagents."
tasks_json="$(python3 - <<'PY' "${repo_root}"
import json
import sys

repo_root = sys.argv[1]
tasks = [
    (
        f"Inspector: inspect only these paths inside {repo_root} using at most 4 read-only tool calls total: "
        "config.yaml, SECURITY.md, README.md, and internal/agent/agent.go if needed. "
        "Do not use search tools, do not delegate again, and return exactly 3 trust or reliability risks with short evidence notes."
    ),
    (
        f"Planner: inspect only these paths inside {repo_root} using at most 3 read-only tool calls total: "
        "README.md, SECURITY.md, and config.yaml. Do not use search tools, do not delegate again, and return exactly 3 safest next steps "
        "a human should take, with one sentence of rationale each."
    ),
]
print(json.dumps(tasks))
PY
)"

timestamp="$(date +%Y%m%d_%H%M%S)"
output_dir="${2:-${SCRIPT_DIR}/artifacts/cooperate_${timestamp}}"
mkdir -p "${output_dir}"

echo "Checking local OpenClio readiness..."
"${SCRIPT_DIR}/status.sh" > "${output_dir}/status.json"

echo "Running cooperation demo task..."
base_url="$(openclio_base_url)"
payload="$(python3 - <<'PY' "${message}" "${objective}" "${tasks_json}"
import json
import sys

message = sys.argv[1]
objective = sys.argv[2]
tasks = json.loads(sys.argv[3])

payload = {
    "objective": objective,
    "tasks": tasks,
}
print(json.dumps(payload))
PY
)"
response_json="$(curl -fsS \
  -H "$(openclio_auth_header)" \
  -H "Content-Type: application/json" \
  -d "${payload}" \
  "${base_url}/api/v1/delegate")"
printf '%s\n' "${response_json}" > "${output_dir}/chat_response.json"

session_id="$(python3 - <<'PY' "${output_dir}/chat_response.json"
import json
import sys
payload = json.load(open(sys.argv[1], "r", encoding="utf-8"))
print(payload.get("session_id", ""))
PY
)"

if [[ -z "${session_id}" ]]; then
  echo "failed to extract session_id from cooperation chat response" >&2
  exit 1
fi

echo "Exporting session ${session_id}..."
"${SCRIPT_DIR}/export_session.sh" "${session_id}" "${output_dir}/session_export.json"

python3 - <<'PY' "${output_dir}" "${message}" "${session_id}"
import json
from pathlib import Path
import sys

out = Path(sys.argv[1])
message = sys.argv[2]
session_id = sys.argv[3]

chat = json.loads((out / "chat_response.json").read_text(encoding="utf-8"))
session = json.loads((out / "session_export.json").read_text(encoding="utf-8"))
status_payload = json.loads((out / "status.json").read_text(encoding="utf-8"))

messages = session.get("messages", [])
stats = session.get("stats", {})
agent_events = chat.get("agent_events") or []

event_summary = {
    "delegation_start_count": sum(1 for e in agent_events if e.get("type") == "delegation_start"),
    "subagent_spawn_count": sum(1 for e in agent_events if e.get("type") == "subagent_spawn"),
    "subagent_tool_count": sum(1 for e in agent_events if e.get("type") == "subagent_tool"),
    "subagent_done_count": sum(1 for e in agent_events if e.get("type") == "subagent_done"),
    "delegation_done_count": sum(1 for e in agent_events if e.get("type") == "delegation_done"),
}

(out / "cooperation_trace.json").write_text(
    json.dumps(
        {
            "session_id": session_id,
            "event_count": len(agent_events),
            "events": agent_events,
            "summary": event_summary,
        },
        indent=2,
    ),
    encoding="utf-8",
)

message_count = stats.get("counts", {}).get("messages", len(messages))
token_count = stats.get("counts", {}).get("tokens", sum((m.get("tokens") or 0) for m in messages))

summary = f"""# OpenClio Sentinel Cooperation Demo Summary

## Task

{message}

## Session

- Session ID: `{session_id}`
- Total messages: {message_count}
- Total tokens: {token_count}
- Provider: `{status_payload.get("overview", {}).get("provider")}`
- Model: `{status_payload.get("overview", {}).get("model")}`

## Cooperation Evidence

- Delegate tool calls observed: 1
- Agent events captured: {len(agent_events)}
- Delegation starts: {event_summary["delegation_start_count"]}
- Subagent spawns: {event_summary["subagent_spawn_count"]}
- Subagent tool events: {event_summary["subagent_tool_count"]}
- Subagent completions: {event_summary["subagent_done_count"]}
- Delegation completions: {event_summary["delegation_done_count"]}

## Final Coordinator Output

{chat.get("response", "").strip()}

## Evidence Files

- `chat_response.json`
- `session_export.json`
- `cooperation_trace.json`
- `status.json`
- `COOPERATION_MODEL.md`
- `SUBMISSION.md`
"""

(out / "SUMMARY.md").write_text(summary, encoding="utf-8")
PY

cp "${SCRIPT_DIR}/COOPERATION_MODEL.md" "${output_dir}/COOPERATION_MODEL.md"
cp "${SCRIPT_DIR}/SUBMISSION.md" "${output_dir}/SUBMISSION.md"

echo
echo "OpenClio Sentinel cooperation demo complete."
echo "Session: ${session_id}"
echo "Artifacts: ${output_dir}"
