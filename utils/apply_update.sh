#!/bin/bash
# After the human says yes to a Coordinator update:
# pull origin/main, rebuild API + UI, restart on this project's PORT / UI_PORT from .env.
# Usage: ./utils/apply_update.sh

set -euo pipefail
DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$(cd "$(dirname "$0")" && pwd)/paths.sh"

cd "$DIR"
echo "⬇  git pull --ff-only origin main"
git fetch --quiet origin main
git pull --ff-only origin main

MODE="$(coordinator_ui_mode)"
echo "📦 UI_MODE=$MODE  PORT=$PORT  UI_PORT=$UI_PORT"

if [ "$MODE" = "vite" ]; then
    if [ ! -d "$DIR/frontend/node_modules" ]; then
        echo "📦 Installing frontend dependencies..."
        (cd "$DIR/frontend" && npm install)
    fi
    echo "📦 Rebuilding frontend..."
    (cd "$DIR/frontend" && npm run build)
fi

echo "📦 Rebuilding API and restarting on :$PORT ..."
"$DIR/utils/backend.sh" restart

if [ "$MODE" = "vite" ]; then
    echo "⚡ Restarting Vite on :$UI_PORT ..."
    "$DIR/utils/frontend.sh" restart
fi

echo "✓ Board $(coordinator_dashboard_url)"
