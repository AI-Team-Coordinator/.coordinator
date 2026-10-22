package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"relay/domain/connectorconfig"
)

// ConnectorConfigRepository implements connectorconfig.Repository in Postgres.
type ConnectorConfigRepository struct {
	db *sql.DB
}

func NewConnectorConfigRepository(db *sql.DB) *ConnectorConfigRepository {
	if db == nil {
		return nil
	}
	return &ConnectorConfigRepository{db: db}
}

func (r *ConnectorConfigRepository) Get(ctx context.Context, connectorID string) (connectorconfig.Meta, bool, error) {
	if r == nil || r.db == nil {
		return connectorconfig.Meta{}, false, fmt.Errorf("connector config repository is not configured")
	}
	connectorID = strings.TrimSpace(connectorID)
	if connectorID == "" {
		return connectorconfig.Meta{}, false, fmt.Errorf("connector id is required")
	}
	var enabled bool
	var raw []byte
	err := r.db.QueryRowContext(ctx, `SELECT enabled, non_secret FROM connector_meta WHERE connector_id = $1`, connectorID).
		Scan(&enabled, &raw)
	if err == sql.ErrNoRows {
		return connectorconfig.Meta{}, false, nil
	}
	if err != nil {
		return connectorconfig.Meta{}, false, err
	}
	ns := map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &ns); err != nil {
			return connectorconfig.Meta{}, false, err
		}
	}
	return connectorconfig.Meta{
		ConnectorID: connectorID,
		Enabled:     enabled,
		NonSecret:   ns,
	}, true, nil
}

func (r *ConnectorConfigRepository) SetEnabled(ctx context.Context, connectorID string, on bool) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("connector config repository is not configured")
	}
	connectorID = strings.TrimSpace(connectorID)
	if connectorID == "" {
		return fmt.Errorf("connector id is required")
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO connector_meta (connector_id, enabled, non_secret, updated_at)
		VALUES ($1, $2, '{}'::jsonb, now())
		ON CONFLICT (connector_id) DO UPDATE SET enabled = EXCLUDED.enabled, updated_at = now()`,
		connectorID, on)
	return err
}

func (r *ConnectorConfigRepository) MergeNonSecret(ctx context.Context, connectorID string, patch map[string]any) error {
	if r == nil || r.db == nil {
		return fmt.Errorf("connector config repository is not configured")
	}
	connectorID = strings.TrimSpace(connectorID)
	if connectorID == "" {
		return fmt.Errorf("connector id is required")
	}
	if patch == nil {
		return nil
	}
	row, exists, err := r.Get(ctx, connectorID)
	if err != nil {
		return err
	}
	merged := map[string]any{}
	if exists && row.NonSecret != nil {
		for k, v := range row.NonSecret {
			merged[k] = v
		}
	}
	for k, v := range patch {
		merged[k] = v
	}
	raw, err := json.Marshal(merged)
	if err != nil {
		return err
	}
	enabled := exists && row.Enabled
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO connector_meta (connector_id, enabled, non_secret, updated_at)
		VALUES ($1, $2, $3::jsonb, now())
		ON CONFLICT (connector_id) DO UPDATE SET non_secret = EXCLUDED.non_secret, updated_at = now()`,
		connectorID, enabled, raw)
	return err
}
