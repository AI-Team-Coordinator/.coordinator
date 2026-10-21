package github

import "relay/domain/marketplace"

func catalogEntry() marketplace.CatalogEntry {
	return marketplace.CatalogEntry{
		ID:           SourceID,
		Name:         "GitHub",
		Detail:       "Pull requests, reviews, and Actions.",
		Icon:         "github",
		Capabilities: []string{"webhook"},
		Connectable:  true,
		SecretKeys:   []string{"webhook_secret"},
		ConfigSchema: []marketplace.ConfigField{
			{
				ID:       "webhook_secret",
				Label:    "Webhook secret",
				Type:     "secret",
				Required: true,
				Help:     "Same value as in the GitHub webhook settings. Relay URL: POST /connectors/github/webhook",
			},
		},
	}
}
