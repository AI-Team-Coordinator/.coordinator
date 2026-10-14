package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"coordinator/model"
)

func (r *FileRepository) relayAccessPath() string {
	return filepath.Join(r.dataDir(), ".relay_access")
}

func (r *FileRepository) GetRelayAccess(_ context.Context) (*model.RelayAccess, error) {
	data, err := os.ReadFile(r.relayAccessPath())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out model.RelayAccess
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out.URL == "" {
		return nil, nil
	}
	return &out, nil
}

func (r *FileRepository) SaveRelayAccess(_ context.Context, access *model.RelayAccess) error {
	if access == nil || access.URL == "" {
		return os.Remove(r.relayAccessPath())
	}
	data, err := json.Marshal(access)
	if err != nil {
		return err
	}
	return writeFileAtomic(r.relayAccessPath(), data)
}

func (r *FileRepository) ClearRelayAccess(_ context.Context) error {
	err := os.Remove(r.relayAccessPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
