package gateway

import (
	"context"

	"relay/infra/store"
	"relay/model"
)

// ChainSink writes to every sink in order; stops on first error.
type ChainSink []Sink

func (c ChainSink) Append(ctx context.Context, ev model.ExternalEvent) error {
	for _, s := range c {
		if s == nil {
			continue
		}
		if err := s.Append(ctx, ev); err != nil {
			return err
		}
	}
	return nil
}

// CompositeSink persists events and enqueues git bus delivery (Outbox handles push).
type CompositeSink struct {
	Bus    Sink // optional sync bus (debug); production uses Outbox + worker
	Store  *store.EventStore
	Outbox *store.OutboxStore
}

func (c CompositeSink) Append(ctx context.Context, ev model.ExternalEvent) error {
	if c.Store != nil {
		if err := c.Store.Insert(ctx, ev); err != nil {
			return err
		}
	}
	if c.Outbox != nil {
		if err := c.Outbox.Enqueue(ctx, ev); err != nil {
			return err
		}
		return nil
	}
	if c.Bus != nil {
		if err := c.Bus.Append(ctx, ev); err != nil {
			return err
		}
	}
	return nil
}
