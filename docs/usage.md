# Usage Reference

Complete command reference for Recall.

## Project Context

### `recall map`

Detects the project type and stores structural metadata.

```bash
recall map
```

Detects language, framework, package manager, entry points, module boundaries, and important directories. Stores the result in the local project database. Run this first in any project before using other commands.

### `recall cache build`

Indexes all source files with summaries and embeddings.

```bash
recall cache build
# Or with custom summary length
recall cache build --sentences 5
```

Walks the project tree, skipping ignored directories. For each file:
1. Computes a SHA256 hash
2. Generates a summary using extractive summarization (LexRank)
3. Stores content for FTS5 indexing
4. Splits into chunks and computes vector embeddings
5. Groups files into heuristic subsystems

### `recall cache inspect`

Views cached data:

```bash
recall cache inspect
```

Output includes:
- Project map (language, framework, signals, entry points)
- Indexed files with truncated summaries
- Detected subsystems
- Stored memories

### `recall cache refresh`

Incremental update: only processes files whose hash has changed.

```bash
recall cache refresh
```

### `recall cache invalidate`

Removes stale entries where the underlying file has changed or been deleted.

```bash
recall cache invalidate
```

### `recall learn "insight"`

Stores a manual insight into project memory.

```bash
recall learn "always use context.WithTimeout for external API calls"
```

These insights appear in `cache inspect` and can be referenced by the brief generator.

## Knowledge Retrieval

### `recall brief "task description"`

Generates an enriched prompt context for AI assistants.

```bash
recall brief "add OAuth login"
recall brief "debug slow database queries"
```

The brief includes:
- Project metadata (language, framework, package manager)
- Top 3 matching task recipes from global.db
- Relevant subsystem summaries
- Recently indexed files matching the task keywords

If no recipes are loaded, prints a warning but still produces the project context section.

### `recall query "question"`

Smart router that tries a local LLM first, then falls back to an enriched brief for cloud assistants.

```bash
# Try local model, delegate if it can't answer
recall query "how do I add health checks to this service"

# Skip local model, just produce enriched brief
recall query --delegate "how do I add health checks"

# Show which local model would be used
recall query --local-only "how do I add health checks"

# Override timeout (default: 30s)
recall query --timeout 10 "quick question"
```

Responses are automatically saved to the global brain. Code blocks in successful responses are extracted as reusable snippets.

### `recall search "terms"`

Full-text search over indexed content with vector re-ranking.

```bash
recall search "auth middleware"
recall search -n 20 "database migration"   # top 20 results
```

Uses SQLite FTS5 for candidate retrieval, then re-ranks by cosine similarity of chunk embeddings.

### `recall search -c "terms"`

Chunk-level semantic search.

```bash
recall search -c "error handling pattern"
```

Returns individual text chunks rather than whole files, useful for finding specific code patterns.

### `recall index <directory>`

Indexes an arbitrary directory into the current project's database.

```bash
recall index ./docs
recall index ../shared-lib
```

Useful for bringing external documentation or shared libraries into the searchable context.

### `recall tldr`

Summarizes piped text without any LLM calls.

```bash
cat long-file.md | recall tldr
# Or with specific sentence count
cat long-file.md | recall tldr --sentences 5
```

Uses the LexRank algorithm for extractive summarization. Purely local, no network, no tokens.

## Global Brain

### `recall brain conversations`

Lists recent interactions with local models.

```bash
recall brain conversations
recall brain conversations -n 50   # show 50
```

Shows: timestamp, model name, whether answered or delegated, and a truncated preview.

### `recall brain snippets`

Lists reusable code blocks extracted from successful local model responses.

```bash
recall brain snippets
```

Each snippet shows: language, framework, usage count, context description, and the code block.

### `recall brain lessons`

Shows patterns the brain has learned about what works.

```bash
recall brain lessons
```

Examples:
- "local model llama3.2 for go tasks" (success rate tracked)
- "delegate qwen3.6-27b to big model for nextjs" (when local model consistently fails)

### `recall brain stats`

Aggregate metrics across all brain data.

```bash
recall brain stats
```

Shows:
- Total conversations
- Answered vs. delegated counts
- Success rate (answered / total)
- Estimated tokens saved
- Top 5 snippets by usage
- Top 5 lessons by success rate
- Per-framework performance breakdown

### `recall brain search "keywords"`

Searches across all brain tables at once.

```bash
recall brain search "auth"
recall brain search "health check" -n 10
```

Returns matching conversations, snippets, and lessons in a single view.

### `recall brain search -v "keywords"`

Vector search over snippets.

```bash
recall brain search -v "middleware pattern"
```

Computes an embedding for the query and returns snippets by cosine similarity rather than keyword matching.

### `recall brain frameworks`

Per-framework breakdown of local model performance.

```bash
recall brain frameworks
```

Shows: framework, total queries, answered, delegated, success rate, and top delegation reason.

## Recipes

### `recall recipes seed`

Loads default recipes into global.db.

```bash
recall recipes seed
```

On first run, copies recipe JSON files from the binary's embedded defaults to `~/.recall/recipes/`. On subsequent runs, syncs new defaults without overwriting user modifications. Skips duplicates by recipe name.

### `recall recipes list`

Shows all loaded recipes.

```bash
recall recipes list
```

Output format: `name (language/framework) [source]` with a total count.

### `recall recipes add --from-file <path>`

Adds a single custom recipe.

```bash
recall recipes add --from-file ./my-auth-recipe.json
```

Validates required fields (`name`, `brief_template`) and exits non-zero on invalid input.

## Local LLM Management

### `recall local status`

Detects and reports local LLM availability.

```bash
recall local status
```

Probes common endpoints (localhost:8080, 8000, 5000, 11434) for OpenAI-compatible APIs. If found, shows:
- Server type (llama.app, generic OpenAI)
- Available models with the auto-selected one marked
- Install instructions if no server is found

### `recall local models`

Lists available models without other status info.

```bash
recall local models
```

### `recall local use <model>`

Locks Recall to a specific model.

```bash
recall local use gemma-4-E4B
recall local use qwen3.6-1b
```

Persists to `~/.recall/config.json`. Exits non-zero if the model doesn't exist in the current server.

The auto-selection logic (when no preference is set):
1. Prefers small models: names containing `1b`, `2b`, `3b`, `4b`, `nano`, `tiny`, `mini`, `small`
2. Avoids large models: names containing `27b`, `35b`, `70b`, `123b`, `198b`, `MoE`
3. Falls back to the first non-large model

## Recording & Stats

### `recall run suggest`

Gated recording of an assistant run.

```bash
recall run suggest \
  --task "add OAuth login" \
  --files-changed "src/lib/auth.ts,src/app/login/page.tsx" \
  --tokens-in 4000 \
  --tokens-out 800 \
  --tests-passed 1
```

Prompts with a one-line preview and waits for `y` (save), `n` (skip), or `i` (save with insight).

### `recall run record`

Direct recording without interactive gate.

```bash
recall run record \
  --task "add OAuth login" \
  --files-changed "src/lib/auth.ts" \
  --tokens-in 4000 \
  --tokens-out 800
```

Useful in scripts and CI where interaction is not possible.

### `recall stats cache`

Shows project cache statistics.

```bash
recall stats cache
```

### `recall stats recipes`

Shows recipe usage statistics.

```bash
recall stats recipes
```

### `recall stats runs`

Shows aggregated run statistics per framework.

```bash
recall stats runs
```

### `recall stats global`

Cross-project global statistics.

```bash
recall stats global
```

### `recall cleanup`

Removes old entries.

```bash
recall cleanup           # default: 30 days
recall cleanup --days 7  # keep last 7 days
```

## Skill Installation

### `recall install-skill --target <assistant>`

Installs a thin skill file into the assistant's configuration.

```bash
recall install-skill --target claude
recall install-skill --target opencode
recall install-skill --target cursor
recall install-skill --target codex
```

Each skill instructs the assistant to:
1. Run `recall brief` at task start and prepend the output to context
2. Run `recall run suggest` at task end to record outcomes

Supported targets:
- **Claude Code**: `~/.claude/skills/recall/SKILL.md`
- **opencode**: `~/.opencode/skills/recall/SKILL.md`
- **Cursor**: `~/.cursor/skills/recall/SKILL.md`
- **Codex**: `~/.codex/skills/recall/SKILL.md`

Environment variables for custom config directories:
- `CLAUDE_CONFIG_DIR`
- `OPENCODE_CONFIG_DIR`
