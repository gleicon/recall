## Technocore

![](pipecamp.jpg)

Technocore is a local and global context storage system for AI-assisted development.

- **Local cache** = "what is true in this repo?"
- **Global cache** = "what patterns keep being true across repos?"

It caches project abstractions, indexes files, provides RAG + vector search,
and generates task briefs to save tokens and avoid unnecessary model roundtrips.

### Commands

#### Project & Cache
- `technocore map` — detect and store the project map
- `technocore cache build` — index the project and build the abstraction cache
- `technocore cache inspect` — inspect cached files, subsystems, and memories
- `technocore cache refresh` — update only changed files
- `technocore cache invalidate` — remove stale entries

#### Briefs & Recipes
- `technocore recipes seed` — load default recipes into global.db
- `technocore recipes add --from-file <path>` — add a custom recipe JSON
- `technocore recipes list` — list all recipes
- `technocore brief "add OAuth login"` — generate a task brief from global recipes + local facts

#### Search & Summarize
- `technocore index ./docs` — index an arbitrary directory
- `technocore search <terms>` — search indexed content (FTS5 + vector re-ranking)
- `technocore search -c <terms>` — chunk-level semantic search
- `cat file.txt | technocore tldr` — summarize text without LLM calls

#### Learning & Stats
- `technocore learn "important insight"` — store an insight into local memory
- `technocore run suggest --task "..." --files-changed "a,b"` — gated run recording
- `technocore run record --task "..." --files-changed "a,b"` — manual run recording
- `technocore stats cache` — show cache statistics
- `technocore stats recipes` — show recipe usage statistics
- `technocore stats runs` — show run statistics
- `technocore cleanup` — remove old entries

#### Skills
- `technocore install-skill --target claude` — install Claude Code skill
- `technocore install-skill --target opencode` — install opencode skill

### Design

Two SQLite databases:
- `~/.technocore/global.db` — reusable patterns, task recipes, framework fingerprints, model behavior stats
- `~/.technocore/projects/<hash>/project.db` — file summaries, chunks, embeddings, subsystem summaries, runs, memories

Vector embeddings are computed with lightweight feature hashing (256 dims) for pure-Go, offline semantic similarity.

Summarization is powered by [tldt](https://github.com/gleicon/tldt) — no API keys, no token costs.

Recipes are stored as JSON files in `~/.technocore/recipes/` and loaded into `global.db` with embeddings for vector search.

### Recipe JSON Format

```json
{
  "name": "add_oauth_nextjs",
  "framework": "nextjs",
  "language": "typescript",
  "signals": ["app/", "middleware.ts", "prisma/schema.prisma"],
  "context_needed": ["auth module", "middleware", "user schema"],
  "avoid": ["do not send all pages"],
  "brief_template": "To add OAuth to this Next.js App Router project:\n1. Check src/lib/auth.ts...",
  "source": "pipecamp-defaults",
  "tags": ["auth", "oauth", "nextjs"]
}
```
