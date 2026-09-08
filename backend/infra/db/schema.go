package db

import (
	"fmt"
)

func (s *Store) initSchema() error {
	if _, err := s.sql.Exec(`
CREATE TABLE IF NOT EXISTS schema_meta (
  key TEXT PRIMARY KEY,
  value TEXT NOT NULL
);`); err != nil {
		return fmt.Errorf("sqlite schema_meta: %w", err)
	}

	var ver string
	err := s.sql.QueryRow(`SELECT value FROM schema_meta WHERE key = 'version'`).Scan(&ver)
	if err != nil || ver != schemaVersion {
		if _, dropErr := s.sql.Exec(`DROP TABLE IF EXISTS events; DROP TABLE IF EXISTS ingest_files;`); dropErr != nil {
			return fmt.Errorf("sqlite drop stale cache: %w", dropErr)
		}
	}

	const ddl = `
CREATE TABLE IF NOT EXISTS ingest_files (
  path TEXT PRIMARY KEY,
  mtime INTEGER NOT NULL,
  size INTEGER NOT NULL,
  lines INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY,
  timestamp INTEGER NOT NULL,
  event TEXT NOT NULL,
  task_id TEXT NOT NULL DEFAULT '',
  branch TEXT NOT NULL DEFAULT '',
  alias TEXT NOT NULL DEFAULT '',
  repo TEXT NOT NULL DEFAULT '',
  service TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT '',
  cost_usd REAL,
  budget_usd REAL,
  ondemand_usd REAL,
  cursor_models_pct REAL,
  other_models_pct REAL,
  usage_plan TEXT NOT NULL DEFAULT '',
  plan_price_usd REAL,
  spend_kind TEXT NOT NULL DEFAULT '',
  source_file TEXT NOT NULL,
  source_line INTEGER NOT NULL,
  UNIQUE(source_file, source_line)
);
CREATE INDEX IF NOT EXISTS idx_events_alias ON events(alias);
CREATE INDEX IF NOT EXISTS idx_events_task ON events(task_id);
CREATE INDEX IF NOT EXISTS idx_events_type ON events(event);
CREATE INDEX IF NOT EXISTS idx_events_ts ON events(timestamp);
CREATE INDEX IF NOT EXISTS idx_events_service ON events(service);
`
	if _, err := s.sql.Exec(ddl); err != nil {
		return fmt.Errorf("sqlite schema: %w", err)
	}
	_, err = s.sql.Exec(`INSERT INTO schema_meta(key, value) VALUES('version', ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value`, schemaVersion)
	return err
}
