package coordinator

import (
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

func computeStats(events []model.Event, members []model.Member) model.Stats {
	stats := model.Stats{}

	for _, m := range members {
		if m.Status == "in_progress" {
			stats.ActiveNow++
		}
	}

	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})

	now := time.Now()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	weekday := int64(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := startOfToday - (weekday-1)*86400

	starts := make(map[string]int64)
	var totalCycleTimeSeconds int64
	var completedCount int

	for _, ev := range events {
		if ev.Event == "task_started" {
			starts[ev.TaskID] = ev.Timestamp
			continue
		}
		if ev.Event != "task_completed" {
			continue
		}

		stats.TotalCompleted++
		completedCount++

		isFix := strings.HasPrefix(ev.TaskID, "FIX-") || strings.HasPrefix(ev.Branch, "fix/")
		if isFix {
			stats.FixesCompleted++
		} else {
			stats.FeaturesCompleted++
		}

		if startTime, ok := starts[ev.TaskID]; ok && ev.Timestamp >= startTime {
			totalCycleTimeSeconds += ev.Timestamp - startTime
		}
		if ev.Timestamp >= startOfToday {
			stats.CompletedToday++
		}
		if ev.Timestamp >= startOfWeek {
			stats.CompletedThisWeek++
		}
		if ev.CostUSD != nil {
			cost := *ev.CostUSD
			stats.CostUSDTotal += cost
			stats.CostTasks++
			if ev.Timestamp >= startOfToday {
				stats.CostUSDToday += cost
			}
			if ev.Timestamp >= startOfWeek {
				stats.CostUSDWeek += cost
			}
		}
		if ev.BudgetUSD != nil {
			budget := *ev.BudgetUSD
			stats.BudgetUSDTotal += budget
			stats.BudgetTasks++
			if ev.Timestamp >= startOfToday {
				stats.BudgetUSDToday += budget
			}
			if ev.Timestamp >= startOfWeek {
				stats.BudgetUSDWeek += budget
			}
			if ev.SpendKind == "infra" {
				if ev.Timestamp >= startOfToday {
					stats.BudgetUSDInfraToday += budget
				}
				if ev.Timestamp >= startOfWeek {
					stats.BudgetUSDInfraWeek += budget
				}
			} else {
				if ev.Timestamp >= startOfToday {
					stats.BudgetUSDProductToday += budget
				}
				if ev.Timestamp >= startOfWeek {
					stats.BudgetUSDProductWeek += budget
				}
			}
		}
		if ev.OnDemandUSD != nil {
			od := *ev.OnDemandUSD
			if ev.Timestamp >= startOfToday {
				stats.OnDemandUSDToday += od
			}
			if ev.Timestamp >= startOfWeek {
				stats.OnDemandUSDWeek += od
			}
		}
	}

	if completedCount > 0 && totalCycleTimeSeconds > 0 {
		stats.AvgCycleTimeMinutes = float64(totalCycleTimeSeconds) / float64(completedCount) / 60.0
	}
	if stats.CostTasks > 0 {
		stats.CostUSDAvg = stats.CostUSDTotal / float64(stats.CostTasks)
	}
	if stats.BudgetTasks > 0 {
		stats.BudgetUSDAvg = stats.BudgetUSDTotal / float64(stats.BudgetTasks)
	}

	addOpenTaskSpend(&stats, members)
	return stats
}

func taskKind(taskID, branch string) string {
	if strings.HasPrefix(taskID, "FIX-") || strings.HasPrefix(branch, "fix/") {
		return "fix"
	}
	return "feature"
}

func computeTasks(events []model.Event, members []model.Member, now time.Time) []model.Task {
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp < events[j].Timestamp
	})

	acc := make(map[string]*model.Task)
	ensure := func(id string) *model.Task {
		if t, ok := acc[id]; ok {
			return t
		}
		t := &model.Task{TaskID: id, Status: "in_progress"}
		acc[id] = t
		return t
	}

	for _, ev := range events {
		if ev.TaskID == "" {
			continue
		}
		if ev.Event != "task_started" && ev.Event != "task_completed" {
			continue
		}
		t := ensure(ev.TaskID)
		if ev.Alias != "" {
			t.Alias = ev.Alias
		}
		if ev.Branch != "" {
			t.Branch = ev.Branch
		}
		t.Kind = taskKind(ev.TaskID, t.Branch)
		switch ev.Event {
		case "task_started":
			if t.CompletedAt > 0 && ev.Timestamp >= t.CompletedAt {
				t.CompletedAt = 0
				t.CostUSD = nil
				t.BudgetUSD = nil
				t.OnDemandUSD = nil
				t.SpendKind = ""
				t.Services = nil
			}
			if t.StartedAt == 0 || ev.Timestamp >= t.StartedAt {
				t.StartedAt = ev.Timestamp
			}
			if t.CompletedAt == 0 {
				t.Status = "in_progress"
			}
		case "task_completed":
			t.CompletedAt = ev.Timestamp
			t.Status = "completed"
			t.CostUSD = ev.CostUSD
			t.BudgetUSD = ev.BudgetUSD
			t.OnDemandUSD = ev.OnDemandUSD
			t.SpendKind = ev.SpendKind
			if t.StartedAt == 0 {
				t.StartedAt = ev.Timestamp
			}
		}
	}

	for _, m := range members {
		if m.Status != "in_progress" || m.TaskID == "" {
			continue
		}
		t := ensure(m.TaskID)
		t.Status = "in_progress"
		t.CompletedAt = 0
		t.Alias = m.Alias
		if m.Branch != "" {
			t.Branch = m.Branch
		}
		if m.TaskTitle != "" {
			t.Title = m.TaskTitle
		}
		if len(m.Services) > 0 {
			t.Services = m.Services
		}
		t.SpendKind = m.SpendKind
		t.CostUSD = m.CostUSD
		t.BudgetUSD = m.BudgetUSD
		t.OnDemandUSD = m.OnDemandUSD
		t.Kind = taskKind(m.TaskID, t.Branch)
		if t.StartedAt == 0 && !m.UpdatedAt.IsZero() {
			t.StartedAt = m.UpdatedAt.Unix()
		}
	}

	nowUnix := now.Unix()
	out := make([]model.Task, 0, len(acc))
	for _, t := range acc {
		if t.Kind == "" {
			t.Kind = taskKind(t.TaskID, t.Branch)
		}
		end := nowUnix
		if t.Status == "completed" && t.CompletedAt > 0 {
			end = t.CompletedAt
		}
		if t.StartedAt > 0 && end >= t.StartedAt {
			t.DurationSeconds = end - t.StartedAt
		}
		out = append(out, *t)
	}

	sort.Slice(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.Status != b.Status {
			return a.Status == "in_progress"
		}
		keyA, keyB := a.CompletedAt, b.CompletedAt
		if a.Status == "in_progress" {
			keyA, keyB = a.StartedAt, b.StartedAt
		}
		if keyA != keyB {
			return keyA > keyB
		}
		return a.TaskID > b.TaskID
	})
	return out
}

func filterTasks(tasks []model.Task, q model.TaskQuery) []model.Task {
	out := make([]model.Task, 0, len(tasks))
	for _, t := range tasks {
		if q.Status != "" && q.Status != "all" && t.Status != q.Status {
			continue
		}
		if q.Alias != "" && t.Alias != q.Alias {
			continue
		}
		if q.Kind != "" && t.Kind != q.Kind {
			continue
		}
		out = append(out, t)
	}
	return out
}

func pageTasks(tasks []model.Task, limit, offset int) []model.Task {
	if offset < 0 {
		offset = 0
	}
	if offset >= len(tasks) {
		return []model.Task{}
	}
	tasks = tasks[offset:]
	if limit > 0 && len(tasks) > limit {
		return tasks[:limit]
	}
	return tasks
}

func addOpenTaskSpend(stats *model.Stats, members []model.Member) {
	for _, m := range members {
		if m.Status != "in_progress" {
			continue
		}
		if m.CostUSD != nil {
			cost := *m.CostUSD
			stats.CostUSDTotal += cost
			stats.CostUSDToday += cost
			stats.CostUSDWeek += cost
			stats.CostUSDOpen += cost
			stats.CostTasks++
		}
		if m.BudgetUSD != nil {
			budget := *m.BudgetUSD
			stats.BudgetUSDTotal += budget
			stats.BudgetUSDToday += budget
			stats.BudgetUSDWeek += budget
			stats.BudgetUSDOpen += budget
			stats.BudgetTasks++
			if m.SpendKind == "infra" {
				stats.BudgetUSDInfraToday += budget
				stats.BudgetUSDInfraWeek += budget
			} else {
				stats.BudgetUSDProductToday += budget
				stats.BudgetUSDProductWeek += budget
			}
		}
		if m.OnDemandUSD != nil {
			od := *m.OnDemandUSD
			stats.OnDemandUSDToday += od
			stats.OnDemandUSDWeek += od
			stats.OnDemandUSDOpen += od
		}
	}
	if stats.CostTasks > 0 {
		stats.CostUSDAvg = stats.CostUSDTotal / float64(stats.CostTasks)
	}
	if stats.BudgetTasks > 0 {
		stats.BudgetUSDAvg = stats.BudgetUSDTotal / float64(stats.BudgetTasks)
	}
}

func detectConflicts(members []model.Member) []model.Conflict {
	conflicts := make([]model.Conflict, 0)

	activeTasks := make(map[string][]string)
	activeBranches := make(map[string][]string)

	for _, m := range members {
		if m.Status != "in_progress" {
			continue
		}
		if m.TaskID != "" {
			activeTasks[m.TaskID] = append(activeTasks[m.TaskID], m.Alias)
		}
		if m.Branch != "" {
			activeBranches[m.Branch] = append(activeBranches[m.Branch], m.Alias)
		}
	}

	for taskID, aliases := range activeTasks {
		if len(aliases) <= 1 {
			continue
		}
		conflicts = append(conflicts, model.Conflict{
			Severity:        "critical",
			Title:           "Duplicate Active Task",
			Description:     "Multiple developers are concurrently working on task " + taskID,
			AffectedAliases: aliases,
		})
	}

	for branch, aliases := range activeBranches {
		if len(aliases) <= 1 {
			continue
		}
		alreadyFlagged := false
		for _, c := range conflicts {
			if strings.Contains(c.Description, branch) {
				alreadyFlagged = true
				break
			}
		}
		if alreadyFlagged {
			continue
		}
		conflicts = append(conflicts, model.Conflict{
			Severity:        "critical",
			Title:           "Branch Collision",
			Description:     "Multiple developers are pushing to the same branch: " + branch,
			AffectedAliases: aliases,
		})
	}

	topicOwners := make(map[string][]string)
	summaryOwners := make(map[string][]string)
	for _, m := range members {
		if m.Status != "in_progress" {
			continue
		}
		if topic := taskTopicKey(m.TaskID); topic != "" {
			topicOwners[topic] = appendUniqueAlias(topicOwners[topic], m.Alias)
		}
		if sum := normalizeTaskSummary(m.TaskSummary); sum != "" {
			summaryOwners[sum] = appendUniqueAlias(summaryOwners[sum], m.Alias)
		}
	}

	for topic, aliases := range topicOwners {
		if len(aliases) <= 1 || aliasesShareExactTask(members, aliases) {
			continue
		}
		conflicts = append(conflicts, model.Conflict{
			Severity:        "warning",
			Title:           "Same Fix Topic",
			Description:     "Multiple developers started work on the same topic: " + topic,
			AffectedAliases: aliases,
		})
	}

	for sum, aliases := range summaryOwners {
		if len(aliases) <= 1 || aliasesShareExactTask(members, aliases) {
			continue
		}
		conflicts = append(conflicts, model.Conflict{
			Severity:        "warning",
			Title:           "Same Task Summary",
			Description:     "Multiple developers described the same in-progress work: " + sum,
			AffectedAliases: aliases,
		})
	}

	return conflicts
}

func taskTopicKey(taskID string) string {
	id := strings.ToUpper(strings.TrimSpace(taskID))
	id = strings.TrimPrefix(id, "FIX-")
	parts := strings.SplitN(id, "-", 4)
	if len(parts) < 4 || strings.TrimSpace(parts[3]) == "" {
		return ""
	}
	return strings.ToLower(parts[3])
}

func normalizeTaskSummary(summary string) string {
	return strings.Join(strings.Fields(strings.ToLower(strings.TrimSpace(summary))), " ")
}

func appendUniqueAlias(aliases []string, alias string) []string {
	alias = strings.TrimSpace(alias)
	if alias == "" {
		return aliases
	}
	for _, existing := range aliases {
		if existing == alias {
			return aliases
		}
	}
	return append(aliases, alias)
}

func aliasesShareExactTask(members []model.Member, aliases []string) bool {
	want := make(map[string]struct{}, len(aliases))
	for _, alias := range aliases {
		want[alias] = struct{}{}
	}
	counts := make(map[string]int)
	for _, m := range members {
		if m.Status != "in_progress" || m.TaskID == "" {
			continue
		}
		if _, ok := want[m.Alias]; !ok {
			continue
		}
		counts[m.TaskID]++
		if counts[m.TaskID] >= 2 {
			return true
		}
	}
	return false
}
