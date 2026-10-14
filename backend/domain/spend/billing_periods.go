package spend

import (
	"sort"
	"strings"
	"time"

	"coordinator/model"
)

func ParseBillingCycleUnix(raw string) int64 {
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

func BillingPeriodsFromSamples(samples []model.UsageSample, now time.Time) []model.BillingPeriod {
	if len(samples) == 0 {
		return nil
	}
	sorted := append([]model.UsageSample(nil), samples...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].TS < sorted[j].TS })

	var out []model.BillingPeriod
	var currentISO string
	var currentStart int64

	closeLast := func(end int64) {
		if len(out) == 0 {
			return
		}
		if end > out[len(out)-1].StartUnix {
			out[len(out)-1].EndUnix = end
		}
	}

	for _, s := range sorted {
		iso := strings.TrimSpace(s.BillingCycleStart)
		if iso == "" {
			continue
		}
		start := ParseBillingCycleUnix(iso)
		if start == 0 {
			continue
		}
		if iso != currentISO {
			closeLast(s.TS)
			currentISO = iso
			currentStart = start
			out = append(out, model.BillingPeriod{
				StartISO:  iso,
				StartUnix: currentStart,
				EndUnix:   now.Unix(),
			})
		}
	}
	if len(out) > 0 {
		out[len(out)-1].EndUnix = now.Unix()
	}
	return out
}

func FilterDaysInPeriod(days []model.HourSpend, from, to int64) []model.HourSpend {
	if from <= 0 && to <= 0 {
		return days
	}
	out := make([]model.HourSpend, 0, len(days))
	for _, d := range days {
		if from > 0 && d.Hour+86400 <= from {
			continue
		}
		if to > 0 && d.Hour > to {
			continue
		}
		out = append(out, d)
	}
	return out
}
