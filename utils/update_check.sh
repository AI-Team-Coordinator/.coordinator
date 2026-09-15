#!/bin/bash
# Daily Coordinator update check (origin/main commit date).
# Cache: $DATA_DIR/.cache/last_update_check
# Usage: ./utils/update_check.sh [--force] [--notice]

set -euo pipefail
DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"
python3 "$DIR/utils/update_check.py" "$@"
