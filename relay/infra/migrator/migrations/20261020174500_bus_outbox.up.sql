CREATE TABLE bus_outbox (
    id BIGSERIAL PRIMARY KEY,
    event_id TEXT NOT NULL,
    payload JSONB NOT NULL,
    status TEXT NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    next_attempt_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT bus_outbox_event_id_key UNIQUE (event_id)
);

CREATE INDEX idx_bus_outbox_pending ON bus_outbox (status, next_attempt_at)
    WHERE status IN ('pending', 'failed');
