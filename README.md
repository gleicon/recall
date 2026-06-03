# Technocore

A local context engine for AI-assisted development. Technocore maintains persistent, searchable knowledge about your projects and coding patterns, reducing redundant context building and token consumption across AI assistant sessions.

## What It Does

Technocore operates as a **context layer** between your codebase and AI assistants:

- **Indexes** project structure, file summaries, and semantic content into local SQLite databases
- **Matches** user prompts against reusable task recipes and historical solutions
- **Routes** simple queries to local LLMs (llama.app, llama.cpp, Ollama) and complex ones to cloud assistants
- **Learns** from interactions: tracks what local models handle well, extracts reusable code snippets, records agent behavior

Nothing leaves your machine unless you delegate to a cloud model.

## Installation

```bash
# macOS / Linux
curl -LsSf https://llama.app/install.sh | sh  # optional: for local LLM support
go install github.com/gleicon/technocore@latest
```

Technocore is a single static binary. No runtime dependencies, no CGO, no API keys required.

## Quick Start

```bash
# 1. Map your project
cd my-project
technocore map

# 2. Load default recipes (framework patterns, common tasks)
technocore recipes seed

# 3. Generate a context-rich brief for an AI assistant
technocore brief "add OAuth login"

# 4. Query with smart routing (local model if available, else enriched brief)
technocore query "how do I structure middleware in this project"
```

## Core Commands

### Project Context

| Command | Purpose |
|---|---|
| `technocore map` | Detect and cache project type, entry points, module boundaries |
| `technocore cache build` | Index all source files with summaries and embeddings |
| `technocore cache inspect` | View cached files, subsystems, and memories |
| `technocore cache refresh` | Incrementally update only changed files |
| `technocore learn "insight"` | Store a manual insight into project memory |

### Knowledge Retrieval

| Command | Purpose |
|---|---|
| `technocore brief "task"` | Generate enriched prompt context from recipes + project state |
| `technocore query "question"` | Smart router: local LLM answer or delegation brief |
| `technocore search "terms"` | FTS5 + vector search over indexed content |
| `technocore search -c "terms"` | Chunk-level semantic search |

### Global Brain

| Command | Purpose |
|---|---|
| `technocore brain conversations` | History of local model interactions |
| `technocore brain snippets` | Reusable code blocks extracted from responses |
| `technocore brain lessons` | What works per framework / model |
| `technocore brain stats` | Aggregate metrics: success rate, tokens saved |
| `technocore brain search "auth"` | Search all brain data |
| `technocore brain frameworks` | Per-framework performance breakdown |

### Recipes

| Command | Purpose |
|---|---|
| `technocore recipes seed` | Load 30+ default recipes (Go, Next.js, Python, Rust, etc.) |
| `technocore recipes list` | Show all loaded recipes with usage counts |
| `technocore recipes add -f my-recipe.json` | Add a custom recipe |

### Local LLM Management

| Command | Purpose |
|---|---|
| `technocore local status` | Detect running local LLM server |
| `technocore local models` | List available models |
| `technocore local use <model>` | Lock to a specific model |

### Recording & Stats

| Command | Purpose |
|---|---|
| `technocore run suggest --task "..."` | Gated recording of an assistant run |
| `technocore run record --task "..."` | Manual recording without prompt |
| `technocore stats cache` | Project cache statistics |
| `technocore stats recipes` | Recipe usage statistics |
| `technocore stats runs` | Aggregated run statistics |
| `technocore stats global` | Cross-project global statistics |
| `technocore stats insights` | Most/least useful recipes |

### Maintenance

| Command | Purpose |
|---|---|
| `technocore bench` | Run performance benchmarks |
| `technocore cleanup` | Remove old cache entries |
| `technocore cleanup project <dir>` | Remove a project's data directory |

## Data Storage

Technocore uses two SQLite databases:

**Global** (`~/.technocore/global.db`):
- Task recipes with vector embeddings
- Framework fingerprints
- Conversation history with local models
- Extracted code snippets
- Agent lessons (what works per framework/model)

**Per-project** (`~/.technocore/projects/<hash>/project.db`):
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
