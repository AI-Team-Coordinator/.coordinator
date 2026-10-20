package bus

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"relay/model"
)

// JSONL appends ExternalEvent lines under progress/external/events/<origin>/<year>.jsonl.
type JSONL struct {
	Root string
}

func (j JSONL) Append(_ context.Context, ev model.ExternalEvent) error {
	year := time.Now().UTC().Format("2006")
	dir := filepath.Join(j.Root, "progress", "external", "events", ev.OriginID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	raw, err := json.Marshal(ev)
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	path := filepath.Join(dir, year+".jsonl")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(raw)
	return err
}
