"""Merge activity pings into task-slot windows. Idle gap > 20 minutes starts a new window."""

from __future__ import annotations

from datetime import datetime, timezone

IDLE_TIMEOUT_SEC = 20 * 60


def ping_research(research: dict, now: datetime) -> None:
    """Same windows as a task slot, stored on nested research."""
    ping_slot(research, now)


def ping_slot(slot: dict, now: datetime) -> None:
    if not isinstance(slot, dict):
        return
    if now.tzinfo is None:
        now = now.replace(tzinfo=timezone.utc)
    iso = now.astimezone(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")
    windows = slot.get("activity_windows")
    if not isinstance(windows, list):
        windows = []
    cleaned = [row for row in windows if isinstance(row, dict)]
    if cleaned:
        last = cleaned[-1]
        ended = parse_iso(last.get("ended_at") or last.get("started_at") or "")
        if ended is not None and (now - ended).total_seconds() <= IDLE_TIMEOUT_SEC:
            last["ended_at"] = iso
            if not last.get("started_at"):
                last["started_at"] = iso
        else:
            cleaned.append({"started_at": iso, "ended_at": iso})
    else:
        cleaned.append({"started_at": iso, "ended_at": iso})
    slot["activity_windows"] = cleaned
    slot["last_activity_at"] = iso
    slot["updated_at"] = iso


def frozen_active_seconds(slot: dict) -> int:
    windows = slot.get("activity_windows") if isinstance(slot, dict) else None
    total = 0
    if isinstance(windows, list):
        for row in windows:
            if not isinstance(row, dict):
                continue
            start = parse_iso(row.get("started_at") or "")
            end = parse_iso(row.get("ended_at") or row.get("started_at") or "")
            if start is None or end is None or end < start:
                continue
            total += int((end - start).total_seconds())
        if total > 0 or windows:
            return max(total, 0)
    started = parse_iso((slot or {}).get("started_at") or "")
    last = parse_iso((slot or {}).get("last_activity_at") or (slot or {}).get("updated_at") or "")
    if started is None or last is None or last < started:
        return 0
    return int((last - started).total_seconds())


def parse_iso(raw: str) -> datetime | None:
    text = (raw or "").strip()
    if not text:
        return None
    if text.endswith("Z"):
        text = text[:-1] + "+00:00"
    try:
        parsed = datetime.fromisoformat(text)
    except ValueError:
        return None
    if parsed.tzinfo is None:
        parsed = parsed.replace(tzinfo=timezone.utc)
    return parsed
