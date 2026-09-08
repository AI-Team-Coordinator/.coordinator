package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"os/exec"
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
		if members[i].Alias == alias && members[i].Status == "in_progress" && members[i].CursorUsage != nil {
			need = true
			break
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
		if members[i].Alias != alias || members[i].Status != "in_progress" || members[i].CursorUsage == nil {
			continue
		}
		delta := model.ComputeUsageDelta(members[i].CursorUsage, end)
		if delta == nil {
			continue
		}
		cost := delta.CostUSD
		budget := delta.BudgetUSD
		ondemand := delta.OnDemandUSD
		cursorPct := delta.CursorModelsPct
		otherPct := delta.OtherModelsPct
		members[i].CostUSD = &cost
		members[i].BudgetUSD = &budget
		members[i].OnDemandUSD = &ondemand
		members[i].CursorModelsPct = &cursorPct
		members[i].OtherModelsPct = &otherPct
		members[i].SpendKind = model.SpendKind(members[i].Services)
	}
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
