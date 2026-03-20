#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=integrations/synthesis/lib.sh
source "${SCRIPT_DIR}/lib.sh"

require_json_tooling
require_openclio_token

session_id="${1:-}"
if [[ -z "${session_id}" ]]; then
  cat >&2 <<'EOF'
usage: integrations/synthesis/export_session.sh SESSION_ID [output.json]
EOF
  exit 1
fi

output_path="${2:-}"
base_url="$(openclio_base_url)"

session_json="$(curl -fsS \
  -H "$(openclio_auth_header)" \
  "${base_url}/api/v1/sessions/${session_id}")"

stats_json="$(curl -fsS \
  -H "$(openclio_auth_header)" \
  "${base_url}/api/v1/sessions/${session_id}/stats")"

merged="$(python3 - <<'PY' "${session_json}" "${stats_json}"
import json
import sys

session_payload = json.loads(sys.argv[1])
stats_payload = json.loads(sys.argv[2])

combined = {
    "session_id": (
        session_payload.get("session", {}) or {}
    ).get("id") or stats_payload.get("session_id") or sys.argv[1],
    "session": session_payload.get("session"),
    "messages": session_payload.get("messages", []),
    "stats": stats_payload,
}
print(json.dumps(combined, indent=2))
PY
)"

if [[ -n "${output_path}" ]]; then
  printf '%s\n' "${merged}" > "${output_path}"
  echo "wrote ${output_path}"
else
  printf '%s\n' "${merged}"
fi
