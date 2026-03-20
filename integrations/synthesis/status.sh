#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
# shellcheck source=integrations/synthesis/lib.sh
source "${SCRIPT_DIR}/lib.sh"

require_json_tooling
require_openclio_token

base_url="$(openclio_base_url)"
config_path="$(openclio_data_dir)/config.yaml"

health_json="$(curl -fsS "${base_url}/api/v1/health")"
overview_json="$(curl -fsS \
  -H "$(openclio_auth_header)" \
  "${base_url}/api/v1/overview")"

python3 - <<'PY' "${health_json}" "${overview_json}" "${base_url}" "${config_path}"
import json
from pathlib import Path
import re
import sys

health = json.loads(sys.argv[1])
overview = json.loads(sys.argv[2])
base_url = sys.argv[3]
config_path = Path(sys.argv[4])

def config_value(key: str) -> str | None:
    if not config_path.exists():
        return None
    text = config_path.read_text(encoding="utf-8")
    match = re.search(rf'(?m)^[ \t]*{re.escape(key)}:[ \t]*"?([^"\n]+)"?[ \t]*$', text)
    if match:
        value = match.group(1).strip()
        if " #" in value:
            value = value.split(" #", 1)[0].rstrip()
        if value.startswith('"') and value.endswith('"'):
            value = value[1:-1]
        return value.strip() or None
    return None

provider = overview.get("model", {}).get("provider") or config_value("provider")
model = overview.get("model", {}).get("name") or config_value("model")
memory_mode = overview.get("tooling", {}).get("memory_mode") or config_value("memory_mode")
if memory_mode is None:
    memory_mode = "unset"

summary = {
    "base_url": base_url,
    "health": health,
    "overview": {
        "provider": provider,
        "model": model,
        "tool_packs": overview.get("tooling", {}).get("packs", []),
        "memory_mode": memory_mode,
        "channels": overview.get("channels", []),
    },
}

print(json.dumps(summary, indent=2))
PY
