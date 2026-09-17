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
	SpendShared     bool
	Research        *Research
	Tasks           []MemberTask
}

// MemberTask is one open slot on a developer's snapshot (in progress or parked).
type MemberTask struct {
	TaskID          string
	Status          string
	Title           string
	Doc             string
	Summary         string
	Branch          string
	Services        []string
	StartedAt       time.Time
	UpdatedAt       time.Time
	LastActivityAt  time.Time
	ParkedAt        time.Time
	ActivityWindows []ActivityWindow
	DurationSeconds int64
	ClockPaused     bool
	Repos           []RepoWork
	CursorUsage     *CursorUsage
	CostUSD         *float64
	BudgetUSD       *float64
	OnDemandUSD     *float64
	CursorModelsPct *float64
	OtherModelsPct  *float64
	SpendKind       string
	SpendShared     bool
	SessionIDs      []string
	Chats           []ChatTab
}

// IsParked is a live task whose branch is set aside so another slot can take the repo.
func (t MemberTask) IsParked() bool {
	return t.Status == "parked"
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
	if (m.Status == "in_progress" || m.Status == "parked") && m.TaskID != "" {
		return []MemberTask{{
			TaskID:          m.TaskID,
			Status:          m.Status,
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
			SpendShared:     m.SpendShared,
		}}
	}
	return nil
}

// ActiveSlots are in-progress claims. Parked slots stay on Pulse but do not exclusive-claim a repo.
func (m Member) ActiveSlots() []MemberTask {
	out := make([]MemberTask, 0)
	for _, slot := range m.Slots() {
		if slot.IsParked() {
			continue
		}
		out = append(out, slot)
	}
	return out
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
	ClockPaused     bool
	SpendShared     bool
	ActivityWindows []ActivityWindow
}
