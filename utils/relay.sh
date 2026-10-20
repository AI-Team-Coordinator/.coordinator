#!/usr/bin/env bash
# Start local Relay for development (default :4322, writes to Common/data).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
RELAY_DIR="$ROOT/relay"
DATA_DIR="${RELAY_DATA:-}"
if [[ -z "$DATA_DIR" ]]; then
  if [[ -f "$ROOT/../Common/data/settings/project_profile.json" ]]; then
    DATA_DIR="$(cd "$ROOT/../Common/data" && pwd)"
  else
    DATA_DIR="$ROOT/../Common/data"
  fi
fi

cmd="${1:-start}"
case "$cmd" in
  start)
    cd "$RELAY_DIR"
    exec go run . -port 4322 -data "$DATA_DIR"
    ;;
  *)
    echo "Usage: $0 start" >&2
    exit 1
    ;;
esac
