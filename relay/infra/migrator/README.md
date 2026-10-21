# Relay database migrations (MigratorV2)

Relay uses a **dedicated PostgreSQL** database (not Core). Same MigratorV2 pattern as Core and BotWhatsAppMD.

## Environment

Either:

- `RELAY_DATABASE_URL` — libpq connection string or `postgres://` URL

Or:

- `RELAY_POSTGRES_HOST`, `RELAY_POSTGRES_USER`, `RELAY_POSTGRES_PASSWORD`, `RELAY_POSTGRES_DB`, `RELAY_POSTGRES_PORT`
- optional `RELAY_POSTGRES_SSLMODE` (default `disable`)

Load from `relay/.env` or `Common/.env` when present (`infra.LoadEnv`).

## Commands

From `.coordinator/relay`:

```bash
go run infra/migrator/main.go          # apply pending
go run infra/migrator/main.go status
go run infra/migrator/cmd/create_migration/main.go add_my_table
```

On **relay process start**, pending migrations run automatically (`infra/db.Open`).

## Schema (initial)

- `connector_secrets` — encrypted marketplace secrets
- `external_events` — normalized events (query + future Pulse/stats)
- `connector_daily_stats` — rollup placeholder

JSONL under `Common/data/progress/external/events/` remains the coordinator-state bus; Postgres is a parallel index for analytics.
