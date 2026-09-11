# Host Cursor pack (beta)

This folder is the **generic** Coordinator voice for a *foreign* project. Install copies it into that project’s `.cursor/`:

- `cursor-rules/*.mdc` → `.cursor/rules/coordinator/`
- `cursor-hooks/*` → `.cursor/hooks/`
- `hooks.json.fragment` merged into `.cursor/hooks.json` (existing host hooks stay)

Run from the Coordinator clone:

```bash
./utils/install_cursor_pack.sh
```

`CURSOR_DIR` comes from `.coordinator/.env`. Do not copy AlinaAssist `safety.mdc`, skills, or prod hooks.

**Do not install this pack into AlinaAssist.** That workspace keeps its own `.cursor/rules` (Common, Coolify, Core). This pack is for other projects.

Paths in the rules mean directories from `.env`, not a folder named `Common`:

| Variable | Typical in-repo value |
|---|---|
| `DATA_DIR` | `../coordinator-data` |
| `DOCS_DIR` | `../docs` |
| `BUS_DIR` | `..` (the product git) |
| `CURSOR_DIR` | `../.cursor` |
