# INSTALL.md — agent playbook (beta)

Author: **Evgeny KOSIVTSOV** (`EK`)  
Created: 2026-09-10 11:38

You are a Cursor agent. The human asked to install **AI Team Coordinator**. This file is the only playbook. Do not install from memory and do not copy AlinaAssist.

This file is **English**. Speak to the human in **their** language. Cursor will translate; do not rewrite these MD files into another language.

Companion files (same folder):

- `README.md` / `ABOUT.md` — overview for business owners and team leaders ([Russian: ABOUT_RU.md](./ABOUT_RU.md))
- `ONBOARDING.md` — what to say in blocks 1–3
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

> Ready to continue? Cursor hooks and ports 4321 / 5175 — OK?

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

Remember the choice. Block 4 depends on it. The web wizard must **not** ask layout again — only confirm.

---

## Block 4 — steps (do not run them yet)

Show the plan **for the chosen** layout. Do not clone.

### If `in-repo`

In the current root (if there is no git yet — `git init` after consent on block 5):

- `.coordinator/` appears (nested git, in the product `.gitignore`);
- `.cursor/rules/coordinator/` and hooks — when `pack/` exists; otherwise say honestly that Coordinator voice in this project will be incomplete;
- `docs/` and `coordinator-data/` — bus in **this** git;
- processes on :4321 and :5175.

### If `workspace-parent` with no moves

The window root stays. `.coordinator/` and `.cursor/` live there. Product git is a child folder; bus and `docs/` live **there**.

### If `workspace-parent` with a new parent

Only if they chose this in block 3:

1. Create the parent directory.
2. Move the current project into it.
3. The human **reopens** Cursor on the parent.
4. Continue as “no moves”.

> That is everything that will hit the disk. Ready to check dependencies?

---

## Block 5 — dependencies

Follow `DEPENDENCIES.md`: facts table first (`go version`, `node -v`, …), **then** offer to install. Ports 4321/5175: free or already this Coordinator.

> Ready to install missing tools and copy the Coordinator onto disk?

After this “yes”, you may write to disk.

---

## Execute (after every “yes”)

Be honest about what already works.

### Already works

1. `git init` at the product root if needed (`in-repo` with no git).
2. Clone into `.coordinator/` from the source above. Do not clone the repo into itself.
3. `cp .coordinator/.env.example .coordinator/.env` and set paths:

`in-repo`:

```text
WORKSPACE_ROOT=..
DATA_DIR=../coordinator-data
DOCS_DIR=../docs
BUS_DIR=..
CURSOR_DIR=../.cursor
PORT=4321
```

`workspace-parent`: `WORKSPACE_ROOT` = window root; `BUS_DIR` / `DATA_DIR` / `DOCS_DIR` inside the product git; `CURSOR_DIR` = `.cursor` at the window root.

4. Append `pack/gitignore.host` to the **product git** `.gitignore` (for `in-repo` that is the same root).
5. Copy seed: `pack/seed/*` → `coordinator-data/settings/` (locale, team, project_profile) and `current_author.example`. Ask briefly for project name / alias if the web wizard does not exist yet.
6. `cd .coordinator && ./utils/run.sh` (npm install if needed, Go build, API + Vite).
7. Give the link: [Coordinator](http://localhost:5175) as an address, not “I will open it for you”.

### Not ready yet (do not pretend)

- `pack/cursor-rules/` — generic rules without AlinaAssist. No pack → do not copy `safety.mdc` or prod hooks from another workspace. Say the board will come up; Coordinator voice in this project will be incomplete.
- Web create-wizard (language, people, `admin` ACL) — when it exists, it opens itself. Until then: seed + short questions in chat (language, alias, name, project title). The installer is admin even if `access` is not read in code yet.
- GitHub org, `gh`, Coolify — do not do these.

Clone/build errors: paste the log, not “try later” without a fact.

---

## Forbidden

- Installing before the last “yes”.
- Creating a parent and `mv` without an explicit `workspace-parent` choice and a confirmed path list.
- Opening a browser / `open http://…` for the human.
- Committing the bus without a separate consent.
- Changing product code (except `.gitignore`, the `.cursor` pack, bus/docs folders).
