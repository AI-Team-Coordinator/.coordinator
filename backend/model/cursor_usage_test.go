package model

import (
	"math"
	"testing"
)

func f64(v float64) *float64 { return &v }

func TestComputeUsageDeltaPlanShare(t *testing.T) {
	start := &CursorUsage{Plan: "pro_plus", PlanPriceUSD: f64(60), BillingCycleStart: "c", CursorModelsPct: 40, OtherModelsPct: 10}
	end := &CursorUsage{Plan: "pro_plus", PlanPriceUSD: f64(60), BillingCycleStart: "c", CursorModelsPct: 50, OtherModelsPct: 15}
	delta := ComputeUsageDelta(start, end)
	if delta == nil || delta.BudgetUSD != 4.5 || delta.CostUSD != 0 {
		t.Fatalf("%+v", delta)
	}
}

func TestComputeUsageDeltaOverflow(t *testing.T) {
	start := &CursorUsage{Plan: "pro_plus", BillingCycleStart: "c", IncludedCents: 7000, OtherModelsPct: 98, CursorModelsPct: 100}
	end := &CursorUsage{Plan: "pro_plus", BillingCycleStart: "c", IncludedCents: 7000, OnDemandCents: 250, OtherModelsPct: 110, CursorModelsPct: 120}
	delta := ComputeUsageDelta(start, end)
	if delta == nil {
		t.Fatal("nil")
	}
	if math.Abs(delta.BudgetUSD-0.6) > 1e-9 {
		t.Fatalf("budget=%v", delta.BudgetUSD)
	}
	if delta.OnDemandUSD != 2.5 || delta.CostUSD != 2.5 {
		t.Fatalf("od=%v cost=%v", delta.OnDemandUSD, delta.CostUSD)
	}
}

func TestComputeUsageDeltaTeamHasPctsNoPrice(t *testing.T) {
	start := &CursorUsage{Plan: "team", BillingCycleStart: "c", CursorModelsPct: 10, OtherModelsPct: 5}
	end := &CursorUsage{Plan: "team", BillingCycleStart: "c", CursorModelsPct: 14, OtherModelsPct: 15}
	delta := ComputeUsageDelta(start, end)
	if delta == nil {
		t.Fatal("nil")
	}
	if delta.BudgetUSD != 0 || delta.HasPlanPrice() {
		t.Fatalf("budget=%v price=%v", delta.BudgetUSD, delta.PlanPriceUSD)
	}
	if delta.CursorModelsPct != 4 || delta.OtherModelsPct != 10 {
		t.Fatalf("pcts cursor=%v other=%v", delta.CursorModelsPct, delta.OtherModelsPct)
	}
	if delta.UsagePlan != "team" {
		t.Fatalf("plan=%s", delta.UsagePlan)
	}
}

func TestSpendKind(t *testing.T) {
	if SpendKind([]string{"Common"}) != "infra" {
		t.Fatal("common")
	}
	if SpendKind([]string{".cursor", "common"}) != "infra" {
		t.Fatal("both")
	}
	if SpendKind([]string{".coordinator"}) != "infra" {
		t.Fatal("coordinator")
	}
	if SpendKind([]string{"Core"}) != "product" {
		t.Fatal("core")
	}
	if SpendKind([]string{"Common", "Core"}) != "product" {
		t.Fatal("mixed")
	}
	if SpendKind(nil) != "" {
		t.Fatal("empty")
	}
}
