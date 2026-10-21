package statebus

import (
	"context"
	"time"

	"relay/infra"
	"relay/infra/store"
)

// RunWorker drains bus_outbox into coordinator-state with retries.
func RunWorker(ctx context.Context, cfg Config, outbox *store.OutboxStore, sink *Sink) {
	if outbox == nil || sink == nil || !cfg.Ready() {
		return
	}
	infra.LogInfo("statebus outbox worker started")
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			drainOnce(ctx, outbox, sink)
		}
	}
}

func drainOnce(ctx context.Context, outbox *store.OutboxStore, sink *Sink) {
	rows, err := outbox.ClaimPending(ctx, 5)
	if err != nil {
		infra.LogWarn("statebus claim: %v", err)
		return
	}
	for _, row := range rows {
		ev, err := store.DecodeOutboxEvent(row.Payload)
		if err != nil {
			_ = outbox.MarkRetry(ctx, row.ID, row.Attempts, err.Error())
			continue
		}
		if err := sink.Append(ctx, ev); err != nil {
			_ = outbox.MarkRetry(ctx, row.ID, row.Attempts, err.Error())
			continue
		}
		if err := outbox.MarkPushed(ctx, row.ID); err != nil {
			infra.LogWarn("statebus mark pushed: %v", err)
		}
	}
}
