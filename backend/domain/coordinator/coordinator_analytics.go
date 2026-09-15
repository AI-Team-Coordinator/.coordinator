package coordinator

import (
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

func computeStats(events []model.Event, members []model.Member) model.Stats {
	return computeStatsSince(events, members, billingCycleStartUnix(members))
}

func countActiveNow(members []model.Member) int {
	n := 0
	for _, m := range members {
		slots := m.Slots()
		n += len(slots)
		if len(slots) == 0 && m.Status == "in_progress" {
			n++
		}
		if m.Research != nil && m.Research.Status == "active" {
			n++
		}
	}
	return n
}

func computeStatsSince(events []model.Event, members []model.Member, cycleStart int64) model.Stats {
	stats := model.Stats{}
	stats.ActiveNow = countActiveNow(members)

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
		if ev.Event == "research_completed" {
			addResearchSpend(&stats, ev, startOfToday, startOfWeek, cycleStart)
			continue
		}
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
		inCycle := inPeriod(ev.Timestamp, cycleStart)
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
			if inCycle {
				stats.CostUSDCycle += cost
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
			if inCycle {
				stats.BudgetUSDCycle += budget
			}
			if ev.SpendKind == "infra" {
				if ev.Timestamp >= startOfToday {
					stats.BudgetUSDInfraToday += budget
				}
				if ev.Timestamp >= startOfWeek {
					stats.BudgetUSDInfraWeek += budget
				}
				if inCycle {
					stats.BudgetUSDInfraCycle += budget
				}
			} else {
				if ev.Timestamp >= startOfToday {
					stats.BudgetUSDProductToday += budget
				}
				if ev.Timestamp >= startOfWeek {
					stats.BudgetUSDProductWeek += budget
				}
				if inCycle {
					stats.BudgetUSDProductCycle += budget
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
			if inCycle {
				stats.OnDemandUSDCycle += od
			}
		}
		addTaskPoolPcts(&stats, ev, startOfToday, cycleStart)
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
	addOpenResearchSpend(&stats, members)
	return stats
}

func inPeriod(ts, start int64) bool {
	return start <= 0 || ts >= start
}

func billingCycleStartUnix(members []model.Member) int64 {
	_, ts := billingCycleFromMembers(members)
	return ts
}

func billingCycleFromMembers(members []model.Member) (string, int64) {
	var iso string
	var best int64
	consider := func(raw string) {
		ts := parseBillingCycleUnix(raw)
		if ts == 0 {
			return
		}
		if ts > best {
			best = ts
			iso = raw
		}
	}
	for _, m := range members {
		if m.CursorUsage != nil {
			consider(m.CursorUsage.BillingCycleStart)
		}
		for _, slot := range m.Slots() {
			if slot.CursorUsage != nil {
				consider(slot.CursorUsage.BillingCycleStart)
			}
		}
		if m.Research != nil && m.Research.CursorUsage != nil {
			consider(m.Research.CursorUsage.BillingCycleStart)
		}
	}
	return iso, best
}

func parseBillingCycleUnix(raw string) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
		if t, err := time.Parse(layout, raw); err == nil {
			return t.Unix()
		}
	}
	return 0
}

func addResearchSpend(stats *model.Stats, ev model.Event, startOfToday, startOfWeek, cycleStart int64) {
	stats.ResearchCompleted++
	if ev.Timestamp >= startOfToday {
		stats.ResearchCompletedToday++
	}
	if ev.Timestamp >= startOfWeek {
		stats.ResearchCompletedWeek++
	}
	inCycle := inPeriod(ev.Timestamp, cycleStart)
	if ev.BudgetUSD != nil {
		budget := *ev.BudgetUSD
		if ev.Timestamp >= startOfToday {
			stats.BudgetUSDResearchToday += budget
		}
		if ev.Timestamp >= startOfWeek {
			stats.BudgetUSDResearchWeek += budget
		}
		if inCycle {
			stats.BudgetUSDResearchCycle += budget
		}
	}
	if ev.CostUSD != nil {
		cost := *ev.CostUSD
		if ev.Timestamp >= startOfToday {
			stats.CostUSDResearchToday += cost
		}
		if ev.Timestamp >= startOfWeek {
			stats.CostUSDResearchWeek += cost
		}
		if inCycle {
			stats.CostUSDResearchCycle += cost
		}
	}
	addResearchPoolPcts(stats, ev.CursorModelsPct, ev.OtherModelsPct, ev.Timestamp >= startOfToday, inCycle, false)
}

func addOpenResearchSpend(stats *model.Stats, members []model.Member) {
	for _, m := range members {
		if m.Research == nil || m.Research.SpendShared {
			continue
		}
		if m.Research.BudgetUSD != nil {
			budget := *m.Research.BudgetUSD
			stats.BudgetUSDResearchToday += budget
			stats.BudgetUSDResearchWeek += budget
			stats.BudgetUSDResearchCycle += budget
			stats.BudgetUSDResearchOpen += budget
		}
		if m.Research.CostUSD != nil {
			cost := *m.Research.CostUSD
			stats.CostUSDResearchToday += cost
			stats.CostUSDResearchWeek += cost
			stats.CostUSDResearchCycle += cost
			stats.CostUSDResearchOpen += cost
		}
		addResearchPoolPcts(stats, m.Research.CursorModelsPct, m.Research.OtherModelsPct, true, true, true)
	}
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
				t.CursorModelsPct = nil
				t.OtherModelsPct = nil
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
			t.CursorModelsPct = ev.CursorModelsPct
			t.OtherModelsPct = ev.OtherModelsPct
			t.SpendKind = ev.SpendKind
			t.ActiveSeconds = ev.ActiveSeconds
			if len(ev.ActivityWindows) > 0 {
				t.ActivityWindows = ev.ActivityWindows
			}
			if t.StartedAt == 0 {
				t.StartedAt = ev.Timestamp
			}
		}
	}

	for _, m := range members {
		for _, slot := range m.Slots() {
			t := ensure(slot.TaskID)
			t.Status = "in_progress"
			t.CompletedAt = 0
			t.Alias = m.Alias
			if slot.Branch != "" {
				t.Branch = slot.Branch
			}
			if slot.Title != "" {
				t.Title = slot.Title
			}
			if len(slot.Services) > 0 {
				t.Services = slot.Services
			}
			t.SpendKind = slot.SpendKind
			t.SpendShared = slot.SpendShared
			t.ActivityWindows = slot.ActivityWindows
			t.CostUSD = slot.CostUSD
			t.BudgetUSD = slot.BudgetUSD
			t.OnDemandUSD = slot.OnDemandUSD
			t.CursorModelsPct = slot.CursorModelsPct
			t.OtherModelsPct = slot.OtherModelsPct
			t.Kind = taskKind(slot.TaskID, t.Branch)
			if t.StartedAt == 0 && !slot.StartedAt.IsZero() {
				t.StartedAt = slot.StartedAt.Unix()
			}
			if t.StartedAt == 0 && !slot.UpdatedAt.IsZero() {
				t.StartedAt = slot.UpdatedAt.Unix()
			}
			last := slot.LastActivityAt
			if last.IsZero() {
				last = slot.UpdatedAt
			}
			d, paused := model.ActiveDuration(slot.ActivityWindows, slot.StartedAt, last, now)
			t.DurationSeconds = d
			t.ClockPaused = paused
		}
	}

	closeOrphanTasks(acc, members, events)

	nowUnix := now.Unix()
	out := make([]model.Task, 0, len(acc))
	for _, t := range acc {
		if t.Kind == "" {
			t.Kind = taskKind(t.TaskID, t.Branch)
		}
		if t.Status == "in_progress" && t.DurationSeconds == 0 && !t.ClockPaused {
			end := nowUnix
			if t.StartedAt > 0 && end >= t.StartedAt {
				t.DurationSeconds = end - t.StartedAt
			}
		} else if t.Status == "completed" {
			if t.ActiveSeconds != nil {
				t.DurationSeconds = *t.ActiveSeconds
			} else if t.StartedAt > 0 && t.CompletedAt >= t.StartedAt {
				t.DurationSeconds = t.CompletedAt - t.StartedAt
			}
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
		slots := m.Slots()
		if len(slots) == 0 && m.Status == "in_progress" {
			if !m.SpendShared {
				addOneOpenSpend(stats, m.CostUSD, m.BudgetUSD, m.OnDemandUSD, m.CursorModelsPct, m.OtherModelsPct, m.SpendKind)
			}
			continue
		}
		for _, slot := range slots {
			if slot.SpendShared {
				continue
			}
			addOneOpenSpend(stats, slot.CostUSD, slot.BudgetUSD, slot.OnDemandUSD, slot.CursorModelsPct, slot.OtherModelsPct, slot.SpendKind)
		}
	}
	if stats.CostTasks > 0 {
		stats.CostUSDAvg = stats.CostUSDTotal / float64(stats.CostTasks)
	}
	if stats.BudgetTasks > 0 {
		stats.BudgetUSDAvg = stats.BudgetUSDTotal / float64(stats.BudgetTasks)
	}
}

func addOneOpenSpend(stats *model.Stats, costUSD, budgetUSD, onDemandUSD, cursorPct, otherPct *float64, spendKind string) {
	if costUSD != nil {
		cost := *costUSD
		stats.CostUSDTotal += cost
		stats.CostUSDToday += cost
		stats.CostUSDWeek += cost
		stats.CostUSDCycle += cost
		stats.CostUSDOpen += cost
		stats.CostTasks++
	}
	if budgetUSD != nil {
		budget := *budgetUSD
		stats.BudgetUSDTotal += budget
		stats.BudgetUSDToday += budget
		stats.BudgetUSDWeek += budget
		stats.BudgetUSDCycle += budget
		stats.BudgetUSDOpen += budget
		stats.BudgetTasks++
		if spendKind == "infra" {
			stats.BudgetUSDInfraToday += budget
			stats.BudgetUSDInfraWeek += budget
			stats.BudgetUSDInfraCycle += budget
		} else {
			stats.BudgetUSDProductToday += budget
			stats.BudgetUSDProductWeek += budget
			stats.BudgetUSDProductCycle += budget
		}
	}
	if onDemandUSD != nil {
		od := *onDemandUSD
		stats.OnDemandUSDToday += od
		stats.OnDemandUSDWeek += od
		stats.OnDemandUSDCycle += od
		stats.OnDemandUSDOpen += od
	}
	addOpenPoolPcts(stats, cursorPct, otherPct, spendKind)
}

func addTaskPoolPcts(stats *model.Stats, ev model.Event, startOfToday, cycleStart int64) {
	inCycle := inPeriod(ev.Timestamp, cycleStart)
	addPeriodPct(ev.CursorModelsPct, ev.Timestamp >= startOfToday, inCycle, &stats.CursorModelsPctToday, &stats.CursorModelsPctCycle)
	addPeriodPct(ev.OtherModelsPct, ev.Timestamp >= startOfToday, inCycle, &stats.OtherModelsPctToday, &stats.OtherModelsPctCycle)
	if !inCycle {
		return
	}
	if ev.SpendKind == "infra" {
		addPct(ev.CursorModelsPct, &stats.CursorModelsPctInfraCycle)
		addPct(ev.OtherModelsPct, &stats.OtherModelsPctInfraCycle)
		return
	}
	addPct(ev.CursorModelsPct, &stats.CursorModelsPctProductCycle)
	addPct(ev.OtherModelsPct, &stats.OtherModelsPctProductCycle)
}

func addResearchPoolPcts(stats *model.Stats, cursor, other *float64, today, cycle, open bool) {
	if cursor != nil {
		n := *cursor
		if today {
			stats.CursorModelsPctResearchToday += n
		}
		if cycle {
			stats.CursorModelsPctResearchCycle += n
		}
		if open {
			stats.CursorModelsPctResearchOpen += n
		}
	}
	if other != nil {
		n := *other
		if today {
			stats.OtherModelsPctResearchToday += n
		}
		if cycle {
			stats.OtherModelsPctResearchCycle += n
		}
		if open {
			stats.OtherModelsPctResearchOpen += n
		}
	}
}

func addOpenPoolPcts(stats *model.Stats, cursor, other *float64, spendKind string) {
	if cursor != nil {
		n := *cursor
		stats.CursorModelsPctToday += n
		stats.CursorModelsPctCycle += n
		stats.CursorModelsPctOpen += n
		if spendKind == "infra" {
			stats.CursorModelsPctInfraCycle += n
		} else {
			stats.CursorModelsPctProductCycle += n
		}
	}
	if other != nil {
		n := *other
		stats.OtherModelsPctToday += n
		stats.OtherModelsPctCycle += n
		stats.OtherModelsPctOpen += n
		if spendKind == "infra" {
			stats.OtherModelsPctInfraCycle += n
		} else {
			stats.OtherModelsPctProductCycle += n
		}
	}
}

func addPeriodPct(v *float64, today, cycle bool, dstToday, dstCycle *float64) {
	if v == nil {
		return
	}
	n := *v
	if today {
		*dstToday += n
	}
	if cycle {
		*dstCycle += n
	}
}

func addPct(v *float64, dst *float64) {
	if v == nil {
		return
	}
	*dst += *v
}

func closeOrphanTasks(acc map[string]*model.Task, members []model.Member, events []model.Event) {
	live := make(map[string]struct{})
	liveStartByAlias := make(map[string]int64)
	for _, m := range members {
		for _, slot := range m.Slots() {
			if slot.TaskID == "" {
				continue
			}
			live[slot.TaskID] = struct{}{}
			ts := int64(0)
			if !slot.StartedAt.IsZero() {
				ts = slot.StartedAt.Unix()
			} else if !slot.UpdatedAt.IsZero() {
				ts = slot.UpdatedAt.Unix()
			}
			if ts == 0 {
				continue
			}
			if prev, ok := liveStartByAlias[m.Alias]; !ok || ts < prev {
				liveStartByAlias[m.Alias] = ts
			}
		}
	}
	nextStart := nextTaskStartByAlias(events)
	for _, t := range acc {
		if t.Status != "in_progress" {
			continue
		}
		if _, ok := live[t.TaskID]; ok {
			continue
		}
		t.Status = "completed"
		closeAt := nextStart[t.TaskID]
		if closeAt == 0 && t.Alias != "" {
			if liveTs := liveStartByAlias[t.Alias]; liveTs > t.StartedAt {
				closeAt = liveTs
			}
		}
		if closeAt == 0 {
			closeAt = t.StartedAt
		}
		t.CompletedAt = closeAt
	}
}

func nextTaskStartByAlias(events []model.Event) map[string]int64 {
	type start struct {
		id, alias string
		ts        int64
	}
	starts := make([]start, 0)
	for _, ev := range events {
		if ev.Event != "task_started" || ev.TaskID == "" {
			continue
		}
		starts = append(starts, start{id: ev.TaskID, alias: ev.Alias, ts: ev.Timestamp})
	}
	out := make(map[string]int64, len(starts))
	for i, s := range starts {
		if s.alias == "" {
			continue
		}
		for j := i + 1; j < len(starts); j++ {
			if starts[j].alias == s.alias && starts[j].id != s.id {
				out[s.id] = starts[j].ts
				break
			}
		}
	}
	return out
}

func detectConflicts(members []model.Member) []model.Conflict {
	conflicts := make([]model.Conflict, 0)

	activeTasks := make(map[string][]string)
	activeBranches := make(map[string][]string)

	for _, m := range members {
		for _, slot := range m.Slots() {
			if slot.TaskID != "" {
				activeTasks[slot.TaskID] = appendUniqueAlias(activeTasks[slot.TaskID], m.Alias)
			}
			b := strings.ToLower(strings.TrimSpace(slot.Branch))
			if b != "" && b != "main" && b != "master" {
				activeBranches[slot.Branch] = appendUniqueAlias(activeBranches[slot.Branch], m.Alias)
			}
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
		for _, slot := range m.Slots() {
			if topic := taskTopicKey(slot.TaskID); topic != "" {
				topicOwners[topic] = appendUniqueAlias(topicOwners[topic], m.Alias)
			}
			if sum := normalizeTaskSummary(slot.Summary); sum != "" {
				summaryOwners[sum] = appendUniqueAlias(summaryOwners[sum], m.Alias)
			}
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

	conflicts = append(conflicts, selfScopeConflicts(members)...)
	conflicts = append(conflicts, peerScopeConflicts(members)...)

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
		if _, ok := want[m.Alias]; !ok {
			continue
		}
		for _, slot := range m.Slots() {
			if slot.TaskID == "" {
				continue
			}
			counts[slot.TaskID]++
			if counts[slot.TaskID] >= 2 {
				return true
			}
		}
	}
	return false
}

func selfScopeConflicts(members []model.Member) []model.Conflict {
	out := make([]model.Conflict, 0)
	for _, m := range members {
		slots := m.Slots()
		if len(slots) < 2 {
			continue
		}
		for i := 0; i < len(slots); i++ {
			for j := i + 1; j < len(slots); j++ {
				if overlap, label := slotScopeOverlap(slots[i].Services, slots[j].Services); overlap {
					out = append(out, model.Conflict{
						Severity:        "critical",
						Title:           "Parallel Slot Overlap",
						Description:     m.Alias + " has two in-progress tasks claiming " + label,
						AffectedAliases: []string{m.Alias},
						Service:         strings.TrimSpace(label),
					})
				}
			}
		}
	}
	return out
}

type peerClaim struct {
	alias, taskID, title, branch, service, serviceKey string
}

func peerScopeConflicts(members []model.Member) []model.Conflict {
	claims := make([]peerClaim, 0)
	for _, m := range members {
		for _, slot := range m.Slots() {
			title := strings.TrimSpace(slot.Title)
			if title == "" {
				title = strings.TrimSpace(slot.Summary)
			}
			if title == "" {
				title = slot.TaskID
			}
			for _, raw := range slot.Services {
				key, kind := scopeKey(raw)
				if kind != "product" {
					continue
				}
				claims = append(claims, peerClaim{
					alias:      m.Alias,
					taskID:     slot.TaskID,
					title:      title,
					branch:     slot.Branch,
					service:    strings.TrimSpace(raw),
					serviceKey: key,
				})
			}
		}
	}
	out := make([]model.Conflict, 0)
	seen := make(map[string]struct{})
	for i := 0; i < len(claims); i++ {
		for j := i + 1; j < len(claims); j++ {
			a, b := claims[i], claims[j]
			if a.alias == b.alias || a.serviceKey != b.serviceKey {
				continue
			}
			fp := peerPairKey(a, b)
			if _, ok := seen[fp]; ok {
				continue
			}
			seen[fp] = struct{}{}
			aliases := []string{a.alias, b.alias}
			sort.Strings(aliases)
			out = append(out, model.Conflict{
				Severity:        "warning",
				Title:           "Peer Scope Overlap",
				Description:     a.alias + " «" + a.title + "» and " + b.alias + " «" + b.title + "» both claim " + a.service,
				AffectedAliases: aliases,
				Service:         a.service,
			})
		}
	}
	return out
}

func peerPairKey(a, b peerClaim) string {
	left, right := a, b
	if a.alias > b.alias || (a.alias == b.alias && a.taskID > b.taskID) {
		left, right = b, a
	}
	return strings.ToLower(left.alias + ":" + left.taskID + "|" + right.alias + ":" + right.taskID + "|" + left.serviceKey)
}

func slotScopeOverlap(a, b []string) (bool, string) {
	seen := make(map[string]string)
	for _, raw := range a {
		key, kind := scopeKey(raw)
		if key == "" || kind == "bus" {
			continue
		}
		seen[key] = kind
	}
	for _, raw := range b {
		key, kind := scopeKey(raw)
		if key == "" || kind == "bus" {
			continue
		}
		if prev, ok := seen[key]; ok && (kind == "product" || prev == "product" || kind == "workspace") {
			return true, raw
		}
	}
	return false, ""
}

func scopeKey(raw string) (key, kind string) {
	n := strings.ToLower(strings.TrimSpace(raw))
	n = strings.TrimPrefix(n, ".")
	n = strings.ReplaceAll(n, "-", "_")
	n = strings.ReplaceAll(n, " ", "_")
	if n == "" {
		return "", ""
	}
	switch n {
	case "common":
		return n, "bus"
	case "cursor", "coordinator":
		return n, "workspace"
	default:
		return n, "product"
	}
}
