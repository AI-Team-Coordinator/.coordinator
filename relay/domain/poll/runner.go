package poll

import (
	"context"
	"fmt"
	"strings"
	"time"

	"relay/domain/activity"
	"relay/domain/config"
	"relay/domain/connectorconfig"
	"relay/domain/gateway"
	"relay/infra"
)

// Job runs on a schedule for one connector source id.
type Job struct {
	SourceID string
	Interval func(cfg *connectorconfig.Service, id string) time.Duration
	Run      func(ctx context.Context, runtime *config.Runtime) error
}

// Start launches all jobs until ctx is cancelled.
func Start(ctx context.Context, gw *gateway.Gateway, runtime *config.Runtime, _ *activity.Recorder, jobs []Job) {
	if gw == nil || runtime == nil {
		return
	}
	for _, job := range jobs {
		if job.Run == nil {
			continue
		}
		go runLoop(ctx, runtime, job)
	}
}

// FindJob returns a scheduled poll job by connector source id.
func FindJob(jobs []Job, sourceID string) (Job, bool) {
	sourceID = strings.TrimSpace(sourceID)
	for _, job := range jobs {
		if job.SourceID == sourceID {
			return job, true
		}
	}
	return Job{}, false
}

// RunNow executes one poll cycle immediately (HTTP trigger from Coordinator).
func RunNow(ctx context.Context, runtime *config.Runtime, job Job) error {
	if job.Run == nil {
		return fmt.Errorf("poll job has no runner")
	}
	if runtime == nil || !runtime.Enabled(job.SourceID) {
		return fmt.Errorf("connector is not enabled")
	}
	ctx = activity.WithTrigger(ctx, activity.TriggerManual)
	return job.Run(ctx, runtime)
}

func runLoop(ctx context.Context, runtime *config.Runtime, job Job) {
	tick := func() {
		if !runtime.Enabled(job.SourceID) {
			return
		}
		runCtx := activity.WithTrigger(ctx, activity.TriggerScheduled)
		if err := job.Run(runCtx, runtime); err != nil {
			infra.LogWarn("poll %s: %v", job.SourceID, err)
		}
	}
	interval := 24 * time.Hour
	if job.Interval != nil && runtime.Config != nil {
		interval = job.Interval(runtime.Config, job.SourceID)
	}
	if interval < time.Minute {
		interval = time.Minute
	}
	infra.LogInfo("poll %s every %s", job.SourceID, interval)
	tick()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tick()
		}
	}
}
