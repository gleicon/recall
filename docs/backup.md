# Backup and Migration

Recall stores all data in SQLite databases under `~/.recall/`. This makes backup straightforward.

## What to Back Up

| Path | Contents | Size |
|---|---|---|
| `~/.recall/global.db` | Recipes, snippets, lessons, conversations, stats | Typically < 50 MB |
| `~/.recall/projects/<hash>/project.db` | Per-project file summaries, chunks, memories | Varies by project size |
| `~/.recall/config.json` | User settings (model preference, timeout) | < 1 KB |
| `~/.recall/recipes/` | Source JSON files for recipes | < 1 MB |

## Simple Backup

```bash
# Create timestamped archive
DATE=$(date +%Y%m%d_%H%M%S)
tar czf "recall-backup-${DATE}.tar.gz" -C "$HOME" .recall/global.db .recall/config.json .recall/recipes

# For project databases (can be large; exclude if you can rebuild)
tar czf "recall-backup-full-${DATE}.tar.gz" -C "$HOME" .recall
```

## Automated Backup

```bash
# ~/.local/bin/recall-backup
#!/bin/bash
BACKUP_DIR="${RECALL_BACKUP_DIR:-$HOME/backups/recall}"
mkdir -p "$BACKUP_DIR"

cp "$HOME/.recall/global.db" "$BACKUP_DIR/global-$(date +%Y%m%d).db"
cp "$HOME/.recall/config.json" "$BACKUP_DIR/config-$(date +%Y%m%d).json"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name 'global-*.db' -mtime +7 -delete
find "$BACKUP_DIR" -name 'config-*.json' -mtime +7 -delete
```

Add to cron:

```bash
0 */6 * * * ~/.local/bin/recall-backup
```

## Restore

```bash
# Restore global database
cp recall-backup-20240115/global.db ~/.recall/global.db

# Restore settings
cp recall-backup-20240115/config.json ~/.recall/config.json

# Rebuild project caches (not backed up)
cd my-project
recall map
recall cache build
```

## Migrating to a New Machine

1. **Copy global.db and config.json**

   ```bash
   scp ~/.recall/global.db newmachine:~/.recall/
   scp ~/.recall/config.json newmachine:~/.recall/
   ```

2. **Reinstall Recall binary**

3. **Re-seed recipes**

   ```bash
   recall recipes seed
   ```

4. **Rebuild project caches**

   ```bash
   cd ~/projects/app1 && recall map && recall cache build
   cd ~/projects/app2 && recall map && recall cache build
   ```

## Selective Export / Import

### Export a Single Project

```bash
PROJECT_HASH=$(echo -n "$PWD" | sha256sum | cut -c1-16)
cp ~/.recall/projects/$PROJECT_HASH/project.db ./recall-project.db
```

### Import a Single Project

```bash
PROJECT_HASH=$(echo -n "$PWD" | sha256sum | cut -c1-16)
mkdir -p ~/.recall/projects/$PROJECT_HASH
cp ./recall-project.db ~/.recall/projects/$PROJECT_HASH/project.db
```

### Export Brain Data Only

```bash
# Dump brain tables from global.db
sqlite3 ~/.recall/global.db <<'EOF'
.mode insert
.output brain-export.sql
SELECT * FROM conversations;
SELECT * FROM snippets;
SELECT * FROM agent_lessons;
.quit
EOF
```

### Import Brain Data Only

```bash
sqlite3 ~/.recall/global.db < brain-export.sql
```

## Compression

Project databases can grow large. Compress old ones:

```bash
# Compress project DBs older than 30 days
find ~/.recall/projects -name 'project.db' -mtime +30 -exec gzip {} \;

# To use a compressed DB, decompress first
gunzip ~/.recall/projects/<hash>/project.db.gz
```

## Corruption Recovery

If a database becomes corrupted:

```bash
# Check integrity
sqlite3 ~/.recall/global.db "PRAGMA integrity_check;"

# Dump and recreate
sqlite3 ~/.recall/global.db ".dump" > /tmp/global_dump.sql
rm ~/.recall/global.db
sqlite3 ~/.recall/global.db < /tmp/global_dump.sql
```

## Size Monitoring

```bash
# Check sizes
du -sh ~/.recall/global.db
du -sh ~/.recall/projects/*

# Total
du -sh ~/.recall
```

## Git-Like Versioning (Experimental)

Track changes to your brain over time:

```bash
# Initialize a git repo for your recall data
cd ~/.recall
git init
git add global.db config.json recipes/
git commit -m "Initial brain state"

# After significant changes
git add -A
git commit -m "Post-refactor brain update"

# Note: SQLite databases don't diff well in git. Consider using sqliterc or git-lfs for large DBs.
```
