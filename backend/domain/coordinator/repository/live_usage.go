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
		if members[i].Status == "in_progress" && members[i].CursorUsage != nil {
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
		if members[i].Status == "in_progress" && members[i].CursorUsage != nil {
			delta := model.ComputeUsageDelta(members[i].CursorUsage, end)
			if delta != nil {
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
		cost := delta.CostUSD
		budget := delta.BudgetUSD
		ondemand := delta.OnDemandUSD
		members[i].Research.CostUSD = &cost
		members[i].Research.BudgetUSD = &budget
		members[i].Research.OnDemandUSD = &ondemand
	}
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
