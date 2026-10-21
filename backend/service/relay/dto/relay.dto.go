package relaydto

type RelayAccessResponse struct {
	Connected bool   `json:"connected"`
	URL       string `json:"url,omitempty"`
	Scope     string `json:"scope,omitempty"`
}

type RelayConnectRequest struct {
	URL   string `json:"url"`
	Token string `json:"token"`
}

type RelaySourceResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Detail  string `json:"detail"`
	Enabled bool   `json:"enabled"`
}

type RelaySourcesResponse struct {
	Sources []RelaySourceResponse `json:"sources"`
}

type RelayConfigField struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Type     string `json:"type"`
	Required bool   `json:"required"`
	Help     string `json:"help,omitempty"`
}

type RelayActionMeta struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type RelayMarketplaceTile struct {
	ID           string             `json:"id"`
	Name         string             `json:"name"`
	Detail       string             `json:"detail"`
	Icon         string             `json:"icon"`
	Capabilities []string           `json:"capabilities"`
	Actions      []RelayActionMeta  `json:"actions,omitempty"`
	Connectable  bool               `json:"connectable"`
	ConfigSchema []RelayConfigField `json:"config_schema"`
	Enabled      bool               `json:"enabled"`
	Connected    bool               `json:"connected"`
}

type RelayMarketplaceResponse struct {
	Connectors []RelayMarketplaceTile `json:"connectors"`
}

type RelayConnectorConfigResponse struct {
	ConnectorID string            `json:"connector_id"`
	Enabled     bool              `json:"enabled"`
	Connected   bool              `json:"connected"`
	NonSecret   map[string]any    `json:"non_secret"`
	Secrets     map[string]string `json:"secrets"`
}

type RelayPutConnectorConfigRequest struct {
	Secrets   map[string]string `json:"secrets"`
	NonSecret map[string]any    `json:"non_secret"`
	Enable    *bool             `json:"enable"`
}

type RelayConnectorActivityEntry struct {
	ID          int64          `json:"id"`
	ConnectorID string         `json:"connector_id"`
	Kind        string         `json:"kind"`
	Level       string         `json:"level"`
	Summary     string         `json:"summary"`
	Detail      map[string]any `json:"detail,omitempty"`
	CreatedAt   string         `json:"created_at"`
}

type RelayConnectorActivityResponse struct {
	Entries []RelayConnectorActivityEntry `json:"entries"`
	Total   int                           `json:"total"`
	Limit   int                           `json:"limit"`
	Offset  int                           `json:"offset"`
}
