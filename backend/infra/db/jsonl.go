package db

import (
	"bufio"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"coordinator/model"
)

type jsonlFile struct {
	Abs   string
	Rel   string
	Alias string
}

func (s *Store) syncLocked(ctx context.Context) error {
	files, err := listJSONL(s.eventsDirs)
	if err != nil {
		return err
	}

	seen := make(map[string]struct{}, len(files))
	for _, file := range files {
		seen[file.Rel] = struct{}{}
		info, statErr := os.Stat(file.Abs)
		if statErr != nil {
			continue
		}
		mtime := info.ModTime().UnixNano()
		size := info.Size()

		var prevMtime, prevSize int64
		err := s.sql.QueryRowContext(ctx, `SELECT mtime, size FROM ingest_files WHERE path = ?`, file.Rel).Scan(&prevMtime, &prevSize)
		if err == nil && prevMtime == mtime && prevSize == size {
			continue
		}

		if err := s.replaceFile(ctx, file, mtime, size); err != nil {
			return err
		}
	}

	rows, err := s.sql.QueryContext(ctx, `SELECT path FROM ingest_files`)
	if err != nil {
		return err
	}
	defer rows.Close()

	stale := make([]string, 0)
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return err
		}
		if _, ok := seen[path]; !ok {
			stale = append(stale, path)
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, path := range stale {
		if _, err := s.sql.ExecContext(ctx, `DELETE FROM events WHERE source_file = ?`, path); err != nil {
			return err
		}
		if _, err := s.sql.ExecContext(ctx, `DELETE FROM ingest_files WHERE path = ?`, path); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) replaceFile(ctx context.Context, file jsonlFile, mtime, size int64) error {
	tx, err := s.sql.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, `DELETE FROM events WHERE source_file = ?`, file.Rel); err != nil {
		return err
	}

	fh, err := os.Open(file.Abs)
	if err != nil {
		return err
	}
	defer fh.Close()

	insert, err := tx.PrepareContext(ctx, `
INSERT INTO events(timestamp, event, task_id, branch, alias, repo, service, status, cost_usd, budget_usd, ondemand_usd, cursor_models_pct, other_models_pct, usage_plan, plan_price_usd, spend_kind, source_file, source_line)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer insert.Close()

	scanner := bufio.NewScanner(fh)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var ev model.Event
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		if ev.Alias == "" {
			ev.Alias = file.Alias
		}
		if _, err := insert.ExecContext(ctx, ev.Timestamp, ev.Event, ev.TaskID, ev.Branch, ev.Alias, ev.Repo, ev.Service, ev.Status,
			floatPtrValue(ev.CostUSD), floatPtrValue(ev.BudgetUSD), floatPtrValue(ev.OnDemandUSD), floatPtrValue(ev.CursorModelsPct), floatPtrValue(ev.OtherModelsPct), ev.UsagePlan, floatPtrValue(ev.PlanPriceUSD), ev.SpendKind,
			file.Rel, lineNo); err != nil {
			return err
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	if _, err := tx.ExecContext(ctx, `
INSERT INTO ingest_files(path, mtime, size, lines) VALUES (?, ?, ?, ?)
ON CONFLICT(path) DO UPDATE SET mtime=excluded.mtime, size=excluded.size, lines=excluded.lines`,
		file.Rel, mtime, size, lineNo); err != nil {
		return err
	}
	return tx.Commit()
}

func listJSONL(dirs []string) ([]jsonlFile, error) {
	nowYear := time.Now().Year()
	minYear := nowYear - hotYearSpan + 1
	out := make([]jsonlFile, 0)
	seen := make(map[string]struct{})

	for _, dir := range dirs {
		if strings.TrimSpace(dir) == "" {
			continue
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() {
				subEntries, err := os.ReadDir(filepath.Join(dir, name))
				if err != nil {
					continue
				}
				for _, sub := range subEntries {
					if sub.IsDir() || !strings.HasSuffix(sub.Name(), ".jsonl") {
						continue
					}
					if year, ok := yearFromName(sub.Name()); ok && year < minYear {
						continue
					}
					abs := filepath.Join(dir, name, sub.Name())
					rel := filepath.ToSlash(filepath.Join(name, sub.Name()))
					if _, dup := seen[rel]; dup {
						continue
					}
					seen[rel] = struct{}{}
					out = append(out, jsonlFile{Abs: abs, Rel: rel, Alias: name})
				}
				continue
			}
			if !strings.HasSuffix(name, ".jsonl") {
				continue
			}
			abs := filepath.Join(dir, name)
			rel := name
			if _, dup := seen[rel]; dup {
				continue
			}
			seen[rel] = struct{}{}
			out = append(out, jsonlFile{
				Abs:   abs,
				Rel:   rel,
				Alias: strings.TrimSuffix(name, ".jsonl"),
			})
		}
	}
	return out, nil
}

func yearFromName(name string) (int, bool) {
	base := strings.TrimSuffix(name, ".jsonl")
	if len(base) != 4 {
		return 0, false
	}
	year, err := strconv.Atoi(base)
	if err != nil {
		return 0, false
	}
	return year, true
}
