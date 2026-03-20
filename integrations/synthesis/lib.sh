#!/usr/bin/env bash
set -euo pipefail

openclio_base_url() {
  if [[ -n "${OPENCLIO_BASE_URL:-}" ]]; then
    printf '%s' "${OPENCLIO_BASE_URL}"
    return
  fi

  local config_path port
  config_path="$(openclio_data_dir)/config.yaml"
  if [[ -f "${config_path}" ]]; then
    port="$(python3 - <<'PY' "${config_path}" 2>/dev/null || true
from pathlib import Path
import re
import sys

text = Path(sys.argv[1]).read_text(encoding="utf-8")
match = re.search(r'(?m)^[ \t]*port:[ \t]*([0-9]+)\b', text)
if match:
    print(match.group(1))
PY
)"
    if [[ -n "${port}" ]]; then
      printf 'http://127.0.0.1:%s' "${port}"
      return
    fi
  fi

  printf '%s' "http://127.0.0.1:18789"
}

openclio_data_dir() {
  if [[ -n "${OPENCLIO_DATA_DIR:-}" ]]; then
    printf '%s' "${OPENCLIO_DATA_DIR}"
    return
  fi
  printf '%s' "${HOME}/.openclio"
}

openclio_token() {
  if [[ -n "${OPENCLIO_TOKEN:-}" ]]; then
    printf '%s' "${OPENCLIO_TOKEN}"
    return
  fi

  local token_path
  token_path="$(openclio_data_dir)/auth.token"
  if [[ -f "${token_path}" ]]; then
    tr -d '\r\n' < "${token_path}"
    return
  fi

  return 1
}

require_openclio_token() {
  if ! openclio_token >/dev/null 2>&1; then
    cat >&2 <<'EOF'
missing OpenClio auth token

Set OPENCLIO_TOKEN or make sure ~/.openclio/auth.token exists.
EOF
    exit 1
  fi
}

openclio_auth_header() {
  printf 'Authorization: Bearer %s' "$(openclio_token)"
}

require_json_tooling() {
  command -v curl >/dev/null 2>&1 || {
    echo "curl is required" >&2
    exit 1
  }
  command -v python3 >/dev/null 2>&1 || {
    echo "python3 is required" >&2
    exit 1
  }
}

openclio_repo_root() {
  local script_dir
  script_dir="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
  cd -- "${script_dir}/../.." && pwd
}
