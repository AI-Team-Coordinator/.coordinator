# AI Team Coordinator onboarding (chat copy)

Author: **Evgeny KOSIVTSOV** (`EK`)  
Created: 2026-09-10 11:38

This file is **English**. Speak to the human in **their** language (first message / UI locale). Do not paste this markdown into the chat. Retell each block in short sentences.

After blocks 1–2, stop per `INSTALL.md`. Block 3 is not theory: **inspect the project first** (procedure in `INSTALL.md`), then explain the two layouts and your recommendation.

---

## Block 1. Ideology

The Coordinator is not a dashboard and not a pile of Cursor rules. It is a **third participant**: the human, the agent (this chat), and the Coordinator.

Its job is to **coordinate work on a project while the codebase grows fast**. The product is built for **teams**. Every install still starts with one person — a lead or an engineer — trying it on their own Cursor chats. If it does not help that person, it will not reach the team.

Git discipline is a tool, not the point. The point is a living map of the work: what is in progress, what overlaps, what is already decided.

Without it, speed turns into lost context — first in two chats on one machine, then between people:

- two streams (chats or people) touch the same module at once;
- one chat fixes a bug while another writes against the old contract;
- tasks live only in a prompt and are gone tomorrow;
- `main` and reviews lag behind the volume of generated code.

How it works (Zero-Trust, local-first):

1. **Voice in chat.** Before edits, the agent checks open tasks and stops on overlapping scope.
2. **Branch discipline.** New work starts from clean `main`. One Cursor tab — one task.
3. **Bus in the product git.** Task status and docs live in the project repository, not in a Coordinator cloud. No third-party SaaS takes your code.
4. **The board.** Pulse on localhost: who is doing what, which branches, where it collides. That screen is not “the Coordinator itself”.

Words worth knowing:

| Term | Meaning |
|---|---|
| **Task** | A unit of work with a name, a branch, and a list of services |
| **Pulse** | The live board of work (one person or a team) |
| **Research** | Reconnaissance without code edits; it is not a task |
| **Bus** | A folder in the product git: people, status, documents |
| **Docs registry** | Project task specifications (`docs/`); institutional memory and scope boundaries |

### Why task docs matter (and why you won't write them by hand)

If your project or team never had a formal documentation registry before, this is the backbone of the system:
- **Why docs are mandatory:** Without written specifications, AI agents hallucinate, lose focus across chat restarts, and silently overwrite adjacent contracts. A task doc anchors the exact boundary of what the agent is allowed to touch.
- **Automated by Cursor instructions:** You never have to manually fill out templates. Once the Coordinator rules are installed into `.cursor/rules/`, Cursor automatically transforms your natural conversation into a versioned task specification in `docs/`, registers the branch and claimed services, and checks with the Coordinator before touching code.
- **Value for solo & team:** For solo developers, it prevents disappearing context and broken `main`. For teams, it is the exact source of truth that allows the Coordinator to prevent collisions before code is written.

The Coordinator does not write features and does not deploy production. It keeps parallel work aligned so a **team** can grow the codebase without losing the plot — starting with the person who installed it.

---

## Block 2. Security and access

What the Coordinator **does** on this machine:

- Places a `.coordinator/` folder (dashboard sources, a separate git).
- Writes rules and hooks into the project’s `.cursor/` — otherwise the agent cannot hear the Coordinator.
- Cursor hooks: before a file write, at session start/end. They may **block** an edit on the wrong branch. That is intentional.
- Starts local processes: API `127.0.0.1:4321`, UI `http://localhost:5175`. The product does not leave this machine.
- Writes settings and status into a folder **inside the product git** (the bus). Commits to that git happen only after consent.

Network:

- Clone/update of the public Coordinator repo (git host).
- Nothing else is required. GitHub org, PATs, prod DBs, Coolify — **not needed**.

What is **not** happening:

- a Coordinator cloud holding your code;
- passwords or tokens requested in chat;
- edits to your product services in this install (only tooling + `.cursor` + the bus).

If hooks or localhost are unacceptable, stop here. Without hooks this is just a local website, not a Coordinator.

---

## Block 3. Why layout can differ

The Coordinator places three things: dashboard sources (`.coordinator/`), Cursor rules (`.cursor/`), and the bus + docs **in the product git**.

Two working layouts:

1. **Inside the current git** — `.coordinator` next to the code. One repository, Cursor opened at its root. Simpler, and this is the default for an empty test.
2. **One level up** — the Cursor root is not git: `.coordinator` sits beside the product folder (several repos in one window). Use this when the git root is not the Cursor window, or you will have more than one repo.

The agent must not move folders or create a parent until the human picks a layout. The recommendation comes from what is actually on disk, not from AlinaAssist habit.
