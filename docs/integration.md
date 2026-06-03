# Integration Guide

How to connect Technocore to AI coding assistants.

## Overview

Technocore provides two integration surfaces:

1. **Skill files** — installed into the assistant's config directory, instructing it to call `technocore brief` before tasks and `technocore run suggest` after
2. **Manual hooks** — shell aliases, editor keybindings, or custom scripts that call `technocore` directly

## Claude Code

### Method 1: Skill File (Recommended)

```bash
technocore install-skill --target claude
```

This writes `~/.claude/skills/technocore/SKILL.md` containing the skill definition. Claude Code reads this file and follows its instructions.

### Method 2: Settings Hook

For deeper integration, Claude Code supports `settings.json` hooks. After installing the skill, the `install-skill` command also registers a `UserPromptSubmit` hook in `~/.claude/settings.json` that silently runs `technocore brief` and prepends the output to the assistant's context.

Verify the hook:

```bash
cat ~/.claude/settings.json | grep -A2 technocore
```

### Method 3: Manual Alias

If you prefer not to install skills:

```bash
# In your shell profile
alias tc="technocore"

# Before asking Claude:
tc brief "add OAuth" | pbcopy  # macOS; paste into Claude prompt
```

## Pi (by Inflection)

Pi does not expose a local skill or hook system. Use the shell-based approach:

### Shell Prepend

```bash
# In your .bashrc or .zshrc
technocore-brief() {
    local task="$1"
    if [ -z "$task" ]; then
        echo "Usage: technocore-brief 'task description'"
        return 1
    fi
    # Generate brief and copy to clipboard
    technocore brief "$task" | pbcopy  # macOS
    echo "Brief copied to clipboard. Paste into Pi."
}

alias tcb='technocore-brief'
```

Usage:

```bash
cd my-project
tcb "how do I structure database migrations"
# Paste into Pi
```

### Pi + Local Model Router

If you run a local model alongside Pi, route simple questions locally first:

```bash
# Check if local model can handle it
technocore query "what's the pattern for middleware in this project"

# If it delegates (prints "DELEGATE"), then ask Pi with the enriched brief
technocore query --delegate "what's the pattern for middleware"
# Copy the enriched brief output and paste into Pi
```

## opencode

opencode uses a skill directory structure similar to Claude Code.

```bash
technocore install-skill --target opencode
```

Writes to: `~/.opencode/skills/technocore/SKILL.md`

opencode discovers skills automatically from this directory. No additional configuration needed.

## Cursor

Cursor's integration is more limited. It does not have a formal skill system, but it does support `.cursorrules` files.

### Method 1: Skill File

```bash
technocore install-skill --target cursor
```

This writes `~/.cursor/skills/technocore/SKILL.md`. Cursor may or may not discover this depending on your version. If it does not:

### Method 2: .cursorrules

Add to your project root `.cursorrules`:

```text
Before generating code, run `technocore brief "<user request>"` in the terminal and consider the project context, recipes, and relevant files it returns.

After completing a task, if files were changed, run `technocore run suggest --task "<summary>" --files-changed "<comma-separated paths>"`.
```

## Codex

Codex (OpenAI's CLI tool) supports skill files in `~/.codex/skills/`.

```bash
technocore install-skill --target codex
```

Writes `~/.codex/skills/technocore/SKILL.md` and registers a `UserPromptSubmit` hook in `~/.codex/hooks.json`.

Verify:

```bash
cat ~/.codex/hooks.json | grep -A2 technocore
```

## VS Code + Copilot

VS Code extensions cannot easily call external CLI tools. Use a VS Code task or custom extension:

### VS Code Task

Add to `.vscode/tasks.json`:

```json
{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Technocore Brief",
            "type": "shell",
            "command": "technocore brief \"${input:task}\"",
            "problemMatcher": [],
            "presentation": {
                "reveal": "always",
                "panel": "new"
            }
        }
    ],
    "inputs": [
        {
            "id": "task",
            "type": "promptString",
            "description": "Task description"
        }
    ]
}
```

Run with `Cmd+Shift+P` → "Tasks: Run Task" → "Technocore Brief". Copy output into the Copilot chat.

## Generic: Any Assistant via Shell

The universal fallback works with any assistant that accepts text input:

```bash
# Generate brief and open in a temp file for copy/paste
technocore brief "add rate limiting" > /tmp/brief.md && $EDITOR /tmp/brief.md
```

Or for automation:

```bash
# Prepend brief to a prompt file
technocore brief "refactor auth" > /tmp/context.txt
cat user-prompt.txt >> /tmp/context.txt
# Send /tmp/context.txt to your assistant via its API or UI
```

## Hook Events Summary

| Assistant | Hook Type | Trigger | Effect |
|---|---|---|---|
| Claude Code | `UserPromptSubmit` | Every user message | Prepends `technocore brief` output |
| Codex | `UserPromptSubmit` | Every user message | Prepends `technocore brief` output |
| opencode | SKILL.md | Task start | Instructs assistant to run `technocore brief` |
| Cursor | SKILL.md / .cursorrules | Task start | Instructs assistant to run `technocore brief` |
| Pi | None (manual) | User action | Shell alias copies brief to clipboard |
| Copilot | None (manual) | VS Code task | User runs task, copies output |

## Custom Hooks

You can build your own hooks using the `query` command's JSON-like output:

```bash
# In a git pre-commit hook: record what changed
#!/bin/bash
changed=$(git diff --name-only HEAD | paste -sd, -)
if [ -n "$changed" ]; then
    technocore run record \
        --task "$(git log -1 --pretty=%s)" \
        --files-changed "$changed" \
        --tokens-in 0 \
        --tokens-out 0
fi
```

Or integrate with your editor:

```vim
" .vimrc / init.vim
command! -nargs=1 TechnocoreBrief
    \ echo system('technocore brief ' . shellescape(<q-args>))
```
