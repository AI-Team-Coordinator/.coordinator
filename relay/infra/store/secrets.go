package store

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"fmt"
	"io"
	"strings"
	"sync"
)

// SecretStore persists per-connector secret fields encrypted at rest (Postgres).
type SecretStore struct {
	db   *sql.DB
	aead cipher.AEAD
	mu   sync.RWMutex
}

func NewSecretStore(db *sql.DB, key []byte) (*SecretStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database is nil")
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("config key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &SecretStore{db: db, aead: aead}, nil
}

func (s *SecretStore) Close() error {
	return nil
}

func (s *SecretStore) SetSecrets(ctx context.Context, connectorID string, secrets map[string]string) error {
	if s == nil {
		return fmt.Errorf("secret store not configured")
	}
	connectorID = strings.TrimSpace(connectorID)
	if connectorID == "" {
		return fmt.Errorf("connector id is required")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range secrets {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		v = strings.TrimSpace(v)
		if v == "" {
			_, err := s.db.ExecContext(ctx, `DELETE FROM connector_secrets WHERE connector_id = $1 AND field_key = $2`, connectorID, k)
			return err
		}
		nonce := make([]byte, s.aead.NonceSize())
		if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
			return err
		}
		ct := s.aead.Seal(nil, nonce, []byte(v), nil)
		_, err := s.db.ExecContext(ctx, `INSERT INTO connector_secrets (connector_id, field_key, ciphertext, nonce, updated_at)
			VALUES ($1, $2, $3, $4, now())
			ON CONFLICT (connector_id, field_key) DO UPDATE SET ciphertext = EXCLUDED.ciphertext, nonce = EXCLUDED.nonce, updated_at = now()`,
			connectorID, k, ct, nonce)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *SecretStore) GetSecret(ctx context.Context, connectorID, fieldKey string) (string, bool) {
	if s == nil {
		return "", false
	}
	connectorID = strings.TrimSpace(connectorID)
	fieldKey = strings.TrimSpace(fieldKey)
	var ct, nonce []byte
	err := s.db.QueryRowContext(ctx, `SELECT ciphertext, nonce FROM connector_secrets WHERE connector_id = $1 AND field_key = $2`,
		connectorID, fieldKey).Scan(&ct, &nonce)
	if err != nil {
		return "", false
	}
	plain, err := s.aead.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", false
	}
	return string(plain), true
}

func (s *SecretStore) ListMasked(ctx context.Context, connectorID string, keys []string) map[string]string {
	out := make(map[string]string, len(keys))
	for _, k := range keys {
		if v, ok := s.GetSecret(ctx, connectorID, k); ok && v != "" {
			out[k] = "••••"
		}
	}
	return out
}

func (s *SecretStore) DeleteConnector(ctx context.Context, connectorID string) error {
	if s == nil {
		return nil
	}
	_, err := s.db.ExecContext(ctx, `DELETE FROM connector_secrets WHERE connector_id = $1`, strings.TrimSpace(connectorID))
	return err
}

func (s *SecretStore) HasAnySecret(ctx context.Context, connectorID string, keys []string) bool {
	for _, k := range keys {
		if v, ok := s.GetSecret(ctx, connectorID, k); ok && strings.TrimSpace(v) != "" {
			return true
		}
	}
	return false
}
