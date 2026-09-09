#!/usr/bin/env python3
import unittest

from cursor_usage import compute_delta, normalize_payload, spend_kind


class NormalizeTests(unittest.TestCase):
    def test_usage_summary_ultra(self):
        snap = normalize_payload(
            {
                "billingCycleStart": "2026-07-04T00:35:51.000Z",
                "membershipType": "ultra",
                "individualUsage": {
                    "plan": {
                        "used": 40000,
                        "limit": 40000,
                        "remaining": 0,
                        "autoPercentUsed": 98.109,
                        "apiPercentUsed": 100,
                    },
                    "onDemand": {"enabled": False, "used": 0},
                },
            }
        )
        self.assertEqual(snap["plan"], "ultra")
        self.assertEqual(snap["plan_price_usd"], 200)
        self.assertEqual(snap["included_cents"], 40000)
        self.assertEqual(snap["ondemand_cents"], 0)
        self.assertAlmostEqual(snap["cursor_models_pct"], 98.109)
        self.assertEqual(snap["other_models_pct"], 100)

    def test_period_usage_cents(self):
        snap = normalize_payload(
            {
                "billingCycleStart": "1768399334000",
                "planUsage": {
                    "totalSpend": 23222,
                    "includedSpend": 12000,
                    "autoPercentUsed": 15.4,
                    "apiPercentUsed": 46.4,
                },
                "spendLimitUsage": {"individualUsed": 250},
            },
            fallback_plan="pro",
        )
        self.assertEqual(snap["plan"], "pro")
        self.assertEqual(snap["included_cents"], 12000)
        self.assertEqual(snap["ondemand_cents"], 250)
        self.assertAlmostEqual(snap["cursor_models_pct"], 15.4)

    def test_empty_payload(self):
        self.assertIsNone(normalize_payload({}))


class DeltaTests(unittest.TestCase):
    def test_money_and_pct(self):
        start = {
            "plan": "ultra",
            "billing_cycle_start": "cycle-a",
            "included_cents": 10000,
            "ondemand_cents": 0,
            "cursor_models_pct": 10.0,
            "other_models_pct": 4.0,
        }
        end = {
            "plan": "ultra",
            "billing_cycle_start": "cycle-a",
            "included_cents": 11250,
            "ondemand_cents": 80,
            "cursor_models_pct": 11.5,
            "other_models_pct": 4.0,
        }
        delta = compute_delta(start, end)
        self.assertEqual(delta["cost_usd"], 0.8)
        self.assertEqual(delta["ondemand_usd"], 0.8)
        self.assertEqual(delta["budget_usd"], 1.5)
        self.assertEqual(delta["plan_price_usd"], 200)
        self.assertEqual(delta["cursor_models_pct"], 1.5)
        self.assertEqual(delta["other_models_pct"], 0.0)
        self.assertEqual(delta["usage_plan"], "ultra")

    def test_cycle_reset_is_unreliable(self):
        start = {
            "billing_cycle_start": "old",
            "included_cents": 40000,
            "ondemand_cents": 0,
            "cursor_models_pct": 90,
            "other_models_pct": 100,
        }
        end = {
            "billing_cycle_start": "new",
            "included_cents": 100,
            "ondemand_cents": 0,
            "cursor_models_pct": 1,
            "other_models_pct": 0,
        }
        self.assertIsNone(compute_delta(start, end))

    def test_negative_clamped(self):
        start = {
            "billing_cycle_start": "c",
            "included_cents": 500,
            "ondemand_cents": 10,
            "cursor_models_pct": 8,
            "other_models_pct": 3,
        }
        end = {
            "billing_cycle_start": "c",
            "included_cents": 400,
            "ondemand_cents": 10,
            "cursor_models_pct": 8,
            "other_models_pct": 3,
        }
        delta = compute_delta(start, end)
        self.assertEqual(delta["cost_usd"], 0.0)
        self.assertEqual(delta["budget_usd"], 0.0)

    def test_subscription_share_pro_plus(self):
        start = {
            "plan": "pro_plus",
            "plan_price_usd": 60,
            "billing_cycle_start": "c",
            "included_cents": 0,
            "ondemand_cents": 0,
            "cursor_models_pct": 40.0,
            "other_models_pct": 10.0,
        }
        end = {
            "plan": "pro_plus",
            "plan_price_usd": 60,
            "billing_cycle_start": "c",
            "included_cents": 0,
            "ondemand_cents": 0,
            "cursor_models_pct": 50.0,
            "other_models_pct": 15.0,
        }
        delta = compute_delta(start, end)
        self.assertEqual(delta["budget_usd"], 4.5)
        self.assertEqual(delta["cost_usd"], 0.0)
        self.assertEqual(delta["ondemand_usd"], 0.0)

    def test_overflow_counts_only_included_then_ondemand(self):
        start = {
            "plan": "pro_plus",
            "billing_cycle_start": "c",
            "included_cents": 7000,
            "ondemand_cents": 0,
            "cursor_models_pct": 100.0,
            "other_models_pct": 98.0,
        }
        end = {
            "plan": "pro_plus",
            "billing_cycle_start": "c",
            "included_cents": 7000,
            "ondemand_cents": 250,
            "cursor_models_pct": 120.0,
            "other_models_pct": 110.0,
        }
        delta = compute_delta(start, end)
        self.assertEqual(delta["budget_usd"], 0.6)
        self.assertEqual(delta["ondemand_usd"], 2.5)
        self.assertEqual(delta["cost_usd"], 2.5)

    def test_spend_kind_infra_for_common_and_cursor(self):
        self.assertEqual(spend_kind(["Common"]), "infra")
        self.assertEqual(spend_kind([".cursor", "common"]), "infra")
        self.assertEqual(spend_kind([".coordinator"]), "infra")
        self.assertEqual(spend_kind(["Core"]), "product")
        self.assertEqual(spend_kind(["Common", "Core"]), "product")
        self.assertIsNone(spend_kind([]))


if __name__ == "__main__":
    unittest.main()
