# Recall

A local context engine for AI-assisted development. Recall maintains persistent, searchable knowledge about your projects and coding patterns, reducing redundant context building and token consumption across AI assistant sessions.

## What It Does

Recall operates as a **context layer** between your codebase and AI assistants:

- **Indexes** project structure, file summaries, and semantic content into local SQLite databases
- **Matches** user prompts against reusable task recipes and historical solutions
- **Routes** simple queries to local LLMs (llama.app, llama.cpp, Ollama) and complex ones to cloud assistants
- **Learns** from interactions: tracks what local models handle well, extracts reusable code snippets, records agent behavior

Nothing leaves your machine unless you delegate to a cloud model.

## Installation

```bash
# macOS / Linux
curl -LsSf https://llama.app/install.sh | sh  # optional: for local LLM support
go install github.com/gleicon/recall@latest
```

Recall is a single static binary. No runtime dependencies, no CGO, no API keys required.

## Quick Start

```bash
# 1. Map your project
cd my-project
recall map

# 2. Load default recipes (framework patterns, common tasks)
recall recipes seed

# 3. Generate a context-rich brief for an AI assistant
recall brief "add OAuth login"

# 4. Query with smart routing (local model if available, else enriched brief)
recall query "how do I structure middleware in this project"
```

## Core Commands

### Project Context

| Command | Purpose |
|---|---|
| `recall map` | Detect and cache project type, entry points, module boundaries |
| `recall cache build` | Index all source files with summaries and embeddings |
| `recall cache inspect` | View cached files, subsystems, and memories |
| `recall cache refresh` | Incrementally update only changed files |
| `recall learn "insight"` | Store a manual insight into project memory |

### Knowledge Retrieval

| Command | Purpose |
|---|---|
| `recall brief "task"` | Generate enriched prompt context from recipes + project state |
| `recall query "question"` | Smart router: local LLM answer or delegation brief |
| `recall search "terms"` | FTS5 + vector search over indexed content |
| `recall search -c "terms"` | Chunk-level semantic search |

### Global Brain

| Command | Purpose |
|---|---|
| `recall brain conversations` | History of local model interactions |
| `recall brain snippets` | Reusable code blocks extracted from responses |
| `recall brain lessons` | What works per framework / model |
| `recall brain stats` | Aggregate metrics: success rate, tokens saved |
| `recall brain search "auth"` | Search all brain data |
| `recall brain frameworks` | Per-framework performance breakdown |

### Recipes

| Command | Purpose |
|---|---|
| `recall recipes seed` | Load 30+ default recipes (Go, Next.js, Python, Rust, etc.) |
| `recall recipes list` | Show all loaded recipes with usage counts |
| `recall recipes add -f my-recipe.json` | Add a custom recipe |

### Local LLM Management

| Command | Purpose |
|---|---|
| `recall local status` | Detect running local LLM server |
| `recall local models` | List available models |
| `recall local use <model>` | Lock to a specific model |

### Recording & Stats

| Command | Purpose |
|---|---|
| `recall run suggest --task "..."` | Gated recording of an assistant run |
| `recall run record --task "..."` | Manual recording without prompt |
| `recall stats cache` | Project cache statistics |
| `recall stats recipes` | Recipe usage statistics |
| `recall stats runs` | Aggregated run statistics |
| `recall stats global` | Cross-project global statistics |
| `recall stats insights` | Most/least useful recipes |

### Maintenance

| Command | Purpose |
|---|---|
| `recall bench` | Run performance benchmarks |
| `recall cleanup` | Remove old cache entries |
| `recall cleanup project <dir>` | Remove a project's data directory |

## Data Storage

Recall uses two SQLite databases:

**Global** (`~/.recall/global.db`):
- Task recipes with vector embeddings
- Framework fingerprints
- Conversation history with local models
- Extracted code snippets
- Agent lessons (what works per framework/model)

**Per-project** (`~/.recall/projects/<hash>/project.db`):
- File summaries and content
- Semantic chunks with embeddings
- Subsystem abstractions
- User insights and memories
- Run history and outcomes

All data is local. No cloud sync, no telemetry.

## Documentation

- [Usage Reference](docs/usage.md) — Complete command reference with examples
- [Integration Guide](docs/integration.md) — Hook into Claude, Pi, opencode, Cursor, Codex
- [Scripting & Automation](docs/scripting.md) — Shell integration, pre-commit hooks, CI
- [Manual Data Management](docs/manual-data.md) — Adding recipes, insights, snippets by hand
- [Backup & Migration](docs/backup.md) — Export, import, and archive strategies
- [Architecture](docs/architecture.md) — How the embedding, search, and routing systems work

## Requirements

- Go 1.22+ (for building from source)
- macOS or Linux
- Optional: Any OpenAI-compatible local LLM server on localhost:8080

## License

MIT
