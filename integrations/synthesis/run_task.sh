#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=integrations/synthesis/lib.sh
source "${SCRIPT_DIR}/lib.sh"

require_json_tooling
require_openclio_token

session_id=""
message=""

while [[ $# -gt 0 ]]; do
  case "$1" in
    -s|--session)
      session_id="${2:-}"
      shift 2
      ;;
    *)
      if [[ -n "${message}" ]]; then
        echo "unexpected extra argument: $1" >&2
        exit 1
      fi
      message="$1"
      shift
      ;;
  esac
done

if [[ -z "${message}" ]]; then
  cat >&2 <<'EOF'
usage: integrations/synthesis/run_task.sh [-s SESSION_ID] "message"
EOF
  exit 1
fi

base_url="$(openclio_base_url)"
repo_root="$(openclio_repo_root)"
connect_timeout="${OPENCLIO_CONNECT_TIMEOUT:-10}"
chat_timeout="${OPENCLIO_CHAT_TIMEOUT:-240}"
chat_mode="${OPENCLIO_CHAT_MODE:-auto}"
sentinel_mode="${OPENCLIO_SENTINEL_MODE:-1}"
payload="$(python3 - <<'PY' "${message}" "${session_id}" "${sentinel_mode}" "${repo_root}"
import json
import re
import sys

message = sys.argv[1]
session_id = sys.argv[2]
sentinel_mode = sys.argv[3] != "0"
repo_root = sys.argv[4]

policy = (
    "You are OpenClio Sentinel, a trusted local operator agent for the Synthesis "
    "hackathon. Work only inside the repository workspace at "
    f"`{repo_root}`. "
    "Do not modify files, do not delete files, do not rewrite config, and do not "
    "run destructive or state-changing shell commands. "
    "You may use the delegate tool only for read-only, role-scoped cooperation "
    "tasks where subagents inspect or plan without mutating the repository. "
    "If the user asks for destructive or mutating actions, refuse direct execution, "
    "explain the trust boundary briefly, and provide a safe read-only hardening or "
    "cleanup plan instead. Prefer concise, auditable answers."
)

destructive_patterns = [
    r"\bdelete\b",
    r"\bremove\b",
    r"\brewrite\b",
    r"\boverwrite\b",
    r"\bedit\b",
    r"\bmodify\b",
    r"\bwrite\b",
    r"\bfix\b.*\bby changing\b",
    r"\brm\b",
    r"\bmv\b",
    r"\bchmod\b",
    r"\bchown\b",
]

effective_message = message
if sentinel_mode:
    is_destructive = any(re.search(pattern, message, re.IGNORECASE) for pattern in destructive_patterns)
    if is_destructive:
        effective_message = (
            f"{policy}\n\n"
            "The human asked for a destructive or mutating action. "
            "Do not perform it. Instead answer in four parts: "
            "(1) what you refuse to do, "
            "(2) why that refusal improves trust, "
            "(3) the safest read-only inspection steps you would take, and "
            "(4) a concrete hardening plan the human can apply manually.\n\n"
            f"Original user request: {message}"
        )
    else:
        effective_message = f"{policy}\n\nUser task: {message}"

payload = {"message": effective_message}
if session_id:
    payload["session_id"] = session_id
print(json.dumps(payload))
PY
)"

echo "Sending task to ${base_url}/api/v1/chat (timeout: ${chat_timeout}s)..." >&2

if [[ "${chat_mode}" == "auto" ]]; then
  if [[ -t 1 ]]; then
    chat_mode="stream"
  else
    chat_mode="buffered"
  fi
fi

if [[ "${chat_mode}" == "buffered" ]]; then
  set +e
  response_json="$(curl -fsS \
    --connect-timeout "${connect_timeout}" \
    --max-time "${chat_timeout}" \
    -H "$(openclio_auth_header)" \
    -H "Content-Type: application/json" \
    -d "${payload}" \
    "${base_url}/api/v1/chat")"
  status=$?
  set -e

  if [[ ${status} -ne 0 ]]; then
    echo "OpenClio buffered chat request failed or timed out after ${chat_timeout}s." >&2
    exit "${status}"
  fi

  printf '%s' "${response_json}" | python3 -c 'import json,sys; print(json.dumps(json.load(sys.stdin), indent=2))'
  exit 0
fi

set +e
curl -NfsS \
  --connect-timeout "${connect_timeout}" \
  --max-time "${chat_timeout}" \
  -H "$(openclio_auth_header)" \
  -H "Content-Type: application/json" \
  -d "${payload}" \
  "${base_url}/api/v1/chat?stream=true" | python3 /dev/fd/3 3<<'PY'
from __future__ import annotations

import json
import sys

current_event = "message"
current_data: list[str] = []
response_parts: list[str] = []
tools_used: list[dict[str, object]] = []
agent_events: list[dict[str, object]] = []
session_id = ""
done = False
line_buffer = ""


def flush_event(event: str, data_lines: list[str]) -> None:
    global session_id, done
    if not data_lines and event != "done":
        return

    payload = "\n".join(data_lines)
    normalized = event or "message"

    if normalized == "tool_use":
        try:
            obj = json.loads(payload)
        except json.JSONDecodeError:
            print(f"[tool] {payload}", file=sys.stderr, flush=True)
            tools_used.append({"name": payload})
            return
        tool_name = str(obj.get("tool", "unknown"))
        print(f"[tool] {tool_name}", file=sys.stderr, flush=True)
        tools_used.append({"name": tool_name})
        return

    if normalized == "agent_event":
        try:
            obj = json.loads(payload)
        except json.JSONDecodeError:
            print(f"[agent] {payload}", file=sys.stderr, flush=True)
            agent_events.append({"type": "unknown", "raw": payload})
            return
        evt_type = str(obj.get("type", "unknown"))
        label = evt_type.replace("_", " ")
        if evt_type == "subagent_spawn":
            detail = str(obj.get("task", ""))
            if detail:
                print(f"[agent] {label}: {detail}", file=sys.stderr, flush=True)
            else:
                print(f"[agent] {label}", file=sys.stderr, flush=True)
        elif evt_type == "subagent_tool":
            detail = str(obj.get("tool_name", ""))
            if detail:
                print(f"[agent] {label}: {detail}", file=sys.stderr, flush=True)
            else:
                print(f"[agent] {label}", file=sys.stderr, flush=True)
        else:
            print(f"[agent] {label}", file=sys.stderr, flush=True)
        agent_events.append(obj)
        return

    if normalized == "error":
        print(f"[error] {payload}", file=sys.stderr, flush=True)
        raise RuntimeError(payload or "OpenClio stream returned an error event")

    if normalized == "done":
        if payload:
            try:
                obj = json.loads(payload)
            except json.JSONDecodeError:
                obj = {}
            session_id = str(obj.get("session_id", session_id))
        done = True
        return

    try:
        obj = json.loads(payload)
    except json.JSONDecodeError:
        text = payload
    else:
        text = obj.get("text", payload)

    text = "" if text is None else str(text)
    if text:
        response_parts.append(text)
        print(text, end="", file=sys.stderr, flush=True)


try:
    for raw_line in sys.stdin:
        line = raw_line.rstrip("\r\n")
        if line.startswith("event:"):
            current_event = line.split(":", 1)[1].strip() or "message"
            continue
        if line.startswith("data:"):
            current_data.append(line.split(":", 1)[1].lstrip())
            continue
        if line == "":
            flush_event(current_event, current_data)
            current_event = "message"
            current_data = []
    if current_data:
        flush_event(current_event, current_data)
except RuntimeError as exc:
    print(file=sys.stderr)
    print(f"OpenClio chat failed: {exc}", file=sys.stderr)
    sys.exit(1)

print(file=sys.stderr)

if not done:
    print("OpenClio stream ended before the done event arrived.", file=sys.stderr)
    sys.exit(1)

if not session_id:
    print("OpenClio stream completed without a session_id.", file=sys.stderr)
    sys.exit(1)

print(json.dumps({
    "response": "".join(response_parts),
    "session_id": session_id,
    "tools_used": tools_used,
    "agent_events": agent_events,
}, indent=2))
PY
status=$?
set -e

if [[ ${status} -ne 0 ]]; then
  echo "OpenClio chat request failed or timed out after ${chat_timeout}s." >&2
  exit "${status}"
fi
