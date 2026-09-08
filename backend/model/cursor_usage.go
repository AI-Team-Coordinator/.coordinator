package model

import (
	"math"
	"strings"
)

// CursorUsage is a period snapshot from the local Cursor session (no secrets).
type CursorUsage struct {
	Plan              string   `json:"plan"`
	PlanPriceUSD      *float64 `json:"plan_price_usd,omitempty"`
	BillingCycleStart string   `json:"billing_cycle_start"`
	IncludedCents     float64  `json:"included_cents"`
	OnDemandCents     float64  `json:"ondemand_cents"`
	CursorModelsPct   float64  `json:"cursor_models_pct"`
	OtherModelsPct    float64  `json:"other_models_pct"`
}

var planPriceUSD = map[string]float64{
	"pro":      20,
	"pro_plus": 60,
	"pro+":     60,
	"ultra":    200,
}

func PlanPriceUSD(plan string) (float64, bool) {
	key := strings.ToLower(strings.TrimSpace(plan))
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	price, ok := planPriceUSD[key]
	return price, ok
}

func IncludedPoolDelta(startPct, endPct float64) float64 {
	startIn := math.Min(math.Max(startPct, 0), 100)
	endIn := math.Min(math.Max(endPct, 0), 100)
	return math.Max(0, endIn-startIn)
}

func SpendKind(services []string) string {
	n := 0
	allInfra := true
	for _, raw := range services {
		label := strings.TrimSpace(raw)
		if label == "" {
			continue
		}
		n++
		if !isInfraService(label) {
			allInfra = false
		}
	}
	if n == 0 {
		return ""
	}
	if allInfra {
		return "infra"
	}
	return "product"
}

func isInfraService(label string) bool {
	key := strings.ToLower(strings.TrimSpace(label))
	key = strings.TrimPrefix(key, ".")
	key = strings.ReplaceAll(key, "-", "_")
	key = strings.ReplaceAll(key, " ", "_")
	return key == "common" || key == "cursor"
}

// ComputeUsageDelta is the same formula as coordinator/utils/cursor_usage.py.
func ComputeUsageDelta(start, end *CursorUsage) *UsageDelta {
	if start == nil || end == nil {
		return nil
	}
	if start.BillingCycleStart != "" && end.BillingCycleStart != "" && start.BillingCycleStart != end.BillingCycleStart {
		return nil
	}

	ondemandDelta := math.Max(0, end.OnDemandCents-start.OnDemandCents)
	cursorDelta := math.Max(0, end.CursorModelsPct-start.CursorModelsPct)
	otherDelta := math.Max(0, end.OtherModelsPct-start.OtherModelsPct)

	plan := strings.TrimSpace(end.Plan)
	if plan == "" {
		plan = strings.TrimSpace(start.Plan)
	}
	price, ok := priceFrom(end)
	if !ok {
		price, ok = priceFrom(start)
	}
	if !ok {
		price, ok = PlanPriceUSD(plan)
	}

	budget := 0.0
	if ok && price > 0 {
		budget = price * (IncludedPoolDelta(start.CursorModelsPct, end.CursorModelsPct) + IncludedPoolDelta(start.OtherModelsPct, end.OtherModelsPct)) / 200
	}

	out := &UsageDelta{
		CostUSD:         round4(ondemandDelta / 100),
		BudgetUSD:       round4(budget),
		OnDemandUSD:     round4(ondemandDelta / 100),
		CursorModelsPct: round4(cursorDelta),
		OtherModelsPct:  round4(otherDelta),
		UsagePlan:       plan,
	}
	if ok {
		p := price
		out.PlanPriceUSD = &p
	}
	return out
}

func priceFrom(snap *CursorUsage) (float64, bool) {
	if snap == nil || snap.PlanPriceUSD == nil {
		return 0, false
	}
	return *snap.PlanPriceUSD, true
}

func round4(v float64) float64 {
	return math.Round(v*10000) / 10000
}

// UsageDelta is live or completed spend for one task window.
type UsageDelta struct {
	CostUSD         float64
	BudgetUSD       float64
	OnDemandUSD     float64
	CursorModelsPct float64
	OtherModelsPct  float64
	UsagePlan       string
	PlanPriceUSD    *float64
}
