package dto

type StatsResponse struct {
	TotalCompleted         int     `json:"total_completed"`
	ActiveNow              int     `json:"active_now"`
	AvgCycleTimeMinutes    float64 `json:"avg_cycle_time_minutes"`
	FeaturesCompleted      int     `json:"features_completed"`
	FixesCompleted         int     `json:"fixes_completed"`
	CompletedToday         int     `json:"completed_today"`
	CompletedThisWeek      int     `json:"completed_this_week"`
	CostUSDToday           float64 `json:"cost_usd_today"`
	CostUSDWeek            float64 `json:"cost_usd_week"`
	CostUSDTotal           float64 `json:"cost_usd_total"`
	CostUSDAvg             float64 `json:"cost_usd_avg"`
	CostTasks              int     `json:"cost_tasks"`
	BudgetUSDToday         float64 `json:"budget_usd_today"`
	BudgetUSDWeek          float64 `json:"budget_usd_week"`
	BudgetUSDTotal         float64 `json:"budget_usd_total"`
	BudgetUSDAvg           float64 `json:"budget_usd_avg"`
	BudgetTasks            int     `json:"budget_tasks"`
	BudgetUSDProductToday  float64 `json:"budget_usd_product_today"`
	BudgetUSDProductWeek   float64 `json:"budget_usd_product_week"`
	BudgetUSDInfraToday    float64 `json:"budget_usd_infra_today"`
	BudgetUSDInfraWeek     float64 `json:"budget_usd_infra_week"`
	OnDemandUSDToday       float64 `json:"ondemand_usd_today"`
	OnDemandUSDWeek        float64 `json:"ondemand_usd_week"`
	BudgetUSDOpen          float64 `json:"budget_usd_open"`
	CostUSDOpen            float64 `json:"cost_usd_open"`
	OnDemandUSDOpen        float64 `json:"ondemand_usd_open"`
	BudgetUSDResearchToday float64 `json:"budget_usd_research_today"`
	BudgetUSDResearchWeek  float64 `json:"budget_usd_research_week"`
	BudgetUSDResearchOpen  float64 `json:"budget_usd_research_open"`
	CostUSDResearchToday   float64 `json:"cost_usd_research_today"`
	CostUSDResearchWeek    float64 `json:"cost_usd_research_week"`
	CostUSDResearchOpen    float64 `json:"cost_usd_research_open"`
	ResearchCompleted      int     `json:"research_completed"`
	ResearchCompletedToday int     `json:"research_completed_today"`
	ResearchCompletedWeek  int     `json:"research_completed_week"`
}

type StatsEnvelope struct {
	Stats StatsResponse `json:"stats"`
}
