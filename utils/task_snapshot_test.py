#!/usr/bin/env python3
import json
import tempfile
import unittest
from pathlib import Path

import task_snapshot as ts


class ParkSlotTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.dir = Path(self.tmp.name)
        self.snap = self.dir / "current.json"
        self.events = self.dir / "events.jsonl"

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_park_keeps_slot_and_releases_root(self) -> None:
        ts.write_snap(
            str(self.snap),
            {
                "alias": "EK",
                "status": "in_progress",
                "task_id": "VOICE",
                "branch": "feat/voice",
                "services": ["LLM"],
                "tasks": [
                    {
                        "task_id": "VOICE",
                        "status": "in_progress",
                        "branch": "feat/voice",
                        "services": ["LLM"],
                        "started_at": "2026-09-17T10:00:00Z",
                    },
                    {
                        "task_id": "TRANSLATE",
                        "status": "in_progress",
                        "branch": "feat/tr",
                        "services": ["LLM", "Core"],
                        "started_at": "2026-09-17T12:00:00Z",
                    },
                ],
            },
        )
        rc = ts.park_once(
            str(self.snap),
            str(self.events),
            "EK",
            "VOICE",
            "1700000000",
            "2026-09-17T21:00:00Z",
        )
        self.assertEqual(rc, 0)
        snap = json.loads(self.snap.read_text())
        by_id = {row["task_id"]: row for row in snap["tasks"]}
        self.assertEqual(by_id["VOICE"]["status"], "parked")
        self.assertEqual(by_id["TRANSLATE"]["status"], "in_progress")
        self.assertEqual(snap["task_id"], "TRANSLATE")
        self.assertEqual(snap["status"], "in_progress")
        row = json.loads(self.events.read_text().strip().splitlines()[-1])
        self.assertEqual(row["event"], "task_parked")
        self.assertEqual(row["task_id"], "VOICE")

    def test_park_idempotent(self) -> None:
        ts.write_snap(
            str(self.snap),
            {
                "alias": "EK",
                "tasks": [{"task_id": "VOICE", "status": "in_progress", "services": ["LLM"]}],
            },
        )
        self.assertEqual(
            ts.park_once(str(self.snap), str(self.events), "EK", "VOICE", "1", "2026-09-17T21:00:00Z"),
            0,
        )
        self.assertEqual(
            ts.park_once(str(self.snap), str(self.events), "EK", "VOICE", "2", "2026-09-17T21:01:00Z"),
            2,
        )
        lines = [ln for ln in self.events.read_text().splitlines() if ln.strip()]
        self.assertEqual(len(lines), 1)


class RelatedSlotTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tmp = tempfile.TemporaryDirectory()
        self.snap = Path(self.tmp.name) / "current.json"

    def tearDown(self) -> None:
        self.tmp.cleanup()

    def test_stores_related_and_keeps_on_empty_resume(self) -> None:
        ts.upsert_started(
            str(self.snap),
            "EK",
            "2026-09-18T00:00:00Z",
            "FIX-1",
            "fix/x",
            "Core",
            "",
            "",
            "broken translate",
            "",
            "20260917-2037-EK-INBOX_MESSAGE_TRANSLATION,FIX-1",
            "docs/20260917-2037-EK-INBOX_MESSAGE_TRANSLATION.md,Common/docs/README.md",
        )
        slot = json.loads(self.snap.read_text())["tasks"][0]
        self.assertEqual(slot["related_tasks"], ["20260917-2037-EK-INBOX_MESSAGE_TRANSLATION"])
        self.assertEqual(
            slot["related_docs"],
            ["docs/20260917-2037-EK-INBOX_MESSAGE_TRANSLATION.md", "docs/README.md"],
        )
        ts.upsert_started(
            str(self.snap),
            "EK",
            "2026-09-18T00:01:00Z",
            "FIX-1",
            "fix/x",
            "Core",
            "",
            "",
            "broken translate",
            "",
            "",
            "",
        )
        slot = json.loads(self.snap.read_text())["tasks"][0]
        self.assertEqual(slot["related_tasks"], ["20260917-2037-EK-INBOX_MESSAGE_TRANSLATION"])
        self.assertEqual(slot["started_at"], "2026-09-18T00:00:00Z")


if __name__ == "__main__":
    unittest.main()
