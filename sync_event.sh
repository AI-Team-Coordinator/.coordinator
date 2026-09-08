#!/bin/bash

# Helper script to instantly update local coordinator state and asynchronously push to Common.
# Usage: ./sync_event.sh <ALIAS> <EVENT_TYPE> [TASK_ID] [BRANCH_NAME] [Service1,Service2] [doc=...] [summary=...]
# Example: ./sync_event.sh EK task_started 20260907-1756 feature/auth Core,InboxPanelWeb
# Example: ./sync_event.sh EK task_started FIX-... fix/avatar Website summary="Restore dark header avatar"
# Example: ./sync_event.sh EK task_completed 20260907-1756

set -e

ALIAS=$1
EVENT_TYPE=$2
TASK_ID=$3
BRANCH_NAME=${4:-""}
SERVICES_CSV=${5:-""}
DOC_PATH=""
SUMMARY=""

if [ -z "$ALIAS" ] || [ -z "$EVENT_TYPE" ]; then
    echo "Usage: $0 <ALIAS> <EVENT_TYPE> [TASK_ID] [BRANCH_NAME] [Service1,Service2] [doc=...] [summary=...]"
    exit 1
fi

if [ "$#" -ge 6 ]; then
    shift 5
    for pair in "$@"; do
        key=${pair%%=*}
        val=${pair#*=}
        case "$key" in
            doc) DOC_PATH=$val ;;
            summary) SUMMARY=$val ;;
        esac
    done
fi

COORDINATOR_ROOT="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=paths.sh
. "$(dirname "$0")/paths.sh"
mkdir -p "$PROGRESS_DIR" "$EVENTS_DIR"

CURRENT_TASK_FILE="$PROGRESS_DIR/.current_task_${ALIAS}"
EVENTS_FILE="$(coordinator_events_file "$ALIAS")"
COORD_DIR="$(cd "$(dirname "$0")" && pwd)"
TIMESTAMP=$(date +%s)
ISO_DATE=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

write_current_task() {
    local status=$1
    local task_id=$2
    local branch=$3
    python3 - "$CURRENT_TASK_FILE" "$ALIAS" "$status" "$task_id" "$branch" "$ISO_DATE" "$SERVICES_CSV" "$COORD_DIR" "$COMMON_ROOT" "$DOC_PATH" "$SUMMARY" <<'PY'
import json, os, sys
path, alias, status, task_id, branch, iso, services_csv, coord_dir, common_root, doc_path, summary = sys.argv[1:]
sys.path.insert(0, coord_dir)
services = [s.strip() for s in services_csv.split(",") if s.strip()]


def find_task_doc(root, tid):
    if not root or not tid:
        return ""
    docs = os.path.join(root, "docs")
    name = tid + ".md"
    candidates = [os.path.join(docs, name)]
    if os.path.isdir(docs):
        for entry in os.listdir(docs):
            candidates.append(os.path.join(docs, entry, name))
    for candidate in candidates:
        if os.path.isfile(candidate):
            return os.path.relpath(candidate, root).replace("\\", "/")
    return ""


if status == "idle":
    payload = {
        "alias": alias,
        "task_id": None,
        "branch": None,
        "status": "idle",
        "services": [],
        "updated_at": iso,
    }
    prev = {}
    if os.path.isfile(path):
        try:
            prev = json.load(open(path))
        except Exception:
            prev = {}
    last_task = task_id or prev.get("task_id") or prev.get("last_task_id")
    last_branch = branch or prev.get("branch") or prev.get("last_branch")
    if last_task:
        payload["last_task_id"] = last_task
    if last_branch:
        payload["last_branch"] = last_branch
else:
    payload = {
        "alias": alias,
        "task_id": task_id or None,
        "branch": branch or None,
        "status": status,
        "services": services,
        "updated_at": iso,
    }
    doc = (doc_path or "").strip()
    if not doc:
        doc = find_task_doc(common_root, task_id)
    if doc:
        payload["doc"] = doc
    text = " ".join((summary or "").split())
    if len(text) > 280:
        text = text[:277].rstrip() + "..."
    if text:
        payload["summary"] = text
    try:
        import cursor_usage
        snap = cursor_usage.take_snapshot()
        if snap:
            payload["cursor_usage"] = snap
    except Exception:
        pass
os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "w") as f:
    json.dump(payload, f, indent=2)
    f.write("\n")
PY
}

append_completed_once() {
    python3 - "$CURRENT_TASK_FILE" "$EVENTS_FILE" "$ALIAS" "$TASK_ID" "$TIMESTAMP" "$ISO_DATE" "$COORD_DIR" <<'PY'
import fcntl, json, os, sys
snapshot, events_file, alias, task_id, ts, iso, coord_dir = sys.argv[1:]
sys.path.insert(0, coord_dir)
os.makedirs(os.path.dirname(events_file), exist_ok=True)
lock_path = events_file + ".lock"
lock = open(lock_path, "a+")
fcntl.flock(lock.fileno(), fcntl.LOCK_EX)

def already():
    if not os.path.isfile(events_file):
        return False
    for raw in reversed(open(events_file).read().splitlines()):
        raw = raw.strip()
        if not raw:
            continue
        try:
            row = json.loads(raw)
        except Exception:
            continue
        if row.get("task_id") != task_id:
            continue
        if row.get("event") == "task_completed":
            return True
        if row.get("event") == "task_started":
            return False
    return False

idle = {
    "alias": alias,
    "task_id": None,
    "branch": None,
    "status": "idle",
    "services": [],
    "updated_at": iso,
}
start_usage = None
services = []
prev_task = task_id or None
prev_branch = None
if os.path.isfile(snapshot):
    try:
        snap = json.load(open(snapshot))
        start_usage = snap.get("cursor_usage")
        services = snap.get("services") or []
        prev_task = task_id or snap.get("task_id") or snap.get("last_task_id")
        prev_branch = snap.get("branch") or snap.get("last_branch")
    except Exception:
        start_usage = None
        services = []
if prev_task:
    idle["last_task_id"] = prev_task
if prev_branch:
    idle["last_branch"] = prev_branch
with open(snapshot, "w") as f:
    json.dump(idle, f, indent=2)
    f.write("\n")

skipped = already()
if not skipped:
    row = {
        "timestamp": int(ts),
        "event": "task_completed",
        "task_id": task_id,
        "alias": alias,
    }
    try:
        import cursor_usage
        kind = cursor_usage.spend_kind(services)
        if kind:
            row["spend_kind"] = kind
        if services:
            row["services"] = services
        end_usage = cursor_usage.take_snapshot()
        delta = cursor_usage.compute_delta(start_usage, end_usage)
        if delta:
            row.update(delta)
    except Exception:
        pass
    with open(events_file, "a") as f:
        f.write(json.dumps(row, ensure_ascii=False) + "\n")

fcntl.flock(lock.fileno(), fcntl.LOCK_UN)
lock.close()
raise SystemExit(2 if skipped else 0)
PY
}

# 1. Update local state files immediately (<10ms, synchronous)
if [ "$EVENT_TYPE" = "task_started" ]; then
    write_current_task "in_progress" "$TASK_ID" "$BRANCH_NAME" "$SERVICES_CSV"
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

# 2. Asynchronously commit, rebase, and push in background (does not block IDE/agent)
(
    cd "$COMMON_ROOT"
    git add "$CURRENT_TASK_FILE" "$EVENTS_FILE"
    git commit -m "chore(progress): $ALIAS $EVENT_TYPE ${TASK_ID:-""}" || true
    git pull --rebase origin main || true
    git push origin main || true
) </dev/null >> "$SYNC_LOG" 2>&1 &
