package bus

import (
	"encoding/json"
	"time"

	"relay/model"
)

// MarshalLine seals the event and returns one JSONL line (including trailing newline).
func MarshalLine(ev model.ExternalEvent) ([]byte, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	ev.Seal(now)
	raw, err := json.Marshal(ev)
	if err != nil {
		return nil, err
	}
	raw = append(raw, '\n')
	return raw, nil
}
