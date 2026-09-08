#!/bin/bash

# Append one coordinator event to Common/data/progress/events/<ALIAS>/<YYYY>.jsonl and push Common.
# Usage: log_event.sh <event> [key=value ...]
# Example: log_event.sh deploy_finished service=core status=finished
#
# Task binding: in-progress snapshot first, then last_task_id kept on idle
# (deploy often runs right after task_completed). jsonl is the last fallback.

set -e

EVENT_TYPE=$1
shift || true

if [ -z "$EVENT_TYPE" ]; then
    echo "Usage: $0 <event> [key=value ...]" >&2
    exit 1
fi

COORDINATOR_ROOT="$(cd "$(dirname "$0")" && pwd)"
# shellcheck source=paths.sh
. "$(dirname "$0")/paths.sh"
mkdir -p "$PROGRESS_DIR" "$EVENTS_DIR"

ALIAS="$(coordinator_current_author)"
if [ -z "$ALIAS" ]; then
    ALIAS="UNK"
fi

TASK_ID=""
BRANCH=""
SERVICE=""
STATUS=""
REPO=""
for pair in "$@"; do
    key=${pair%%=*}
    val=${pair#*=}
    case "$key" in
        service) SERVICE=$val ;;
        status) STATUS=$val ;;
        repo) REPO=$val ;;
        task_id) TASK_ID=$val ;;
        branch) BRANCH=$val ;;
        alias) ALIAS=$val ;;
    esac
done

CURRENT_TASK_FILE="$PROGRESS_DIR/.current_task_${ALIAS}"
EVENTS_FILE="$(coordinator_events_file "$ALIAS")"
TIMESTAMP=$(date +%s)

python3 - "$EVENTS_FILE" "$CURRENT_TASK_FILE" "$TIMESTAMP" "$EVENT_TYPE" "$ALIAS" "$TASK_ID" "$BRANCH" "$SERVICE" "$STATUS" "$REPO" <<'PY'
import json, os, sys

path, snapshot, ts, event, alias, task_id, branch, service, status, repo = sys.argv[1:]

def from_snapshot(path):
    if not os.path.isfile(path):
        return "", ""
    try:
        snap = json.load(open(path))
    except Exception:
        return "", ""
    if snap.get("status") == "in_progress":
        return (snap.get("task_id") or ""), (snap.get("branch") or "")
    return (snap.get("last_task_id") or snap.get("task_id") or ""), (
        snap.get("last_branch") or snap.get("branch") or ""
    )

def from_jsonl(path):
    if not os.path.isfile(path):
        return "", ""
    try:
        lines = open(path).read().splitlines()
    except OSError:
        return "", ""
    for raw in reversed(lines):
        raw = raw.strip()
        if not raw:
            continue
        try:
            row = json.loads(raw)
        except Exception:
            continue
        if row.get("event") not in ("task_started", "task_completed", "repo_merged"):
            continue
        tid = (row.get("task_id") or "").strip()
        if not tid:
            continue
        return tid, (row.get("branch") or "")
    return "", ""

if not task_id or not branch:
    snap_id, snap_branch = from_snapshot(snapshot)
    if not task_id:
        task_id = snap_id
    if not branch:
        branch = snap_branch
if not task_id:
    jsonl_id, jsonl_branch = from_jsonl(path)
    task_id = jsonl_id
    if not branch:
        branch = jsonl_branch

row = {
    "timestamp": int(ts),
    "event": event,
    "task_id": task_id,
    "alias": alias,
}
if branch:
    row["branch"] = branch
if service:
    row["service"] = service
if status:
    row["status"] = status
if repo:
    row["repo"] = repo
os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "a") as f:
    f.write(json.dumps(row, ensure_ascii=False) + "\n")
PY

(
    cd "$COMMON_ROOT"
    git add "$EVENTS_FILE"
    git commit -m "chore(progress): $ALIAS $EVENT_TYPE ${SERVICE:-$TASK_ID}" || true
    git pull --rebase origin main || true
    git push origin main || true
) </dev/null >> "$SYNC_LOG" 2>&1 &
