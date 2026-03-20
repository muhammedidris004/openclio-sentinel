#!/usr/bin/env bash
set -euo pipefail

artifact_dir="${1:-}"
if [[ -z "${artifact_dir}" ]]; then
  cat >&2 <<'EOF'
usage: integrations/synthesis/verify_artifact.sh ARTIFACT_DIR
EOF
  exit 1
fi

for required in SUMMARY.md chat_response.json session_export.json conversation_log.json tool_trace.json status.json TRUST_MODEL.md ARCHITECTURE.md SUBMISSION.md; do
  if [[ ! -f "${artifact_dir}/${required}" ]]; then
    echo "missing required artifact: ${required}" >&2
    exit 1
  fi
done

python3 - <<'PY' "${artifact_dir}"
import json
from pathlib import Path
import sys

artifact = Path(sys.argv[1])
summary = (artifact / "SUMMARY.md").read_text(encoding="utf-8")
session_export = json.loads((artifact / "session_export.json").read_text(encoding="utf-8"))
tool_trace = json.loads((artifact / "tool_trace.json").read_text(encoding="utf-8"))
status = json.loads((artifact / "status.json").read_text(encoding="utf-8"))

messages = session_export.get("messages", [])
stats = session_export.get("stats", {})
message_count = stats.get("counts", {}).get("messages", len(messages))
token_count = stats.get("counts", {}).get("tokens", sum((m.get("tokens") or 0) for m in messages))

problems = []
if message_count < 2:
    problems.append(f"expected at least 2 messages, got {message_count}")
if token_count <= 0:
    problems.append(f"expected positive token count, got {token_count}")
if tool_trace.get("tool_count", 0) <= 0:
    problems.append("expected at least one tool call in tool_trace.json")
if status.get("health", {}).get("status") != "ok":
    problems.append("status.json health is not ok")
for required_phrase in ("Trust Findings", "Runtime Evidence", "Agent Output"):
    if required_phrase not in summary:
        problems.append(f"SUMMARY.md missing section: {required_phrase}")

if problems:
    for p in problems:
        print(f"- {p}")
    sys.exit(1)

print(json.dumps({
    "artifact": str(artifact),
    "message_count": message_count,
    "token_count": token_count,
    "tool_count": tool_trace.get("tool_count", 0),
    "blocked_count": tool_trace.get("blocked_count", 0),
    "health": status.get("health", {}).get("status"),
}, indent=2))
PY
