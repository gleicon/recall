# Manual Data Management

Recall accumulates data automatically, but you can also add, edit, and curate data manually.

## Adding Recipes

Recipes are the core reusable knowledge unit. They describe how to perform a common task for a specific framework.

### Recipe JSON Format

```json
{
  "name": "add_oauth_nextjs",
  "framework": "nextjs",
  "language": "typescript",
  "signals": ["app/", "middleware.ts", "prisma/schema.prisma"],
  "context_needed": ["auth module", "middleware", "user schema"],
  "avoid": ["do not send all pages", "do not send generated .next"],
  "brief_template": "To add OAuth to this Next.js App Router project:\n1. Check src/lib/auth.ts for existing provider setup\n2. Check src/middleware.ts for route protection\n3. Check prisma/schema.prisma for user/session tables\n4. Add callback route at src/app/api/auth/callback/route.ts\n5. Add login UI at src/app/login/page.tsx",
  "source": "user-created",
  "tags": ["auth", "oauth", "nextjs"]
}
```

### Validation Rules

- `name` is required, must be unique within global.db
- `brief_template` is required — this is the text shown in briefs
- `framework` should match what `recall map` detects (e.g., `go`, `nextjs`, `fastapi`)
- `language` is the primary language string
- `signals` are file/directory names whose presence indicates this recipe is relevant
- `context_needed` lists files or concepts the assistant should check
- `avoid` lists things the assistant should not send (reduces token waste)
- `source` is metadata (e.g., `recall-defaults`, `user-created`, `team-shared`)
- `tags` are searchable keywords

### Adding a Recipe

```bash
# Write the JSON file
cat > ~/my-recipe.json <<'EOF'
{
  "name": "add_circuit_breaker_go",
  "framework": "go",
  "language": "go",
  "signals": ["go.mod", "internal/"],
  "context_needed": ["HTTP client", "retry logic", "metrics"],
  "avoid": ["do not change existing API signatures"],
  "brief_template": "To add circuit breaker pattern:\n1. Check existing HTTP client wrapper\n2. Choose library: sony/gobreaker or custom\n3. Configure thresholds (failures, timeout, half-open)\n4. Add metrics export for circuit state\n5. Write tests for open, closed, and half-open transitions",
  "source": "user-created",
  "tags": ["resilience", "go", "circuit-breaker"]
}
EOF

# Add to global.db
recall recipes add --from-file ~/my-recipe.json

# Verify
recall recipes list | grep circuit_breaker
```

### Editing Recipes

Recall does not have an edit command. Recipes are edited as JSON files and re-added:

```bash
# The ON CONFLICT clause updates existing recipes by name
recall recipes add --from-file ~/updated-recipe.json
```

Or use a database tool directly:

```bash
sqlite3 ~/.recall/global.db "UPDATE task_recipes SET brief_template = 'new text' WHERE name = 'add_circuit_breaker_go';"
```

### Recipe Directory

Default recipes live in `~/.recall/recipes/` as individual `.json` files. You can:

```bash
# Add new files here
cp ~/my-custom-recipe.json ~/.recall/recipes/

# Re-seed to load them (skips duplicates, updates changed ones)
recall recipes seed
```

## Adding Insights

Store institutional knowledge about a project:

```bash
# Single insight
recall learn "always pass context.Context as first argument to service methods"

# Multiple insights from a file
while read -r line; do
    [ -n "$line" ] && recall learn "$line"
done < ~/project-insights.txt

# View stored insights
recall cache inspect | grep -A1 "^\[insight\]"
```

## Adding Code Snippets

Snippets are extracted automatically from local model responses, but you can add them manually:

```bash
# Direct SQLite insert (requires understanding the schema)
sqlite3 ~/.recall/global.db <<'EOF'
INSERT INTO snippets (name, language, framework, code, context, source)
VALUES (
  'generic_retry_loop',
  'go',
  'go',
  'for i := 0; i < maxRetries; i++ {\n    err := doWork()\n    if err == nil {\n        break\n    }\n    if i == maxRetries-1 {\n        return err\n    }\n    time.Sleep(backoff * time.Duration(i+1))\n}',
  'Retry loop with linear backoff',
  'user-added'
);
EOF
```

## Adding Agent Lessons

Teach the brain about what works:

```bash
sqlite3 ~/.recall/global.db <<'EOF'
INSERT INTO agent_lessons (pattern, framework, model_name, success_rate, context)
VALUES (
  'small models handle go refactoring well',
  'go',
  'llama3.2-1b',
  0.9,
  'go tasks are well-scoped and the local model produces correct imports'
);
EOF
```

## Adding Conversations

Record a past interaction you had with an assistant:

```bash
sqlite3 ~/.recall/global.db <<'EOF'
INSERT INTO conversations (task, prompt, response, model_name, delegated, delegation_reason)
VALUES (
  'refactor auth middleware',
  'how should I refactor auth middleware to use dependency injection',
  'Use functional options pattern. Create an AuthMiddleware struct that accepts a UserStore interface. This allows testing with a mock store and switching between JWT, session, or OAuth backends without changing middleware signatures.',
  'claude',
  1,
  'needed architectural reasoning beyond local model capacity'
);
EOF
```

## Bulk Import from Markdown

If you have a markdown file of tips and patterns:

```bash
#!/bin/bash
FILE="$1"
FRAMEWORK="$2"

awk '/^## /{name=$2} /^- /{print name "|" substr($0,3)}' "$FILE" | while IFS='|' read -r name tip; do
    recall learn "[$name] $tip"
done
```

Example input (`tips.md`):

```markdown
## API Design
- Always return structured errors: `{ error: { code, message, details } }`
- Use 409 Conflict for duplicate resource creation attempts

## Testing
- Table-driven tests for all service methods
- Use httptest.Server for HTTP client tests, not real network calls
```

## Exporting Data

```bash
# Export all recipes as JSON
sqlite3 ~/.recall/global.db "SELECT json_object('name', name, 'framework', framework, 'brief_template', brief_template) FROM task_recipes;" > recipes-export.jsonl

# Export all snippets
cd ~/.recall
sqlite3 global.db ".mode json" "SELECT * FROM snippets;" > snippets-export.json

# Export conversation history
sqlite3 global.db ".mode csv" "SELECT task, model_name, delegated, created_at FROM conversations;" > conversations.csv
```

## Importing Data

```bash
# Import recipes from JSONL
while read -r line; do
    echo "$line" > /tmp/recipe.json
    recall recipes add --from-file /tmp/recipe.json
done < recipes-import.jsonl
```
