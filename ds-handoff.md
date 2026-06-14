# Handoff — recall (local AI context engine)

## Repo
`/Users/gleicon/code/go/src/github.com/gleicon/recall`  
Branch: `main` — working tree clean as of handoff.

---

## What this project is

`recall` is a CLI tool that caches project context locally so AI-assisted dev sessions cost fewer tokens and preserve cross-session knowledge. Core loop:

1. `recall map` — scans codebase, stores project metadata + signals in per-project SQLite DB
2. `recall brief <task>` — emits a compressed context brief (subsystems, recipes, relevant files, prior conversations) for injection into AI prompts
3. `recall query <prompt>` — enriches prompt with context, routes to local LLM or prints DELEGATE for big model
4. `recall feedback --good/--bad` — marks last conversation; drives quality signal

Data lives in `~/.recall/`:
- `global.db` — snippets, lessons, conversations, recipes
- `projects/<hash>/project.db` — per-project map, files, embeddings

---

## Work done this session

### Fixes applied (all committed, clean tree)

| Commit | What |
|---|---|
| `8348630` | Port list + JSON parse struct deduplication in `internal/llm/client.go` |
| `c4f53b1` | Multi-endpoint detection (LM Studio, Ollama, vLLM, llama.cpp); `--endpoint` CLI flag; `detect_timeout` config; `ProbeAll()` for `recall local status` |
| `e34d10b` | Version stamping via ldflags (`make`) + `debug.ReadBuildInfo()` fallback for `go install` |
| `77e7c41` | `recall map` stores timestamp; `recall status` JSON output |
| `8e836eb` | Real vector search (FTS5 + cosine rerank); brain recycling in brief; token-recall metrics; deslop pass |

### This session's mechanical cleanup (also committed in `8348630`)
- `defaultEndpoints []string` var — `Detect()` and `ProbeAll()` share it (was hardcoded twice)
- `parseModelList(body []byte)` helper — eliminates identical anonymous struct in `detectOpenAI` + `probeEndpoint`
- `local status` no longer double-probes: builds `*llm.Client` from first reachable `ProbeResult` instead of calling `llm.Detect()` again

---

## Architecture snapshot

```
cmd/
  brief.go       — FTS5+vector file search, brain injection, tldt compress, token stats
  query.go       — enriched prompt builder + local LLM routing
  brain.go       — global knowledge search (snippets, lessons, conversations)
  feedback.go    — mark last conversation good/bad
  status.go      — JSON status (version, mapped, mapped_at, files_indexed)
  local.go       — local LLM management (status/models/use)
  root.go        — version, PersistentPreRun wires endpoint/timeout from config

internal/
  llm/client.go          — OpenAI-compat detection, ProbeAll, Query/QueryStream/GetEmbedding
  embeddings/pluggable.go — ComputeSmart (LLM embeds → FNV fallback), ComputeSmartWithClient
  search/search.go        — NewEngine(db, embedModel); FTS5 candidates → cosine rerank
  cache/conversations.go  — Conversation struct + FindSimilarConversations + feedback helpers
  cache/cache.go          — Manager, indexFile stores embeddings, EmbedModel field
  db/schema.go            — embedding BLOB on files/conversations; accepted/feedback_note; embed_model
  config/config.go        — Settings: LocalModel, EmbedModel, LocalEndpoint, QueryTimeout, DetectTimeout
```

---

## Open items / next steps

1. **Tag + release** — last `go install @v0.1.3` showed "dev" (tag predated `debug.ReadBuildInfo` fix). Need a new tag (`v0.1.4` or higher) after committing these changes so `go install @latest` returns the real version. Do: `git tag v0.1.4 && git push origin v0.1.4`.

2. **Makefile + goreleaser** — `.goreleaser.yaml` has `ldflags: -X cmd.Version={{.Version}}` but it references `cmd.Version` not `github.com/gleicon/recall/cmd.Version`. Verify the full package path is correct.

3. **Embed model UX** — `recall local use <model>` sets `local_model`; there's no `recall local embed <model>` equivalent for the embedding model. Users must edit `~/.recall/config.json` directly to set `embed_model`. Consider adding `localEmbedCmd`.

4. **`approxTokens` scope** — defined in `brief.go`, called from `query.go` (same `cmd` package — fine). If a third package ever needs it, move to `internal/token/token.go`.

5. **Test coverage** — `internal/embeddings/pluggable_test.go` exists but search, cache, and LLM probe paths have no tests. Suggested next: table-driven tests for `parseModelList` and `pickSmallestModel`.

6. **README** — version section was updated per prior session but should be re-checked to reflect the full endpoint detection list (Ollama port 11434, etc.).

---

## Suggested commands for next session

```bash
# Cut a release
git tag v0.1.4 && git push origin v0.1.4

# Verify go install picks up version
go install github.com/gleicon/recall@v0.1.4
recall --version   # should show v0.1.4

# Add embed model CLI subcommand
# → follow pattern of cmd/local.go localUseCmd
# → save to settings.EmbedModel via cfg.SaveSettings

# Run existing tests
go test ./...
```

Skill to run: `/ds-go-review` after any new code — project uses Tiger Style enforcement.
