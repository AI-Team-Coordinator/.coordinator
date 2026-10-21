package github

import (
	"relay/app"
	"relay/domain/config"
	"relay/domain/connector"
)

// Options registers the stock GitHub webhook connector (pack / any instance).
func Options() []app.Option {
	return []app.Option{
		app.WithCatalogEntry(catalogEntry()),
		app.WithConnectorInit(func(runtime *config.Runtime, cfg app.Config) connector.Connector {
			return Adapter{
				WebhookSecret: cfg.GitHubWebhookSecret,
				Secret:        runtime.GitHubWebhookSecret,
			}
		}),
	}
}
