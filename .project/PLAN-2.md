# Plan: Hooks, Recipes, Cross-Project Learning, Benchmarking

## Wave 1: Hook Integration

Real hook integration means assistant automatically calls pipecamp without user prompting.

### Task 1.1: Claude Code hooks
- Create `skills/claude/hook.sh` — shell script that runs `pipecamp brief` and captures output
- Update `install-skill` to write `~/.claude/settings.local.json` hook entry
- Hook runs on `UserPromptSubmit` event

### Task 1.2: Codex hooks  
- Same pattern as Claude Code (`~/.codex/settings.json`)

### Task 1.3: opencode advisory plugin
- opencode uses SKILL.md only — already done, but add a `plugin.md` for deeper integration

### Task 1.4: Cursor integration
- Cursor uses `.cursorrules` file. Add a `.cursorrules` snippet that instructs Cursor to use pipecamp.
- Also support Cursor's command palette approach.

## Wave 2: Recipe Expansion

Add 15 more recipes:
- Django CRUD, Django auth
- Rails scaffold, Rails migration
- Kubernetes deployment, service, ingress
- Terraform AWS module
- React Native navigation
- Flutter state management
- SvelteKit routing
- Vue Pinia store
- Docker Compose stack
- Redis caching layer
- PostgreSQL indexing
- API rate limiting
- Webhook endpoint
- Cron job/scheduled task
- Background worker queue

## Wave 3: Cross-Project Learning

### Task 3.1: Aggregate stats across all projects
- Walk `~/.pipecamp/projects/*/project.db`
- Read `runs` tables from each
- Insert aggregated data into `global.db` `model_behavior_stats`

### Task 3.2: Cross-project insights command
- `pipecamp stats global` — show per-framework stats across all projects
- `pipecamp stats insights` — show most/least useful recipes globally

### Task 3.3: Update recipe scores from global stats
- When recipe is used, increment `use_count`
- Compute `avg_score` from run success rates

## Wave 4: Benchmarking

### Task 4.1: Benchmark command
- `pipecamp bench` — run benchmarks:
  - `brief` latency with 100, 500, 1000 recipes
  - `search` latency with 100, 1000, 10000 files
  - Vector computation throughput

### Task 4.2: Optimizations based on results
- If scan is too slow at 1000 recipes, add FTS5 index on task_recipes or cache vectors in memory
- If search is slow, optimize chunking strategy

## Dependency Graph

```
1.1 Claude hooks
1.2 Codex hooks
1.3 opencode plugin
1.4 Cursor rules
   ↓
2.1 Recipe JSON files (15 new)
   ↓
3.1 Cross-project aggregation
3.2 Stats commands
3.3 Recipe scoring
   ↓
4.1 Benchmark command
4.2 Optimizations
```
