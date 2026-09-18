#!/usr/bin/env python3
"""Candidate related docs (README registry) and tasks (event log) for a FIX."""

from __future__ import annotations

import json
import os
import re
import sys


def workspace_roots() -> tuple[str, str]:
    here = os.path.dirname(os.path.abspath(__file__))
    coord = os.path.dirname(here)
    common = os.path.abspath(os.path.join(coord, "..", "Common"))
    return coord, common


def tokens(text: str) -> list[str]:
    return [t for t in re.split(r"[^a-z0-9а-яё_+-]+", (text or "").lower()) if len(t) >= 3]


def score(hay: str, needles: list[str]) -> int:
    blob = hay.lower()
    return sum(1 for n in needles if n in blob)


def scan_readme(docs_root: str, needles: list[str]) -> list[dict]:
    readme = os.path.join(docs_root, "README.md")
    if not os.path.isfile(readme):
        return []
    out: list[dict] = []
    seen: set[str] = set()
    link = re.compile(r"\[([^\]]+)\]\(\./([^)]+\.md)\)")
    for line in open(readme, encoding="utf-8", errors="replace"):
        n = score(line, needles)
        if n == 0:
            continue
        for title, rel in link.findall(line):
            path = "docs/" + rel.lstrip("./")
            if path in seen:
                continue
            seen.add(path)
            out.append({"path": path, "title": title.strip(), "hits": n})
    out.sort(key=lambda r: (-r["hits"], r["path"]))
    return out[:12]


def scan_events(events_dir: str, needles: list[str]) -> list[dict]:
    if not os.path.isdir(events_dir):
        return []
    acc: dict[str, dict] = {}
    for dirpath, _, files in os.walk(events_dir):
        for name in files:
            if not name.endswith(".jsonl"):
                continue
            path = os.path.join(dirpath, name)
            try:
                lines = open(path, encoding="utf-8", errors="replace")
            except OSError:
                continue
            for raw in lines:
                raw = raw.strip()
                if not raw:
                    continue
                try:
                    row = json.loads(raw)
                except Exception:
                    continue
                if row.get("event") not in ("task_started", "task_completed", "task_parked"):
                    continue
                tid = str(row.get("task_id") or "").strip()
                if not tid:
                    continue
                blob = " ".join(
                    [
                        tid,
                        str(row.get("summary") or ""),
                        str(row.get("branch") or ""),
                        str(row.get("service") or ""),
                    ]
                )
                n = score(blob, needles)
                if n == 0:
                    continue
                prev = acc.get(tid)
                if prev is None or n >= prev["hits"]:
                    acc[tid] = {
                        "task_id": tid,
                        "summary": str(row.get("summary") or "").strip(),
                        "branch": str(row.get("branch") or "").strip(),
                        "hits": n,
                    }
    out = list(acc.values())
    out.sort(key=lambda r: (-r["hits"], r["task_id"]))
    return out[:12]


def main() -> None:
    query = " ".join(sys.argv[1:]).strip()
    if not query:
        raise SystemExit('usage: suggest_related.py "<words>" [Service,...]')
    needles = tokens(query)
    if not needles:
        print(json.dumps({"docs": [], "tasks": []}, ensure_ascii=False, indent=2))
        return
    _, common = workspace_roots()
    docs_root = os.path.join(common, "docs")
    events_dir = os.path.join(common, "data", "progress", "events")
    print(
        json.dumps(
            {"docs": scan_readme(docs_root, needles), "tasks": scan_events(events_dir, needles)},
            ensure_ascii=False,
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
