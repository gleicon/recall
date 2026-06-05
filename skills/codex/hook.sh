#!/usr/bin/env bash
# recall-hook.sh — UserPromptSubmit hook for Claude Code
# Auto-injects project brief and matching recipes into assistant context.
# Exits 0 silently if recall is not in PATH.

set -euo pipefail

# Require recall in PATH
if ! command -v recall >/dev/null 2>&1; then
  exit 0
fi

# Read JSON from stdin (Claude Code sends event as JSON on stdin)
INPUT=$(cat)

# Extract prompt text
if command -v jq >/dev/null 2>&1; then
  PROMPT=$(printf '%s' "$INPUT" | jq -r '.prompt // empty' 2>/dev/null || true)
else
  PROMPT=$(printf '%s' "$INPUT" | python3 -c \
    "import json,sys; d=json.load(sys.stdin); print(d.get('prompt',''), end='')" 2>/dev/null || true)
fi

# Empty prompt — no-op
if [ -z "$PROMPT" ]; then
  exit 0
fi

# Run recall brief in the current working directory
BRIEF=$(recall brief "$PROMPT" 2>/dev/null || true)

if [ -z "$BRIEF" ]; then
  exit 0
fi

# Output hookSpecificOutput JSON for Claude Code to inject as additionalContext
# Use python3 for JSON encoding to handle special chars safely
printf '%s' "$BRIEF" | python3 -c "
import json, sys
content = sys.stdin.read()
output = {
  'hookSpecificOutput': {
    'hookEventName': 'UserPromptSubmit',
    'additionalContext': '[Recall Brief]\\n' + content
  }
}
print(json.dumps(output))
"
