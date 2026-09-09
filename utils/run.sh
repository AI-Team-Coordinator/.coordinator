#!/bin/bash
# Local dashboard: detached Go API (:4321) + Vite HMR (:5175).
# Look at http://localhost:5175 — API stays up after the chat ends.
# After Go edits: ./utils/backend.sh restart
# After UI edits: Vite reloads itself; do not restart.

set -e

DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"

echo "=================================================="
echo " AI Team Coordinator — local UI"
echo "=================================================="

"$DIR/utils/backend.sh" start
"$DIR/utils/frontend.sh" start

echo "Dashboard: http://localhost:5175"
echo "API:       http://127.0.0.1:4321/health"
