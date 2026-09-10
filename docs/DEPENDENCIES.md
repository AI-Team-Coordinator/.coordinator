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
| **Node.js + npm** | Vite UI | `node -v` && `npm -v` | Node 20+ |
| **Python 3** | Hooks and daemonize | `python3 --version` | 3.10+ |

Ports that must be free **or** already used by this Coordinator:

- `4321` — API
- `5175` — UI

Check: `lsof -nP -tiTCP:4321 -sTCP:LISTEN` and the same for `5175`. If another process owns the port — stop and ask.

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
