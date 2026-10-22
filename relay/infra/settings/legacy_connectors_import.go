package settings

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"relay/domain/connectorconfig"
)

const legacyConnectorsFileName = "relay_connectors.json"

type legacyConnectorsFile struct {
	Version    int                             `json:"version"`
	Enabled    []string                        `json:"enabled"`
	Connectors map[string]legacyConnectorEntry `json:"connectors"`
}

type legacyConnectorEntry struct {
	NonSecret map[string]any `json:"non_secret,omitempty"`
}

// ImportLegacyConnectorFile copies data from relay_connectors.json into Postgres when a connector has no meta row yet.
func ImportLegacyConnectorFile(ctx context.Context, repo connectorconfig.Repository, dataDir string) error {
	if repo == nil || strings.TrimSpace(dataDir) == "" {
		return nil
	}
	path := filepath.Join(dataDir, "settings", legacyConnectorsFileName)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var file legacyConnectorsFile
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	enabledSet := map[string]bool{}
	for _, e := range file.Enabled {
		enabledSet[strings.TrimSpace(e)] = true
	}
	ids := map[string]struct{}{}
	for id := range file.Connectors {
		ids[strings.TrimSpace(id)] = struct{}{}
	}
	for id := range enabledSet {
		ids[id] = struct{}{}
	}
	for id := range ids {
		if id == "" {
			continue
		}
		_, exists, err := repo.Get(ctx, id)
		if err != nil || exists {
			continue
		}
		if ent, ok := file.Connectors[id]; ok && len(ent.NonSecret) > 0 {
			if err := repo.MergeNonSecret(ctx, id, ent.NonSecret); err != nil {
				return err
			}
		}
		if enabledSet[id] {
			if err := repo.SetEnabled(ctx, id, true); err != nil {
				return err
			}
		}
	}
	return nil
}
