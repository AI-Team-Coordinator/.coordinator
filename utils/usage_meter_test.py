#!/usr/bin/env python3
import unittest

from usage_meter import bind_target


class BindTargetTests(unittest.TestCase):
    def test_research_session(self):
        snap = {
            "research": {"status": "active", "session_id": "r1", "summary": "how hours"},
            "tasks": [{"task_id": "T1", "session_ids": ["t1"]}],
        }
        self.assertEqual(bind_target(snap, "r1"), ("", "research", "how hours"))

    def test_task_session_ids(self):
        snap = {
            "tasks": [
                {"task_id": "T1", "session_id": "old", "session_ids": ["a", "b"], "summary": "voice"},
            ]
        }
        self.assertEqual(bind_target(snap, "b"), ("T1", "task", "voice"))

    def test_unbound(self):
        self.assertEqual(bind_target({"tasks": []}, "x"), ("", "", ""))


if __name__ == "__main__":
    unittest.main()
