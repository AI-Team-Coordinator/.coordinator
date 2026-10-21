package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const fileName = "relay_connectors.json"

// ConnectorsFile is git-safe enablement and non-secret fields (Common/data/settings).
type ConnectorsFile struct {
	Version    int                       `json:"version"`
	Enabled    []string                  `json:"enabled"`
	Connectors map[string]ConnectorEntry `json:"connectors"`
}

type ConnectorEntry struct {
	NonSecret map[string]any `json:"non_secret,omitempty"`
}

type RelayConnectors struct {
	dataDir string
	mu      sync.RWMutex
}

func NewRelayConnectors(dataDir string) *RelayConnectors {
	return &RelayConnectors{dataDir: dataDir}
}

func (r *RelayConnectors) path() string {
	return filepath.Join(r.dataDir, "settings", fileName)
}

func (r *RelayConnectors) Load() (ConnectorsFile, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	path := r.path()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ConnectorsFile{Version: 1, Enabled: nil, Connectors: map[string]ConnectorEntry{}}, nil
		}
		return ConnectorsFile{}, err
	}
	var out ConnectorsFile
	if err := json.Unmarshal(data, &out); err != nil {
		return ConnectorsFile{}, err
	}
	if out.Version == 0 {
		out.Version = 1
	}
	if out.Connectors == nil {
		out.Connectors = map[string]ConnectorEntry{}
	}
	return out, nil
}

func (r *RelayConnectors) Save(file ConnectorsFile) error {
	if file.Version == 0 {
		file.Version = 1
	}
	if file.Connectors == nil {
		file.Connectors = map[string]ConnectorEntry{}
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	r.mu.Lock()
	defer r.mu.Unlock()
	dir := filepath.Dir(r.path())
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(r.path(), data, 0o644)
}

func (r *RelayConnectors) IsEnabled(id string) bool {
	file, err := r.Load()
	if err != nil {
		return false
	}
	id = strings.TrimSpace(id)
	for _, e := range file.Enabled {
		if strings.TrimSpace(e) == id {
			return true
		}
	}
	return false
}

func (r *RelayConnectors) SetEnabled(id string, on bool) error {
	file, err := r.Load()
	if err != nil {
		return err
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return nil
	}
	next := make([]string, 0, len(file.Enabled)+1)
	found := false
	for _, e := range file.Enabled {
		e = strings.TrimSpace(e)
		if e == "" {
			continue
		}
		if e == id {
			found = true
			if on {
				next = append(next, id)
			}
			continue
		}
		next = append(next, e)
	}
	if on && !found {
		next = append(next, id)
	}
	file.Enabled = next
	return r.Save(file)
}

func (r *RelayConnectors) NonSecret(id string) map[string]any {
	file, err := r.Load()
	if err != nil {
		return nil
	}
	ent, ok := file.Connectors[id]
	if !ok || ent.NonSecret == nil {
		return map[string]any{}
	}
	return ent.NonSecret
}

func (r *RelayConnectors) MergeNonSecret(id string, patch map[string]any) error {
	file, err := r.Load()
	if err != nil {
		return err
	}
	ent := file.Connectors[id]
	if ent.NonSecret == nil {
		ent.NonSecret = map[string]any{}
	}
	for k, v := range patch {
		ent.NonSecret[k] = v
	}
	file.Connectors[id] = ent
	return r.Save(file)
}
