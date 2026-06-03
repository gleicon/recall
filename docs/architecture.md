# Architecture

This document describes how Technocore's subsystems work. It is intended for contributors and advanced users who want to understand or extend the system.

## Overview

Technocore is a context engine with three layers:

1. **Ingestion Layer** — Detects project structure, indexes files, extracts summaries
2. **Knowledge Layer** — Stores recipes, snippets, lessons, and conversation history
3. **Retrieval Layer** — Matches user prompts against stored knowledge, routes to local or cloud models

All data is stored in SQLite. All computation is local. No cloud APIs are required for core functionality.

## Data Stores

### Global Database (`~/.technocore/global.db`)

| Table | Purpose |
|---|---|
| `task_recipes` | Reusable task descriptions with framework/language matching and embeddings |
| `patterns` | Framework fingerprints (signal files, conventions) |
| `prompt_templates` | Reusable prompt structures |
| `framework_fingerprints` | Detected framework metadata |
| `model_behavior_stats` | Aggregated run outcomes for learning |
| `conversations` | History of local model interactions |
| `snippets` | Reusable code blocks with semantic embeddings |
| `agent_lessons` | Learned patterns about what works per framework/model |

### Project Database (`~/.technocore/projects/<hash>/project.db`)

| Table | Purpose |
|---|---|
| `files` | Indexed file paths, hashes, content, summaries |
| `file_search` | SQLite FTS5 virtual table for full-text search |
| `chunks` | Text chunks with vector embeddings |
| `chunk_search` | FTS5 virtual table for chunk-level search |
| `file_summaries` | Per-file abstraction summaries |
| `subsystem_summaries` | Heuristic groupings (e.g., "auth", "api") |
| `runs` | Recorded assistant run metadata |
| `memories` | User insights and notes |
| `project_map` | Detected project metadata (single row) |

## Embeddings

Technocore uses a **pluggable embedding system**:

### Default: Feature Hashing

When no local model is available, embeddings are computed with a deterministic feature hashing algorithm:

1. Tokenize text into lowercase words
2. Hash each word with FNV-1a into a bucket (0-255)
3. Increment the bucket value
4. L2-normalize the resulting 256-dimensional vector

Properties:
- Pure Go, no dependencies
- Deterministic (same text always produces same embedding)
- Sub-millisecond computation
- No model weights or inference runtime
- Good for keyword-level similarity, weak on semantic relationships

### Local Model Embeddings

When an OpenAI-compatible local server is detected (llama.app, llama.cpp, etc), Technocore can request embeddings via the `/v1/embeddings` endpoint. This requires:
- A running local model server
- An embedding-capable model loaded (e.g., `nomic-embed-text`)

The `ComputeSmart()` function tries the local model first, falls back to feature hashing on any error.

## Search

### File Search

Two-stage hybrid search:

1. **Candidate retrieval**: SQLite FTS5 query over `file_search` or `chunk_search`
2. **Re-ranking**: Compute cosine similarity between query embedding and candidate content embedding

Final score: `0.5 + 0.5 * cosine_similarity`

### Recipe Matching

1. Compute query embedding
2. If project framework is known, pre-filter to recipes matching that framework
3. Compute cosine similarity between query and each recipe embedding
4. Apply scoring bonuses:
   - Framework match: +0.3
   - Signal overlap (file exists in project): +0.1 per matched signal
5. Sort by score, return top-k

## Project Detection

The detector (`internal/project/detector.go`) analyzes the project root for characteristic files:

| File | Inference |
|---|---|
| `package.json` | Node.js/TypeScript project |
| `go.mod` | Go project |
| `Cargo.toml` | Rust project |
| `pyproject.toml`, `requirements.txt` | Python project |
| `prisma/schema.prisma` | Prisma ORM usage |

Frameworks are identified by dependency names in `package.json` or `pyproject.toml`.

## Brief Generation

The `brief` command assembles context from multiple sources:

1. **Project map** — language, framework, package manager, entry points, signals
2. **Recipes** — top 3 matching recipes by vector + framework + signal scoring
3. **Subsystems** — stored subsystem summaries from cache build
4. **Relevant files** — files whose summary or path matches the task keywords

Output is plain Markdown, structured for easy consumption by AI assistants.

## Query Routing

The `query` command implements a decision tree:

```
User prompt
    |
    v
Build brief (always)
    |
    v
Detect local LLM?
    |-- No -> Output brief + DELEGATE marker
    |
    v
Send to local model with system context + brief
    |
    v
Parse response
    |-- Contains "DELEGATE" -> Output brief + reason
    |-- Otherwise -> Output answer + extract snippets
    |
    v
Save conversation to brain
```

## Learning

Technocore learns from three sources:

### Recipe Feedback

Each time a recipe appears in a brief, its `use_count` increments. Over time, frequently used recipes bubble to the top of `stats recipes`.

### Conversation Outcomes

Every `query` stores:
- Prompt and response
- Whether the local model answered or delegated
- Delegation reason (extracted from "DELEGATE" response)
- Model name used
- Framework of the project

This enables per-framework success rate tracking.

### Snippet Extraction

Successful local model responses are scanned for fenced code blocks (```language\ncode```). Each block is stored as a snippet with:
- Detected language
- Project framework
- Surrounding context (truncated prompt/response)
- Vector embedding

## Build Pipeline

```
go.mod / package.json detected
    |
    v
Project map stored
    |
    v
File walk (respecting .gitignore patterns)
    |
    v
For each source file:
    - Compute SHA256 hash
    - Generate summary (LexRank extractive summarization)
    - Chunk text (1500 char chunks with line awareness)
    - Compute embedding per chunk
    - Store in files, chunks, file_search, chunk_search tables
    |
    v
Heuristic subsystem grouping (by directory structure)
```

## Dependencies

| Package | Purpose | License |
|---|---|---|
| `modernc.org/sqlite` | Pure Go SQLite (no CGO) | BSD-3 |
| `github.com/spf13/cobra` | CLI framework | Apache-2.0 |
| `github.com/gleicon/tldt` | Extractive text summarization | MIT |

No neural network libraries, no vector database clients, no cloud SDKs.

## Extension Points

To add a new embedding backend:

1. Implement the embedding interface in `internal/embeddings/`
2. Register it in `ComputeSmart()` with fallback logic

To add a new recipe source:

1. Add JSON files to `recipes/` directory
2. Run `technocore recipes seed`

To add a new assistant skill target:

1. Create `skills/<name>/SKILL.md`
2. Add case to `cmd/install_skill.go`

To add a new project detector heuristic:

1. Extend `internal/project/detector.go`
2. Add framework detection logic based on characteristic files
