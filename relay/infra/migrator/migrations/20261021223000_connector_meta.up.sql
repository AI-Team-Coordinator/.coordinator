-- Durable connector enablement + non-secret fields (survives RELAY_DATA file loss on redeploy).

CREATE TABLE connector_meta (
    connector_id TEXT PRIMARY KEY,
    enabled BOOLEAN NOT NULL DEFAULT false,
    non_secret JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
