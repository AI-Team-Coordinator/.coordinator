package repository

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"coordinator/model"
)

func (r *FileRepository) relayIssuedTokensPath() string {
	return filepath.Join(r.dataDir(), ".relay_issued_tokens.json")
}

func (r *FileRepository) GetRelayIssuedTokens(_ context.Context) (*model.RelayIssuedTokenFile, error) {
	path := r.relayIssuedTokensPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &model.RelayIssuedTokenFile{Version: 1, Tokens: nil}, nil
		}
		return nil, err
	}
	var out model.RelayIssuedTokenFile
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	if out.Version == 0 {
		out.Version = 1
	}
	return &out, nil
}

func (r *FileRepository) SaveRelayIssuedTokens(_ context.Context, file *model.RelayIssuedTokenFile) error {
	if file == nil {
		file = &model.RelayIssuedTokenFile{Version: 1}
	}
	if file.Version == 0 {
		file.Version = 1
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	return writeFileAtomic(r.relayIssuedTokensPath(), data)
}
