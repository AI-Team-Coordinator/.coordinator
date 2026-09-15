# AI Team Coordinator

> **Air Traffic Control for AI-Assisted Software Development**  
> Zero-Trust, Git-Native team coordination when your codebase grows faster than human attention.

[Читать на русском (ABOUT_RU.md)](./docs/ABOUT_RU.md)

---

## Executive Summary: The AI Coding Paradox

Software development has entered a new era. With AI coding assistants like Cursor and Claude, writing code is 5 to 10 times faster and practically free. 

However, many technology teams and business leaders are discovering a painful paradox: **as individual code generation speeds up, overall team delivery often slows down.**

Why? Because writing code became instant, but **coordinating what is being written became exponentially harder.**

When multiple developers (or a single developer working across multiple AI chat sessions) generate hundreds of lines of code simultaneously:
- Different AI chats modify adjacent parts of the system without knowing about each other.
- Code is written against outdated assumptions or broken interfaces.
- Senior engineers spend their days untangling merge conflicts and reviewing massive AI-generated pull requests instead of building business features.
- Daily standup meetings become obsolete within an hour.

**AI Team Coordinator** solves this fundamental bottleneck. It acts as an active **Air Traffic Controller** for software development, preventing collisions and keeping the team’s work aligned in real time.

### Why Cursor: Built on the Global Category Leader
**Cursor is today's undisputed global leader in AI-native software development** and the gold standard for hundreds of thousands of engineers, startups, and tech enterprises worldwide. Its unprecedented adoption and code-generation speeds are precisely what exposed the coordination bottleneck: writing code is now instant, but aligning parallel streams with traditional tools is impossible.

The Coordinator does not compete with Cursor or propose "yet another IDE" — it sits natively on top of the world leader, solving its primary blind spot: multi-stream team alignment and proactive collision prevention.

---

## The 4 Hidden Costs AI Coders Create for Your Business

If your development team is actively using AI tools, you are likely experiencing these four invisible profit leaks:

### 1. "Ghost" Collisions and Broken Releases
An engineer opens three AI chat windows to work on three features at once. Chat A updates the user authentication logic; Chat B builds a new billing screen assuming the old authentication logic. Both features seem to work perfectly in isolation. But when combined, they silently break in production. 

Finding and fixing these architectural collisions after the fact takes days of expensive senior developer time.

### 2. Disappearing Context and Lost Decisions
When a developer interacts with an AI agent, critical architectural reasoning happens inside an ephemeral chat window. Once that chat is closed, all that reasoning vanishes into thin air. 

The business is left with newly generated code, but nobody remembers *why* it was designed that way, what tradeoffs were considered, or what edge cases were handled.

### 3. The Code Review Bottleneck & "Blind Merging"
Before AI, an engineer might submit a 100-line pull request that a tech lead could review carefully in 15 minutes. Today, AI assistants generate 1,500 lines in minutes. 

Senior engineers physically cannot read that much code every day. Teams face an impossible choice: either stop everything to review code all day (halting feature delivery), or hit "Merge" blindly and hope nothing breaks.

### 4. Standups and Trackers That Expire in 60 Minutes
Traditional project management tools (Jira, Linear, Trello) were designed for human working speeds. A task was picked up in the morning and finished in two days. 

In an AI-native workflow, a task starts, morphs, and produces multiple branches in 45 minutes. By 11:00 AM, the morning standup notes are already inaccurate.

---

## What AI Team Coordinator Does: The Third Teammate

Most developer tools are passive: they wait until code is already broken in Git or rejected by CI/CD before sounding an alarm.

The **Coordinator** is active. It acts as a dedicated **third participant in the development workflow**:

$$\text{Developer} \longleftrightarrow \text{AI Agent} \longleftrightarrow \textbf{Coordinator}$$

### Proactive Collision Prevention (Before Code Is Written)
Before an AI agent touches a single line of code, the Coordinator checks what the rest of the team (and the developer's other AI chats) are currently doing. 

If an AI agent attempts to edit a module or service that is already claimed by an ongoing task, the Coordinator intervenes immediately inside the conversation:
> *"Coordinator stopped edits: Task 'Refactor Billing API' is already open on this module. Complete that work or resolve the scope first."*

Work stops **before** conflicting code is created, saving hours of cleanup.

### Living Map of Work (Pulse)
The Coordinator provides a real-time visual dashboard (**Pulse**) running locally for the team:
- **Who is doing what right now:** Active tasks, associated branches, and claimed services.
- **Intentions over file lists:** Clear human-readable task summaries, not cryptic commit hashes.
- **Exploratory research vs. delivery:** Clear separation between active development tasks and early research sessions.

### Spec-Driven Institutional Memory (Automated via Cursor)
Instead of relying on vague conversational prompts that cause AI models to hallucinate and expand scope uncontrollably, the Coordinator anchors every task to a concise, versioned specification in the project's documentation registry (`docs/`).

**Developers don't need to write docs by hand:** once the Coordinator rules are installed into `.cursor/rules/`, Cursor automatically structures the task passport, names the branch, lists the claimed services, and commits the intent before modifying a single line of code.

The business permanently retains the "why" behind every change. When new developers or new AI agents touch that code six months later, the full historical context is instantly accessible.

---

## Zero-Trust Security: Why Leaders and CISOs Trust It

Many AI orchestration tools require uploading your company's proprietary code, database credentials, and internal roadmaps to a third-party cloud. For serious businesses, this is an unacceptable compliance and intellectual property risk.

AI Team Coordinator is built on a **Zero-Trust, Local-First philosophy**:

1. **Your Code Never Leaves Your Perimeter:** The Coordinator does not send your code, prompts, or environment variables to any external SaaS servers.
2. **Git-Native Coordination:** The coordination data lives directly within your team’s existing private Git repository. Your Git permissions dictate who can see and do what.
3. **Zero Extra Cloud Infrastructure:** No expensive external databases or dedicated orchestration clusters to maintain.
4. **AI FinOps Visibility:** Real-time tracking of AI model consumption and costs per task, replacing unexpected monthly cloud bills with predictable operational visibility.

---

## How Teams Adopt It: Zero Friction

Adopting the Coordinator does not require an organizational overhaul or weeks of training:

1. **Starts with One Engineer (Day 1):** A team lead or senior developer installs the Coordinator in under 5 minutes to organize their own parallel AI sessions. They immediately feel the relief of clean branches and zero lost context.
2. **Expands to the Team Naturally (Week 1):** Once proven on one machine, teammates connect to the same project repository. The team instantly gains a shared, real-time map of all AI-assisted work without learning complex new software.
3. **Compounds as the Codebase Grows (Month 1+):** As the codebase expands, the Coordinator prevents tech debt from accumulating, keeping developer velocity high and senior engineers focused on strategic architecture.

---

## Quick Start & Installation

Setting up the Coordinator is completely automated through the AI coding environment itself — no complex manual terminal configuration required.

### Install in Cursor (Recommended)

Open Cursor in your project folder (or a fresh test project) and send this prompt in the chat:

```text
Install the Coordinator.
Playbook: https://github.com/AI-Team-Coordinator/.coordinator/blob/main/docs/INSTALL.md
```

The agent will explain how it works, inspect your project and existing Cursor rules, propose the cleanest layout, verify required tools, pick a free localhost port, and set up the local dashboard on that port (default [http://127.0.0.1:4321](http://127.0.0.1:4321)).

`docs/INSTALL.md` is the **agent playbook** (not a human terminal checklist). People paste the prompt above; the agent follows that file.

---

## Development & Local Architecture

For developers working directly on the Coordinator codebase:

### Architecture
- **Backend:** Go daemon (`127.0.0.1:$PORT`, default `4321`) managing JSONL event bus, SQLite cache, and Git synchronization.
- **Frontend:** React SPA. **Installs** serve the built UI from the same `$PORT`. **Alina Assist** development uses Vite HMR on `$UI_PORT` (default `5175`).
- **IDE Layer:** Cursor lifecycle hooks (`beforeFileEdit`, `sessionStart`, `afterAgentResponse`) and rule engines.

### Running Locally
```bash
cp .env.example .env   # configure workspace paths
# Alina Assist: UI_MODE=vite (Vite on UI_PORT). Installs: UI_MODE=static (board on PORT).
./utils/run.sh
# after Go edits (and after UI edits in static mode):
./utils/backend.sh restart
```

### Environment Variables

| Variable | Meaning | Example |
|---|---|---|
| `WORKSPACE_ROOT` | Clone root containing all repos | `..` |
| `DATA_DIR` | Settings, progress, event bus (`jsonl`), SQLite cache | `../coordinator-data` |
| `DOCS_DIR` | Task specifications and documentation registry | `../docs` |
| `BUS_DIR` | Git coordination bus repository | `..` |
| `CURSOR_DIR` | Cursor IDE rules and hooks | `../.cursor` |
| `PORT` | Local API port; in static mode this is also the board | `4321` |
| `UI_PORT` | Vite port when `UI_MODE=vite` | `5175` |
| `UI_MODE` | `static` (install default) or `vite` (Alina Assist) | `static` |

### Updates

Once a day the `sessionStart` hook compares local `.coordinator` to `origin/main` (commit date on the public repo). Cache: `$DATA_DIR/.cache/last_update_check`. If HEAD is behind, Otto asks in chat whether to update. Do not pull until the human says yes; then `git pull --ff-only` in `.coordinator` and `./utils/backend.sh restart`. Manual: `./utils/update_check.sh [--force] [--notice]`.
