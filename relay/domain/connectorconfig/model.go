package connectorconfig

// Meta is durable connector state stored on the relay instance (Postgres).
// Secrets live in connector_secrets; this holds enablement and non-secret fields only.
type Meta struct {
	ConnectorID string
	Enabled     bool
	NonSecret   map[string]any
}
