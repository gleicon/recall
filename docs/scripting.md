# Scripting and Automation

Technocore is designed for shell integration. Every command exits with conventional codes and produces machine-parseable output.

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error (invalid input, missing data, etc.) |
| Non-zero | Subcommand-specific errors passed through |

## Parsing Output

Most commands produce plain text. Use `grep`, `awk`, `sed`, or `jq` (for JSON recipe files) to parse.

```bash
# Extract just the recipe names
technocore recipes list | grep '^- ' | sed 's/^- //' | cut -d' ' -f1

# Check if any recipes exist
if technocore recipes list | grep -q 'Total: 0'; then
    technocore recipes seed
fi

# Get the framework of the current project
technocore cache inspect | grep '^framework:' | cut -d' ' -f2
```

## Pre-commit Hook

Automatically record what changed before each commit:

```bash
# .git/hooks/pre-commit (or via husky)
#!/bin/bash
set -e

CHANGED=$(git diff --cached --name-only | paste -sd ',' -)
if [ -z "$CHANGED" ]; then
    exit 0
fi

# Only record if we're in a project with a technocore map
if [ -f "go.mod" ] || [ -f "package.json" ] || [ -f "Cargo.toml" ]; then
    technocore run record \
        --task "pre-commit changes" \
        --files-changed "$CHANGED" \
        --tokens-in 0 \
        --tokens-out 0 2>/dev/null || true
fi
```

## CI Integration

Cache build in CI to speed up brief generation:

```yaml
# .github/workflows/ci.yml
jobs:
  build:
    steps:
      - uses: actions/checkout@v4

      - name: Cache Technocore index
        uses: actions/cache@v4
        with:
          path: ~/.technocore/projects
          key: technocore-${{ hashFiles('**/*.go', 'go.mod') }}

      - name: Build project cache
        run: |
          technocore map
          technocore cache build

      - name: Generate brief for PR description
        run: |
          echo "## Context"
          technocore brief "review this PR" >> $GITHUB_STEP_SUMMARY
```

## Daily Standup / Journal

Record daily work automatically:

```bash
# ~/bin/daily-technocore
#!/bin/bash
DIR="${1:-.}"
cd "$DIR"

# Build cache if stale
technocore cache refresh 2>/dev/null || technocore cache build

# Record today's activity
if git rev-parse --git-dir > /dev/null 2>&1; then
    FILES=$(git diff --name-only HEAD@{1.day.ago} 2>/dev/null | paste -sd ',' -)
    if [ -n "$FILES" ]; then
        technocore run record \
            --task "daily work" \
            --files-changed "$FILES" \
            --tokens-in 0 \
            --tokens-out 0
    fi
fi
```

Add to cron:

```bash
0 18 * * * ~/bin/daily-technocore ~/projects/my-app
```

## FZF Integration

Search your brain with fzf:

```bash
# ~/.bashrc or ~/.zshrc
ftechnocore() {
    local selection
    selection=$(technocore brain search "$1" 2>/dev/null | fzf --height 40% --reverse)
    if [ -n "$selection" ]; then
        echo "$selection"
    fi
}
```

## Recipe Generator

Convert a successful assistant interaction into a recipe:

```bash
# ~/bin/make-recipe
#!/bin/bash
NAME="$1"
FRAMEWORK="$2"
TEMPLATE="$3"

if [ -z "$NAME" ] || [ -z "$FRAMEWORK" ] || [ -z "$TEMPLATE" ]; then
    echo "Usage: make-recipe <name> <framework> <template-file>"
    exit 1
fi

LANGUAGE=$(technocore cache inspect | grep '^language:' | cut -d' ' -f2)

cat > "/tmp/${NAME}.json" <<EOF
{
  "name": "$NAME",
  "framework": "$FRAMEWORK",
  "language": "$LANGUAGE",
  "signals": [],
  "context_needed": [],
  "avoid": [],
  "brief_template": $(jq -Rs . < "$TEMPLATE"),
  "source": "user-created",
  "tags": ["$FRAMEWORK"]
}
EOF

technocore recipes add --from-file "/tmp/${NAME}.json"
```

## Backup Before Major Changes

Before a large refactor, snapshot your context:

```bash
#!/bin/bash
PROJECT=$(basename "$PWD")
DATE=$(date +%Y%m%d)
BACKUP_DIR="$HOME/.technocore/backups/$PROJECT-$DATE"
mkdir -p "$BACKUP_DIR"

cp ~/.technocore/global.db "$BACKUP_DIR/"
PROJECT_HASH=$(echo -n "$PWD" | sha256sum | cut -c1-16)
cp ~/.technocore/projects/$PROJECT_HASH/project.db "$BACKUP_DIR/"

echo "Backup saved to $BACKUP_DIR"
```

## Query from Scripts

Use `technocore query` in scripts to decide whether to use a local or cloud model:

```bash
#!/bin/bash
PROMPT="$1"

# Try local model first
OUTPUT=$(technocore query "$PROMPT" 2>&1)

if echo "$OUTPUT" | grep -q "DELEGATE"; then
    # Local model couldn't handle it
    # Send enriched brief to cloud API
    BRIEF=$(echo "$OUTPUT" | sed -n '/^# Context/,/^---$/p')
    curl -X POST https://api.anthropic.com/v1/messages \
        -H "x-api-key: $ANTHROPIC_API_KEY" \
        -H "content-type: application/json" \
        -d "{\"messages\": [{\"role\": \"user\", \"content\": \"$BRIEF\n\n$PROMPT\"}]}"
else
    # Local model answered; print result
    echo "$OUTPUT"
fi
```

## Environment Variables

| Variable | Effect |
|---|---|
| `HOME` | Base for `~/.technocore/` directory |
| `CLAUDE_CONFIG_DIR` | Override Claude Code config path for skill install |
| `OPENCODE_CONFIG_DIR` | Override opencode config path |

## Batch Processing

Process multiple files through `tldr`:

```bash
for f in docs/*.md; do
    echo "=== $(basename $f) ==="
    cat "$f" | technocore tldr --sentences 2
    echo
done
```

## Custom Shell Completion

Generate completion scripts for your shell:

```bash
# Bash (add to .bashrc)
source <(technocore completion bash)

# Zsh (add to .zshrc)
source <(technocore completion zsh)

# Fish
 technocore completion fish > ~/.config/fish/completions/technocore.fish
```

Note: Completion generation requires Cobra's built-in completion support, which is available if built with the standard cobra CLI.
