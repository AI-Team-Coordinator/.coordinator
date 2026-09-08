package model

// Stats holds computed cycle-time and velocity metrics.
type Stats struct {
	TotalCompleted        int
	ActiveNow             int
	AvgCycleTimeMinutes   float64
	FeaturesCompleted     int
	FixesCompleted        int
	CompletedToday        int
	CompletedThisWeek     int
	CostUSDToday          float64
	CostUSDWeek           float64
	CostUSDTotal          float64
	CostUSDAvg            float64
	CostTasks             int
	BudgetUSDToday        float64
	BudgetUSDWeek         float64
	BudgetUSDTotal        float64
	BudgetUSDAvg          float64
	BudgetTasks           int
	BudgetUSDProductToday float64
	BudgetUSDProductWeek  float64
	BudgetUSDInfraToday   float64
	BudgetUSDInfraWeek    float64
	OnDemandUSDToday      float64
	OnDemandUSDWeek       float64
	BudgetUSDOpen         float64
	CostUSDOpen           float64
	OnDemandUSDOpen       float64
}
