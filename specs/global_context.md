Translate it by separating project-specific facts from reusable patterns.

Local cache = “what is true in this repo?”
Global cache = “what patterns keep being true across repos?”

Global cache should store

Framework patterns
- Next.js App Router auth structure
- Prisma migration conventions
- Go CLI layout
- Rust workspace layout
- FastAPI service patterns
Task recipes
- add OAuth to Next.js
- add health check to Go service
- add sqlite migrations
- add GitHub Actions CI
- add structured logging
Reusable brief templates
- debugging brief
- feature implementation brief
- refactor brief
- security review brief
- test generation brief
Known tool behavior
- Claude Code works better with file map + constraints
- Codex works better with explicit edit targets
- small models are good at summarizing, weak at architecture
Token heuristics
- auth tasks usually need routes + middleware + schema
- DB tasks need models + migrations + config
- UI tasks need component + style + state files

What should not go global

Secrets
.env files
private customer data
private code snippets
proprietary architecture details
full repo chunks by default
sensitive logs

Global cache should usually store schemas, summaries, fingerprints, and recipes, not raw private code.

Architecture

Local Project Cache
  - raw chunks
  - file summaries
  - project map
  - repo-specific memories
  - run history
Global Cache
  - reusable patterns
  - task recipes
  - framework maps
  - prompt templates
  - token-performance statistics

At runtime:

User task
  ↓
Detect project type
  ↓
Fetch global recipe
  ↓
Fetch local project facts
  ↓
Merge into brief
  ↓
Send to Claude/Codex
  ↓
Record outcome
  ↓
Update local + global stats

Example:

pipecamp brief "add OAuth login"

Global cache contributes:

OAuth recipe for Next.js App Router:
- check middleware
- check auth provider
- check user/session schema
- check callback route
- check env vars
- check login UI

Local cache contributes:

This repo:
- uses Prisma
- auth is in src/lib/auth.ts
- middleware is src/middleware.ts
- login page is src/app/login/page.tsx
- env config is in src/env.ts

Final brief becomes much better and smaller.

Practical implementation

Use two SQLite DBs:

~/.alt/global.db
~/.alt/projects/<project-hash>/project.db

Global tables:

patterns
task_recipes
prompt_templates
framework_fingerprints
model_behavior_stats
tool_profiles

Local tables:

files
chunks
file_summaries
subsystem_summaries
runs
run_context_items
memories

The key object: a reusable recipe

{
  "kind": "task_recipe",
  "name": "add_oauth_nextjs_app_router",
  "applies_to": {
    "language": "typescript",
    "framework": "nextjs",
    "signals": ["app/", "middleware.ts", "prisma/schema.prisma"]
  },
  "context_needed": [
    "auth module",
    "middleware",
    "user schema",
    "env config",
    "login page",
    "callback route"
  ],
  "avoid": [
    "do not send all pages",
    "do not send generated .next",
    "do not send node_modules"
  ],
  "brief_template": "..."
}

How it learns

After every run, record:

Task type
Framework
Files included
Files actually changed
Tests passed
Follow-up needed
Input tokens
Output tokens
Accepted/rejected

Then update global heuristics:

For Next.js OAuth tasks:
- middleware.ts was useful 82% of time
- prisma/schema.prisma was useful 76%
- unrelated UI components were useful 9%

So next time, retrieval is smarter.

The important shift:

Local cache remembers the project.
Global cache remembers how to work on projects like it.
