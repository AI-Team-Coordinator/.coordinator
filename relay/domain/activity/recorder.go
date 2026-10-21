package activity

import (
	"context"

	"relay/infra"
	"relay/infra/store"
)

// Recorder writes connector activity lines to Postgres.
type Recorder struct {
	Store *store.ActivityStore
}

func (r *Recorder) Record(ctx context.Context, connectorID, kind, level, summary string, detail map[string]any) {
	if r == nil || r.Store == nil {
		return
	}
	if err := r.Store.Insert(ctx, connectorID, kind, level, summary, detail); err != nil {
		infra.LogWarn("connector activity %s: %v", connectorID, err)
	}
}
