#!/usr/bin/env python3
"""Daily Coordinator update check.

Source of truth: commit date of origin/main on the public .coordinator repo.
Cache: $DATA_DIR/.cache/last_update_check (gitignored with the rest of .cache).
sessionStart reads session_notice() — fail open, never block the chat.
"""

from __future__ import annotations

import json
import os
import subprocess
import sys
from datetime import datetime, timedelta, timezone
from pathlib import Path

TTL = timedelta(hours=24)
FETCH_TIMEOUT = 20

_PACK_HOOKS = Path(__file__).resolve().parent.parent / "pack" / "cursor-hooks"
if str(_PACK_HOOKS) not in sys.path:
    sys.path.insert(0, str(_PACK_HOOKS))
from coord_paths import load_paths  # noqa: E402


def _now() -> datetime:
    return datetime.now(timezone.utc)


def _parse_iso(raw: str) -> datetime | None:
    text = (raw or "").strip()
    if not text:
        return None
    try:
        return datetime.fromisoformat(text.replace("Z", "+00:00"))
    except ValueError:
        return None


def _git(coord: Path, *args: str, timeout: int = 8) -> str:
    env = os.environ.copy()
    env["GIT_TERMINAL_PROMPT"] = "0"
    out = subprocess.check_output(
        ["git", "-C", str(coord), *args],
        stderr=subprocess.DEVNULL,
        timeout=timeout,
        env=env,
    )
    return out.decode().strip()


def _chat_language(settings: Path) -> str:
    path = settings / "locale.json"
    if not path.is_file():
        return "en"
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return "en"
    lang = str(data.get("chat_language") or "en").strip().lower()
    return lang if lang else "en"


def _cache_path(data_dir: Path) -> Path:
    return data_dir / ".cache" / "last_update_check"


def _load_cache(path: Path) -> dict:
    if not path.is_file():
        return {}
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return {}
    return data if isinstance(data, dict) else {}


def _write_cache(path: Path, data: dict) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    payload = dict(data)
    raw = json.dumps(payload, indent=2, ensure_ascii=False) + "\n"
    tmp = path.with_suffix(".tmp")
    tmp.write_text(raw, encoding="utf-8")
    tmp.replace(path)


def _fresh(cache: dict, now: datetime) -> bool:
    checked = _parse_iso(str(cache.get("checked_at") or ""))
    if checked is None:
        return False
    if checked.tzinfo is None:
        checked = checked.replace(tzinfo=timezone.utc)
    return now - checked < TTL


def _probe(coord: Path) -> dict:
    local_sha = _git(coord, "rev-parse", "HEAD")
    local_date = _git(coord, "log", "-1", "--format=%cI", "HEAD")
    _git(coord, "fetch", "--quiet", "origin", "main", timeout=FETCH_TIMEOUT)
    remote_sha = _git(coord, "rev-parse", "origin/main")
    remote_date = _git(coord, "log", "-1", "--format=%cI", "origin/main")
    behind = int(_git(coord, "rev-list", "--count", "HEAD..origin/main") or "0")
    return {
        "local_sha": local_sha,
        "local_date": local_date,
        "remote_sha": remote_sha,
        "remote_date": remote_date,
        "behind": behind,
        "update_available": behind > 0,
    }


def run_check(*, force: bool = False) -> dict:
    paths = load_paths()
    coord = paths["COORDINATOR_ROOT"]
    cache_file = _cache_path(paths["DATA_DIR"])
    now = _now()
    prev = _load_cache(cache_file)

    try:
        local_sha = _git(coord, "rev-parse", "HEAD")
    except Exception:
        local_sha = str(prev.get("local_sha") or "")

    if not force and _fresh(prev, now):
        out = dict(prev)
        out["local_sha"] = local_sha
        if local_sha and local_sha == out.get("remote_sha") and out.get("update_available"):
            out["update_available"] = False
            out["behind"] = 0
            _write_cache(cache_file, out)
        return out

    try:
        probed = _probe(coord)
    except Exception as err:
        out = dict(prev)
        out["checked_at"] = now.isoformat().replace("+00:00", "Z")
        out["error"] = str(err)[:200]
        out["local_sha"] = local_sha
        _write_cache(cache_file, out)
        return out

    out = {
        "checked_at": now.isoformat().replace("+00:00", "Z"),
        **probed,
    }
    _write_cache(cache_file, out)
    return out


def format_notice(data: dict) -> str:
    if not data.get("update_available"):
        return ""
    remote_date = str(data.get("remote_date") or "")
    day = remote_date[:10] if remote_date else ""
    sha = str(data.get("remote_sha") or "")[:7]
    behind = data.get("behind") or 0
    lang = _chat_language(load_paths()["SETTINGS_DIR"])
    if lang == "ru":
        when = f" от {day}" if day else ""
        extra = f", {behind} коммит(ов)" if behind else ""
        return (
            f"Отто: вышло обновление координатора (репа origin/main{when}"
            f"{extra}, {sha}). Обновить? По «да»: из `.coordinator` запусти "
            "`./utils/apply_update.sh` — pull, пересборка API и UI, рестарт на "
            "PORT/UI_PORT из `.env` этого проекта."
        )
    when = f" dated {day}" if day else ""
    extra = f", {behind} commit(s)" if behind else ""
    return (
        f"Otto: a Coordinator update is available (origin/main{when}"
        f"{extra}, {sha}). Update? On yes: from `.coordinator` run "
        "`./utils/apply_update.sh` — pull, rebuild API and UI, restart on this "
        "project's PORT/UI_PORT from `.env`."
    )


def session_notice() -> str:
    try:
        return format_notice(run_check(force=False))
    except Exception:
        return ""


def main() -> int:
    force = "--force" in sys.argv[1:]
    notice = "--notice" in sys.argv[1:]
    data = run_check(force=force)
    if notice:
        text = format_notice(data)
        if text:
            print(text)
        return 0
    json.dump(data, sys.stdout, indent=2, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    try:
        raise SystemExit(main())
    except Exception:
        sys.exit(0)
