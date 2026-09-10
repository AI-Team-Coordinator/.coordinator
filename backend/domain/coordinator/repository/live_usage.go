package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"coordinator/model"
)

const usageCacheTTL = 30 * time.Second

func (r *FileRepository) attachLiveUsage(members []model.Member) {
	alias, err := r.CurrentAuthor(context.Background())
	if err != nil || alias == "" {
		return
	}
	need := false
	for i := range members {
		if members[i].Alias != alias {
			continue
		}
		if len(members[i].Slots()) > 0 {
			need = true
		}
		if members[i].Research != nil {
			need = true
		}
	}
	if !need {
		return
	}
	end := r.cachedUsageSnapshot()
	if end == nil {
		return
	}
	for i := range members {
		if members[i].Alias != alias {
			continue
		}
		slots := members[i].Slots()
		for si := range slots {
			if slots[si].CursorUsage == nil {
				continue
			}
			delta := model.ComputeUsageDelta(slots[si].CursorUsage, end)
			if delta == nil {
				continue
			}
			applyUsageDelta(&slots[si].CostUSD, &slots[si].BudgetUSD, &slots[si].OnDemandUSD, &slots[si].CursorModelsPct, &slots[si].OtherModelsPct, delta)
			slots[si].SpendKind = model.SpendKind(slots[si].Services)
		}
		if len(slots) > 0 {
			members[i].Tasks = slots
			last := slots[len(slots)-1]
			members[i].CostUSD = last.CostUSD
			members[i].BudgetUSD = last.BudgetUSD
			members[i].OnDemandUSD = last.OnDemandUSD
			members[i].CursorModelsPct = last.CursorModelsPct
			members[i].OtherModelsPct = last.OtherModelsPct
			members[i].SpendKind = last.SpendKind
			if slots[len(slots)-1].CursorUsage != nil {
				members[i].CursorUsage = slots[len(slots)-1].CursorUsage
			}
		}
		if members[i].Research == nil {
			continue
		}
		if members[i].Research.CursorUsage == nil {
			if r.persistResearchCursorUsage(alias, end) {
				start := *end
				members[i].Research.CursorUsage = &start
			}
		}
		if members[i].Research.CursorUsage == nil {
			continue
		}
		delta := model.ComputeUsageDelta(members[i].Research.CursorUsage, end)
		if delta == nil {
			continue
		}
		applyUsageDelta(&members[i].Research.CostUSD, &members[i].Research.BudgetUSD, &members[i].Research.OnDemandUSD, &members[i].Research.CursorModelsPct, &members[i].Research.OtherModelsPct, delta)
	}
}

func applyUsageDelta(cost, budget, ondemand, cursorPct, otherPct **float64, delta *model.UsageDelta) {
	if delta == nil {
		return
	}
	c := delta.CostUSD
	o := delta.OnDemandUSD
	cp := delta.CursorModelsPct
	op := delta.OtherModelsPct
	*cost = &c
	*ondemand = &o
	*cursorPct = &cp
	*otherPct = &op
	if delta.HasPlanPrice() {
		b := delta.BudgetUSD
		*budget = &b
		return
	}
	*budget = nil
}

func (r *FileRepository) persistResearchCursorUsage(alias string, usage *model.CursorUsage) bool {
	if usage == nil || alias == "" {
		return false
	}
	path := filepath.Join(r.progressDir(), ".current_task_"+alias)
	raw, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	var snap map[string]any
	if err := json.Unmarshal(raw, &snap); err != nil {
		return false
	}
	research, _ := snap["research"].(map[string]any)
	if research == nil || research["status"] != "active" {
		return false
	}
	if research["cursor_usage"] != nil {
		return true
	}
	encoded, err := json.Marshal(usage)
	if err != nil {
		return false
	}
	var usageVal any
	if err := json.Unmarshal(encoded, &usageVal); err != nil {
		return false
	}
	research["cursor_usage"] = usageVal
	snap["research"] = research
	out, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return false
	}
	out = append(out, '\n')
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return false
	}
	return true
}

func (r *FileRepository) CursorUsageSnapshot() *model.CursorUsage {
	return r.cachedUsageSnapshot()
}

func (r *FileRepository) cachedUsageSnapshot() *model.CursorUsage {
	r.usageMu.Lock()
	defer r.usageMu.Unlock()
	if r.usageSnap != nil && time.Since(r.usageAt) < usageCacheTTL {
		return r.usageSnap
	}
	snap := r.fetchUsageSnapshot()
	r.usageAt = time.Now()
	if snap != nil {
		r.usageSnap = snap
	}
	return r.usageSnap
}

func (r *FileRepository) fetchUsageSnapshot() *model.CursorUsage {
	script := r.appScript("cursor_usage.py")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "python3", script, "snapshot")
	out, err := cmd.Output()
	if err != nil {
		return nil
	}
	out = bytes.TrimSpace(out)
	if len(out) == 0 {
		return nil
	}
	var snap model.CursorUsage
	if err := json.Unmarshal(out, &snap); err != nil {
		return nil
	}
	return &snap
}
