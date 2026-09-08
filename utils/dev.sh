#!/bin/bash

# Optional Vite HMR for React source (port 5175). Not the dashboard URL.
# To look at the coordinator, use ./run.sh → http://localhost:4321

set -e

DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$(cd "$(dirname "$0")" && pwd)/paths.sh"
BACKEND_DIR="$DIR/backend"
WEB_DIR="$DIR/web"
BIN="$BACKEND_DIR/coordinator-server"
API_PORT="${PORT:-4321}"

echo "=================================================="
echo " AI Team Coordinator — dev (Vite HMR)"
echo "=================================================="

if ! lsof -nP -iTCP:"$API_PORT" -sTCP:LISTEN >/dev/null 2>&1; then
    echo "📦 Compiling Go API server..."
    (cd "$BACKEND_DIR" && go build -o coordinator-server .)
    echo "🌐 API on http://127.0.0.1:$API_PORT"
    echo "📂 data $DATA_DIR"
    "$BIN" -port "$API_PORT" -web "$WEB_DIR" &
    API_PID=$!
    trap 'kill "$API_PID" 2>/dev/null || true' EXIT INT TERM
    sleep 1
else
    echo "🌐 API already running on http://127.0.0.1:$API_PORT"
fi

if [ ! -d "$WEB_DIR/node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    (cd "$WEB_DIR" && npm install)
fi

echo "⚡ Vite HMR on http://127.0.0.1:5175 (do not test here — use ./run.sh :4321)"
cd "$WEB_DIR"
exec npm run dev
