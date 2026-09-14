package repository

import (
	"testing"
	"time"

	"coordinator/model"
)

func TestLiveSlotSharedIgnoresResearchWithoutWindows(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	slot := model.MemberTask{
		TaskID:    "TASK-A",
		StartedAt: now.Add(-time.Hour),
		ActivityWindows: []model.ActivityWindow{{
			StartedAt: now.Add(-40 * time.Minute),
			EndedAt:   now.Add(-5 * time.Minute),
		}},
	}
	research := &model.Research{Status: "active", StartedAt: now.Add(-2 * time.Hour)}
	if liveSlotShared(slot, []model.MemberTask{slot}, research, now) {
		t.Fatal("research without activity_windows must not mark the slot shared")
	}
}

func TestLiveSlotSharedWhenResearchWindowsOverlap(t *testing.T) {
	now := time.Date(2026, 9, 14, 12, 0, 0, 0, time.UTC)
	slot := model.MemberTask{
		TaskID: "TASK-A",
		ActivityWindows: []model.ActivityWindow{{
			StartedAt: now.Add(-30 * time.Minute),
			EndedAt:   now.Add(-2 * time.Minute),
		}},
	}
	research := &model.Research{
		Status: "active",
		ActivityWindows: []model.ActivityWindow{{
			StartedAt: now.Add(-10 * time.Minute),
			EndedAt:   now.Add(-1 * time.Minute),
		}},
	}
	if !liveSlotShared(slot, []model.MemberTask{slot}, research, now) {
		t.Fatal("overlapping research windows should share the slot")
	}
	if !liveResearchShared(*research, []model.MemberTask{slot}, now) {
		t.Fatal("research should be shared too")
	}
}

func TestApplyLiveSpendSharingNilsDollars(t *testing.T) {
	now := time.Now()
	budget := 1.5
	members := []model.Member{{
		Tasks: []model.MemberTask{
			{
				TaskID:          "A",
				BudgetUSD:       &budget,
				CostUSD:         &budget,
				ActivityWindows: []model.ActivityWindow{{StartedAt: now.Add(-20 * time.Minute), EndedAt: now.Add(-1 * time.Minute)}},
			},
			{
				TaskID:          "B",
				BudgetUSD:       &budget,
				ActivityWindows: []model.ActivityWindow{{StartedAt: now.Add(-15 * time.Minute), EndedAt: now.Add(-2 * time.Minute)}},
			},
		},
	}}
	applyLiveSpendSharing(members)
	if !members[0].Tasks[0].SpendShared || members[0].Tasks[0].BudgetUSD != nil {
		t.Fatalf("slot A %+v", members[0].Tasks[0])
	}
	if !members[0].Tasks[1].SpendShared || members[0].Tasks[1].BudgetUSD != nil {
		t.Fatalf("slot B %+v", members[0].Tasks[1])
	}
}
