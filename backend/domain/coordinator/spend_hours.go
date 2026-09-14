package coordinator

import (
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

const hourlySpendWindow = 30 * 24 * time.Hour
const dailySpendDays = 30

type liveParticipant struct {
	kind, id, title, alias string
	windows                []model.ActivityWindow
}

func BuildHourlySpend(samples []model.UsageSample, members []model.Member, now time.Time, window time.Duration) []model.HourSpend {
	if window <= 0 {
		window = hourlySpendWindow
	}
	if now.IsZero() {
		now = time.Now()
	}
	loc := now.Location()
	hourNow := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), 0, 0, 0, loc)
	start := time.Date(now.Add(-window).Year(), now.Add(-window).Month(), now.Add(-window).Day(), now.Add(-window).Hour(), 0, 0, 0, loc)

	sorted := append([]model.UsageSample(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TS < sorted[j].TS })

	live := collectLiveParticipants(members, now)
	var out []model.HourSpend
	for h := start; !h.After(hourNow); h = h.Add(time.Hour) {
		hEnd := h.Add(time.Hour)
		hs, he := h.Unix(), hEnd.Unix()
		var before *model.UsageSample
		var inHour []model.UsageSample
		for i := range sorted {
			ts := sorted[i].TS
			if ts <= hs {
				before = &sorted[i]
				continue
			}
			if ts < he {
				inHour = append(inHour, sorted[i])
			}
		}
		if len(inHour) == 0 {
			continue
		}
		startSnap := before
		if startSnap == nil {
			startSnap = &inHour[0]
		}
		lastIn := inHour[len(inHour)-1]
		delta := model.ComputeUsageDelta(model.CursorUsageFromSample(*startSnap), model.CursorUsageFromSample(lastIn))
		if delta == nil {
			continue
		}
		if delta.CursorModelsPct == 0 && delta.OtherModelsPct == 0 && delta.OnDemandUSD == 0 {
			continue
		}
		hour := model.HourSpend{
			Hour:            hs,
			Alias:           hourAlias(inHour),
			CursorModelsPct: delta.CursorModelsPct,
			OtherModelsPct:  delta.OtherModelsPct,
			OnDemandUSD:     delta.OnDemandUSD,
		}
		if delta.HasPlanPrice() {
			b := delta.BudgetUSD
			hour.BudgetUSD = &b
		}
		hour.Participants = hourParticipants(inHour, live, h, hEnd)
		hour.Aliases = spendAliases(hour.Alias, hour.Participants)
		shared := len(hour.Participants) > 1
		for i := range hour.Participants {
			hour.Participants[i].Shared = shared
		}
		out = append(out, hour)
	}
	return out
}

func hourAlias(inHour []model.UsageSample) string {
	var alias string
	for _, s := range inHour {
		if s.Alias == "" {
			continue
		}
		if alias == "" {
			alias = s.Alias
			continue
		}
		if s.Alias != alias {
			return ""
		}
	}
	return alias
}

func spendAliases(hourAlias string, parts []model.HourParticipant) []string {
	seen := map[string]struct{}{}
	var out []string
	add := func(raw string) {
		a := strings.TrimSpace(raw)
		if a == "" {
			return
		}
		if _, ok := seen[a]; ok {
			return
		}
		seen[a] = struct{}{}
		out = append(out, a)
	}
	add(hourAlias)
	for _, p := range parts {
		add(p.Alias)
	}
	sort.Strings(out)
	return out
}

func BuildDailySpend(hours []model.HourSpend, now time.Time, days int) []model.HourSpend {
	if days <= 0 {
		days = dailySpendDays
	}
	if now.IsZero() {
		now = time.Now()
	}
	loc := now.Location()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	byDay := map[int64]*model.HourSpend{}
	parts := map[int64]map[string]model.HourParticipant{}
	for _, hour := range hours {
		t := time.Unix(hour.Hour, 0).In(loc)
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc).Unix()
		agg, ok := byDay[day]
		if !ok {
			agg = &model.HourSpend{Hour: day}
			byDay[day] = agg
			parts[day] = map[string]model.HourParticipant{}
		}
		agg.CursorModelsPct += hour.CursorModelsPct
		agg.OtherModelsPct += hour.OtherModelsPct
		agg.OnDemandUSD += hour.OnDemandUSD
		if hour.BudgetUSD != nil {
			v := 0.0
			if agg.BudgetUSD != nil {
				v = *agg.BudgetUSD
			}
			v += *hour.BudgetUSD
			agg.BudgetUSD = &v
		}
		if agg.Alias == "" {
			agg.Alias = hour.Alias
		} else if hour.Alias != "" && hour.Alias != agg.Alias {
			agg.Alias = ""
		}
		for _, p := range hour.Participants {
			key := p.Kind + "\x00" + p.ID
			if p.ID == "" {
				key = p.Kind + "\x00" + p.Alias
			}
			if _, exists := parts[day][key]; !exists {
				parts[day][key] = p
			}
		}
	}
	out := make([]model.HourSpend, 0, days)
	start := today.AddDate(0, 0, -(days - 1))
	for d := start; !d.After(today); d = d.AddDate(0, 0, 1) {
		key := d.Unix()
		row := model.HourSpend{Hour: key}
		if agg, ok := byDay[key]; ok {
			row = *agg
			ps := make([]model.HourParticipant, 0, len(parts[key]))
			for _, p := range parts[key] {
				ps = append(ps, p)
			}
			sort.Slice(ps, func(i, j int) bool {
				if ps[i].Kind != ps[j].Kind {
					return ps[i].Kind < ps[j].Kind
				}
				return ps[i].ID < ps[j].ID
			})
			row.Participants = ps
			row.Aliases = spendAliases(row.Alias, ps)
		}
		out = append(out, row)
	}
	return out
}

func collectLiveParticipants(members []model.Member, now time.Time) []liveParticipant {
	var live []liveParticipant
	for _, m := range members {
		for _, slot := range m.Slots() {
			title := slot.Title
			if title == "" {
				title = slot.Summary
			}
			if title == "" {
				title = slot.TaskID
			}
			live = append(live, liveParticipant{
				kind:    "task",
				id:      slot.TaskID,
				title:   title,
				alias:   m.Alias,
				windows: model.SlotWindows(slot.ActivityWindows, slot.StartedAt, slot.LastActivityAt, slot.UpdatedAt, now),
			})
		}
		if m.Research == nil || m.Research.Status != "active" {
			continue
		}
		id := m.Research.SessionID
		if id == "" {
			id = "research:" + m.Alias
		}
		live = append(live, liveParticipant{
			kind:    "research",
			id:      id,
			title:   m.Research.Summary,
			alias:   m.Alias,
			windows: model.ResearchWindows(m.Research.ActivityWindows, now),
		})
	}
	return live
}

func hourParticipants(inHour []model.UsageSample, live []liveParticipant, hourStart, hourEnd time.Time) []model.HourParticipant {
	parts := make(map[string]model.HourParticipant)
	add := func(p model.HourParticipant) {
		if p.Kind == "" {
			return
		}
		if p.ID == "" && p.Alias == "" {
			return
		}
		key := p.Kind + "\x00" + p.ID
		if p.ID == "" {
			key = p.Kind + "\x00" + p.Alias
		}
		prev, ok := parts[key]
		if !ok || (prev.Title == "" && p.Title != "") {
			parts[key] = p
		}
	}
	for _, s := range inHour {
		kind, id := sampleTarget(s)
		if kind == "" {
			continue
		}
		add(model.HourParticipant{Kind: kind, ID: id, Title: s.Title, Alias: s.Alias})
	}
	for _, lp := range live {
		if !windowsHitHour(lp.windows, hourStart, hourEnd) {
			continue
		}
		add(model.HourParticipant{Kind: lp.kind, ID: lp.id, Title: lp.title, Alias: lp.alias})
	}
	out := make([]model.HourParticipant, 0, len(parts))
	for _, p := range parts {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Kind != out[j].Kind {
			return out[i].Kind < out[j].Kind
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func sampleTarget(s model.UsageSample) (kind, id string) {
	kind = s.Kind
	if kind == "research" {
		id = s.SessionID
		if id == "" && s.Alias != "" {
			id = "research:" + s.Alias
		}
		return kind, id
	}
	if s.TaskID != "" {
		if kind == "" {
			kind = "task"
		}
		return kind, s.TaskID
	}
	return "", ""
}

func windowsHitHour(windows []model.ActivityWindow, hourStart, hourEnd time.Time) bool {
	for _, w := range windows {
		if model.WindowOverlapsRange(w, hourStart, hourEnd) {
			return true
		}
	}
	return false
}

func applyCompletedSpendSharing(tasks []model.Task, hours []model.HourSpend) {
	for i := range tasks {
		if tasks[i].Status != "completed" {
			continue
		}
		if !completedExclusive(tasks[i].TaskID, hours) {
			tasks[i].SpendShared = true
			tasks[i].BudgetUSD = nil
			tasks[i].CostUSD = nil
			tasks[i].OnDemandUSD = nil
		}
	}
}

func completedExclusive(taskID string, hours []model.HourSpend) bool {
	if taskID == "" {
		return false
	}
	seen := false
	for _, hour := range hours {
		hasMe := false
		for _, p := range hour.Participants {
			if p.Kind == "task" && p.ID == taskID {
				hasMe = true
				break
			}
		}
		if !hasMe {
			continue
		}
		seen = true
		if len(hour.Participants) > 1 {
			return false
		}
	}
	return seen
}
