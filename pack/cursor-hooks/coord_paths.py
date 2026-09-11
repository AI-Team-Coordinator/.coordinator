"""Resolve Coordinator bus paths from .coordinator/.env.

Used by host hooks in <workspace>/.cursor/hooks/ after install.
Also works when run from this pack folder (walks up to .coordinator).
"""

from __future__ import annotations

import os
from pathlib import Path


def _parse_env(path: Path) -> dict[str, str]:
    out: dict[str, str] = {}
    if not path.is_file():
        return out
    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, _, val = line.partition("=")
        key = key.strip()
        val = val.strip().strip("'").strip('"')
        if key and key not in os.environ:
            out[key] = val
        elif key and key in os.environ:
            out[key] = os.environ[key]
    for key, val in os.environ.items():
        if key in {"WORKSPACE_ROOT", "DATA_DIR", "DOCS_DIR", "BUS_DIR", "CURSOR_DIR", "PORT"}:
            out[key] = val
    return out


def _abs(coord_root: Path, value: str) -> Path:
    p = Path(value).expanduser()
    if p.is_absolute():
        return p.resolve()
    return (coord_root / p).resolve()


def find_coordinator_root(start: Path | None = None) -> Path:
    here = (start or Path(__file__).resolve()).parent
    for p in [here, *here.parents]:
        if p.name == ".coordinator" and ((p / ".env").is_file() or (p / ".env.example").is_file()):
            return p
        nested = p / ".coordinator"
        if nested.is_dir() and ((nested / ".env").is_file() or (nested / ".env.example").is_file()):
            return nested
    raise FileNotFoundError("could not find .coordinator (looked for .env / .env.example)")


def load_paths(start: Path | None = None) -> dict[str, Path]:
    coord = find_coordinator_root(start)
    env = _parse_env(coord / ".env")
    workspace = _abs(coord, env.get("WORKSPACE_ROOT", ".."))
    bus = _abs(coord, env.get("BUS_DIR", env.get("COMMON_DIR", "..")))
    if env.get("DATA_DIR"):
        data = _abs(coord, env["DATA_DIR"])
    else:
        data = bus / "coordinator-data"
        if not data.is_dir() and (bus / "data").is_dir():
            data = bus / "data"
    docs = _abs(coord, env["DOCS_DIR"]) if env.get("DOCS_DIR") else bus / "docs"
    cursor = _abs(coord, env.get("CURSOR_DIR", "../.cursor"))
    return {
        "COORDINATOR_ROOT": coord,
        "WORKSPACE_ROOT": workspace,
        "BUS_DIR": bus,
        "DATA_DIR": data,
        "DOCS_DIR": docs,
        "CURSOR_DIR": cursor,
        "PROGRESS_DIR": data / "progress",
        "SETTINGS_DIR": data / "settings",
        "AUTHOR_FILE": data / ".current_author",
        "CACHE_DIR": data / ".cache",
    }


def current_alias(paths: dict[str, Path] | None = None) -> str:
    paths = paths or load_paths()
    for path in (paths["AUTHOR_FILE"], paths["CURSOR_DIR"] / ".current_author"):
        if not path.is_file():
            continue
        for line in path.read_text(encoding="utf-8").splitlines():
            line = line.strip()
            if line and not line.startswith("#"):
                return line
    return ""


if __name__ == "__main__":
    p = load_paths()
    for k in ("COORDINATOR_ROOT", "WORKSPACE_ROOT", "BUS_DIR", "DATA_DIR", "DOCS_DIR", "CURSOR_DIR"):
        print(f"{k}={p[k]}")
    print(f"ALIAS={current_alias(p) or '(none)'}")
