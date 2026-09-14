package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"coordinator/model"
)

func (r *FileRepository) usageMeterPath() string {
	return filepath.Join(r.dataDir(), ".cache", "usage_meter.jsonl")
}

func (r *FileRepository) LoadUsageSamples() []model.UsageSample {
	path := r.usageMeterPath()
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	out := make([]model.UsageSample, 0, 256)
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var row model.UsageSample
		if json.Unmarshal(line, &row) != nil || row.TS <= 0 {
			continue
		}
		out = append(out, row)
	}
	return out
}

func (r *FileRepository) AppendUsageSample(sample model.UsageSample) {
	if sample.TS == 0 {
		sample.TS = time.Now().Unix()
	}
	raw, err := json.Marshal(sample)
	if err != nil {
		return
	}
	path := r.usageMeterPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.Write(append(raw, '\n'))
}

func sampleFromCursorUsage(snap *model.CursorUsage, alias string) model.UsageSample {
	row := model.UsageSample{TS: time.Now().Unix(), Alias: alias}
	if snap == nil {
		return row
	}
	row.Plan = snap.Plan
	row.BillingCycleStart = snap.BillingCycleStart
	row.OnDemandCents = snap.OnDemandCents
	row.CursorModelsPct = snap.CursorModelsPct
	row.OtherModelsPct = snap.OtherModelsPct
	row.PlanPriceUSD = snap.PlanPriceUSD
	return row
}
