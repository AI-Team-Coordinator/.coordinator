package model

import "time"

// Member is the current workspace state of one developer.
type Member struct {
	Alias       string
	Name        string
	Role        string
	Focus       []string
	Services    []string
	Status      string
	TaskID      string
	TaskTitle   string
	TaskDoc     string
	TaskSummary string
	Branch      string
	UpdatedAt   time.Time
	Repos       []RepoWork
	GitReport   *GitReport

	CursorUsage     *CursorUsage
	CostUSD         *float64
	BudgetUSD       *float64
	OnDemandUSD     *float64
	CursorModelsPct *float64
	OtherModelsPct  *float64
	SpendKind       string
}
