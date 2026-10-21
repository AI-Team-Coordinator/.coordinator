package crypto

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const configKeyFile = ".relay_config_key"
const keyLen = 32

// LoadConfigKey returns a 32-byte AES key from RELAY_CONFIG_KEY or dataDir/.relay_config_key.
func LoadConfigKey(dataDir string) ([]byte, error) {
	if v := strings.TrimSpace(os.Getenv("RELAY_CONFIG_KEY")); v != "" {
		raw, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("RELAY_CONFIG_KEY: invalid base64")
		}
		if len(raw) != keyLen {
			return nil, fmt.Errorf("RELAY_CONFIG_KEY: want %d bytes after decode", keyLen)
		}
		return raw, nil
	}
	path := filepath.Join(dataDir, configKeyFile)
	if data, err := os.ReadFile(path); err == nil {
		raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(data)))
		if err != nil || len(raw) != keyLen {
			return nil, fmt.Errorf("invalid %s", configKeyFile)
		}
		return raw, nil
	}
	key := make([]byte, keyLen)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	enc := base64.StdEncoding.EncodeToString(key)
	if err := os.WriteFile(path, []byte(enc+"\n"), 0o600); err != nil {
		return nil, err
	}
	return key, nil
}
