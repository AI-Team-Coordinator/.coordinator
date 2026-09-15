#!/bin/bash
# Vite HMR on $UI_PORT (default 5175). Alina Assist only (UI_MODE=vite).
# Does not restart the Go API.
# Usage: ./utils/frontend.sh {start|stop|status}

set -e

DIR="$(cd "$(dirname "$0")/.." && pwd)"
export COORDINATOR_ROOT="$DIR"
# shellcheck source=paths.sh
. "$(cd "$(dirname "$0")" && pwd)/paths.sh"

FRONTEND_DIR="$DIR/frontend"
CACHE="$DIR/.cache"
PIDFILE="$CACHE/frontend.pid"
LOG="$CACHE/frontend.log"
DAEMONIZE="$DIR/utils/daemonize.py"
VITE_PORT="${UI_PORT:-5175}"

assert_vite_mode() {
    if [ "$(coordinator_ui_mode)" != "vite" ]; then
        echo "Vite is only for Alina Assist (UI_MODE=vite). Board: $(coordinator_dashboard_url)"
        exit 1
    fi
}

listen_pids() {
    lsof -nP -tiTCP:"$VITE_PORT" -sTCP:LISTEN 2>/dev/null || true
}

vite_up() {
    local pids
    pids=$(listen_pids)
    [ -n "$pids" ]
}

cmd_status() {
    if vite_up; then
        echo "ok :$VITE_PORT $(listen_pids)"
        return 0
    fi
    echo "down :$VITE_PORT"
    return 1
}

cmd_stop() {
    local pids
    pids=$(listen_pids)
    if [ -z "$pids" ]; then
        rm -f "$PIDFILE"
        return 0
    fi
    echo "🔄 Stopping Vite on :$VITE_PORT (PID $pids)..."
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
    fi
    rm -f "$PIDFILE"
}

cmd_start() {
    assert_vite_mode
    if vite_up; then
        echo "⚡ Vite already on http://127.0.0.1:$VITE_PORT"
        return 0
    fi
    if [ ! -d "$FRONTEND_DIR/node_modules" ]; then
        echo "📦 Installing frontend dependencies..."
        (cd "$FRONTEND_DIR" && npm install)
    fi
    mkdir -p "$CACHE"
    echo "⚡ Starting Vite on http://127.0.0.1:$VITE_PORT (detached)..."
    python3 "$DAEMONIZE" "$PIDFILE" "$LOG" "$FRONTEND_DIR" -- npm run dev
    local i
    for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do
        if vite_up; then
            echo "✓ UI http://127.0.0.1:$VITE_PORT  log $LOG"
            return 0
        fi
        sleep 0.4
    done
    echo "Vite failed to listen on :$VITE_PORT. Last log:"
    tail -n 40 "$LOG" 2>/dev/null || true
    return 1
}

case "${1:-start}" in
    start) cmd_start ;;
    stop) cmd_stop ;;
    status) cmd_status ;;
    *)
        echo "Usage: $0 {start|stop|status}"
        exit 1
        ;;
esac
