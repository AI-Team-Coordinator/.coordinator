#!/usr/bin/env python3
"""Persist coordinator work tasks (backlog) on the progress bus."""

from __future__ import annotations

import json
import os
import sys
from datetime import datetime, timezone

from task_snapshot import (
    find_task_doc,
    load_snap,
    normalize_related_docs,
    normalize_related_tasks,
    slots_from,
)


def work_tasks_dir(progress_dir: str) -> str:
    return os.path.join(progress_dir, "work_tasks")


def work_task_path(progress_dir: str, task_id: str) -> str:
    safe = (task_id or "").strip().replace("/", "_")
    return os.path.join(work_tasks_dir(progress_dir), safe + ".json")


def _iso_now() -> str:
    return datetime.now(timezone.utc).strftime("%Y-%m-%dT%H:%M:%SZ")


def load_work_task(progress_dir: str, task_id: str) -> dict:
    path = work_task_path(progress_dir, task_id)
    if not os.path.isfile(path):
        return {}
    try:
        data = json.loads(open(path).read())
    except Exception:
        return {}
    return data if isinstance(data, dict) else {}


def write_work_task(progress_dir: str, row: dict) -> None:
    tid = str(row.get("task_id") or "").strip()
    if not tid:
        return
    path = work_task_path(progress_dir, tid)
    os.makedirs(os.path.dirname(path), exist_ok=True)
    with open(path, "w") as f:
        json.dump(row, f, indent=2, ensure_ascii=False)
        f.write("\n")


def canonical_branch(alias: str, branch: str) -> str:
    b = (branch or "").strip()
    if not b:
        return ""
    low = b.lower()
    if low in ("main", "master"):
        return b
    owner = (alias or "").strip().lower()
    if not owner:
        return b
    prefix = owner + "/"
    if low.startswith(prefix):
        return b[len(prefix) :]
    return b


def upsert_on_planned(
    progress_dir: str,
    common_root: str,
    alias: str,
    task_id: str,
    branch: str,
    services_csv: str,
    doc_path: str,
    summary: str,
    assignee: str,
    related_tasks: str,
    related_docs: str,
) -> dict:
    iso = _iso_now()
    existing = load_work_task(progress_dir, task_id)
    services = [s.strip() for s in (services_csv or "").split(",") if s.strip()]
    doc = (doc_path or "").strip() or find_task_doc(common_root, task_id)
    text = " ".join((summary or "").split())
    assignee = (assignee or alias or "").strip()
    row = {
        "task_id": task_id,
        "status": "backlog",
        "created_by": (existing.get("created_by") or alias or "").strip(),
        "assignee": assignee,
        "branch": canonical_branch(assignee, branch or existing.get("branch") or ""),
        "services": services or existing.get("services") or [],
        "created_at": existing.get("created_at") or iso,
        "updated_at": iso,
    }
    if doc:
        row["doc"] = doc
    elif existing.get("doc"):
        row["doc"] = existing["doc"]
    if text:
        row["summary"] = text[:280]
    elif existing.get("summary"):
        row["summary"] = existing["summary"]
    tasks = normalize_related_tasks(related_tasks, task_id)
    docs = normalize_related_docs(related_docs)
    if not tasks:
        tasks = normalize_related_tasks(existing.get("related_tasks") or [], task_id)
    if not docs:
        docs = normalize_related_docs(existing.get("related_docs") or [])
    if tasks:
        row["related_tasks"] = tasks
    if docs:
        row["related_docs"] = docs
    write_work_task(progress_dir, row)
    return row


def upsert_on_started(
    progress_dir: str,
    common_root: str,
    alias: str,
    task_id: str,
    branch: str,
    services_csv: str,
    doc_path: str,
    summary: str,
    related_tasks: str,
    related_docs: str,
) -> dict:
    iso = _iso_now()
    existing = load_work_task(progress_dir, task_id)
    services = [s.strip() for s in (services_csv or "").split(",") if s.strip()]
    doc = (doc_path or "").strip() or find_task_doc(common_root, task_id)
    text = " ".join((summary or "").split())
    canon = canonical_branch(alias, branch or existing.get("branch") or "")
    row = {
        "task_id": task_id,
        "status": "open",
        "created_by": (existing.get("created_by") or alias or "").strip(),
        "assignee": (existing.get("assignee") or alias or "").strip(),
        "branch": canon,
        "services": services or existing.get("services") or [],
        "created_at": existing.get("created_at") or iso,
        "updated_at": iso,
    }
    if doc:
        row["doc"] = doc
    elif existing.get("doc"):
        row["doc"] = existing["doc"]
    if text:
        row["summary"] = text[:280]
    elif existing.get("summary"):
        row["summary"] = existing["summary"]
    tasks = normalize_related_tasks(related_tasks, task_id)
    docs = normalize_related_docs(related_docs)
    if not tasks:
        tasks = normalize_related_tasks(existing.get("related_tasks") or [], task_id)
    if not docs:
        docs = normalize_related_docs(existing.get("related_docs") or [])
    if tasks:
        row["related_tasks"] = tasks
    if docs:
        row["related_docs"] = docs
    write_work_task(progress_dir, row)
    return row


def mark_done(progress_dir: str, task_id: str) -> None:
    existing = load_work_task(progress_dir, task_id)
    if not existing:
        return
    iso = _iso_now()
    existing["status"] = "done"
    existing["updated_at"] = iso
    existing["completed_at"] = iso
    write_work_task(progress_dir, existing)


def _cli_planned() -> None:
    if len(sys.argv) < 14 or sys.argv[1] != "planned":
        raise SystemExit(1)
    (
        _,
        _,
        progress_dir,
        common_root,
        alias,
        _iso,
        task_id,
        branch,
        services_csv,
        doc_path,
        summary,
        assignee,
        related_tasks,
        related_docs,
    ) = sys.argv[:14]
    upsert_on_planned(
        progress_dir,
        common_root,
        alias,
        task_id,
        branch,
        services_csv,
        doc_path,
        summary,
        assignee,
        related_tasks,
        related_docs,
    )


if __name__ == "__main__":
    _cli_planned()
