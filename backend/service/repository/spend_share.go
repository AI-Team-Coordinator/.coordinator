package repository

import (
	"time"

	"coordinator/model"
)

func applyLiveSpendSharing(members []model.Member) {
	now := time.Now()
	for i := range members {
		slots := members[i].Slots()
		for si := range slots {
			if liveSlotShared(slots[si], slots, members[i].Research, now) {
				markSharedMemberTask(&slots[si])
			}
		}
		if len(members[i].Tasks) > 0 {
			members[i].Tasks = slots
			newest := members[i].Tasks[len(members[i].Tasks)-1]
			members[i].CostUSD = newest.CostUSD
			members[i].BudgetUSD = newest.BudgetUSD
			members[i].OnDemandUSD = newest.OnDemandUSD
			members[i].CursorModelsPct = newest.CursorModelsPct
			members[i].OtherModelsPct = newest.OtherModelsPct
			members[i].SpendKind = newest.SpendKind
			members[i].SpendShared = newest.SpendShared
		} else if len(slots) == 1 && slots[0].SpendShared {
			members[i].SpendShared = true
			members[i].BudgetUSD = nil
			members[i].CostUSD = nil
			members[i].OnDemandUSD = nil
		}
		if members[i].Research != nil && liveResearchShared(*members[i].Research, slots, now) {
			markSharedResearch(members[i].Research)
		}
	}
}

func liveSlotShared(slot model.MemberTask, slots []model.MemberTask, research *model.Research, now time.Time) bool {
	wins := slotWindows(slot, now)
	for _, other := range slots {
		if other.TaskID == slot.TaskID {
			continue
		}
		if model.WindowsOverlap(wins, slotWindows(other, now)) {
			return true
		}
	}
	if research == nil || research.Status != "active" {
		return false
	}
	return model.WindowsOverlap(wins, researchWindows(*research, now))
}

func liveResearchShared(research model.Research, slots []model.MemberTask, now time.Time) bool {
	wins := researchWindows(research, now)
	if len(wins) == 0 {
		return false
	}
	for _, slot := range slots {
		if model.WindowsOverlap(wins, slotWindows(slot, now)) {
			return true
		}
	}
	return false
}

func slotWindows(slot model.MemberTask, now time.Time) []model.ActivityWindow {
	return model.SpendWindows(slot.ActivityWindows, slot.StartedAt, slot.LastActivityAt, slot.UpdatedAt, now)
}

func researchWindows(research model.Research, now time.Time) []model.ActivityWindow {
	return model.ResearchSpendWindows(research.ActivityWindows)
}

func markSharedMemberTask(slot *model.MemberTask) {
	slot.SpendShared = true
	slot.BudgetUSD = nil
	slot.CostUSD = nil
	slot.OnDemandUSD = nil
}

func markSharedResearch(research *model.Research) {
	research.SpendShared = true
	research.BudgetUSD = nil
	research.CostUSD = nil
	research.OnDemandUSD = nil
}
