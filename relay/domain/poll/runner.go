package poll

import (
	"context"
	"time"

	"relay/domain/config"
	"relay/domain/gateway"
	"relay/infra"
	"relay/infra/settings"
)

// Job runs on a schedule for one connector source id.
type Job struct {
	SourceID string
	Interval func(file settings.ConnectorsFile, id string) time.Duration
	Run      func(ctx context.Context, runtime *config.Runtime, settings *settings.RelayConnectors) error
}

// Start launches all jobs until ctx is cancelled.
func Start(ctx context.Context, gw *gateway.Gateway, runtime *config.Runtime, relaySettings *settings.RelayConnectors, jobs []Job) {
	if gw == nil || runtime == nil || relaySettings == nil {
		return
	}
	for _, job := range jobs {
		if job.Run == nil {
			continue
		}
		go runLoop(ctx, gw, runtime, relaySettings, job)
	}
}

func runLoop(ctx context.Context, gw *gateway.Gateway, runtime *config.Runtime, relaySettings *settings.RelayConnectors, job Job) {
	tick := func() {
		if !runtime.Enabled(job.SourceID) {
			return
		}
		if err := job.Run(ctx, runtime, relaySettings); err != nil {
			infra.LogWarn("poll %s: %v", job.SourceID, err)
		}
	}
	file, err := relaySettings.Load()
	if err != nil {
		infra.LogWarn("poll %s settings: %v", job.SourceID, err)
		return
	}
	interval := 24 * time.Hour
	if job.Interval != nil {
		interval = job.Interval(file, job.SourceID)
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
