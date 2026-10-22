package repository

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"coordinator/model"
)

type externalEventTriageFile struct {
	Entries map[string]model.ExternalEventTriage `json:"entries"`
}

func (r *FileRepository) externalEventTriagePath() string {
	return filepath.Join(r.dataDir(), ".cache", "external_event_triage.json")
}

func externalEventTriageKey(source, externalID string) string {
	return strings.TrimSpace(source) + "\x00" + strings.TrimSpace(externalID)
}

func (r *FileRepository) loadExternalEventTriageFile() (externalEventTriageFile, error) {
	path := r.externalEventTriagePath()
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return externalEventTriageFile{Entries: map[string]model.ExternalEventTriage{}}, nil
		}
		return externalEventTriageFile{}, err
	}
	var f externalEventTriageFile
	if err := json.Unmarshal(raw, &f); err != nil {
		return externalEventTriageFile{}, err
	}
	if f.Entries == nil {
		f.Entries = map[string]model.ExternalEventTriage{}
	}
	return f, nil
}

var triageMu sync.Mutex

func (r *FileRepository) GetExternalEventTriage(source, externalID string) (*model.ExternalEventTriage, bool) {
	triageMu.Lock()
	defer triageMu.Unlock()
	f, err := r.loadExternalEventTriageFile()
	if err != nil {
		return nil, false
	}
	key := externalEventTriageKey(source, externalID)
	row, ok := f.Entries[key]
	if !ok {
		return nil, false
	}
	return &row, true
}

func (r *FileRepository) ListExternalEventTriage() map[string]model.ExternalEventTriage {
	triageMu.Lock()
	defer triageMu.Unlock()
	f, err := r.loadExternalEventTriageFile()
	if err != nil {
		return map[string]model.ExternalEventTriage{}
	}
	return f.Entries
}

func (r *FileRepository) UpsertExternalEventTriage(row model.ExternalEventTriage) error {
	triageMu.Lock()
	defer triageMu.Unlock()
	f, err := r.loadExternalEventTriageFile()
	if err != nil {
		return err
	}
	key := externalEventTriageKey(row.Source, row.ExternalID)
	if key == "\x00" || strings.TrimSpace(row.Source) == "" {
		return os.ErrInvalid
	}
	now := time.Now().UTC().Format(time.RFC3339)
	row.UpdatedAt = now
	f.Entries[key] = row
	path := r.externalEventTriagePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, raw, 0o644)
}
