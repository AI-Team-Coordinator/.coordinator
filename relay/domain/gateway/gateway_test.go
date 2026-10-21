package gateway

import (
	"context"
	"errors"
	"testing"

	"relay/domain/connector"
	"relay/model"
)

type stubConnector struct {
	id      string
	events  []model.ExternalEvent
	normErr error
}

func (s stubConnector) Descriptor() connector.Descriptor {
	return connector.Descriptor{ID: s.id, Name: s.id, Detail: "stub", Enabled: true}
}

func (s stubConnector) Normalize(context.Context, connector.Delivery) ([]model.ExternalEvent, error) {
	if s.normErr != nil {
		return nil, s.normErr
	}
	return s.events, nil
}

func (s stubConnector) Read(context.Context, string) (model.ExternalEvent, error) {
	return model.ExternalEvent{}, errors.New("not implemented")
}

type memSink struct {
	lines []model.ExternalEvent
}

func (m *memSink) Append(_ context.Context, ev model.ExternalEvent) error {
	m.lines = append(m.lines, ev)
	return nil
}

func TestPublishRequiresRegisteredSource(t *testing.T) {
	reg, err := connector.NewRegistry(stubConnector{id: "github"})
	if err != nil {
		t.Fatal(err)
	}
	sink := &memSink{}
	g := &Gateway{Registry: reg, Sink: sink}
	err = g.Publish(context.Background(), "jira", []model.ExternalEvent{{Kind: "x"}})
	if !errors.Is(err, ErrUnknownConnector) {
		t.Fatalf("err=%v", err)
	}
}

func TestWebhookPublishesNormalized(t *testing.T) {
	reg, err := connector.NewRegistry(stubConnector{
		id: "github",
		events: []model.ExternalEvent{
			{ID: "a", Source: "github", Kind: "ci.passed", Class: model.ClassInfo},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	sink := &memSink{}
	g := &Gateway{Registry: reg, Sink: sink}
	if err := g.Webhook(context.Background(), "github", connector.Delivery{}); err != nil {
		t.Fatal(err)
	}
	if len(sink.lines) != 1 || sink.lines[0].Kind != "ci.passed" {
		t.Fatalf("%+v", sink.lines)
	}
}
