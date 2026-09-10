#!/bin/bash

# Helper script to instantly update local coordinator state and asynchronously
# push snapshots to origin/coordinator-state (not Common main).
# Usage: ./sync_event.sh <ALIAS> <EVENT_TYPE> [TASK_ID] [BRANCH_NAME] [Service1,Service2] [doc=...] [summary=...]
# Example: ./sync_event.sh EK task_started 20260907-1756 feature/auth Core,InboxPanelWeb
# Example: ./sync_event.sh EK task_started FIX-... fix/avatar Website summary="Restore dark header avatar"
# Example: ./sync_event.sh EK task_completed 20260907-1756
# Example: ./sync_event.sh EK research_started summary="How coordinator logs off-task chats"
# Example: ./sync_event.sh EK research_completed

set -e

ALIAS=$1
EVENT_TYPE=$2
DOC_PATH=""
SUMMARY=""
SESSION_ID=""
FINDINGS=""
POS=()

if [ -z "$ALIAS" ] || [ -z "$EVENT_TYPE" ]; then
    echo "Usage: $0 <ALIAS> <EVENT_TYPE> [TASK_ID] [BRANCH_NAME] [Service1,Service2] [doc=...] [summary=...]"
    exit 1
fi

shift 2
for arg in "$@"; do
    case "$arg" in
        doc=*|summary=*|session_id=*|findings=*)
            key=${arg%%=*}
            val=${arg#*=}
            case "$key" in
                doc) DOC_PATH=$val ;;
                summary) SUMMARY=$val ;;
                session_id) SESSION_ID=$val ;;
                findings) FINDINGS=$val ;;
            esac
            ;;
        *)
            POS+=("$arg")
            ;;
    esac
done
TASK_ID=${POS[0]:-}
BRANCH_NAME=${POS[1]:-}
SERVICES_CSV=${POS[2]:-}

# shellcheck source=paths.sh
. "$(dirname "$0")/paths.sh"
mkdir -p "$PROGRESS_DIR" "$EVENTS_DIR"

CURRENT_TASK_FILE="$PROGRESS_DIR/.current_task_${ALIAS}"
EVENTS_FILE="$(coordinator_events_file "$ALIAS")"
COORD_DIR="$(cd "$(dirname "$0")" && pwd)"
TIMESTAMP=$(date +%s)
ISO_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

write_current_task() {
    python3 "$COORD_DIR/task_snapshot.py" start "$CURRENT_TASK_FILE" "$ALIAS" "$ISO_DATE" "$TASK_ID" "$BRANCH_NAME" "$SERVICES_CSV" "$COORD_DIR" "$COMMON_ROOT" "$DOC_PATH" "$SUMMARY" "$SESSION_ID"
}

# Nested .research on the snapshot — does not replace in_progress / idle task fields.
# Does not bump root updated_at (that drives task duration on Pulse).
# Exit 2 = already in the requested state (no jsonl line).
apply_research() {
    local action=$1
    python3 - "$CURRENT_TASK_FILE" "$EVENTS_FILE" "$ALIAS" "$ISO_DATE" "$TIMESTAMP" "$action" "$SUMMARY" "$SESSION_ID" "$FINDINGS" "$COORD_DIR" "$DATA_DIR" <<'PY'
import json, os, sys

path, events_file, alias, iso, ts, action, summary, session_id, findings, coord_dir, data_dir = sys.argv[1:]
sys.path.insert(0, coord_dir)

snap = {}
if os.path.isfile(path):
    try:
        snap = json.load(open(path))
    except Exception:
        snap = {}

research = snap.get("research") if isinstance(snap.get("research"), dict) else {}
active = research.get("status") == "active"


def write_snap():
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(snap, f, indent=2)
        f.write("\n")


def append_event(event, extra=None):
    os.makedirs(os.path.dirname(events_file), exist_ok=True)
    row = {"timestamp": int(ts), "event": event, "alias": alias, "spend_kind": "research"}
    if extra:
        row.update({k: v for k, v in extra.items() if v not in (None, "")})
    with open(events_file, "a") as f:
        f.write(json.dumps(row, ensure_ascii=False) + "\n")


def clip(text, n=280):
    text = " ".join((text or "").split())
    if len(text) > n:
        return text[: n - 3].rstrip() + "..."
    return text


def cache_last_text():
    cache_path = os.path.join(data_dir, ".cache", "composer_sessions.json")
    if not os.path.isfile(cache_path):
        return ""
    try:
        cache = json.load(open(cache_path))
    except Exception:
        return ""
    sessions = cache.get("sessions") if isinstance(cache, dict) else None
    if not isinstance(sessions, dict):
        return ""
    sid = (research.get("session_id") or session_id or "").strip()
    if sid and isinstance(sessions.get(sid), dict):
        return str(sessions[sid].get("last_text") or "")
    for rec in sessions.values():
        if isinstance(rec, dict) and rec.get("alias") == alias:
            return str(rec.get("last_text") or "")
    return ""


def take_usage():
    try:
        import cursor_usage
        return cursor_usage.take_snapshot()
    except Exception:
        return None


if action == "start":
    if active:
        if session_id and not research.get("session_id"):
            snap["research"]["session_id"] = session_id
            write_snap()
        raise SystemExit(2)
    snap.setdefault("alias", alias)
    snap.setdefault("status", "idle")
    snap.setdefault("services", snap.get("services") or [])
    text = clip(summary)
    rec = {"status": "active", "started_at": iso}
    if text:
        rec["summary"] = text
    if session_id:
        rec["session_id"] = session_id
    usage = take_usage()
    if usage:
        rec["cursor_usage"] = usage
    snap["research"] = rec
    write_snap()
    extra = {"summary": text} if text else None
    if session_id:
        extra = extra or {}
        extra["session_id"] = session_id
    append_event("research_started", extra)
    raise SystemExit(0)

if action == "stop":
    if not active:
        raise SystemExit(2)
    extra = {}
    topic = clip(research.get("summary") or summary)
    if topic:
        extra["summary"] = topic
    notes = clip(findings or cache_last_text(), 500)
    if notes:
        extra["findings"] = notes
    usage_start = research.get("cursor_usage") if isinstance(research.get("cursor_usage"), dict) else None
    end_usage = take_usage()
    try:
        import cursor_usage
        delta = cursor_usage.compute_delta(usage_start, end_usage)
        if delta:
            extra.update(delta)
    except Exception:
        pass
    snap.pop("research", None)
    write_snap()
    append_event("research_completed", extra or None)
    raise SystemExit(0)

raise SystemExit(1)
PY
}

append_completed_once() {
    python3 "$COORD_DIR/task_snapshot.py" complete "$CURRENT_TASK_FILE" "$EVENTS_FILE" "$ALIAS" "$TASK_ID" "$TIMESTAMP" "$ISO_DATE" "$COORD_DIR"
}

# 1. Update local state files immediately (<10ms, synchronous)
if [ "$EVENT_TYPE" = "research_started" ] || [ "$EVENT_TYPE" = "research_completed" ]; then
    action="start"
    if [ "$EVENT_TYPE" = "research_completed" ]; then
        action="stop"
    fi
    set +e
    apply_research "$action"
    rc=$?
    set -e
    if [ "$rc" -eq 2 ]; then
        exit 0
    fi
    if [ "$rc" -ne 0 ]; then
        exit "$rc"
    fi
    EVENT_JSON=""
elif [ "$EVENT_TYPE" = "task_started" ]; then
    write_current_task
    EVENT_JSON="{\"timestamp\": $TIMESTAMP, \"event\": \"task_started\", \"task_id\": \"$TASK_ID\", \"branch\": \"$BRANCH_NAME\", \"alias\": \"$ALIAS\"}"
elif [ "$EVENT_TYPE" = "task_completed" ]; then
    set +e
    append_completed_once
    rc=$?
    set -e
    if [ "$rc" -eq 2 ]; then
        exit 0
    fi
    if [ "$rc" -ne 0 ]; then
        exit "$rc"
    fi
    EVENT_JSON=""
else
    EVENT_JSON="{\"timestamp\": $TIMESTAMP, \"event\": \"$EVENT_TYPE\", \"task_id\": \"$TASK_ID\", \"alias\": \"$ALIAS\"}"
fi

if [ -n "$EVENT_JSON" ]; then
    echo "$EVENT_JSON" >> "$EVENTS_FILE"
fi

# 2. Push snapshots to origin/coordinator-state (Common main stays clean)
(
    "$COORD_DIR/coordinator_state.sh" push "chore(progress): $ALIAS $EVENT_TYPE ${TASK_ID:-}"
) </dev/null >> "$SYNC_LOG" 2>&1 &
