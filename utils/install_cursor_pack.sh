#!/bin/bash
# Copy the generic Coordinator Cursor pack into the host .cursor.
# Does not overwrite host rules outside rules/coordinator/.
# Merges hooks.json (adds Coordinator entries; keeps existing host hooks).
#
# Usage (from the .coordinator clone):
#   ./utils/install_cursor_pack.sh
#   CURSOR_DIR=/path/to/.cursor ./utils/install_cursor_pack.sh
#
# Do not run this against AlinaAssist — that workspace keeps its own rules.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
COORD_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
PACK="$COORD_ROOT/pack"
# shellcheck source=paths.sh
. "$SCRIPT_DIR/paths.sh"

DEST="${CURSOR_DIR:-}"
if [ -z "$DEST" ]; then
    echo "install_cursor_pack.sh: CURSOR_DIR is empty (set it in .coordinator/.env)" >&2
    exit 1
fi

RULES_SRC="$PACK/cursor-rules"
HOOKS_SRC="$PACK/cursor-hooks"
FRAGMENT="$PACK/hooks.json.fragment"

if [ ! -d "$RULES_SRC" ] || [ ! -d "$HOOKS_SRC" ] || [ ! -f "$FRAGMENT" ]; then
    echo "install_cursor_pack.sh: pack is incomplete under $PACK" >&2
    exit 1
fi

RULES_DEST="$DEST/rules/coordinator"
HOOKS_DEST="$DEST/hooks"
mkdir -p "$RULES_DEST" "$HOOKS_DEST"

# Replace only the Coordinator rule pack (this folder). Host rules elsewhere stay.
find "$RULES_DEST" -maxdepth 1 -type f -name '*.mdc' -delete
cp "$RULES_SRC"/*.mdc "$RULES_DEST/"

cp "$HOOKS_SRC"/coord_paths.py "$HOOKS_DEST/"
cp "$HOOKS_SRC"/coordinator-check-branch.py "$HOOKS_DEST/"
cp "$HOOKS_SRC"/coordinator-check-branch.sh "$HOOKS_DEST/"
cp "$HOOKS_SRC"/coordinator-research.py "$HOOKS_DEST/"
cp "$HOOKS_SRC"/coordinator-log-merge.py "$HOOKS_DEST/"
chmod +x "$HOOKS_DEST/coordinator-check-branch.sh"
chmod +x "$HOOKS_DEST/coordinator-check-branch.py" "$HOOKS_DEST/coordinator-research.py" "$HOOKS_DEST/coordinator-log-merge.py"

python3 - "$DEST/hooks.json" "$FRAGMENT" <<'PY'
import json, os, sys

host_path, frag_path = sys.argv[1:3]
frag = json.loads(open(frag_path, encoding="utf-8").read())
if os.path.isfile(host_path):
    try:
        host = json.loads(open(host_path, encoding="utf-8").read())
    except Exception:
        host = {"version": 1, "hooks": {}}
else:
    host = {"version": 1, "hooks": {}}
if not isinstance(host, dict):
    host = {"version": 1, "hooks": {}}
hooks = host.setdefault("hooks", {})
if not isinstance(hooks, dict):
    hooks = {}
    host["hooks"] = hooks
host.setdefault("version", frag.get("version", 1))

def command_key(entry):
    if not isinstance(entry, dict):
        return ""
    return str(entry.get("command") or "").strip()

for event, additions in (frag.get("hooks") or {}).items():
    if not isinstance(additions, list):
        continue
    existing = hooks.get(event)
    if not isinstance(existing, list):
        existing = []
        hooks[event] = existing
    have = {command_key(e) for e in existing}
    for item in additions:
        key = command_key(item)
        if not key or key in have:
            continue
        existing.append(item)
        have.add(key)

os.makedirs(os.path.dirname(host_path) or ".", exist_ok=True)
with open(host_path, "w", encoding="utf-8") as f:
    json.dump(host, f, indent=2)
    f.write("\n")
PY

echo "Coordinator pack installed:"
echo "  rules: $RULES_DEST"
echo "  hooks: $HOOKS_DEST"
echo "  hooks.json merged: $DEST/hooks.json"
