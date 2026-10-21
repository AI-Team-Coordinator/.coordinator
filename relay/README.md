# Relay pack (upstream)

Reference implementation and **pack** for [Coordinator Relay](https://github.com/Alina-Assist/coordinator-relay) instances.

- Version: `PACK_VERSION` in this directory.
- **Edit shared logic here**, not in project instances.
- Future: Go module `relay-core`; until then, instances sync via `../utils/relay_sync_pack.sh`.

## Workflow (no `relay-core` module yet)

1. Change code under this directory (`app/`, `domain/`, `infra/`, `model/`).
2. Bump `PACK_VERSION` when the pack release is meaningful.
3. Commit to the **Coordinator** repo (`.coordinator`).
4. From workspace root: `./.coordinator/utils/relay_sync_pack.sh` → updates `.coordinator-relay/`.
5. In the instance: `go test ./...`, commit **coordinator-relay**, rebuild (local `go build` or Coolify deploy).

Full checklist (Russian): `Common/docs/CoordinatorRelay/20261020-1925-EK-RELAY_PACK_SYNC.md`.

Project-only code (e.g. Alina connectors) lives in the instance under `connectors/` — the sync script does not overwrite that folder.

Coordinator dashboard proxies this service over HTTP; it does not embed the relay binary.
