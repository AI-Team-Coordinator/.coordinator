package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"relay/model"
)

const (
	OutboxPending = "pending"
	OutboxPushed  = "pushed"
	OutboxFailed  = "failed"
)

type OutboxStore struct {
	db *sql.DB
}

func NewOutboxStore(db *sql.DB) *OutboxStore {
	if db == nil {
		return nil
	}
	return &OutboxStore{db: db}
}

func (s *OutboxStore) Enqueue(ctx context.Context, ev model.ExternalEvent) error {
	if s == nil || s.db == nil {
		return nil
	}
	eventID := strings.TrimSpace(ev.ID)
	if eventID == "" {
		return fmt.Errorf("event id is required")
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `
		INSERT INTO bus_outbox (event_id, payload, status, next_attempt_at, updated_at)
		VALUES ($1, $2::jsonb, $3, now(), now())
		ON CONFLICT (event_id) DO NOTHING`,
		eventID, string(raw), OutboxPending,
	)
	return err
}

type OutboxRow struct {
	ID      int64
	EventID string
	Payload []byte
	Attempts int
}

func (s *OutboxStore) ClaimPending(ctx context.Context, limit int) ([]OutboxRow, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, event_id, payload, attempts
		FROM bus_outbox
		WHERE status IN ($1, $2) AND next_attempt_at <= now()
		ORDER BY id
		LIMIT $3`, OutboxPending, OutboxFailed, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []OutboxRow
	for rows.Next() {
		var r OutboxRow
		if err := rows.Scan(&r.ID, &r.EventID, &r.Payload, &r.Attempts); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

func (s *OutboxStore) MarkPushed(ctx context.Context, id int64) error {
	_, err := s.db.ExecContext(ctx, `
		UPDATE bus_outbox SET status = $1, updated_at = now(), last_error = NULL WHERE id = $2`,
		OutboxPushed, id)
	return err
}

func (s *OutboxStore) MarkRetry(ctx context.Context, id int64, attempts int, errMsg string) error {
	delay := retryDelay(attempts)
	_, err := s.db.ExecContext(ctx, `
		UPDATE bus_outbox
		SET status = $1, attempts = $2, last_error = $3, next_attempt_at = now() + ($4 * interval '1 second'), updated_at = now()
		WHERE id = $5`,
		OutboxFailed, attempts+1, truncateErr(errMsg), int(delay.Seconds()), id)
	return err
}

func retryDelay(attempts int) time.Duration {
	if attempts < 1 {
		return 30 * time.Second
	}
	if attempts > 6 {
		return 30 * time.Minute
	}
	return time.Duration(attempts*attempts) * time.Minute
}

func truncateErr(s string) string {
	s = strings.TrimSpace(s)
	if len(s) > 500 {
		return s[:500]
	}
	return s
}

func DecodeOutboxEvent(payload []byte) (model.ExternalEvent, error) {
	var ev model.ExternalEvent
	if err := json.Unmarshal(payload, &ev); err != nil {
		return model.ExternalEvent{}, err
	}
	return ev, nil
}
