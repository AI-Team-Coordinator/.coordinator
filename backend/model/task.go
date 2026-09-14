package model

// Task is one coordinator task folded from start/complete events and live snapshots.
type Task struct {
	TaskID          string
	Title           string
	Status          string
	Kind            string
	Alias           string
	Branch          string
	Services        []string
	StartedAt       int64
	CompletedAt     int64
	DurationSeconds int64
	ClockPaused     bool
	ActiveSeconds   *int64
	CostUSD         *float64
	BudgetUSD       *float64
	OnDemandUSD     *float64
	CursorModelsPct *float64
	OtherModelsPct  *float64
	SpendKind       string
	SpendShared     bool
}

// TaskQuery filters the aggregated task list. Limit 0 means no page cap.
type TaskQuery struct {
	Status string
	Alias  string
	Kind   string
	Limit  int
	Offset int
}
