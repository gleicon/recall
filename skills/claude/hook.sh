#!/usr/bin/env bash
# technocore-hook.sh — UserPromptSubmit hook for Claude Code
# Auto-injects project brief and matching recipes into assistant context.
# Exits 0 silently if technocore is not in PATH.

set -euo pipefail

# Require technocore in PATH
if ! command -v technocore >/dev/null 2>&1; then
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

# Run technocore brief in the current working directory
BRIEF=$(technocore brief "$PROMPT" 2>/dev/null || true)

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
    'additionalContext': '[Pipecamp Brief]\\n' + content
  }
}
print(json.dumps(output))
"
