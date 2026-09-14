package model

// UsageSample is one local Cursor quota tick (no secrets).
type UsageSample struct {
	TS                int64    `json:"ts"`
	Alias             string   `json:"alias,omitempty"`
	SessionID         string   `json:"session_id,omitempty"`
	TaskID            string   `json:"task_id,omitempty"`
	Kind              string   `json:"kind,omitempty"`
	Title             string   `json:"title,omitempty"`
	Plan              string   `json:"plan,omitempty"`
	BillingCycleStart string   `json:"billing_cycle_start,omitempty"`
	OnDemandCents     float64  `json:"ondemand_cents,omitempty"`
	CursorModelsPct   float64  `json:"cursor_models_pct,omitempty"`
	OtherModelsPct    float64  `json:"other_models_pct,omitempty"`
	PlanPriceUSD      *float64 `json:"plan_price_usd,omitempty"`
}

// HourSpend is the account-meter delta for one local hour (or day, Hour = midnight) plus who was in frame.
type HourSpend struct {
	Hour            int64
	Alias           string
	Aliases         []string
	CursorModelsPct float64
	OtherModelsPct  float64
	BudgetUSD       *float64
	OnDemandUSD     float64
	Participants    []HourParticipant
}

// HourParticipant is a task or research that pinged during the hour.
type HourParticipant struct {
	Kind   string
	ID     string
	Title  string
	Alias  string
	Shared bool
}

func CursorUsageFromSample(s UsageSample) *CursorUsage {
	return &CursorUsage{
		Plan:              s.Plan,
		PlanPriceUSD:      s.PlanPriceUSD,
		BillingCycleStart: s.BillingCycleStart,
		OnDemandCents:     s.OnDemandCents,
		CursorModelsPct:   s.CursorModelsPct,
		OtherModelsPct:    s.OtherModelsPct,
	}
}
