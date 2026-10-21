package repository

import (
	"bufio"
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// ExternalEventRow is one normalized external fact from progress/external jsonl.
type ExternalEventRow struct {
	SchemaVersion int            `json:"schema_version"`
	ID            string         `json:"id"`
	Source        string         `json:"source"`
	ExternalID    string         `json:"external_id"`
	ReceivedAt    string         `json:"received_at,omitempty"`
	OccurredAt    string         `json:"occurred_at,omitempty"`
	Class         string         `json:"class"`
	Kind          string         `json:"kind"`
	Title         string         `json:"title,omitempty"`
	Summary       string         `json:"summary,omitempty"`
	URL           string         `json:"url,omitempty"`
	Tags          []string       `json:"tags,omitempty"`
	Data          map[string]any `json:"data,omitempty"`
}

const pulseExternalEventCap = 100

// LoadExternalEvents reads jsonl under data/progress/external/events, newest first (Pulse cap).
func (r *FileRepository) LoadExternalEvents() ([]ExternalEventRow, error) {
	all, err := r.loadExternalEventsSorted()
	if err != nil {
		return nil, err
	}
	if len(all) > pulseExternalEventCap {
		all = all[:pulseExternalEventCap]
	}
	return all, nil
}

// ListExternalEvents returns a page filtered by source (empty = all). sources lists distinct connector ids.
func (r *FileRepository) ListExternalEvents(source string, limit, offset int) ([]ExternalEventRow, int, []string, error) {
	all, err := r.loadExternalEventsSorted()
	if err != nil {
		return nil, 0, nil, err
	}
	sources := distinctSources(all)
	source = strings.TrimSpace(source)
	filtered := all
	if source != "" {
		filtered = make([]ExternalEventRow, 0, len(all))
		for _, ev := range all {
			if strings.EqualFold(ev.Source, source) {
				filtered = append(filtered, ev)
			}
		}
	}
	total := len(filtered)
	if offset < 0 {
		offset = 0
	}
	if offset > total {
		offset = total
	}
	if limit < 1 {
		limit = 20
	}
	end := offset + limit
	if end > total {
		end = total
	}
	page := filtered[offset:end]
	return page, total, sources, nil
}

func (r *FileRepository) loadExternalEventsSorted() ([]ExternalEventRow, error) {
	root := filepath.Join(r.dataPath, "progress", "external", "events")
	if !externalEventsRootExists(root) {
		return nil, nil
	}
	byKey := make(map[string]ExternalEventRow)
	var paths []string
	if err := walkJSONL(root, &paths); err != nil {
		return nil, err
	}
	sort.Strings(paths)
	for _, path := range paths {
		if err := readJSONLLines(path, byKey); err != nil {
			return nil, err
		}
	}
	out := make([]ExternalEventRow, 0, len(byKey))
	for _, ev := range byKey {
		out = append(out, ev)
	}
	sort.Slice(out, func(i, j int) bool {
		return eventTime(out[i]).After(eventTime(out[j]))
	})
	return out, nil
}

func distinctSources(events []ExternalEventRow) []string {
	seen := make(map[string]struct{})
	for _, ev := range events {
		s := strings.TrimSpace(ev.Source)
		if s == "" {
			continue
		}
		seen[s] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func externalEventsRootExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func walkJSONL(root string, out *[]string) error {
	return filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if strings.HasSuffix(d.Name(), ".jsonl") {
			*out = append(*out, path)
		}
		return nil
	})
}

func readJSONLLines(path string, byKey map[string]ExternalEventRow) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 2<<20)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		var ev ExternalEventRow
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		key := strings.TrimSpace(ev.Source) + "\x00" + strings.TrimSpace(ev.ExternalID)
		if key == "\x00" {
			key = strings.TrimSpace(ev.ID)
		}
		byKey[key] = ev
	}
	return sc.Err()
}

func eventTime(ev ExternalEventRow) time.Time {
	for _, raw := range []string{ev.ReceivedAt, ev.OccurredAt} {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		if t, err := time.Parse(time.RFC3339, raw); err == nil {
			return t
		}
	}
	return time.Time{}
}
