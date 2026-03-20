#!/usr/bin/env bash
set -euo pipefail

base_url="${SYNTHESIS_BASE_URL:-https://synthesis.devfolio.co}"

required_vars=(
  HUMAN_NAME
  HUMAN_EMAIL
  HUMAN_BACKGROUND
  HUMAN_CRYPTO_EXPERIENCE
  HUMAN_AI_AGENT_EXPERIENCE
  HUMAN_CODING_COMFORT
  HUMAN_PROBLEM_TO_SOLVE
)

for var_name in "${required_vars[@]}"; do
  if [[ -z "${!var_name:-}" ]]; then
    echo "missing required env var: ${var_name}" >&2
    exit 1
  fi
done

agent_name="${AGENT_NAME:-OpenClio}"
agent_description="${AGENT_DESCRIPTION:-A local-first memory-native AI agent runtime focused on trusted execution, private memory, and human-controlled agent workflows.}"
agent_image="${AGENT_IMAGE:-}"
agent_model="${AGENT_MODEL:-gpt-oss:20b}"

payload="$(
python3 - <<'PY'
import json
import os

payload = {
    "name": os.environ.get("AGENT_NAME", "OpenClio"),
    "description": os.environ.get(
        "AGENT_DESCRIPTION",
        "A local-first memory-native AI agent runtime focused on trusted execution, private memory, and human-controlled agent workflows.",
    ),
    "agentHarness": "other",
    "agentHarnessOther": "openclio",
    "model": os.environ.get("AGENT_MODEL", "gpt-oss:20b"),
    "humanInfo": {
        "name": os.environ["HUMAN_NAME"],
        "email": os.environ["HUMAN_EMAIL"],
        "socialMediaHandle": os.environ.get("HUMAN_SOCIAL_HANDLE", ""),
        "background": os.environ["HUMAN_BACKGROUND"],
        "cryptoExperience": os.environ["HUMAN_CRYPTO_EXPERIENCE"],
        "aiAgentExperience": os.environ["HUMAN_AI_AGENT_EXPERIENCE"],
        "codingComfort": int(os.environ["HUMAN_CODING_COMFORT"]),
        "problemToSolve": os.environ["HUMAN_PROBLEM_TO_SOLVE"],
    },
}

image = os.environ.get("AGENT_IMAGE", "").strip()
if image:
    payload["image"] = image

print(json.dumps(payload))
PY
)"

echo "Registering ${agent_name} against ${base_url} ..."
echo

curl -X POST "${base_url}/register" \
  -H "Content-Type: application/json" \
  -d "${payload}"

echo
