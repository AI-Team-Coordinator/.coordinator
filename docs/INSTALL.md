# INSTALL.md — agent playbook (beta)

Author: **Evgeny KOSIVTSOV** (`EK`)  
Created: 2026-09-10 11:38

**If you are a person reading this file:** do not run the steps. Open Cursor in the project and paste this prompt (use this file’s path or GitHub URL as the playbook):

```text
Install the Coordinator.
Playbook: <path or URL of this INSTALL.md>
```

You confirm each block. The agent executes. There is no separate human install checklist — one playbook, this file.

---

You are a Cursor agent. The human asked to install **AI Team Coordinator**. This file is the only playbook. Do not install from memory and do not copy AlinaAssist.

This file is **English**. Speak to the human in **their** language. Cursor will translate; do not rewrite these MD files into another language.

Companion files (same folder):

- `../README.md` — overview for people ([Russian: ABOUT_RU.md](./ABOUT_RU.md))
- `ONBOARDING.md` — what to say in blocks 1–4
- `DEPENDENCIES.md` — tool checks

Stop after every block: **“Ready to continue?”** Until an explicit yes — **no writes to disk** (no `git clone` / `brew install`).

Do not treat this INSTALL as a “foreign empty project” while you are inside AlinaAssist, unless the human said they are testing in another folder.

---

## Repo source

The root of this playbook is the parent of the `docs/` directory.

- The human gave a **local path** to this file → clone/copy **this git** (not GitHub if clone fails or the repo is still private).
- The human gave a **URL** `…/docs/INSTALL.md` → `https://github.com/AI-Team-Coordinator/.coordinator.git` (or SSH if that is how they work).
- The destination directory is always `.coordinator/` relative to the **install root** (see layout).

If `.coordinator/` already exists and is not a leftover empty dir: stop. This is not a first install.

---

## Block 1 — ideology

Read `ONBOARDING.md`, retell block 1 briefly (including why task docs registry matters and that Cursor generates specs automatically without manual bureaucracy). Do not paste tables whole.

> Ready to continue?

---

## Block 2 — security

Retell `ONBOARDING.md` block 2: hooks, localhost, git bus, what is absent (no cloud, no PAT in chat).

> Ready to continue? Cursor hooks and a free localhost port for the board (default 4321) — OK? Vite :5175 is only for Alina Assist development, not for this install.

---

## Block 3 — layout (inspect the project, do not install)

**Read-only** first. Create nothing and do not `mv`.

### 3.1. What to inspect

At the Cursor workspace root:

```bash
pwd
ls -la
git rev-parse --is-inside-work-tree 2>/dev/null || true
git rev-parse --show-toplevel 2>/dev/null || true
```

Also, without writing:

- whether **direct children** of the root have `.git` (`ls -d */.git` or equivalent);
- whether Cursor is opened **inside** a repo (toplevel strictly above `pwd`);
- whether `.coordinator/` is already there.

From files: empty / `git init` / `package.json` | `go.mod` | `src/` at the root.

### 3.2. How to decide (one recommendation)

| What you see | Recommend | Config name |
|---|---|---|
| Empty or almost empty; no git or a fresh `git init` | Put the Coordinator **inside this folder** | `in-repo` |
| Cursor root = one git root, product code at that root | Same: **inside** | `in-repo` |
| Cursor root is **not** git, one child git | Already “one level up”: `.coordinator` beside that git, **do not move folders** | `workspace-parent` |
| Git toplevel is **above** the Cursor root | Suggest **reopen Cursor at the repo root** and `in-repo`. If they want to keep the nested window — `workspace-parent` (may create a parent) | `in-repo` first |
| Several sibling `.git` dirs | `.coordinator` at the window root; pick the bus repo later. Full multi-repo is not beta | `workspace-parent` |
| `.coordinator/` already exists | Do not install | — |

An empty-folder test always defaults to **`in-repo`**.

### 3.3. What to tell the human

1. One paragraph: “I see …”.
2. The recommendation and why.
3. The other option in one sentence (for an empty folder: “one level up is for several repos in one window later”).
4. If the recommendation needs **creating a parent and moving** the current folder — list the paths that will appear and that Cursor must then be opened on the parent. No “yes” on that list — do not do it.

Retell why two layouts exist (`ONBOARDING.md` block 3), briefly.

> Install the way I recommend? If not, say `inside` or `one level up`.

Remember the choice. Later blocks depend on it. The setup wizard must **not** ask layout again — only confirm.

---

## Block 4 — existing Cursor rules (inspect, do not copy)

**Read-only.** The Coordinator voice lives in `.cursor`. A live project often already has rules. Align with them; do not overwrite the host.

Retell `ONBOARDING.md` block 4 in short sentences, then inspect the **intended** `.cursor` for the layout chosen in block 3:

- `in-repo` → `<install root>/.cursor`
- `workspace-parent` → `<Cursor window root>/.cursor`

### 4.1. What to inspect

Without writing:

- does `.cursor/` exist;
- `rules/**/*.mdc` (names, which are `alwaysApply`);
- whether `rules/coordinator/` already exists;
- `hooks.json` and `hooks/` if present.

### 4.2. How to decide

| What you see | Tell the human | Next |
|---|---|---|
| No `.cursor` | Coordinator will create `.cursor` and put its pack under `rules/coordinator/` | Continue |
| `.cursor` exists, no `rules/coordinator/` | Host rules **stay**. Coordinator adds only `rules/coordinator/` and merges hooks — no replace of `hooks.json` | Continue |
| Host has product rules (`safety`, domain, frontend, …) | Those files are not touched. Name a few so it is visible | Continue |
| `rules/coordinator/` already there | This is not a first install of the voice | Stop unless they asked to update |
| Another alwaysApply protocol fights Coordinator (other branch lock, other “one chat = one task”, another coordinator) | Name the files. Do not copy the pack on top | Stop and ask |
| `pack/cursor-rules/` missing in **this** clone | Dashboard can still come up; voice in this project will be incomplete. Do **not** copy another team’s `safety.mdc` / prod hooks | Continue, be honest |

### 4.3. What to tell the human

1. What already lives in `.cursor` (or that it is empty).
2. What the Coordinator will **add** (`rules/coordinator/`, hook scripts, merge into `hooks.json`).
3. What it will **not** touch (everything else under `.cursor/rules/`).
4. Any conflict, in one sentence.

> Ready to add Coordinator rules next to yours? Nothing is copied yet.

---

## Block 5 — steps (do not run them yet)

Show the plan **for the chosen** layout. Do not clone.

### If `in-repo`

In the current root (if there is no git yet — `git init` after consent on block 6):

- `.coordinator/` appears (nested git, in the product `.gitignore`);
- `.cursor/rules/coordinator/` and hooks via `./utils/install_cursor_pack.sh`; host rules outside that folder stay;
- host rules outside `rules/coordinator/` stay as agreed in block 4;
- `docs/` and `coordinator-data/` — bus in **this** git;
- processes on a **free** API port (default `4321`; if busy, the next free one). The board is that same port (built UI). Do not start Vite.

### If `workspace-parent` with no moves

The window root stays. `.coordinator/` and `.cursor/` live there. Product git is a child folder; bus and `docs/` live **there**. Same `.cursor` rule: add `rules/coordinator/`, do not replace the host pack.

### If `workspace-parent` with a new parent

Only if they chose this in block 3:

1. Create the parent directory.
2. Move the current project into it.
3. The human **reopens** Cursor on the parent.
4. Continue as “no moves”.

> That is everything that will hit the disk. Ready to check dependencies?

---

## Block 6 — dependencies

Follow `DEPENDENCIES.md`: facts table first (`go version`, `node -v`, …), **then** offer to install. After clone you will run `./utils/free_port.sh 4321`, propose that port (and the next free ones if 4321 is taken — another Coordinator is likely using it), and write it to `.env` as `PORT=`.

> Ready to install missing tools and copy the Coordinator onto disk?

After this “yes”, you may write to disk.

---

## Execute (after every “yes”)

Be honest about what already works.

### Already works

1. `git init` at the product root if needed (`in-repo` with no git).
2. Clone into `.coordinator/` from the source above. Do not clone the repo into itself.
3. `cp .coordinator/.env.example .coordinator/.env` and set paths. Pick a free API port (`./.coordinator/utils/free_port.sh 4321` after the clone, or from the install root `./utils/free_port.sh 4321`). Propose it. If 4321 is busy, say another Coordinator is probably already on this machine and use the next free port. Write `PORT=` and `UI_MODE=static`. Do **not** start Vite and do **not** write `UI_MODE=vite`.

`in-repo`:

```text
WORKSPACE_ROOT=..
DATA_DIR=../coordinator-data
DOCS_DIR=../docs
BUS_DIR=..
CURSOR_DIR=../.cursor
PORT=<free port>
UI_MODE=static
```

`workspace-parent`: `WORKSPACE_ROOT` = window root; `BUS_DIR` / `DATA_DIR` / `DOCS_DIR` inside the product git; `CURSOR_DIR` = `.cursor` at the window root.

4. Append `pack/gitignore.host` to the **product git** `.gitignore` (for `in-repo` that is the same root).
5. Copy seed: `pack/seed/*` → `coordinator-data/settings/` (locale, `coordinator.json` with the assistant's name, empty team, empty project_profile) and `current_author.example`. Do **not** ask for project name, alias, or people in chat — the first setup screen confirms that. Do **not** set `setup.completed`.
6. From `.coordinator` run `./utils/install_cursor_pack.sh` (uses `CURSOR_DIR` from `.env`). It copies `pack/cursor-rules/` → `.cursor/rules/coordinator/`, Coordinator hooks → `.cursor/hooks/`, and **merges** `hooks.json`. Host rules outside `rules/coordinator/` stay. Do not copy AlinaAssist `safety.mdc` or prod skills.
7. `cd .coordinator && ./utils/run.sh` (npm install if needed, Go build, **static** UI on `PORT`). Do not start Vite. Later, `sessionStart` may ask to update from `origin/main` (once a day, cache `$DATA_DIR/.cache/last_update_check`). Do not `git pull` until the human says yes.
8. Give the link: [Coordinator](http://127.0.0.1:PORT) with the **real** port from `.env`, as an address, not “I will open it for you”. If Simple Browser is not open yet — once, by hand, at that URL. The first screen is a confirm form: **Solo** vs **Team** (two cards), language, project name, installer as admin. Default is Solo — states stay on this machine, no git bus. Team needs an existing or separate repo for snapshots (in a monorepo that is the same git). Layout was already chosen in chat — the form only shows it. Alias is filled from the name (first + last initials, or the first two letters).

### Not ready yet (do not pretend)

- `pack/cursor-rules/` missing — should not happen on a current clone. If it is missing, do not copy `safety.mdc` or prod hooks from another workspace. Say the board will come up; Coordinator voice in this project will be incomplete.
- GitHub org, `gh`, Coolify — do not do these.

Clone/build errors: paste the log, not “try later” without a fact.

---

## Forbidden

- Installing before the last “yes”.
- Creating a parent and `mv` without an explicit `workspace-parent` choice and a confirmed path list.
- Opening a browser / `open http://…` for the human.
- Committing the bus without a separate consent.
- Changing product code (except `.gitignore`, the `.cursor` pack under `rules/coordinator/`, hook merge, bus/docs folders).
- Overwriting host Cursor rules outside `rules/coordinator/`.
- Running `install_cursor_pack.sh` against AlinaAssist (that workspace keeps its own `.cursor`).
