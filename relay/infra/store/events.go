package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"

	"relay/model"
)

// EventStore persists normalized external events for query and future stats.
type EventStore struct {
	db *sql.DB
}

func NewEventStore(db *sql.DB) *EventStore {
	if db == nil {
		return nil
	}
	return &EventStore{db: db}
}

func (s *EventStore) Insert(ctx context.Context, ev model.ExternalEvent) error {
	if s == nil || s.db == nil {
		return nil
	}
	eventID := strings.TrimSpace(ev.ID)
	if eventID == "" {
		return fmt.Errorf("event id is required")
	}
	payload := ev.Data
	if payload == nil {
		payload = map[string]any{}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	occurred := parseTimePtr(ev.OccurredAt)
	received := parseTime(ev.ReceivedAt)
	if received.IsZero() {
		received = time.Now().UTC()
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO external_events (
			event_id, source, external_id, class, kind, title, summary, url, occurred_at, received_at, payload, tags
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12)
		ON CONFLICT (event_id) DO NOTHING`,
		eventID,
		strings.TrimSpace(ev.Source),
		nullString(ev.ExternalID),
		strings.TrimSpace(ev.Class),
		strings.TrimSpace(ev.Kind),
		nullString(ev.Title),
		nullString(ev.Summary),
		nullString(ev.URL),
		occurred,
		received,
		string(raw),
		pq.Array(ev.Tags),
	)
	return err
}

func parseTime(raw string) time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Time{}
	}
	return t.UTC()
}

func parseTimePtr(raw string) *time.Time {
	t := parseTime(raw)
	if t.IsZero() {
		return nil
	}
	return &t
}

func nullString(s string) interface{} {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return s
}
