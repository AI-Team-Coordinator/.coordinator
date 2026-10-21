package gateway

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"relay/domain/connector"
	"relay/model"
)

// ErrIgnored means the connector deliberately skipped this delivery (not an error).
var ErrIgnored = errors.New("connector: delivery ignored")

// ErrUnknownConnector is returned when source id is not registered.
var ErrUnknownConnector = errors.New("connector: unknown source")

// Sink appends standardized events to the coordinator-state bus.
type Sink interface {
	Append(ctx context.Context, ev model.ExternalEvent) error
}

// Gateway is the internal API between connectors and Relay storage.
// Connectors may be in-process (Go) or out-of-process (POST /connectors/{id}/events).
type EnabledFunc func(source string) bool

type Gateway struct {
	Registry *connector.Registry
	Sink     Sink
	IsEnabled EnabledFunc
}

func (g *Gateway) connectorActive(source string) bool {
	if g.IsEnabled != nil && !g.IsEnabled(source) {
		return false
	}
	return true
}

// Publish writes already-normalized events. Each event.Source must match the connector id when routed.
func (g *Gateway) Publish(ctx context.Context, source string, events []model.ExternalEvent) error {
	if g == nil || g.Sink == nil {
		return fmt.Errorf("gateway not configured")
	}
	source = strings.TrimSpace(source)
	if source == "" {
		return fmt.Errorf("source is required")
	}
	if _, ok := g.Registry.Get(source); !ok {
		return ErrUnknownConnector
	}
	if !g.connectorActive(source) {
		return ErrUnknownConnector
	}
	for i := range events {
		ev := events[i]
		if strings.TrimSpace(ev.Source) == "" {
			ev.Source = source
		}
		if ev.Source != source {
			return fmt.Errorf("event source %q does not match connector %q", ev.Source, source)
		}
		if err := g.Sink.Append(ctx, ev); err != nil {
			return err
		}
	}
	return nil
}

// Webhook runs a connector's Normalize path and publishes the result.
func (g *Gateway) Webhook(ctx context.Context, source string, d connector.Delivery) error {
	c, ok := g.Registry.Get(source)
	if !ok {
		return ErrUnknownConnector
	}
	if !g.connectorActive(source) {
		return ErrUnknownConnector
	}
	events, err := c.Normalize(ctx, d)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return ErrIgnored
	}
	return g.Publish(ctx, source, events)
}
