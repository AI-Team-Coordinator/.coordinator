package repository

import (
	"context"

	"coordinator/model"
)

// EventStore is the query cache over jsonl. SQLite implementation lives in infra/db.
type EventStore interface {
	Sync(ctx context.Context) error
	List(ctx context.Context, q model.EventQuery) ([]model.Event, int, error)
	Close() error
}
