# Connector architecture (relay pack)

## Split of responsibility

```text
relay pack (stock)                    instance repo (e.g. .coordinator-relay/connectors/*)
─────────────────────                 ────────────────────────────────────────────────
HTTP /api/v1/*                        register.go → app.Options()
marketplace catalog + secrets         catalog.go (tile, schema, declared actions[])
connector.Connector (Normalize)       adapter + domain + infra
poll.Start scheduler                  WithPoll → poll job
connector.ActionRegistry              WithActions (optional extra actions)
POST …/connectors/{id}/actions/{action}   (no per-connector HTTP in instance)
```

**Core rule:** instance connectors never register their own `http.HandleFunc` on the relay mux. They register **capabilities** (webhook adapter, poll job, actions) and **catalog metadata**. The pack exposes uniform routes.

## Connector interface (events)

`connector.Connector` — normalize inbound deliveries → `[]model.ExternalEvent`. Used for webhooks and manual ingest.

## Poll connectors

`app.WithPoll(factory)` — periodic job. Pack auto-registers action `poll_now` → same `Run` as the scheduler.

## On-demand actions

`connector.ActionRegistry` — `POST /api/v1/connectors/{id}/actions/{action}` (master bearer). Catalog `actions[]` describes what the Coordinator UI may show.

Custom actions: `app.WithActions(func(reg, gw, runtime, settings, pollJobs) { reg.Register(...) })`.

## Stock vs instance

- **Stock** (`connectors/github/`): webhook only, synced with pack.
- **Instance** (`llm_model_lifecycle/`): poll + Alina-specific infra; not overwritten by `relay_sync_pack.sh`.
