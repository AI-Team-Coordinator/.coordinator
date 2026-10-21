package connector

import (
	"context"

	"relay/model"
)

// Delivery is one inbound signal before a connector normalizes it (webhook, poll batch, …).
type Delivery struct {
	Event string
	Body  []byte
	Sig   string
	// Headers carries connector-specific metadata (e.g. GitHub event name lives in Event).
	Headers map[string]string
}

// Descriptor is the public catalog entry for GET /sources.
type Descriptor struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Detail  string `json:"detail"`
	Enabled bool   `json:"enabled"`
}

// Connector turns an external system into the standardized ExternalEvent stream.
// Relay core only sees []model.ExternalEvent after Normalize; bus layout is fixed.
type Connector interface {
	Descriptor() Descriptor
	Normalize(ctx context.Context, d Delivery) ([]model.ExternalEvent, error)
	Read(ctx context.Context, externalID string) (model.ExternalEvent, error)
}
