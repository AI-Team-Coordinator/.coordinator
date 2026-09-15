# Machine dependencies (beta)

Author: **Evgeny KOSIVTSOV** (`EK`)  
Created: 2026-09-10 11:38

This file is **English**. Speak to the human in **their** language.

Agent: **check first, install nothing.** Show the table. Install only after “Ready to continue?” from `INSTALL.md`.

macOS: offer missing tools via `brew`. Linux: the distro package manager. Windows is out of beta — say so.

Do not install global extras “just in case” (`gh`, Docker, Coolify).

---

## Required

| Tool | Why | How to check | Minimum |
|---|---|---|---|
| **git** | Clone `.coordinator`, product-git bus | `git --version` | 2.30+ |
| **Cursor** | Rules and hooks | The human is already in Cursor | `hooks.json` support |
| **Go** | Build the API | `go version` | 1.22+ (this repo is `go 1.25`) |
| **Node.js + npm** | Build the UI (static dist on install; Vite only in Alina Assist) | `node -v` && `npm -v` | Node 20+ |
| **Python 3** | Hooks and daemonize | `python3 --version` | 3.10+ |

Ports: one **API/UI** port must be free **or** already used by **this** Coordinator. Default `4321`. Several Coordinators on one machine each get their own `PORT` in `.env`.

Check: `./utils/free_port.sh 4321` (after `.coordinator` exists) or `lsof -nP -tiTCP:4321 -sTCP:LISTEN`. If another process owns the candidate port — propose the next free one; do not kill a neighbour Coordinator.

Do **not** require `:5175` for an install. Vite is Alina Assist development only.

---

## Optional (do not install in beta)

| Tool | When it is needed |
|---|---|
| **GitHub CLI (`gh`)** | Org, creating repos, inviting a second person |
| Docker / Coolify | Deploying the *product*, not the Coordinator |

---

## Install commands (macOS, after consent)

```bash
brew install git go node python3
```

Re-check versions. If `go` or `node` exist but are old — tell the human; do not silently downgrade.
