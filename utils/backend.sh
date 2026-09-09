#!/bin/bash
# Go API on :4321, detached from the Cursor chat (new session / nohup).
# Does not touch Vite :5175.
# Usage: ./utils/backend.sh {start|stop|restart|status}

set -e

DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$(cd "$(dirname "$0")" && pwd)/paths.sh"

PORT="${PORT:-4321}"
BACKEND_DIR="$DIR/backend"
WEB_DIR="$DIR/web"
BIN="$BACKEND_DIR/coordinator-server"
CACHE="$DIR/.cache"
PIDFILE="$CACHE/backend.pid"
LOG="$CACHE/backend.log"
DAEMONIZE="$DIR/utils/daemonize.py"

healthy() {
    local code
    code=$(curl -sf -o /dev/null -w "%{http_code}" "http://127.0.0.1:${PORT}/health" 2>/dev/null || true)
    [ "$code" = "200" ]
}

listen_pids() {
    lsof -nP -tiTCP:"$PORT" -sTCP:LISTEN 2>/dev/null || true
}

stop_port() {
    local pids
    pids=$(listen_pids)
    if [ -z "$pids" ]; then
        rm -f "$PIDFILE"
        return 0
    fi
    echo "🔄 Stopping coordinator API on :$PORT (PID $pids)..."
    # shellcheck disable=SC2086
    kill $pids 2>/dev/null || true
    local i
    for i in 1 2 3 4 5 6 7 8 9 10; do
        if [ -z "$(listen_pids)" ]; then
            rm -f "$PIDFILE"
            return 0
        fi
        sleep 0.3
    done
    pids=$(listen_pids)
    if [ -n "$pids" ]; then
        # shellcheck disable=SC2086
        kill -9 $pids 2>/dev/null || true
        sleep 0.3
    fi
    rm -f "$PIDFILE"
}

cmd_status() {
    if healthy; then
        echo "ok :$PORT $(listen_pids)"
        return 0
    fi
    echo "down :$PORT"
    return 1
}

cmd_start() {
    if healthy; then
        echo "🌐 API already on http://127.0.0.1:$PORT"
        return 0
    fi
    echo "📦 Compiling Go API server..."
    (cd "$BACKEND_DIR" && go build -o coordinator-server .)
    mkdir -p "$CACHE"
    echo "📂 data $DATA_DIR"
    echo "🌐 Starting API on http://127.0.0.1:$PORT (detached)..."
    python3 "$DAEMONIZE" "$PIDFILE" "$LOG" "$BACKEND_DIR" -- "$BIN" -port "$PORT" -web "$WEB_DIR"
    local i
    for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15; do
        if healthy; then
            echo "✓ API http://127.0.0.1:$PORT  log $LOG"
            return 0
        fi
        sleep 0.4
    done
    echo "API failed to become healthy. Last log:"
    tail -n 40 "$LOG" 2>/dev/null || true
    return 1
}

cmd_stop() {
    stop_port
}

cmd_restart() {
    stop_port
    cmd_start
}

case "${1:-start}" in
    start) cmd_start ;;
    stop) cmd_stop ;;
    restart) cmd_restart ;;
    status) cmd_status ;;
    *)
        echo "Usage: $0 {start|stop|restart|status}"
        exit 1
        ;;
esac
