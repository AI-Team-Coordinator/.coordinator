#!/usr/bin/env python3
"""Snapshot Cursor period usage for coordinator task cost (local session only)."""

from __future__ import annotations

import json
import os
import sqlite3
import sys
import urllib.error
import urllib.request
from pathlib import Path
from typing import Any

TIMEOUT_SEC = 8
USAGE_SUMMARY_URL = "https://cursor.com/api/usage-summary"
PERIOD_USAGE_URL = "https://api2.cursor.sh/aiserver.v1.DashboardService/GetCurrentPeriodUsage"

# Seat price. 100% Cursor Models + 100% Other Models = this amount.
PLAN_PRICE_USD = {
    "pro": 20.0,
    "pro_plus": 60.0,
    "pro+": 60.0,
    "ultra": 200.0,
}


def _state_db_path() -> Path | None:
    if sys.platform == "darwin":
        path = Path.home() / "Library/Application Support/Cursor/User/globalStorage/state.vscdb"
    elif sys.platform == "win32":
        appdata = os.environ.get("APPDATA", "")
        path = Path(appdata) / "Cursor/User/globalStorage/state.vscdb" if appdata else None
    else:
        path = Path.home() / ".config/Cursor/User/globalStorage/state.vscdb"
    if path is None or not path.is_file():
        return None
    return path


def _item(conn: sqlite3.Connection, key: str) -> str | None:
    row = conn.execute("SELECT value FROM ItemTable WHERE key = ?", (key,)).fetchone()
    if not row or row[0] is None:
        return None
    value = row[0]
    if isinstance(value, bytes):
        value = value.decode("utf-8", errors="replace")
    text = str(value).strip()
    return text or None


def _jwt_sub(token: str) -> str | None:
    parts = token.split(".")
    if len(parts) < 2:
        return None
    payload = parts[1]
    pad = "=" * (-len(payload) % 4)
    try:
        raw = json.loads(_b64url(payload + pad))
    except Exception:
        return None
    sub = raw.get("sub") or raw.get("user_id")
    if isinstance(sub, str) and sub.strip():
        return sub.strip()
    return None


def _b64url(data: str) -> bytes:
    import base64

    return base64.urlsafe_b64decode(data.encode("ascii"))


def _local_session() -> tuple[str | None, str | None, str | None]:
    db_path = _state_db_path()
    if db_path is None:
        return None, None, None
    conn = sqlite3.connect(f"file:{db_path}?mode=ro", uri=True)
    try:
        token = _item(conn, "cursorAuth/accessToken")
        plan = _item(conn, "cursorAuth/stripeMembershipType")
        user_id = _jwt_sub(token) if token else None
        return token, user_id, plan
    finally:
        conn.close()


def plan_price_usd(plan: str) -> float | None:
    key = (plan or "").strip().lower().replace(" ", "_").replace("-", "_")
    return PLAN_PRICE_USD.get(key)


def included_pool_delta(start_pct: float, end_pct: float) -> float:
    """Share of a pool still inside the included 100%. Overflow is on-demand."""
    start_in = min(max(start_pct, 0.0), 100.0)
    end_in = min(max(end_pct, 0.0), 100.0)
    return max(0.0, end_in - start_in)


def _to_number(value: Any) -> float | None:
    if value is None or isinstance(value, bool):
        return None
    if isinstance(value, (int, float)):
        if value != value:  # NaN
            return None
        return float(value)
    if isinstance(value, str):
        text = value.strip().replace("$", "").replace(",", "")
        if not text:
            return None
        try:
            return float(text)
        except ValueError:
            return None
    return None


def _cycle_start(payload: dict[str, Any]) -> str:
    for key in ("billingCycleStart", "billing_cycle_start", "startOfMonth"):
        raw = payload.get(key)
        if raw is None:
            continue
        return str(raw)
    return ""


def normalize_payload(payload: dict[str, Any], fallback_plan: str = "") -> dict[str, Any] | None:
    if not isinstance(payload, dict) or not payload:
        return None

    plan = str(payload.get("membershipType") or payload.get("plan") or fallback_plan or "").strip()
    included = None
    ondemand = None
    cursor_pct = None
    other_pct = None

    plan_usage = payload.get("planUsage")
    if isinstance(plan_usage, dict):
        included = _to_number(plan_usage.get("includedSpend"))
        if included is None:
            included = _to_number(plan_usage.get("totalSpend"))
        cursor_pct = _to_number(plan_usage.get("autoPercentUsed"))
        other_pct = _to_number(plan_usage.get("apiPercentUsed"))

    spend_limit = payload.get("spendLimitUsage")
    if isinstance(spend_limit, dict):
        ondemand = _to_number(spend_limit.get("individualUsed"))
        if ondemand is None:
            ondemand = _to_number(spend_limit.get("totalSpend"))
        if ondemand is None:
            ondemand = _to_number(spend_limit.get("pooledUsed"))

    individual = payload.get("individualUsage")
    if isinstance(individual, dict):
        bucket = individual.get("plan")
        if isinstance(bucket, dict):
            if included is None:
                included = _to_number(bucket.get("used"))
            if cursor_pct is None:
                cursor_pct = _to_number(bucket.get("autoPercentUsed"))
            if other_pct is None:
                other_pct = _to_number(bucket.get("apiPercentUsed"))
        demand = individual.get("onDemand")
        if isinstance(demand, dict) and ondemand is None:
            ondemand = _to_number(demand.get("used"))

    if included is None and cursor_pct is None and other_pct is None and ondemand is None:
        return None

    snap = {
        "plan": plan,
        "billing_cycle_start": _cycle_start(payload),
        "included_cents": included if included is not None else 0.0,
        "ondemand_cents": ondemand if ondemand is not None else 0.0,
        "cursor_models_pct": cursor_pct if cursor_pct is not None else 0.0,
        "other_models_pct": other_pct if other_pct is not None else 0.0,
    }
    price = plan_price_usd(plan)
    if price is not None:
        snap["plan_price_usd"] = price
    return snap


def compute_delta(start: dict[str, Any] | None, end: dict[str, Any] | None) -> dict[str, Any] | None:
    if not isinstance(start, dict) or not isinstance(end, dict):
        return None
    start_cycle = str(start.get("billing_cycle_start") or "")
    end_cycle = str(end.get("billing_cycle_start") or "")
    if start_cycle and end_cycle and start_cycle != end_cycle:
        return None

    def cents(row: dict[str, Any], key: str) -> float:
        n = _to_number(row.get(key))
        return n if n is not None else 0.0

    def pct(row: dict[str, Any], key: str) -> float:
        n = _to_number(row.get(key))
        return n if n is not None else 0.0

    ondemand_delta = max(0.0, cents(end, "ondemand_cents") - cents(start, "ondemand_cents"))
    cursor_start = pct(start, "cursor_models_pct")
    cursor_end = pct(end, "cursor_models_pct")
    other_start = pct(start, "other_models_pct")
    other_end = pct(end, "other_models_pct")
    cursor_delta = max(0.0, cursor_end - cursor_start)
    other_delta = max(0.0, other_end - other_start)
    plan = str(end.get("plan") or start.get("plan") or "").strip()
    price = _to_number(end.get("plan_price_usd"))
    if price is None:
        price = _to_number(start.get("plan_price_usd"))
    if price is None:
        price = plan_price_usd(plan)

    cursor_included = included_pool_delta(cursor_start, cursor_end)
    other_included = included_pool_delta(other_start, other_end)
    budget = 0.0
    if price is not None and price > 0:
        budget = price * (cursor_included + other_included) / 200.0

    out = {
        "cost_usd": round(ondemand_delta / 100.0, 4),
        "budget_usd": round(budget, 4),
        "ondemand_usd": round(ondemand_delta / 100.0, 4),
        "cursor_models_pct": round(cursor_delta, 4),
        "other_models_pct": round(other_delta, 4),
        "usage_plan": plan,
    }
    if price is not None:
        out["plan_price_usd"] = price
    return out


INFRA_SERVICE_KEYS = frozenset({"common", "cursor"})


def normalize_service_key(label: str) -> str:
    text = (label or "").strip().lower().replace("-", "_").replace(" ", "_")
    return text.lstrip(".")


def is_infra_service(label: str) -> bool:
    return normalize_service_key(label) in INFRA_SERVICE_KEYS


def spend_kind(services: list[Any] | None) -> str | None:
    labels = [str(item).strip() for item in (services or []) if str(item).strip()]
    if not labels:
        return None
    if all(is_infra_service(item) for item in labels):
        return "infra"
    return "product"


def _http_json(method: str, url: str, headers: dict[str, str], body: bytes | None = None) -> dict[str, Any] | None:
    req = urllib.request.Request(url, data=body, headers=headers, method=method)
    try:
        with urllib.request.urlopen(req, timeout=TIMEOUT_SEC) as resp:
            raw = resp.read()
    except (urllib.error.URLError, TimeoutError, ValueError, OSError):
        return None
    try:
        data = json.loads(raw)
    except json.JSONDecodeError:
        return None
    return data if isinstance(data, dict) else None


def fetch_raw(token: str, user_id: str | None) -> dict[str, Any] | None:
    headers = {
        "Accept": "application/json",
        "Content-Type": "application/json",
    }
    if user_id:
        cookie = f"{user_id}::{token}"
        summary = _http_json(
            "GET",
            USAGE_SUMMARY_URL,
            {**headers, "Cookie": f"WorkosCursorSessionToken={cookie}"},
        )
        if summary:
            return summary

    rpc = _http_json(
        "POST",
        PERIOD_USAGE_URL,
        {**headers, "Authorization": f"Bearer {token}"},
        b"{}",
    )
    return rpc


def take_snapshot() -> dict[str, Any] | None:
    try:
        token, user_id, local_plan = _local_session()
        if not token:
            return None
        raw = fetch_raw(token, user_id)
        return normalize_payload(raw or {}, fallback_plan=local_plan or "")
    except Exception:
        return None


def main(argv: list[str]) -> int:
    cmd = argv[1] if len(argv) > 1 else "snapshot"
    if cmd == "snapshot":
        snap = take_snapshot()
        if not snap:
            return 1
        json.dump(snap, sys.stdout, ensure_ascii=False)
        sys.stdout.write("\n")
        return 0
    if cmd == "delta":
        if len(argv) != 4:
            print("usage: cursor_usage.py delta <start.json> <end.json>", file=sys.stderr)
            return 2
        with open(argv[2], encoding="utf-8") as fh:
            start = json.load(fh)
        with open(argv[3], encoding="utf-8") as fh:
            end = json.load(fh)
        delta = compute_delta(start, end)
        if not delta:
            return 1
        json.dump(delta, sys.stdout, ensure_ascii=False)
        sys.stdout.write("\n")
        return 0
    print("usage: cursor_usage.py [snapshot|delta]", file=sys.stderr)
    return 2


if __name__ == "__main__":
    raise SystemExit(main(sys.argv))
