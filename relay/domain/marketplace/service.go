package marketplace

import (
	"context"
	"fmt"
	"strings"

	"relay/domain/config"
	"relay/infra/settings"
	"relay/infra/store"
)

type Tile struct {
	ID           string        `json:"id"`
	Name         string        `json:"name"`
	Detail       string        `json:"detail"`
	Icon         string        `json:"icon"`
	Capabilities []string      `json:"capabilities"`
	Connectable  bool          `json:"connectable"`
	ConfigSchema []ConfigField `json:"config_schema"`
	Enabled      bool          `json:"enabled"`
	Connected    bool          `json:"connected"`
}

type ConfigView struct {
	ConnectorID string            `json:"connector_id"`
	Enabled     bool              `json:"enabled"`
	Connected   bool              `json:"connected"`
	NonSecret   map[string]any    `json:"non_secret"`
	Secrets     map[string]string `json:"secrets"`
}

type Service struct {
	Runtime  *config.Runtime
	Settings *settings.RelayConnectors
	Secrets  *store.SecretStore
}

func (s *Service) List(ctx context.Context) ([]Tile, error) {
	var tiles []Tile
	for _, ent := range allCatalogEntries() {
		enabled := s.Settings != nil && s.Settings.IsEnabled(ent.ID)
		connected := false
		if s.Runtime != nil {
			connected = ent.Connectable && s.Runtime.Connected(ctx, ent.ID, ent.SecretKeys)
		}
		tiles = append(tiles, Tile{
			ID:           ent.ID,
			Name:         ent.Name,
			Detail:       ent.Detail,
			Icon:         ent.Icon,
			Capabilities: ent.Capabilities,
			Connectable:  ent.Connectable,
			ConfigSchema: ent.ConfigSchema,
			Enabled:      enabled,
			Connected:    connected,
		})
	}
	return tiles, nil
}

func (s *Service) GetConfig(ctx context.Context, id string) (*ConfigView, error) {
	ent, ok := EntryByID(id)
	if !ok {
		return nil, fmt.Errorf("unknown connector")
	}
	enabled := s.Settings != nil && s.Settings.IsEnabled(id)
	nonSecret := map[string]any{}
	if s.Settings != nil {
		nonSecret = s.Settings.NonSecret(id)
	}
	masked := map[string]string{}
	if s.Secrets != nil {
		masked = s.Secrets.ListMasked(ctx, id, ent.SecretKeys)
	}
	connected := false
	if s.Runtime != nil {
		connected = ent.Connectable && s.Runtime.Connected(ctx, id, ent.SecretKeys)
	}
	return &ConfigView{
		ConnectorID: id,
		Enabled:     enabled,
		Connected:   connected,
		NonSecret:   nonSecret,
		Secrets:     masked,
	}, nil
}

type PutConfigRequest struct {
	Secrets   map[string]string `json:"secrets"`
	NonSecret map[string]any    `json:"non_secret"`
	Enable    *bool             `json:"enable"`
}

func (s *Service) PutConfig(ctx context.Context, id string, req PutConfigRequest) (*ConfigView, error) {
	ent, ok := EntryByID(id)
	if !ok {
		return nil, fmt.Errorf("unknown connector")
	}
	if !ent.Connectable {
		return nil, fmt.Errorf("connector is not connectable yet")
	}
	if req.Secrets != nil && s.Secrets != nil {
		filtered := map[string]string{}
		for k, v := range req.Secrets {
			k = strings.TrimSpace(k)
			if k == "" {
				continue
			}
			filtered[k] = strings.TrimSpace(v)
		}
		if err := s.Secrets.SetSecrets(ctx, id, filtered); err != nil {
			return nil, err
		}
	}
	if req.NonSecret != nil && s.Settings != nil {
		if err := s.Settings.MergeNonSecret(id, req.NonSecret); err != nil {
			return nil, err
		}
	}
	if req.Enable != nil && s.Settings != nil {
		if err := s.Settings.SetEnabled(id, *req.Enable); err != nil {
			return nil, err
		}
	}
	return s.GetConfig(ctx, id)
}

func (s *Service) Disable(ctx context.Context, id string, clearSecrets bool) (*ConfigView, error) {
	if _, ok := EntryByID(id); !ok {
		return nil, fmt.Errorf("unknown connector")
	}
	if s.Settings != nil {
		if err := s.Settings.SetEnabled(id, false); err != nil {
			return nil, err
		}
	}
	if clearSecrets && s.Secrets != nil {
		if err := s.Secrets.DeleteConnector(ctx, id); err != nil {
			return nil, err
		}
	}
	return s.GetConfig(ctx, id)
}
