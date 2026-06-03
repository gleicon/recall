# Specification: Pipecamp Recipes, Learning Loop, and Skill Installer

## Problem

Developers waste tokens and time because AI coding assistants rediscover the same project facts and framework patterns on every task. Pipecamp already indexes local projects and generates briefs, but its global cache is empty — there are no reusable recipes, no learning from past runs, and no assistant integration. This spec adds a curated recipe seed, a vector-based retrieval system, a feedback loop that learns which context was useful, and thin assistant skills so the tool accelerates tasks automatically and records outcomes with minimal friction.

## Scope

**In scope:**
- Global recipe storage with vector embeddings and signal-based matching
- Initial seed of ~15 hand-curated recipes across Next.js, Go, Python/FastAPI, Rust, and generic tasks
- `brief` command upgraded to use vector RAG over recipes
- `recipes seed` and `recipes add` commands
- Learning loop: `run suggest` (gated) and `run record` (manual)
- Skill installer: `install-skill --target <assistant>`
- Per-assistant skill files that auto-inject briefs at task start and suggest recording at task end

**Out of scope:**
- Importing editor snippet collections automatically
- Web UI or server mode
- Real-time sync between machines
- Multi-user or team sharing
- Automatic code generation from recipes

## Users

**Developer (primary user):**
- Runs `pipecamp` as a CLI tool alongside their editor
- Wants faster, cheaper model interactions by reusing cached context
- Occasionally confirms or skips a one-line prompt to record a task outcome
- Edits recipe JSON files in `~/.pipecamp/recipes/` to customize

**AI assistant (secondary consumer):**
- Receives prepended brief context from the pipecamp skill
- Reports files changed back to pipecamp via the skill
- Does not need to know pipecamp exists; the skill handles everything

## Functional Requirements

FR-1: The system shall store task recipes in `global.db` with the following fields: `name`, `framework`, `language`, `signals`, `context_needed`, `avoid`, `brief_template`, `source`, `tags`, `embedding`.

FR-2: The system shall compute a 256-dim normalized feature-hashing embedding for each recipe at seed time, storing it as a BLOB.

FR-3: The system shall provide a `recipes seed` command that reads JSON files from `~/.pipecamp/recipes/` and inserts them into `global.db` with embeddings, skipping duplicates by `name`.

FR-4: The system shall ship a default set of at least 15 recipe JSON files in a `recipes/` directory within the repository.

FR-5: The system shall, on first run of `recipes seed`, copy the repository default recipes to `~/.pipecamp/recipes/` if the directory does not already exist.

FR-6: The system shall provide a `recipes add --from-file <path>` command that validates and inserts a single recipe JSON file into `global.db`.

FR-7: The `brief` command shall retrieve top-k recipes from `global.db` by computing the cosine similarity between the task prompt embedding and each recipe embedding.

FR-8: The `brief` command shall re-rank retrieved recipes by framework match (+0.3) and signal overlap (+0.1 per matched signal), capping the combined score at 1.0.

FR-9: The `brief` command shall include the top 3 re-ranked recipes in the generated brief output.

FR-10: The system shall provide a `run suggest` command that accepts `task`, `files_changed`, `tokens_in`, `tokens_out`, `tests_passed`, and `follow_up_needed`.

FR-11: The `run suggest` command shall print a one-line preview and wait for a single-key response: `y` (save), `n` (skip), `i` (save with user-provided insight).

FR-12: When the user responds `y` or `i`, the system shall write a record to both the local `runs` table and the global `model_behavior_stats` table.

FR-13: The system shall provide a `run record` command that allows fully manual entry of all run fields without the interactive gate.

FR-14: The system shall provide an `install-skill --target <assistant>` command that copies the correct skill files into the assistant's config directory.

FR-15: The installed skill shall, at the start of every user task, silently execute `pipecamp brief "<user prompt>"` and prepend the result to the assistant's planning context.

FR-16: The installed skill shall, at the end of every user task, execute `pipecamp run suggest` with the assistant's best-effort file-change list and report the user's `y/n/i` response.

FR-17: The installed skill shall not record any file content, only file paths, task type, token counts, and optional user-written insights.

FR-18: The system shall provide a `stats recipes` command that shows recipe hit counts and average usefulness scores.

FR-19: The system shall provide a `stats runs` command that shows per-framework token efficiency and file usefulness statistics.

## Non-Functional Requirements

NFR-1: Latency: `brief` command shall complete in P99 < 500 ms on a project with 10,000 indexed files and 100 recipes, running on a 2022 M1 MacBook Air.

NFR-2: Scale: The system shall support at least 1,000 recipes and 50 projects on a single machine without degradation.

NFR-3: Availability: The CLI tool is a local binary; there is no server or uptime requirement.

NFR-4: Data retention: Global recipe data and run stats are retained indefinitely. Local project memories may be cleaned with `cleanup --days N` (default 30).

NFR-5: Portability: The system shall compile and run on macOS and Linux with Go 1.22+ and no CGO dependencies.

NFR-6: Privacy: The global database shall never store raw source code, secrets, `.env` files, or proprietary architecture details.

## Interfaces

**CLI commands:**
- `pipecamp recipes seed` — load recipes from disk into global DB
- `pipecamp recipes add --from-file <path>` — add one recipe
- `pipecamp recipes list` — list all recipes in global DB
- `pipecamp brief "<task>"` — generate brief with vector RAG recipes
- `pipecamp run suggest --task "..." --files-changed "a,b" --tokens-in N --tokens-out M` — gated recording
- `pipecamp run record --task "..." --files-changed "a,b" ...` — manual recording
- `pipecamp stats recipes` — recipe usage statistics
- `pipecamp stats runs` — run statistics
- `pipecamp install-skill --target claude|opencode|cursor` — install assistant skill

**External systems:**
- AI assistants (Claude Code, opencode, Cursor, Codex) via installed skill files
- Git (indirectly, through assistant file-change tracking — no direct git dependency)

**Data formats at boundaries:**
- Recipe JSON files (see FR-1 for schema)
- Skill files: Markdown skill definitions per assistant's skill format
- Brief output: plain Markdown text

## Constraints

- Primary language: Go 1.22+
- Database: SQLite via `modernc.org/sqlite` (pure Go, no CGO)
- Embeddings: feature hashing, 256 dims, no external model or API key
- Summarization: `github.com/gleicon/tldt` library
- Build toolchain: standard `go build`
- Testing: `go test ./...`
- Forbidden approaches: no CGO, no external vector DB, no cloud API calls, no web server

## Technical Profile

- Primary language: Go 1.22+
- Runtime target: CLI binary for macOS and Linux
- Build toolchain: standard Go modules (`go mod tidy`, `go build`)
- Testing framework: built-in `testing` package, `go test ./...`
- Dependencies: `modernc.org/sqlite`, `github.com/spf13/cobra`, `github.com/gleicon/tldt`

## Acceptance Criteria

AC-1: Given a fresh install, when the user runs `pipecamp recipes seed`, then `~/.pipecamp/recipes/` is populated with defaults and `global.db` contains at least 15 recipes with non-null embeddings.

AC-2: Given `global.db` contains a recipe named `add_oauth_nextjs`, when the user runs `pipecamp brief "add OAuth login"` in a Next.js project, then the brief output includes the `add_oauth_nextjs` recipe.

AC-3: Given `global.db` contains 100 recipes, when the user runs `pipecamp brief "add health check"` in a Go project, then the command completes in under 500 ms.

AC-4: Given the user has just completed a task, when `pipecamp run suggest --task "add OAuth" --files-changed "src/lib/auth.ts" --tokens-in 4000 --tokens-out 800` is run, then a one-line prompt appears and pressing `y` stores a record in both local and global stats tables.

AC-5: Given a stored run for a Next.js OAuth task, when the user runs `pipecamp stats runs`, then the output shows framework="nextjs", task_type="add OAuth", and the token reduction percentage.

AC-6: Given a clean machine, when the user runs `pipecamp install-skill --target claude`, then a skill file exists at `~/.claude/skills/pipecamp/` and references the `pipecamp brief` and `pipecamp run suggest` commands.

AC-7: Given the skill is installed, when the assistant starts a new task, then `pipecamp brief` is executed silently and its output is prepended to the assistant's context.

AC-8: Given the skill is installed, when the assistant finishes a task, then `pipecamp run suggest` is executed and the user's `y/n/i` response is respected.

AC-9: Given a recipe JSON file with an invalid field, when the user runs `pipecamp recipes add --from-file bad.json`, then the command exits with a non-zero status and prints a validation error.

AC-10: Given two identical `recipes seed` runs, when the second run executes, then no duplicate recipes are inserted into `global.db`.

### Coverage

| FR | AC |
|---|---|
| FR-1 | AC-1 |
| FR-2 | AC-1 |
| FR-3 | AC-1, AC-10 |
| FR-4 | AC-1 |
| FR-5 | AC-1 |
| FR-6 | AC-9 |
| FR-7 | AC-2 |
| FR-8 | AC-2 |
| FR-9 | AC-2 |
| FR-10 | AC-4 |
| FR-11 | AC-4 |
| FR-12 | AC-4, AC-5 |
| FR-13 | — (manual path, implicitly covered by AC-4) |
| FR-14 | AC-6 |
| FR-15 | AC-7 |
| FR-16 | AC-8 |
| FR-17 | AC-8 |
| FR-18 | AC-5 |
| FR-19 | AC-5 |

## Open Questions

1. What is the exact file path and format for each assistant's skill directory? (e.g., Claude Code uses `.claude/skills/`, but what about opencode, Cursor, Codex?)
2. Should recipe JSON files support multiple `brief_template` variants per framework version (e.g., Next.js App Router vs Pages Router)?
3. Should `run suggest` auto-detect token counts from a model API, or should the skill/assistant report them?
4. What should happen if `pipecamp brief` finds no relevant recipes — fallback to generic project map only, or print a warning?
