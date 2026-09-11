#!/usr/bin/env python3
"""afterShellExecution: log repo_merged when a product main is merged/pushed."""

from __future__ import annotations

import json
import os
import re
import subprocess
import sys
import time
from pathlib import Path

from coord_paths import current_alias, load_paths

SKIP = {".cursor", ".coordinator"}
DEDUP_SEC = 180


def main() -> None:
    try:
        payload = json.load(sys.stdin)
    except Exception:
        print("{}")
        return
    try:
        maybe_log(payload.get("command") or "", payload.get("output") or "", payload.get("cwd") or "")
    except Exception:
        pass
    print("{}")


def maybe_log(command: str, output: str, cwd: str) -> None:
    if not looks_like_main_publish(command, output):
        return
    p = load_paths(Path(__file__))
    repo_dir = extract_repo_dir(command, cwd, p["WORKSPACE_ROOT"])
    if repo_dir is None:
        return
    folder = repo_dir.name
    if folder in SKIP:
        return
    if folder in {p["COORDINATOR_ROOT"].name}:
        return
    service = service_for_folder(folder, p)
    if not service:
        return
    if recently_logged(service["id"], service["name"], p):
        return
    log_event = p["COORDINATOR_ROOT"] / "utils" / "log_event.sh"
    if not log_event.is_file():
        return
    subprocess.run(
        [str(log_event), "repo_merged", f"service={service['name']}", f"repo={folder}"],
        check=False,
        stdout=subprocess.DEVNULL,
        stderr=subprocess.DEVNULL,
        env={**os.environ, "WORKSPACE_ROOT": str(p["WORKSPACE_ROOT"])},
    )


def looks_like_main_publish(command: str, output: str) -> bool:
    if re.search(r"! \[rejected\]|pre-receive hook declined|Merge conflict|Automatic merge failed", output):
        return False
    if re.search(r"\bmain\s+->\s+main\b", output) and re.search(r"\bgit\s+push\b", command, re.I):
        return True
    if not re.search(r"\bgit\s+merge\b", command, re.I):
        return False
    if not re.search(r"Merge made by|Fast-forward", output):
        return False
    if re.search(r"merge\s+origin/main\b", command, re.I):
        return False
    if re.search(r"merge\s+main\b", command, re.I) and not re.search(r"checkout\s+main\b", command, re.I):
        return False
    if re.search(r"merge\s+[^\n]*\b(?:feat|fix)/", command, re.I):
        return True
    return bool(re.search(r"checkout\s+main\b", command, re.I) and re.search(r"\bgit\s+merge\b", command, re.I))


def extract_repo_dir(command: str, cwd: str, workspace: Path) -> Path | None:
    candidates: list[Path] = []
    if cwd:
        candidates.append(Path(cwd).expanduser())
    for match in re.finditer(r"(?:^|[;&\n]|\|\||&&)\s*cd\s+([^\s;&|]+)", command):
        candidates.append(resolve_path(match.group(1).strip("\"'"), workspace))
    match = re.search(r"git\s+-C\s+([^\s]+)", command)
    if match:
        candidates.append(resolve_path(match.group(1).strip("\"'"), workspace))
    for path in reversed(candidates):
        if (path / ".git").exists():
            return path
    return None


def resolve_path(raw: str, workspace: Path) -> Path:
    path = Path(raw).expanduser()
    if not path.is_absolute():
        path = (workspace / path).resolve()
    return path


def service_for_folder(folder: str, p: dict[str, Path]) -> dict[str, str] | None:
    profile = p["SETTINGS_DIR"] / "project_profile.json"
    if not profile.is_file():
        return {"id": folder.lower(), "name": folder}
    data = json.loads(profile.read_text(encoding="utf-8"))
    for svc in data.get("services") or []:
        if svc.get("group") == "workspace" or svc.get("kind") == "workspace":
            continue
        if (svc.get("repo") or "").strip() == folder:
            return {"id": svc.get("id") or folder.lower(), "name": svc.get("name") or folder}
    return {"id": folder.lower(), "name": folder}


def recently_logged(service_id: str, service_name: str, p: dict[str, Path]) -> bool:
    alias = current_alias(p)
    if not alias:
        return False
    year = time.strftime("%Y")
    path = p["PROGRESS_DIR"] / "events" / alias / f"{year}.jsonl"
    if not path.is_file():
        return False
    now = int(time.time())
    try:
        lines = path.read_text(encoding="utf-8").splitlines()[-30:]
    except OSError:
        return False
    for line in reversed(lines):
        line = line.strip()
        if not line:
            continue
        try:
            row = json.loads(line)
        except json.JSONDecodeError:
            continue
        if row.get("event") != "repo_merged":
            continue
        svc = row.get("service") or ""
        if svc not in {service_id, service_name}:
            continue
        ts = int(row.get("timestamp") or 0)
        return now - ts <= DEDUP_SEC
    return False


if __name__ == "__main__":
    main()
