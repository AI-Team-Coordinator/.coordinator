package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"coordinator/model"

	_ "modernc.org/sqlite"
)

const (
	schemaVersion = "6"
	hotYearSpan   = 3
	dbFileName    = "coordinator.sqlite"
)

// Store is a local SQLite cache of coordinator events. jsonl on disk remains
// the git source of truth; this file is gitignored and rebuilt when missing.
type Store struct {
	sql        *sql.DB
	eventsDirs []string
	mu         sync.Mutex
}

func Open(cacheDir string, eventsDirs ...string) (*Store, error) {
	if err := os.MkdirAll(cacheDir, 0o755); err != nil {
		return nil, fmt.Errorf("create cache dir: %w", err)
	}

	dsn := filepath.Join(cacheDir, dbFileName)
	conn, err := sql.Open("sqlite", dsn+"?_pragma=busy_timeout(5000)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	conn.SetMaxOpenConns(1)

	store := &Store{sql: conn, eventsDirs: eventsDirs}
	if err := store.initSchema(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return store, nil
}

func (s *Store) Close() error {
	if s == nil || s.sql == nil {
		return nil
	}
	return s.sql.Close()
}

func (s *Store) Sync(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.syncLocked(ctx)
}

func (s *Store) List(ctx context.Context, q model.EventQuery) ([]model.Event, int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.syncLocked(ctx); err != nil {
		return nil, 0, err
	}

	where, args := eventFilters(q)

	countQ := "SELECT COUNT(*) FROM events" + where
	var total int
	if err := s.sql.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	listQ := "SELECT timestamp, event, task_id, branch, alias, repo, service, status, cost_usd, budget_usd, ondemand_usd, cursor_models_pct, other_models_pct, usage_plan, plan_price_usd, spend_kind, summary, findings, active_seconds FROM events" + where + " ORDER BY timestamp DESC"
	if q.Limit > 0 {
		listQ += " LIMIT ?"
		args = append(args, q.Limit)
		if q.Offset > 0 {
			listQ += " OFFSET ?"
			args = append(args, q.Offset)
		}
	}

	rows, err := s.sql.QueryContext(ctx, listQ, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := make([]model.Event, 0)
	for rows.Next() {
		var ev model.Event
		var cost, budget, ondemand, cursorPct, otherPct, planPrice sql.NullFloat64
		var active sql.NullInt64
		if err := rows.Scan(
			&ev.Timestamp, &ev.Event, &ev.TaskID, &ev.Branch, &ev.Alias, &ev.Repo, &ev.Service, &ev.Status,
			&cost, &budget, &ondemand, &cursorPct, &otherPct, &ev.UsagePlan, &planPrice, &ev.SpendKind, &ev.Summary, &ev.Findings, &active,
		); err != nil {
			return nil, 0, err
		}
		ev.CostUSD = nullFloatPtr(cost)
		ev.BudgetUSD = nullFloatPtr(budget)
		ev.OnDemandUSD = nullFloatPtr(ondemand)
		ev.CursorModelsPct = nullFloatPtr(cursorPct)
		ev.OtherModelsPct = nullFloatPtr(otherPct)
		ev.PlanPriceUSD = nullFloatPtr(planPrice)
		ev.ActiveSeconds = nullInt64Ptr(active)
		items = append(items, ev)
	}
	return items, total, rows.Err()
}

func eventFilters(q model.EventQuery) (string, []any) {
	clauses := make([]string, 0, 4)
	args := make([]any, 0, 4)
	if q.Alias != "" {
		clauses = append(clauses, "alias = ?")
		args = append(args, q.Alias)
	}
	if q.Event != "" {
		clauses = append(clauses, "event = ?")
		args = append(args, q.Event)
	}
	if q.Service != "" {
		clauses = append(clauses, "service = ?")
		args = append(args, q.Service)
	}
	if q.TaskID != "" {
		clauses = append(clauses, "task_id = ?")
		args = append(args, q.TaskID)
	}
	if q.Since > 0 {
		clauses = append(clauses, "timestamp >= ?")
		args = append(args, q.Since)
	}
	if len(clauses) == 0 {
		return "", args
	}
	sql := " WHERE " + clauses[0]
	for i := 1; i < len(clauses); i++ {
		sql += " AND " + clauses[i]
	}
	return sql, args
}

func nullFloatPtr(v sql.NullFloat64) *float64 {
	if !v.Valid {
		return nil
	}
	n := v.Float64
	return &n
}

func floatPtrValue(v *float64) any {
	if v == nil {
		return nil
	}
	return *v
}

func int64PtrValue(v *int64) any {
	if v == nil {
		return nil
	}
	return *v
}

func nullInt64Ptr(v sql.NullInt64) *int64 {
	if !v.Valid {
		return nil
	}
	n := v.Int64
	return &n
}
