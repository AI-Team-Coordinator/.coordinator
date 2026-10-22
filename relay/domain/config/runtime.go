package config

import (
	"context"

	"relay/domain/connectorconfig"
	"relay/infra/settings"
	"relay/infra/store"
)

// Runtime merges connector config from Postgres with encrypted secrets.
type Runtime struct {
	Config   *connectorconfig.Service
	Secrets  *store.SecretStore
	Instance *settings.InstanceFiles
}

func (r *Runtime) Enabled(id string) bool {
	if r == nil || r.Config == nil {
		return false
	}
	return r.Config.IsEnabled(id)
}

func (r *Runtime) Secret(ctx context.Context, connectorID, field string) string {
	if r == nil || r.Secrets == nil {
		return ""
	}
	v, ok := r.Secrets.GetSecret(ctx, connectorID, field)
	if !ok {
		return ""
	}
	return v
}

func (r *Runtime) GitHubWebhookSecret(ctx context.Context) string {
	return r.Secret(ctx, "github", "webhook_secret")
}

func (r *Runtime) Connected(ctx context.Context, id string, secretKeys []string) bool {
	if !r.Enabled(id) {
		return false
	}
	if len(secretKeys) == 0 {
		return true
	}
	if r.Secrets == nil {
		return false
	}
	return r.Secrets.HasAnySecret(ctx, id, secretKeys)
}
