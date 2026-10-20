package adapter

import (
	"context"

	"relay/model"
)

// Delivery is one inbound webhook before an adapter normalizes it.
type Delivery struct {
	Event string
	Body  []byte
	Sig   string
}

// Adapter turns an external payload into bus records.
// Normalize is the async path. Read fetches one object on demand.
type Adapter interface {
	ID() string
	Normalize(ctx context.Context, d Delivery) ([]model.ExternalEvent, error)
	Read(ctx context.Context, externalID string) (model.ExternalEvent, error)
}
