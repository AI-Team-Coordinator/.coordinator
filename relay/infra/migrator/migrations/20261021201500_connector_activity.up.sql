-- Connector activity log (operational checks, not coordinator ExternalEvents).

CREATE TABLE connector_activity (
    id BIGSERIAL PRIMARY KEY,
    connector_id TEXT NOT NULL,
    kind TEXT NOT NULL,
    level TEXT NOT NULL DEFAULT 'info',
    summary TEXT NOT NULL,
    detail JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_connector_activity_connector_created
    ON connector_activity (connector_id, created_at DESC);

CREATE INDEX idx_connector_activity_created
    ON connector_activity (created_at DESC);
