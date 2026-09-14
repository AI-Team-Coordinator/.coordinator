package dto

type StatsResponse struct {
	TotalCompleted               int                 `json:"total_completed"`
	ActiveNow                    int                 `json:"active_now"`
	AvgCycleTimeMinutes          float64             `json:"avg_cycle_time_minutes"`
	FeaturesCompleted            int                 `json:"features_completed"`
	FixesCompleted               int                 `json:"fixes_completed"`
	CompletedToday               int                 `json:"completed_today"`
	CompletedThisWeek            int                 `json:"completed_this_week"`
	CostUSDToday                 float64             `json:"cost_usd_today"`
	CostUSDWeek                  float64             `json:"cost_usd_week"`
	CostUSDCycle                 float64             `json:"cost_usd_cycle"`
	CostUSDTotal                 float64             `json:"cost_usd_total"`
	CostUSDAvg                   float64             `json:"cost_usd_avg"`
	CostTasks                    int                 `json:"cost_tasks"`
	BudgetUSDToday               float64             `json:"budget_usd_today"`
	BudgetUSDWeek                float64             `json:"budget_usd_week"`
	BudgetUSDCycle               float64             `json:"budget_usd_cycle"`
	BudgetUSDTotal               float64             `json:"budget_usd_total"`
	BudgetUSDAvg                 float64             `json:"budget_usd_avg"`
	BudgetTasks                  int                 `json:"budget_tasks"`
	BudgetUSDProductToday        float64             `json:"budget_usd_product_today"`
	BudgetUSDProductWeek         float64             `json:"budget_usd_product_week"`
	BudgetUSDProductCycle        float64             `json:"budget_usd_product_cycle"`
	BudgetUSDInfraToday          float64             `json:"budget_usd_infra_today"`
	BudgetUSDInfraWeek           float64             `json:"budget_usd_infra_week"`
	BudgetUSDInfraCycle          float64             `json:"budget_usd_infra_cycle"`
	OnDemandUSDToday             float64             `json:"ondemand_usd_today"`
	OnDemandUSDWeek              float64             `json:"ondemand_usd_week"`
	OnDemandUSDCycle             float64             `json:"ondemand_usd_cycle"`
	BudgetUSDOpen                float64             `json:"budget_usd_open"`
	CostUSDOpen                  float64             `json:"cost_usd_open"`
	OnDemandUSDOpen              float64             `json:"ondemand_usd_open"`
	BudgetUSDResearchToday       float64             `json:"budget_usd_research_today"`
	BudgetUSDResearchWeek        float64             `json:"budget_usd_research_week"`
	BudgetUSDResearchCycle       float64             `json:"budget_usd_research_cycle"`
	BudgetUSDResearchOpen        float64             `json:"budget_usd_research_open"`
	CostUSDResearchToday         float64             `json:"cost_usd_research_today"`
	CostUSDResearchWeek          float64             `json:"cost_usd_research_week"`
	CostUSDResearchCycle         float64             `json:"cost_usd_research_cycle"`
	CostUSDResearchOpen          float64             `json:"cost_usd_research_open"`
	ResearchCompleted            int                 `json:"research_completed"`
	ResearchCompletedToday       int                 `json:"research_completed_today"`
	ResearchCompletedWeek        int                 `json:"research_completed_week"`
	BillingCycleStart            string              `json:"billing_cycle_start,omitempty"`
	PlanPriceUSD                 *float64            `json:"plan_price_usd,omitempty"`
	CursorModelsPctToday         float64             `json:"cursor_models_pct_today"`
	OtherModelsPctToday          float64             `json:"other_models_pct_today"`
	CursorModelsPctCycle         float64             `json:"cursor_models_pct_cycle"`
	OtherModelsPctCycle          float64             `json:"other_models_pct_cycle"`
	CursorModelsPctOpen          float64             `json:"cursor_models_pct_open"`
	OtherModelsPctOpen           float64             `json:"other_models_pct_open"`
	CursorModelsPctProductCycle  float64             `json:"cursor_models_pct_product_cycle"`
	OtherModelsPctProductCycle   float64             `json:"other_models_pct_product_cycle"`
	CursorModelsPctInfraCycle    float64             `json:"cursor_models_pct_infra_cycle"`
	OtherModelsPctInfraCycle     float64             `json:"other_models_pct_infra_cycle"`
	CursorModelsPctResearchToday float64             `json:"cursor_models_pct_research_today"`
	OtherModelsPctResearchToday  float64             `json:"other_models_pct_research_today"`
	CursorModelsPctResearchCycle float64             `json:"cursor_models_pct_research_cycle"`
	OtherModelsPctResearchCycle  float64             `json:"other_models_pct_research_cycle"`
	CursorModelsPctResearchOpen  float64             `json:"cursor_models_pct_research_open"`
	OtherModelsPctResearchOpen   float64             `json:"other_models_pct_research_open"`
	Hours                        []HourSpendResponse `json:"hours,omitempty"`
	Days                         []HourSpendResponse `json:"days,omitempty"`
}

type HourSpendResponse struct {
	Hour            int64                     `json:"hour"`
	Alias           string                    `json:"alias,omitempty"`
	Aliases         []string                  `json:"aliases,omitempty"`
	CursorModelsPct float64                   `json:"cursor_models_pct"`
	OtherModelsPct  float64                   `json:"other_models_pct"`
	BudgetUSD       *float64                  `json:"budget_usd,omitempty"`
	OnDemandUSD     float64                   `json:"ondemand_usd"`
	Participants    []HourParticipantResponse `json:"participants"`
}

type HourParticipantResponse struct {
	Kind   string `json:"kind"`
	ID     string `json:"id,omitempty"`
	Title  string `json:"title,omitempty"`
	Alias  string `json:"alias,omitempty"`
	Shared bool   `json:"shared,omitempty"`
}

type StatsEnvelope struct {
	Stats StatsResponse `json:"stats"`
}
