-- Relay DB: connector secrets (encrypted blobs), event store, daily stats rollup.

CREATE TABLE connector_secrets (
    connector_id TEXT NOT NULL,
    field_key TEXT NOT NULL,
    ciphertext BYTEA NOT NULL,
    nonce BYTEA NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (connector_id, field_key)
);

CREATE TABLE external_events (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL,
    source TEXT NOT NULL,
    external_id TEXT,
    class TEXT NOT NULL,
    kind TEXT NOT NULL,
    title TEXT,
    summary TEXT,
    url TEXT,
    occurred_at TIMESTAMPTZ,
    received_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    payload JSONB NOT NULL DEFAULT '{}'::jsonb,
    tags TEXT[] NOT NULL DEFAULT '{}',
    CONSTRAINT external_events_event_id_key UNIQUE (event_id)
);

CREATE INDEX idx_external_events_source_received ON external_events (source, received_at DESC);
CREATE INDEX idx_external_events_kind_received ON external_events (kind, received_at DESC);
CREATE INDEX idx_external_events_class_received ON external_events (class, received_at DESC);

CREATE TABLE connector_daily_stats (
    day DATE NOT NULL,
    connector_id TEXT NOT NULL,
    events_total INT NOT NULL DEFAULT 0,
    by_class JSONB NOT NULL DEFAULT '{}'::jsonb,
    by_kind JSONB NOT NULL DEFAULT '{}'::jsonb,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (day, connector_id)
);
