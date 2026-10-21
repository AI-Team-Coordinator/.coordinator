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

// ActivityStore persists connector operational activity (poll checks, errors).
type ActivityStore struct {
	db *sql.DB
}

func NewActivityStore(db *sql.DB) *ActivityStore {
	if db == nil {
		return nil
	}
	return &ActivityStore{db: db}
}

func (s *ActivityStore) Insert(ctx context.Context, connectorID, kind, level, summary string, detail map[string]any) error {
	if s == nil || s.db == nil {
		return nil
	}
	connectorID = strings.TrimSpace(connectorID)
	kind = strings.TrimSpace(kind)
	level = strings.TrimSpace(level)
	if connectorID == "" || kind == "" || summary == "" {
		return fmt.Errorf("connector activity: connector_id, kind, and summary are required")
	}
	if level == "" {
		level = "info"
	}
	if detail == nil {
		detail = map[string]any{}
	}
	raw, err := json.Marshal(detail)
	if err != nil {
		return err
	}
	_, err = s.db.ExecContext(ctx, `INSERT INTO connector_activity (connector_id, kind, level, summary, detail)
		VALUES ($1, $2, $3, $4, $5::jsonb)`,
		connectorID, kind, level, summary, string(raw))
	return err
}

func (s *ActivityStore) List(ctx context.Context, connectorID string, limit, offset int) ([]model.ConnectorActivityEntry, int, error) {
	if s == nil || s.db == nil {
		return nil, 0, nil
	}
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}
	connectorID = strings.TrimSpace(connectorID)

	var total int
	countSQL := `SELECT COUNT(*) FROM connector_activity`
	listSQL := `SELECT id, connector_id, kind, level, summary, detail, created_at FROM connector_activity`
	args := []any{}
	where := ""
	if connectorID != "" {
		where = ` WHERE connector_id = $1`
		args = append(args, connectorID)
	}
	if err := s.db.QueryRowContext(ctx, countSQL+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listArgs := append([]any{}, args...)
	n := len(listArgs)
	listSQL += where + fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, n+1, n+2)
	listArgs = append(listArgs, limit, offset)

	rows, err := s.db.QueryContext(ctx, listSQL, listArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []model.ConnectorActivityEntry
	for rows.Next() {
		var (
			id          int64
			cid         string
			kind        string
			level       string
			summary     string
			detailRaw   []byte
			created     time.Time
		)
		if err := rows.Scan(&id, &cid, &kind, &level, &summary, &detailRaw, &created); err != nil {
			return nil, 0, err
		}
		detail := map[string]any{}
		if len(detailRaw) > 0 {
			_ = json.Unmarshal(detailRaw, &detail)
		}
		out = append(out, model.ConnectorActivityEntry{
			ID:          id,
			ConnectorID: cid,
			Kind:        kind,
			Level:       level,
			Summary:     summary,
			Detail:      detail,
			CreatedAt:   created.UTC().Format(time.RFC3339),
		})
	}
	return out, total, rows.Err()
}
