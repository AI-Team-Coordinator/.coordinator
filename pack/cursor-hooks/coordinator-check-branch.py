#!/usr/bin/env python3
"""PreToolUse: block product edits on the wrong branch or without an open slot."""

from __future__ import annotations

import json
import subprocess
import sys
from pathlib import Path

from coord_paths import current_alias, load_paths


def extract_path(payload: dict) -> str:
    args = payload.get("arguments") or payload.get("tool_input") or payload.get("input") or payload
    if isinstance(args, dict):
        return str(args.get("path") or payload.get("path") or "").strip()
    return str(payload.get("path") or "").strip()


def allow() -> None:
    print('{"permission":"allow"}')
    raise SystemExit(0)


def deny(msg: str) -> None:
    body = json.dumps({"permission": "deny", "user_message": msg, "agent_message": msg})
    print(body)
    raise SystemExit(0)


def is_under(path: Path, root: Path) -> bool:
    try:
        path.resolve().relative_to(root.resolve())
        return True
    except (ValueError, OSError):
        return False


def git_toplevel(dir_path: Path) -> Path | None:
    try:
        out = subprocess.check_output(
            ["git", "-C", str(dir_path), "rev-parse", "--show-toplevel"],
            stderr=subprocess.DEVNULL,
            text=True,
        ).strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return None
    return Path(out) if out else None


def git_branch(repo: Path) -> str:
    try:
        return subprocess.check_output(
            ["git", "-C", str(repo), "branch", "--show-current"],
            stderr=subprocess.DEVNULL,
            text=True,
        ).strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return ""


def norm(s: str) -> str:
    s = (s or "").strip().lower().lstrip(".")
    return s.replace("-", "_").replace(" ", "_")


def slots(snap: dict) -> list[dict]:
    raw = snap.get("tasks")
    if isinstance(raw, list) and raw:
        return [t for t in raw if isinstance(t, dict) and t.get("task_id")]
    if snap.get("status") == "in_progress" and snap.get("task_id"):
        return [snap]
    return []


def claims(task: dict, repo: str) -> bool:
    svcs = task.get("services") or []
    if not svcs:
        return True
    rn = norm(repo)
    return any(norm(str(svc)) == rn for svc in svcs)


def load_snap(path: Path) -> dict | None:
    if not path.is_file():
        return None
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except Exception:
        return None
    return data if isinstance(data, dict) else None


def main() -> None:
    try:
        payload = json.load(sys.stdin)
    except Exception:
        allow()
    if not isinstance(payload, dict):
        allow()

    raw = extract_path(payload)
    if not raw:
        allow()

    file_path = Path(raw)
    if not file_path.is_absolute():
        file_path = (Path.cwd() / file_path).resolve()
    else:
        file_path = file_path.resolve()

    try:
        paths = load_paths(Path(__file__))
    except Exception:
        allow()

    for root in (
        paths["CURSOR_DIR"],
        paths["COORDINATOR_ROOT"],
        paths["DATA_DIR"],
        paths["DOCS_DIR"],
    ):
        if root and is_under(file_path, root):
            allow()

    repo = git_toplevel(file_path.parent)
    if repo is None:
        allow()

    repo_name = repo.name
    if repo_name in {".cursor", ".coordinator"}:
        allow()

    author = current_alias(paths)
    if not author:
        deny(
            "No current author. Set DATA_DIR/.current_author (first non-comment line = alias) before editing product code."
        )

    state_file = paths["PROGRESS_DIR"] / f".current_task_{author}"
    snap = load_snap(state_file)
    branch = git_branch(repo)

    if branch in {"main", "master"}:
        if snap:
            for task in slots(snap):
                tb = (task.get("branch") or "").strip()
                if claims(task, repo_name) and tb == branch:
                    allow()
        deny(
            f"Direct editing of product code on '{branch}' in '{repo_name}' is blocked by AI Coordinator. "
            "Start a task with a dedicated feature branch first."
        )

    if not snap:
        deny(
            f"No open task slot claims '{repo_name}' for author '{author}'. "
            "Editing is blocked. Start or extend a task that lists this service."
        )

    matched = [t for t in slots(snap) if claims(t, repo_name)]
    if not matched:
        deny(
            f"No open task slot claims '{repo_name}' for author '{author}'. "
            "Editing is blocked. Start or extend a task that lists this service."
        )
    for task in matched:
        tb = (task.get("branch") or "").strip()
        if tb == branch:
            allow()
    task = matched[0]
    deny(
        f"Branch mismatch in '{repo_name}': open slot '{task.get('task_id')}' expects branch "
        f"'{task.get('branch')}', but current branch is '{branch}'. Switch before editing."
    )


if __name__ == "__main__":
    main()
