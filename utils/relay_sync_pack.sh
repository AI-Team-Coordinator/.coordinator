#!/usr/bin/env bash
# Sync upstream relay pack (.coordinator/relay) into a project instance (.coordinator-relay).
#
# Workflow (until relay-core Go module exists):
#   1) Edit pack in .coordinator/relay, bump PACK_VERSION if needed, commit .coordinator
#   2) Run this script from workspace root
#   3) cd .coordinator-relay && git diff && go test ./... && commit coordinator-relay
#   4) Rebuild: go build locally or coolify_deploy.sh coordinator_relay
#
# Docs: Common/docs/CoordinatorRelay/20261020-1925-EK-RELAY_PACK_SYNC.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
PACK="${ROOT}/.coordinator/relay"
INSTANCE="${ROOT}/.coordinator-relay"
META="${INSTANCE}/.relay_pack_meta.json"

if [[ ! -d "$PACK" ]]; then
  echo "pack not found: $PACK" >&2
  exit 1
fi
if [[ ! -d "$INSTANCE" ]]; then
  echo "instance dir missing: $INSTANCE (create it first)" >&2
  exit 1
fi

PACK_VER="$(tr -d '[:space:]' < "${PACK}/PACK_VERSION")"
echo "Syncing relay pack v${PACK_VER} -> ${INSTANCE}"

rsync -a \
  --exclude '.git' \
  --exclude 'connectors/' \
  --exclude 'main.go' \
  --exclude 'README.md' \
  --exclude '.relay_pack_meta.json' \
  --exclude '.gitignore' \
  --exclude '.env' \
  --exclude '.env.example' \
  "${PACK}/" "${INSTANCE}/"

STAMP="$(date '+%Y-%m-%d')"
if command -v python3 >/dev/null 2>&1; then
  python3 - <<PY
import json, pathlib
p = pathlib.Path("${META}")
data = json.loads(p.read_text()) if p.exists() else {}
data["pack_source"] = "coordinator/relay"
data["pack_version"] = int("${PACK_VER}")
data["project_id"] = data.get("project_id", "alina-assist")
data["synced_at"] = "${STAMP}"
p.write_text(json.dumps(data, indent=2) + "\n")
PY
else
  echo "warning: python3 not found; update .relay_pack_meta.json manually" >&2
fi

if [[ -d "${PACK}/connectors" ]]; then
  mkdir -p "${INSTANCE}/connectors"
  rsync -a "${PACK}/connectors/" "${INSTANCE}/connectors/"
fi

echo "Done. Review: cd ${INSTANCE} && git status && go test ./..."
