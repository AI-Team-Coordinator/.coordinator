# Stock connectors (relay pack)

In-process connectors shipped with the pack. Synced into project instances via `relay_sync_pack.sh` (merged into `connectors/`).

| Connector | ID | Type |
|-----------|-----|------|
| [github](./github/) | `github` | Webhook — PR reviews, CI |

Project-specific connectors live in the **instance** repo (e.g. `.coordinator-relay/connectors/llm_model_lifecycle/`), not overwritten by sync.

Wire stock connectors in pack `main.go`: `app.New(cfg, github.Options()...)`.
