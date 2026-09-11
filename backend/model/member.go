package model

import "time"

// Member is the current workspace state of one developer.
type Member struct {
	Alias       string
	Name        string
	Role        string
	Access      string
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

	Research *Research
	Tasks    []MemberTask
}

// MemberTask is one in-progress slot on a developer's snapshot.
type MemberTask struct {
	TaskID          string
	Title           string
	Doc             string
	Summary         string
	Branch          string
	Services        []string
	StartedAt       time.Time
	UpdatedAt       time.Time
	DurationSeconds int64
	Repos           []RepoWork
	CursorUsage     *CursorUsage
	CostUSD         *float64
	BudgetUSD       *float64
	OnDemandUSD     *float64
	CursorModelsPct *float64
	OtherModelsPct  *float64
	SpendKind       string
	SessionIDs      []string
	Chats           []ChatTab
}

// ChatTab is a Cursor Composer chat bound to a task or research.
type ChatTab struct {
	Title     string
	SessionID string
}

// Slots returns open tasks. Legacy snapshots with only root fields yield one slot.
func (m Member) Slots() []MemberTask {
	if len(m.Tasks) > 0 {
		return m.Tasks
	}
	if m.Status == "in_progress" && m.TaskID != "" {
		return []MemberTask{{
			TaskID:          m.TaskID,
			Title:           m.TaskTitle,
			Doc:             m.TaskDoc,
			Summary:         m.TaskSummary,
			Branch:          m.Branch,
			Services:        m.Services,
			UpdatedAt:       m.UpdatedAt,
			StartedAt:       m.UpdatedAt,
			Repos:           m.Repos,
			CursorUsage:     m.CursorUsage,
			CostUSD:         m.CostUSD,
			BudgetUSD:       m.BudgetUSD,
			OnDemandUSD:     m.OnDemandUSD,
			CursorModelsPct: m.CursorModelsPct,
			OtherModelsPct:  m.OtherModelsPct,
			SpendKind:       m.SpendKind,
		}}
	}
	return nil
}

// Research is an off-task Cursor chat running in parallel with (or without) a task.
type Research struct {
	Status          string
	Summary         string
	StartedAt       time.Time
	SessionID       string
	Chat            *ChatTab
	CursorUsage     *CursorUsage
	CostUSD         *float64
	BudgetUSD       *float64
	OnDemandUSD     *float64
	CursorModelsPct *float64
	OtherModelsPct  *float64
	DurationSeconds int64
}
