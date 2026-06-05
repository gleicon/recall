# Recall

Accelerate tasks with cached project context and record outcomes for continuous improvement.

## Before Task

When a user asks you to do something (fix, add, refactor, debug), run:

```bash
recall brief "<exact user prompt>"
```

Prepend the output to your planning context. Use the project map, subsystems, recipes, and relevant files to avoid rediscovering facts.

## After Task

When you finish a task and files have changed, run:

```bash
recall run suggest --task "<short summary>" --files-changed "file1,file2" --tokens-in <N> --tokens-out <M>
```

Respect the user's single-key response:
- `y` → the run was recorded
- `n` → skip recording
- `i` → the user will type an insight, then it is recorded

## Rules

- Never record file contents. Only paths, task type, and token counts.
- Do not run `recall` if the binary is not found in PATH.
- If `recall brief` returns "No project map found", suggest the user run `recall map`.
