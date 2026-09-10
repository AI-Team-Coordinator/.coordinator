#!/usr/bin/env python3
"""Upsert / complete coordinator task slots without wiping sibling tasks or research."""

from __future__ import annotations

import fcntl
import json
import os
import sys


def main() -> None:
    if len(sys.argv) < 2:
        raise SystemExit(1)
    action = sys.argv[1]
    if action == "start":
        (
            _,
            _,
            path,
            alias,
            iso,
            task_id,
            branch,
            services_csv,
            coord_dir,
            common_root,
            doc_path,
            summary,
        ) = sys.argv[:12]
        session_id = sys.argv[12] if len(sys.argv) > 12 else ""
        sys.path.insert(0, coord_dir)
        upsert_started(
            path, alias, iso, task_id, branch, services_csv, common_root, doc_path, summary, session_id
        )
        return
    if action == "complete":
        _, _, path, events_file, alias, task_id, ts, iso, coord_dir = sys.argv[:9]
        sys.path.insert(0, coord_dir)
        raise SystemExit(complete_once(path, events_file, alias, task_id, ts, iso))
    raise SystemExit(1)


def load_snap(path: str) -> dict:
    if not os.path.isfile(path):
        return {}
    try:
        data = json.loads(open(path).read())
    except Exception:
        return {}
    return data if isinstance(data, dict) else {}


def write_snap(path: str, snap: dict) -> None:
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(snap, f, indent=2, ensure_ascii=False)
        f.write("\n")


def find_task_doc(root: str, tid: str) -> str:
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


def merge_session_ids(existing: dict, new_sid: str) -> list[str]:
    ids: list[str] = []
    seen: set[str] = set()

    def add(raw) -> None:
        sid = str(raw or "").strip()
        if not sid or sid in seen:
            return
        seen.add(sid)
        ids.append(sid)

    raw = existing.get("session_ids") if isinstance(existing, dict) else None
    if isinstance(raw, list):
        for item in raw:
            add(item)
    if isinstance(existing, dict):
        add(existing.get("session_id"))
    add(new_sid)
    return ids


def clip(text: str, n: int = 280) -> str:
    text = " ".join((text or "").split())
    if len(text) > n:
        return text[: n - 3].rstrip() + "..."
    return text


def take_usage(coord_dir: str):
    try:
        sys.path.insert(0, coord_dir)
        import cursor_usage

        return cursor_usage.take_snapshot()
    except Exception:
        return None


def spend_kind(services: list, coord_dir: str):
    try:
        sys.path.insert(0, coord_dir)
        import cursor_usage

        return cursor_usage.spend_kind(services)
    except Exception:
        return None


def usage_delta(start, end, coord_dir: str):
    try:
        sys.path.insert(0, coord_dir)
        import cursor_usage

        return cursor_usage.compute_delta(start, end)
    except Exception:
        return None


def legacy_slot(snap: dict) -> dict | None:
    if snap.get("status") != "in_progress":
        return None
    tid = snap.get("task_id")
    if not tid:
        return None
    slot = {
        "task_id": tid,
        "branch": snap.get("branch"),
        "services": snap.get("services") or [],
        "updated_at": snap.get("updated_at") or "",
        "started_at": snap.get("started_at") or snap.get("updated_at") or "",
    }
    if snap.get("doc"):
        slot["doc"] = snap["doc"]
    if snap.get("summary"):
        slot["summary"] = snap["summary"]
    if isinstance(snap.get("cursor_usage"), dict):
        slot["cursor_usage"] = snap["cursor_usage"]
    return slot


def slots_from(snap: dict) -> list[dict]:
    raw = snap.get("tasks")
    if isinstance(raw, list) and raw:
        out = []
        for item in raw:
            if isinstance(item, dict) and item.get("task_id"):
                out.append(item)
        if out:
            return out
    one = legacy_slot(snap)
    return [one] if one else []


def mirror_root(snap: dict, slot: dict | None, iso: str) -> None:
    snap["updated_at"] = iso
    if not slot:
        snap["status"] = "idle"
        snap["task_id"] = None
        snap["branch"] = None
        snap["services"] = []
        snap.pop("doc", None)
        snap.pop("summary", None)
        snap.pop("cursor_usage", None)
        snap["tasks"] = []
        return
    snap["status"] = "in_progress"
    snap["task_id"] = slot.get("task_id")
    snap["branch"] = slot.get("branch")
    snap["services"] = slot.get("services") or []
    if slot.get("doc"):
        snap["doc"] = slot["doc"]
    else:
        snap.pop("doc", None)
    if slot.get("summary"):
        snap["summary"] = slot["summary"]
    else:
        snap.pop("summary", None)
    if isinstance(slot.get("cursor_usage"), dict):
        snap["cursor_usage"] = slot["cursor_usage"]
    else:
        snap.pop("cursor_usage", None)


def upsert_started(
    path: str,
    alias: str,
    iso: str,
    task_id: str,
    branch: str,
    services_csv: str,
    common_root: str,
    doc_path: str,
    summary: str,
    session_id: str = "",
) -> None:
    coord_dir = os.path.dirname(os.path.abspath(__file__))
    snap = load_snap(path)
    research = snap.get("research") if isinstance(snap.get("research"), dict) else None
    git_report = snap.get("git_report")
    last_task = snap.get("last_task_id")
    last_branch = snap.get("last_branch")
    slots = slots_from(snap)

    services = [s.strip() for s in (services_csv or "").split(",") if s.strip()]
    doc = (doc_path or "").strip() or find_task_doc(common_root, task_id)
    text = clip(summary)
    usage = take_usage(coord_dir)

    slot = {
        "task_id": task_id,
        "branch": branch or None,
        "services": services,
        "started_at": iso,
        "updated_at": iso,
    }
    if doc:
        slot["doc"] = doc
    if text:
        slot["summary"] = text
    if usage:
        slot["cursor_usage"] = usage
    sid = (session_id or "").strip()
    replaced = False
    for i, existing in enumerate(slots):
        if existing.get("task_id") == task_id:
            slot["started_at"] = existing.get("started_at") or iso
            if not slot.get("cursor_usage") and isinstance(existing.get("cursor_usage"), dict):
                slot["cursor_usage"] = existing["cursor_usage"]
            ids = merge_session_ids(existing, sid)
            if ids:
                slot["session_ids"] = ids
                slot["session_id"] = ids[0]
            slots[i] = slot
            replaced = True
            break
    if not replaced:
        if sid:
            slot["session_id"] = sid
            slot["session_ids"] = [sid]
        slots.append(slot)

    out = {"alias": alias, "tasks": slots}
    if last_task:
        out["last_task_id"] = last_task
    if last_branch:
        out["last_branch"] = last_branch
    if isinstance(research, dict) and research.get("status") == "active":
        out["research"] = research
    if git_report is not None:
        out["git_report"] = git_report
    mirror_root(out, slot, iso)
    write_snap(path, out)


def already_completed(events_file: str, task_id: str) -> bool:
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


def complete_once(path: str, events_file: str, alias: str, task_id: str, ts: str, iso: str) -> int:
    coord_dir = os.path.dirname(os.path.abspath(__file__))
    os.makedirs(os.path.dirname(events_file), exist_ok=True)
    lock_path = events_file + ".lock"
    lock = open(lock_path, "a+")
    fcntl.flock(lock.fileno(), fcntl.LOCK_EX)
    try:
        snap = load_snap(path)
        slots = slots_from(snap)
        research = snap.get("research") if isinstance(snap.get("research"), dict) else None
        git_report = snap.get("git_report")
        done = None
        remain = []
        for slot in slots:
            if slot.get("task_id") == task_id and done is None:
                done = slot
                continue
            remain.append(slot)
        if done is None and snap.get("task_id") == task_id:
            done = legacy_slot(snap) or {"task_id": task_id, "services": snap.get("services") or []}

        out = {"alias": alias, "tasks": remain}
        prev_task = task_id or snap.get("task_id") or snap.get("last_task_id")
        prev_branch = None
        if done:
            prev_branch = done.get("branch")
        prev_branch = prev_branch or snap.get("branch") or snap.get("last_branch")
        if prev_task:
            out["last_task_id"] = prev_task
        if prev_branch:
            out["last_branch"] = prev_branch
        if isinstance(research, dict) and research.get("status") == "active":
            out["research"] = research
        if git_report is not None and remain:
            out["git_report"] = git_report
        mirror_root(out, remain[-1] if remain else None, iso)
        write_snap(path, out)

        skipped = already_completed(events_file, task_id)
        if skipped:
            return 2
        row = {
            "timestamp": int(ts),
            "event": "task_completed",
            "task_id": task_id,
            "alias": alias,
        }
        services = (done or {}).get("services") or []
        kind = spend_kind(services, coord_dir)
        if kind:
            row["spend_kind"] = kind
        if services:
            row["services"] = services
        start_usage = (done or {}).get("cursor_usage") if isinstance((done or {}).get("cursor_usage"), dict) else None
        end_usage = take_usage(coord_dir)
        delta = usage_delta(start_usage, end_usage, coord_dir)
        if delta:
            row.update(delta)
        with open(events_file, "a") as f:
            f.write(json.dumps(row, ensure_ascii=False) + "\n")
        return 0
    finally:
        fcntl.flock(lock.fileno(), fcntl.LOCK_UN)
        lock.close()


if __name__ == "__main__":
    main()
