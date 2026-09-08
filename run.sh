#!/bin/bash

# Launch script for AI Team Coordinator Go Engine and React Dashboard.
# Usage: ./run.sh [PORT]
#        ./run.sh --dev     Go API + Vite HMR (edits refresh immediately)

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$DIR/paths.sh"

if [ "${1:-}" = "dev" ] || [ "${1:-}" = "--dev" ]; then
    exec "$DIR/dev.sh"
fi

PORT=${1:-${PORT:-4321}}
BACKEND_DIR="$DIR/backend"
WEB_DIR="$DIR/web"
BIN="$BACKEND_DIR/coordinator-server"

echo "=================================================="
echo " 🚀 AI Team Coordinator Dashboard"
echo "=================================================="

echo "📦 Compiling Go API server..."
(cd "$BACKEND_DIR" && go build -o coordinator-server .)

if [ ! -d "$WEB_DIR/dist" ]; then
    echo "📦 Building React web assets..."
    (cd "$WEB_DIR" && npm run build)
fi

PID=$(lsof -ti :$PORT 2>/dev/null || true)
if [ -n "$PID" ]; then
    echo "🔄 Stopping previous instance on port $PORT (PID $PID)..."
    kill "$PID" 2>/dev/null || true
    sleep 1
fi

echo "📂 data $DATA_DIR"
echo "🌐 Starting server on http://localhost:$PORT..."

(sleep 1 && open "http://localhost:$PORT") &

exec "$BIN" -port "$PORT" -web "$WEB_DIR"
