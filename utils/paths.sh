#!/bin/bash
# Resolve data-bus paths from .coordinator/.env.
# Paths in .env are relative to the coordinator app root unless absolute.
# This file lives in utils/; the app root is the parent directory.

_coord_script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
if [ "$(basename "$_coord_script_dir")" = "utils" ]; then
    _coord_default_root="$(cd "$_coord_script_dir/.." && pwd)"
else
    _coord_default_root="$_coord_script_dir"
fi
if [ -n "${COORDINATOR_ROOT:-}" ] && [ "$(basename "$COORDINATOR_ROOT")" = "utils" ]; then
    COORDINATOR_ROOT="$(cd "$COORDINATOR_ROOT/.." && pwd)"
fi
COORDINATOR_ROOT="${COORDINATOR_ROOT:-$_coord_default_root}"

_coord_load_env() {
    local f="$COORDINATOR_ROOT/.env"
    if [ ! -f "$f" ]; then
        return 0
    fi
    while IFS= read -r line || [ -n "$line" ]; do
        case "$line" in
            ''|\#*) continue ;;
        esac
        key=${line%%=*}
        val=${line#*=}
        key=$(printf '%s' "$key" | sed 's/[[:space:]]*$//')
        val=$(printf '%s' "$val" | sed 's/^[[:space:]]*//;s/^["'\'']//;s/["'\'']$//')
        if [ -n "$key" ] && [ -z "${!key+x}" ]; then
            export "$key=$val"
        fi
    done < "$f"
}

_coord_abs() {
    local p=$1
    if [ -z "$p" ]; then
        echo ""
        return
    fi
    if [ "${p#/}" != "$p" ]; then
        echo "$p"
        return
    fi
    (cd "$COORDINATOR_ROOT" && cd "$p" && pwd)
}

_coord_load_env

WORKSPACE_ROOT="$(_coord_abs "${WORKSPACE_ROOT:-..}")"
COMMON_ROOT="$(_coord_abs "${BUS_DIR:-${COMMON_DIR:-../Common}}")"
DATA_DIR="${DATA_DIR:-}"
if [ -n "$DATA_DIR" ]; then
    DATA_DIR="$(_coord_abs "$DATA_DIR")"
else
    DATA_DIR="${COMMON_ROOT}/data"
fi
PROGRESS_DIR="${DATA_DIR}/progress"
EVENTS_DIR="${PROGRESS_DIR}/events"
SETTINGS_DIR="${DATA_DIR}/settings"
AUTHOR_FILE="${DATA_DIR}/.current_author"
SYNC_LOG="${PROGRESS_DIR}/.sync.log"
PORT="${PORT:-4321}"
UI_PORT="${UI_PORT:-5175}"
export PORT UI_PORT

coordinator_is_alina_assist() {
    python3 - "$SETTINGS_DIR/project_profile.json" <<'PY'
import json, os, sys
path = sys.argv[1]
ok = False
if os.path.isfile(path):
    try:
        data = json.load(open(path))
        proj = data.get("project") or {}
        pid = str(proj.get("id") or "").strip().lower()
        name = str(proj.get("name") or "").strip().lower()
        ok = pid == "alina-assist" or name == "alina assist"
    except Exception:
        pass
print("yes" if ok else "no")
PY
}

coordinator_ui_mode() {
    local raw
    raw=$(printf '%s' "${UI_MODE:-}" | tr '[:upper:]' '[:lower:]')
    if [ "$raw" = "vite" ] || [ "$raw" = "static" ]; then
        echo "$raw"
        return
    fi
    if [ "$(coordinator_is_alina_assist)" = "yes" ]; then
        echo "vite"
    else
        echo "static"
    fi
}

coordinator_dashboard_url() {
    if [ "$(coordinator_ui_mode)" = "vite" ]; then
        echo "http://localhost:${UI_PORT}"
    else
        echo "http://127.0.0.1:${PORT}"
    fi
}

coordinator_current_author() {
    local alias=""
    if [ -f "$AUTHOR_FILE" ]; then
        alias=$(grep -v '^#' "$AUTHOR_FILE" | grep -v '^[[:space:]]*$' | head -1 | tr -d '[:space:]')
    fi
    local cursor_dir
    cursor_dir="$(_coord_abs "${CURSOR_DIR:-../.cursor}")"
    if [ -z "$alias" ] && [ -f "${cursor_dir}/.current_author" ]; then
        alias=$(grep -v '^#' "${cursor_dir}/.current_author" | grep -v '^[[:space:]]*$' | head -1 | tr -d '[:space:]')
    fi
    echo "$alias"
}

coordinator_collaboration() {
    python3 - "$SETTINGS_DIR/coordinator.json" <<'PY'
import json, os, sys
path = sys.argv[1]
mode = "team"
if os.path.isfile(path):
    try:
        data = json.load(open(path))
        raw = str(data.get("collaboration") or "").strip().lower()
        if raw == "solo":
            mode = "solo"
    except Exception:
        pass
print(mode)
PY
}

coordinator_events_file() {
    local alias=$1
    local year
    year=$(date +%Y)
    mkdir -p "${EVENTS_DIR}/${alias}"
    echo "${EVENTS_DIR}/${alias}/${year}.jsonl"
}
