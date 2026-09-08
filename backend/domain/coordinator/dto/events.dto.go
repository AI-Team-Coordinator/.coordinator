package dto

type EventResponse struct {
	Timestamp       int64    `json:"timestamp"`
	Event           string   `json:"event"`
	TaskID          string   `json:"task_id"`
	Branch          string   `json:"branch,omitempty"`
	Alias           string   `json:"alias,omitempty"`
	Repo            string   `json:"repo,omitempty"`
	Service         string   `json:"service,omitempty"`
	Status          string   `json:"status,omitempty"`
	CostUSD         *float64 `json:"cost_usd,omitempty"`
	BudgetUSD       *float64 `json:"budget_usd,omitempty"`
	OnDemandUSD     *float64 `json:"ondemand_usd,omitempty"`
	CursorModelsPct *float64 `json:"cursor_models_pct,omitempty"`
	OtherModelsPct  *float64 `json:"other_models_pct,omitempty"`
	UsagePlan       string   `json:"usage_plan,omitempty"`
	PlanPriceUSD    *float64 `json:"plan_price_usd,omitempty"`
	SpendKind       string   `json:"spend_kind,omitempty"`
	Summary         string   `json:"summary,omitempty"`
	Findings        string   `json:"findings,omitempty"`
}

type EventsResponse struct {
	Events []EventResponse `json:"events"`
	Total  int             `json:"total"`
	Offset int             `json:"offset"`
	Limit  int             `json:"limit"`
}

type EventsQuery struct {
	Days    int
	Limit   int
	Offset  int
	Alias   string
	Event   string
	Service string
}
