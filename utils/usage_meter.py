#!/usr/bin/env python3
"""Append local Cursor quota samples (no secrets) for hourly spend charts."""

from __future__ import annotations

import json
import os
import time
from pathlib import Path
from typing import Any

DEBOUNCE_SEC = 25
METER_NAME = "usage_meter.jsonl"
KEEP_DAYS = 45


def meter_path(cache_dir: str | Path) -> Path:
    return Path(cache_dir) / METER_NAME


def _number(value: Any) -> float | None:
    if value is None or isinstance(value, bool):
        return None
    if isinstance(value, (int, float)):
        if value != value:
            return None
        return float(value)
    return None


def record_session(
    cache_dir: str | Path,
    *,
    alias: str,
    session_id: str,
    snap: dict[str, Any] | None,
    fetch_usage,
) -> bool:
    """Bind this Composer tab to a slot/research and append a meter sample."""
    task_id, kind, title = bind_target(snap, session_id)
    prev = last_sample(cache_dir)
    now = time.time()
    usage = None
    if prev and now - int(prev.get("ts") or 0) < DEBOUNCE_SEC:
        usage = prev
    else:
        try:
            usage = fetch_usage()
        except Exception:
            usage = prev
    return record_from_usage(
        cache_dir,
        usage,
        alias=alias,
        session_id=session_id,
        task_id=task_id,
        kind=kind,
        title=title,
    )


def bind_target(snap: dict[str, Any] | None, session_id: str) -> tuple[str, str, str]:
    sid = (session_id or "").strip()
    if not sid or not isinstance(snap, dict):
        return "", "", ""
    research = snap.get("research")
    if isinstance(research, dict) and research.get("status") == "active":
        if str(research.get("session_id") or "").strip() == sid:
            return "", "research", str(research.get("summary") or "").strip()
    tasks = snap.get("tasks")
    if isinstance(tasks, list):
        for slot in tasks:
            if not isinstance(slot, dict):
                continue
            ids = {str(slot.get("session_id") or "").strip()}
            raw = slot.get("session_ids")
            if isinstance(raw, list):
                ids.update(str(item or "").strip() for item in raw)
            if sid not in ids:
                continue
            tid = str(slot.get("task_id") or "").strip()
            title = str(slot.get("summary") or slot.get("doc") or tid).strip()
            return tid, "task", title
    return "", "", ""


def last_sample(cache_dir: str | Path) -> dict[str, Any] | None:
    path = meter_path(cache_dir)
    if not path.is_file():
        return None
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError:
        return None
    for raw in reversed(lines):
        raw = raw.strip()
        if not raw:
            continue
        try:
            row = json.loads(raw)
        except json.JSONDecodeError:
            continue
        if isinstance(row, dict):
            return row
    return None


def append_sample(cache_dir: str | Path, rec: dict[str, Any]) -> bool:
    """Write one sample. Reuses last meter fields if they are fresh (debounce)."""
    cache = Path(cache_dir)
    now = int(time.time())
    prev = last_sample(cache)
    if prev and now - int(prev.get("ts") or 0) < DEBOUNCE_SEC:
        for key in (
            "plan",
            "billing_cycle_start",
            "included_cents",
            "ondemand_cents",
            "cursor_models_pct",
            "other_models_pct",
            "plan_price_usd",
        ):
            if rec.get(key) is None and prev.get(key) is not None:
                rec[key] = prev[key]
        if rec.get("cursor_models_pct") is None and rec.get("other_models_pct") is None:
            return False
    rec["ts"] = now
    rec = {k: v for k, v in rec.items() if v is not None and v != ""}
    if "cursor_models_pct" not in rec and "other_models_pct" not in rec and "ondemand_cents" not in rec:
        return False
    cache.mkdir(parents=True, exist_ok=True)
    path = meter_path(cache)
    with open(path, "a", encoding="utf-8") as fh:
        fh.write(json.dumps(rec, ensure_ascii=False) + "\n")
    prune(path, now)
    return True


def record_from_usage(
    cache_dir: str | Path,
    snap: dict[str, Any] | None,
    *,
    alias: str = "",
    session_id: str = "",
    task_id: str = "",
    kind: str = "",
    title: str = "",
) -> bool:
    if not isinstance(snap, dict):
        return False
    rec: dict[str, Any] = {
        "alias": alias or None,
        "session_id": session_id or None,
        "task_id": task_id or None,
        "kind": kind or None,
        "title": title or None,
        "plan": snap.get("plan") or None,
        "billing_cycle_start": snap.get("billing_cycle_start") or None,
        "included_cents": _number(snap.get("included_cents")),
        "ondemand_cents": _number(snap.get("ondemand_cents")),
        "cursor_models_pct": _number(snap.get("cursor_models_pct")),
        "other_models_pct": _number(snap.get("other_models_pct")),
        "plan_price_usd": _number(snap.get("plan_price_usd")),
    }
    return append_sample(cache_dir, rec)


def prune(path: Path, now: int) -> None:
    cutoff = now - KEEP_DAYS * 86400
    try:
        size = path.stat().st_size
    except OSError:
        return
    if size < 256_000:
        return
    try:
        lines = path.read_text(encoding="utf-8").splitlines()
    except OSError:
        return
    kept = []
    for raw in lines:
        try:
            row = json.loads(raw)
        except json.JSONDecodeError:
            continue
        if int(row.get("ts") or 0) >= cutoff:
            kept.append(raw)
    tmp = path.with_suffix(".tmp")
    tmp.write_text("\n".join(kept) + ("\n" if kept else ""), encoding="utf-8")
    os.replace(tmp, path)
