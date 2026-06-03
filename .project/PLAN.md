# Plan: Pipecamp Recipes, Learning Loop, and Skill Installer

## Overview

Implement the three milestones defined in `.project/SPEC.md`:
1. Global recipe seed with vector embeddings and `brief` RAG upgrade
2. Learning / feedback loop (`run suggest`, `run record`, stats)
3. Assistant skill installer (`install-skill`) with thin skill files

## Phase 1: Schema & Infrastructure

**Goal:** Update the database schema to support recipes with embeddings and run stats.

### Task 1.1: Add `embedding` to `task_recipes`
- **File:** `internal/db/schema.go`
- **Action:** Add `embedding BLOB` to `task_recipes` table. Add `embedding BLOB` to `patterns` table (future-proofing).
- **Depends on:** Nothing
- **Verify:** Fresh `global.db` created with `embedding` column

### Task 1.2: Add `use_count` and `avg_score` to `task_recipes`
- **Action:** Add `use_count INTEGER DEFAULT 0` and `avg_score REAL DEFAULT 0` for stats.
- **Depends on:** 1.1
- **Verify:** Schema inspection shows new columns

## Phase 2: Recipe Data Model & Persistence

**Goal:** Build a clean Go struct and CRUD layer for recipes.

### Task 2.1: Define `Recipe` struct
- **File:** `internal/recipes/recipe.go`
- **Fields:** Name, Framework, Language, Signals, ContextNeeded, Avoid, BriefTemplate, Source, Tags, Embedding
- **Depends on:** 1.2
- **Verify:** Compiles

### Task 2.2: Implement `StoreRecipe`
- **Action:** Insert recipe into `task_recipes`. Compute embedding via `embeddings.Compute(name+" "+briefTemplate)`. Skip if `name` exists.
- **Depends on:** 2.1
- **Verify:** Insert + dedup works

### Task 2.3: Implement `FindRecipesByVector`
- **Action:** Scan all recipes, compute cosine similarity with query embedding, return top-k.
- **Depends on:** 2.2
- **Verify:** `FindRecipesByVector("add OAuth", 3)` returns OAuth recipe

### Task 2.4: Implement `FindRecipesByFramework`
- **Action:** Fallback query `WHERE framework = ?`.
- **Depends on:** 2.2
- **Verify:** Returns framework-specific recipes

## Phase 3: Recipe Seeding

**Goal:** Ship default recipes and load them into the DB.

### Task 3.1: Create `recipes/` directory and JSON files
- **Files:**
  - `recipes/nextjs_oauth.json`
  - `recipes/nextjs_prisma_migration.json`
  - `recipes/nextjs_middleware.json`
  - `recipes/go_healthcheck.json`
  - `recipes/go_sqlite_migrations.json`
  - `recipes/go_structured_logging.json`
  - `recipes/fastapi_crud.json`
  - `recipes/fastapi_pydantic.json`
  - `recipes/rust_axum_route.json`
  - `recipes/rust_sqlx_migration.json`
  - `recipes/github_actions_ci.json`
  - `recipes/add_dockerfile.json`
  - `recipes/debug_slow_query.json`
  - `recipes/python_add_tests.json`
  - `recipes/go_add_tests.json`
- **Depends on:** Nothing
- **Verify:** All 15 files exist and validate against JSON schema

### Task 3.2: Implement `recipes seed` command
- **File:** `cmd/recipes.go`
- **Action:** Check if `~/.pipecamp/recipes/` exists. If not, copy repo `recipes/` there. Read all `.json` files, validate, call `StoreRecipe` for each.
- **Depends on:** 3.1, 2.3
- **Verify:** `pipecamp recipes seed` → `global.db` has 15 recipes

### Task 3.3: Implement `recipes add --from-file` command
- **Action:** Validate single JSON, call `StoreRecipe`.
- **Depends on:** 2.2
- **Verify:** Invalid JSON exits non-zero with clear error

### Task 3.4: Implement `recipes list` command
- **Action:** Print all recipes from `global.db`.
- **Depends on:** 2.2
- **Verify:** Lists names + frameworks

## Phase 4: Upgrade `brief` to Vector RAG

**Goal:** Make `brief` use recipe embeddings + framework/signal boosting.

### Task 4.1: Refactor `brief` command
- **File:** `cmd/brief.go`
- **Action:**
  1. Compute query embedding
  2. Call `FindRecipesByVector(query, 10)`
  3. For each recipe: score = cosineSimilarity + (frameworkMatch ? 0.3 : 0) + (matchedSignals * 0.1), cap at 1.0
  4. Sort, take top 3
  5. Include in brief output under `## Recipes`
- **Depends on:** 2.3, 3.2
- **Verify:** `pipecamp brief "add OAuth"` shows `nextjs_oauth` recipe in a Go repo? No — should show only if framework matches or vector similarity is high enough.

### Task 4.2: Fallback behavior
- **Action:** If no recipes match (all scores < 0.3), print a subtle warning to stderr: `No relevant recipes found. Run 'pipecamp recipes seed' to load defaults.`
- **Depends on:** 4.1
- **Verify:** Empty global.db → brief prints warning but still outputs project map

## Phase 5: Learning Loop — Data Model

**Goal:** Store run metadata and compute stats.

### Task 5.1: Extend `runs` and `model_behavior_stats` schemas
- **File:** `internal/db/schema.go`
- **Action:** Ensure `runs` and `model_behavior_stats` tables cover all fields from FR-10.
- **Depends on:** Nothing (schema already mostly there)
- **Verify:** Fresh DB has correct columns

### Task 5.2: Implement `StoreRun`
- **File:** `internal/cache/cache.go`
- **Action:** Insert into both `runs` (local) and `model_behavior_stats` (global).
- **Depends on:** 5.1
- **Verify:** Insert + read back works

### Task 5.3: Implement run stats aggregation
- **Action:** SQL queries for:
  - Per-framework average token reduction
  - Per-task-type file usefulness (files_changed / files_included ratio)
  - Recipe hit counts
- **Depends on:** 5.2
- **Verify:** `pipecamp stats runs` prints meaningful numbers after test data inserted

## Phase 6: Learning Loop — CLI

**Goal:** Build `run suggest` and `run record` commands.

### Task 6.1: Implement `run suggest` (gated)
- **File:** `cmd/run.go`
- **Action:**
  1. Accept flags: `--task`, `--files-changed`, `--tokens-in`, `--tokens-out`, `--tests-passed`, `--follow-up-needed`
  2. Print one-line preview
  3. Read single char from stdin: `y`, `n`, `i`
  4. `y` → call `StoreRun`
  5. `i` → prompt for insight text, append to `content` field of a `memory`, then call `StoreRun`
  6. `n` → exit 0, no DB write
- **Depends on:** 5.2
- **Verify:** Interactive test with `y` and `i` paths

### Task 6.2: Implement `run record` (manual)
- **Action:** Same flags as `run suggest`, but skip the gate — directly write to DB.
- **Depends on:** 5.2
- **Verify:** Non-interactive, exits 0, DB has row

### Task 6.3: Implement `stats recipes`
- **Action:** Query `task_recipes` for `use_count` and `avg_score`. Print table.
- **Depends on:** 2.2
- **Verify:** Shows recipe usage

### Task 6.4: Implement `stats runs`
- **Action:** Query `model_behavior_stats` and `runs`. Print per-framework token efficiency and file usefulness.
- **Depends on:** 5.3
- **Verify:** Shows aggregated stats

## Phase 7: Skill Installer

**Goal:** Install thin assistant skills.

### Task 7.1: Research assistant skill formats
- **Action:** Determine exact file paths and formats for:
  - Claude Code: `~/.claude/skills/pipecamp/` with `SKILL.md` + `hooks/`
  - opencode: `.opencode/skills/pipecamp/` with `SKILL.md`
  - Cursor: `.cursor/skills/pipecamp/` (if supported)
  - Codex: `~/.codex/skills/pipecamp/`
- **Depends on:** Nothing
- **Verify:** File exists in correct location after install

### Task 7.2: Create skill templates
- **Files:**
  - `skills/claude/SKILL.md`
  - `skills/opencode/SKILL.md`
  - `skills/cursor/SKILL.md` (or skip if unsupported)
- **Content:** Markdown skill definition that instructs assistant to:
  - At task start: `pipecamp brief "<user prompt>"` → prepend to context
  - At task end: `pipecamp run suggest --task "<summary>" --files-changed "..."` → respect user response
- **Depends on:** 7.1
- **Verify:** Skill file is valid markdown and references correct commands

### Task 7.3: Implement `install-skill` command
- **File:** `cmd/install_skill.go`
- **Action:**
  1. `--target claude` → copy `skills/claude/` to `~/.claude/skills/pipecamp/`
  2. `--target opencode` → copy to `.opencode/skills/pipecamp/`
  3. `--target cursor` → copy to `.cursor/skills/pipecamp/`
  4. Create directories if missing
  5. Print success message with path
- **Depends on:** 7.2
- **Verify:** `pipecamp install-skill --target claude` → skill files exist

### Task 7.4: Auto-detect assistant config dirs
- **Action:** Use env vars (`$CLAUDE_CONFIG_DIR`, `$OPENCODE_CONFIG_DIR`) with fallbacks to `~/.claude/` and `~/.opencode/`.
- **Depends on:** 7.3
- **Verify:** Respects env overrides

## Phase 8: Integration & Verification

**Goal:** Everything works end-to-end.

### Task 8.1: Build and smoke test
- **Action:** `go build`, `pipecamp recipes seed`, `pipecamp brief "add OAuth"`, `pipecamp run suggest ...`, `pipecamp stats runs`
- **Depends on:** All previous tasks
- **Verify:** No panics, outputs look correct

### Task 8.2: Acceptance Criteria Checklist
- **AC-1:** `recipes seed` → 15 recipes in DB with embeddings ✅
- **AC-2:** `brief "add OAuth"` in Next.js context → shows OAuth recipe ✅
- **AC-3:** `brief` with 100 recipes → < 500 ms ✅
- **AC-4:** `run suggest` → gate works, `y` stores record ✅
- **AC-5:** `stats runs` → shows framework stats ✅
- **AC-6:** `install-skill --target claude` → skill files exist ✅
- **AC-7:** Skill prepend brief at task start ✅ (manual verification)
- **AC-8:** Skill suggest recording at task end ✅ (manual verification)
- **AC-9:** `recipes add --from-file bad.json` → non-zero exit + error ✅
- **AC-10:** Duplicate `recipes seed` → no dupes ✅

### Task 8.3: Documentation update
- **Action:** Update `README.md` with new commands: `recipes`, `run`, `stats`, `install-skill`
- **Depends on:** 8.2
- **Verify:** README is accurate

## Task Dependency Graph

```
1.1 Schema embedding
   ↓
1.2 Schema stats cols
   ↓
2.1 Recipe struct
   ↓
2.2 StoreRecipe
   ↓
2.3 FindRecipesByVector
   ↓
2.4 FindRecipesByFramework
   ↓
3.1 JSON recipe files
   ↓
3.2 recipes seed cmd ←→ 2.3
3.3 recipes add cmd ←→ 2.2
3.4 recipes list cmd ←→ 2.2
   ↓
4.1 brief RAG ←→ 2.3, 3.2
4.2 brief fallback ←→ 4.1
   ↓
5.1 runs schema
   ↓
5.2 StoreRun
   ↓
5.3 Stats aggregation
   ↓
6.1 run suggest ←→ 5.2
6.2 run record ←→ 5.2
6.3 stats recipes ←→ 2.2
6.4 stats runs ←→ 5.3
   ↓
7.1 Research skill formats
   ↓
7.2 Skill templates
   ↓
7.3 install-skill cmd ←→ 7.2
7.4 Auto-detect dirs ←→ 7.3
   ↓
8.1 Build + smoke test ←→ ALL
8.2 AC checklist ←→ 8.1
8.3 README update ←→ 8.2
```

## Waves

**Wave 1 (Schema + Recipes):** Tasks 1.1–1.2, 2.1–2.4, 3.1–3.4, 4.1–4.2
**Wave 2 (Learning):** Tasks 5.1–5.3, 6.1–6.4
**Wave 3 (Skills):** Tasks 7.1–7.4
**Wave 4 (Integration):** Tasks 8.1–8.3

## Risks & Mitigations

| Risk | Mitigation |
|---|---|
| Skill format changes across assistants | Keep skill templates minimal; document that formats may need updating |
| Vector scan of 1000+ recipes is slow | Add `WHERE framework = ?` pre-filter before cosine scan |
| User forgets to run `recipes seed` | Print warning in `brief` if no recipes exist |
| Interactive `run suggest` breaks scripts | `run record` is the non-interactive alternative |
