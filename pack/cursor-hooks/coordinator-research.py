#!/usr/bin/env python3
"""Bind Cursor composer sessions to coordinator research and auto-close them."""

from __future__ import annotations

import json
import os
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

from coord_paths import current_alias, load_paths

LAST_TEXT_MAX = 500


def paths() -> dict[str, Path]:
    return load_paths(Path(__file__))


try:
    sys.path.insert(0, str(paths()["COORDINATOR_ROOT"] / "utils"))
    from activity_clock import ping_slot
except Exception:
    ping_slot = None


def main() -> None:
    event = sys.argv[1] if len(sys.argv) > 1 else ""
    try:
        payload = json.load(sys.stdin)
    except Exception:
        payload = {}
    if not isinstance(payload, dict):
        payload = {}
    try:
        if event == "sessionStart":
            handle_session_start(payload)
            return
        if event == "preToolUse":
            handle_pre_tool(payload)
            return
        if event == "afterAgentResponse":
            handle_after_response(payload)
            return
        if event == "sessionEnd":
            handle_session_end(payload)
            return
    except Exception:
        pass
    print("{}")


def handle_session_start(payload: dict) -> None:
    sid = str(payload.get("session_id") or "").strip()
    out: dict = {}
    if sid:
        title = composer_title(sid)
        alias = current_alias(paths())
        if alias:
            write_cache(alias, sid, int(datetime.now(timezone.utc).timestamp()), "", title)
        out["env"] = {"COORDINATOR_SESSION_ID": sid}
        if title:
            out["env"]["COORDINATOR_CHAT_TITLE"] = title
        out["additional_context"] = (
            f"Coordinator composer session_id={sid}. "
            + (f"Chat title: {title}. " if title else "")
            + "If this chat is off-task research, pass session_id="
            f"{sid} to research_started. "
            "If starting or resuming a task, pass session_id="
            f"{sid} to task_started."
        )
    print(json.dumps(out, ensure_ascii=False))
    if sid:
        bump_bound_slot(sid, datetime.now(timezone.utc))


def handle_pre_tool(payload: dict) -> None:
    print("{}")
    sid = os.environ.get("COORDINATOR_SESSION_ID", "").strip()
    if not sid:
        sid = str(payload.get("session_id") or "").strip()
    if not sid:
        return
    bump_bound_slot(sid, datetime.now(timezone.utc))


def handle_after_response(payload: dict) -> None:
    print("{}")
    sid = os.environ.get("COORDINATOR_SESSION_ID", "").strip()
    if not sid:
        sid = str(payload.get("session_id") or "").strip()
    if not sid:
        return
    p = paths()
    alias = current_alias(p)
    now = datetime.now(timezone.utc)
    text = clip(str(payload.get("text") or ""), LAST_TEXT_MAX)
    write_cache(alias, sid, int(now.timestamp()), text, composer_title(sid))
    snap = load_snapshot(alias)
    if not isinstance(snap, dict):
        return
    research = snap.get("research") if isinstance(snap.get("research"), dict) else None
    if isinstance(research, dict) and research.get("status") == "active":
        bound = str(research.get("session_id") or "").strip()
        if not bound:
            research["session_id"] = sid
            snap["research"] = research
            write_snapshot(alias, snap)
    bump_bound_slot(sid, now)


def handle_session_end(payload: dict) -> None:
    print("{}")
    sid = str(payload.get("session_id") or os.environ.get("COORDINATOR_SESSION_ID") or "").strip()
    if not sid:
        return
    alias = current_alias(paths())
    snap = load_snapshot(alias)
    research = (snap or {}).get("research") if isinstance(snap, dict) else None
    if not isinstance(research, dict) or research.get("status") != "active":
        return
    bound = str(research.get("session_id") or "").strip()
    if bound and bound != sid:
        return
    if not bound:
        rec = read_cache_session(sid)
        if not rec or rec.get("alias") != alias:
            return
    findings = clip(
        "Tab closed. " + (read_cache_session(sid) or {}).get("last_text") or research.get("summary") or "",
        LAST_TEXT_MAX,
    )
    close_research(alias, findings)


def close_research(alias: str, findings: str) -> None:
    p = paths()
    sync = p["COORDINATOR_ROOT"] / "utils" / "sync_event.sh"
    if not sync.is_file() or not alias:
        return
    args = [str(sync), alias, "research_completed"]
    if findings:
        args.append(f"findings={findings}")
    subprocess.Popen(
        args,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        cwd=str(p["WORKSPACE_ROOT"]),
        env={**os.environ, "WORKSPACE_ROOT": str(p["WORKSPACE_ROOT"])},
        start_new_session=True,
    )


def snapshot_path(alias: str) -> Path:
    return paths()["PROGRESS_DIR"] / f".current_task_{alias}"


def load_snapshot(alias: str) -> dict | None:
    path = snapshot_path(alias)
    if not path.is_file():
        return None
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return None
    return data if isinstance(data, dict) else None


def write_snapshot(alias: str, snap: dict) -> None:
    path = snapshot_path(alias)
    path.write_text(json.dumps(snap, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def slot_session_ids(slot: dict) -> set[str]:
    ids: set[str] = set()
    sid = str(slot.get("session_id") or "").strip()
    if sid:
        ids.add(sid)
    raw = slot.get("session_ids")
    if isinstance(raw, list):
        for item in raw:
            value = str(item or "").strip()
            if value:
                ids.add(value)
    return ids


def bump_bound_slot(sid: str, now: datetime) -> None:
    alias = current_alias(paths())
    snap = load_snapshot(alias)
    if not isinstance(snap, dict):
        return
    bump_slot_activity(snap, alias, sid, now)


def bump_slot_activity(snap: dict, alias: str, sid: str, now: datetime) -> None:
    iso = now.strftime("%Y-%m-%dT%H:%M:%SZ")
    tasks = snap.get("tasks")
    if not isinstance(tasks, list) or not tasks:
        return
    changed = False
    for slot in tasks:
        if not isinstance(slot, dict):
            continue
        if sid not in slot_session_ids(slot):
            continue
        if ping_slot:
            ping_slot(slot, now)
        else:
            slot["updated_at"] = iso
        changed = True
        if snap.get("task_id") == slot.get("task_id"):
            snap["updated_at"] = slot.get("updated_at") or iso
    if changed:
        write_snapshot(alias, snap)


def cache_file() -> Path:
    return paths()["CACHE_DIR"] / "composer_sessions.json"


def read_cache() -> dict:
    path = cache_file()
    if not path.is_file():
        return {"sessions": {}}
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return {"sessions": {}}
    if not isinstance(data, dict):
        return {"sessions": {}}
    if not isinstance(data.get("sessions"), dict):
        data["sessions"] = {}
    return data


def read_cache_session(sid: str) -> dict | None:
    rec = read_cache().get("sessions", {}).get(sid)
    return rec if isinstance(rec, dict) else None


def write_cache(alias: str, sid: str, ts: int, text: str, title: str = "") -> None:
    data = read_cache()
    sessions = data.setdefault("sessions", {})
    prev = sessions.get(sid) if isinstance(sessions.get(sid), dict) else {}
    rec = {
        "alias": alias,
        "last_agent_at": ts,
        "last_text": text or prev.get("last_text") or "",
    }
    name = (title or prev.get("title") or "").strip()
    if name:
        rec["title"] = name
    sessions[sid] = rec
    path = cache_file()
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(data, indent=2, ensure_ascii=False) + "\n", encoding="utf-8")


def composer_db() -> Path | None:
    home = Path.home()
    candidates = [
        home / "Library/Application Support/Cursor/User/globalStorage/state.vscdb",
        home / ".config/Cursor/User/globalStorage/state.vscdb",
        home / "AppData/Roaming/Cursor/User/globalStorage/state.vscdb",
    ]
    for path in candidates:
        if path.is_file():
            return path
    return None


def composer_title(sid: str) -> str:
    sid = (sid or "").strip()
    if not sid:
        return ""
    cached = read_cache_session(sid)
    if cached and str(cached.get("title") or "").strip():
        return str(cached["title"]).strip()
    db = composer_db()
    if not db:
        return ""
    try:
        import sqlite3

        con = sqlite3.connect(f"file:{db}?mode=ro", uri=True)
        row = con.execute(
            "SELECT value FROM composerHeaders WHERE composerId=?", (sid,)
        ).fetchone()
        if row and row[0]:
            raw = row[0]
            if isinstance(raw, bytes):
                raw = raw.decode("utf-8", "replace")
            header = json.loads(raw)
            name = str(header.get("name") or "").strip()
            if name:
                return name
        row = con.execute(
            "SELECT value FROM cursorDiskKV WHERE key=?", (f"composerData:{sid}",)
        ).fetchone()
        if row and row[0]:
            raw = row[0]
            if isinstance(raw, bytes):
                raw = raw.decode("utf-8", "replace")
            data = json.loads(raw)
            name = str(data.get("name") or "").strip()
            if name:
                return name
    except Exception:
        return ""
    return ""


def clip(text: str, n: int) -> str:
    text = " ".join((text or "").split())
    if len(text) > n:
        return text[: n - 3].rstrip() + "..."
    return text


if __name__ == "__main__":
    main()
