package coordinator

import (
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

const (
	eventCoordinatorWarning = "coordinator_warning"
	eventCoordinatorStop    = "coordinator_stop"
)

func noticeEventName(severity string) string {
	if strings.EqualFold(strings.TrimSpace(severity), "critical") {
		return eventCoordinatorStop
	}
	return eventCoordinatorWarning
}

func conflictFingerprint(c model.Conflict) string {
	aliases := append([]string{}, c.AffectedAliases...)
	sort.Strings(aliases)
	return strings.ToLower(c.Title + "|" + strings.Join(aliases, ",") + "|" + c.Service + "|" + c.Description)
}

func noticeFingerprints(events []model.Event) map[string]struct{} {
	known := make(map[string]struct{})
	for _, ev := range events {
		if ev.Event != eventCoordinatorWarning && ev.Event != eventCoordinatorStop {
			continue
		}
		if ev.Findings != "" {
			known[ev.Findings] = struct{}{}
		}
		if ev.Summary != "" {
			known[ev.Event+"|"+ev.Summary] = struct{}{}
		}
	}
	return known
}

func pendingNotices(alias, taskID string, conflicts []model.Conflict, known map[string]struct{}) []model.Event {
	if alias == "" || len(conflicts) == 0 {
		return nil
	}
	now := time.Now().Unix()
	out := make([]model.Event, 0)
	for _, c := range conflicts {
		if strings.TrimSpace(c.Title) == "" {
			continue
		}
		kind := noticeEventName(c.Severity)
		fp := conflictFingerprint(c)
		summary := c.Title + ": " + c.Description
		if _, ok := known[fp]; ok {
			continue
		}
		if _, ok := known[kind+"|"+summary]; ok {
			continue
		}
		out = append(out, model.Event{
			Timestamp: now,
			Event:     kind,
			TaskID:    taskID,
			Alias:     alias,
			Service:   c.Service,
			Summary:   summary,
			Findings:  fp,
		})
	}
	return out
}

func observerTaskID(members []model.Member, alias string) string {
	for _, m := range members {
		if m.Alias != alias {
			continue
		}
		if id := strings.TrimSpace(m.TaskID); id != "" {
			return id
		}
		slots := m.Slots()
		if n := len(slots); n > 0 {
			return slots[n-1].TaskID
		}
	}
	return ""
}
