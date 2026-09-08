#!/bin/bash

# Rebuild Go + UI and serve the dashboard on :4321 (or PORT / first arg).
# Usage: ./run.sh [PORT]
# Look at http://localhost:4321 — not Vite :5175.

set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$DIR/paths.sh"

PORT=${1:-${PORT:-4321}}
BACKEND_DIR="$DIR/backend"
WEB_DIR="$DIR/web"
BIN="$BACKEND_DIR/coordinator-server"

echo "=================================================="
echo " 🚀 AI Team Coordinator Dashboard"
echo "=================================================="

echo "📦 Compiling Go API server..."
(cd "$BACKEND_DIR" && go build -o coordinator-server .)

if [ ! -d "$WEB_DIR/node_modules" ]; then
    echo "📦 Installing frontend dependencies..."
    (cd "$WEB_DIR" && npm install)
fi

echo "📦 Building React web assets..."
(cd "$WEB_DIR" && npm run build)

stop_port() {
    local port="$1"
    local pids
    pids=$(lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)
    if [ -z "$pids" ]; then
        return 0
    fi
    echo "🔄 Stopping previous instance on port $port (PID $pids)..."
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
    local i
    for i in 1 2 3 4 5 6 7 8 9 10; do
        if ! lsof -nP -tiTCP:"$port" -sTCP:LISTEN >/dev/null 2>&1; then
            return 0
        fi
        sleep 0.3
    done
    echo "⚠️  Port $port still busy, killing..."
    pids=$(lsof -nP -tiTCP:"$port" -sTCP:LISTEN 2>/dev/null || true)
    if [ -n "$pids" ]; then
        # shellcheck disable=SC2086
        kill -9 $pids 2>/dev/null || true
        sleep 0.3
    fi
}

stop_port "$PORT"
stop_port 5175

echo "📂 data $DATA_DIR"
echo "🌐 Starting server on http://localhost:$PORT..."

(sleep 1 && open "http://localhost:$PORT") &

exec "$BIN" -port "$PORT" -web "$WEB_DIR"
