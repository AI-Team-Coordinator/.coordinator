#!/bin/bash
# Local dashboard. Detached API on $PORT.
# Alina Assist (UI_MODE=vite): also Vite HMR on $UI_PORT — look at that URL.
# Installs (static): built dist served on $PORT — look at http://127.0.0.1:$PORT
# After Go edits: ./utils/backend.sh restart
# After UI edits (vite): Vite reloads itself. After UI edits (static): backend.sh restart rebuilds dist.

set -e

DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$DIR/utils/paths.sh"

echo "=================================================="
echo " AI Team Coordinator — local UI"
echo "=================================================="

"$DIR/utils/backend.sh" start
if [ "$(coordinator_ui_mode)" = "vite" ]; then
    "$DIR/utils/frontend.sh" start
fi

echo "Dashboard: $(coordinator_dashboard_url)"
echo "API:       http://127.0.0.1:${PORT}/health"
