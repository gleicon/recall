# Backup and Migration

Technocore stores all data in SQLite databases under `~/.technocore/`. This makes backup straightforward.

## What to Back Up

| Path | Contents | Size |
|---|---|---|
| `~/.technocore/global.db` | Recipes, snippets, lessons, conversations, stats | Typically < 50 MB |
| `~/.technocore/projects/<hash>/project.db` | Per-project file summaries, chunks, memories | Varies by project size |
| `~/.technocore/config.json` | User settings (model preference, timeout) | < 1 KB |
| `~/.technocore/recipes/` | Source JSON files for recipes | < 1 MB |

## Simple Backup

```bash
# Create timestamped archive
DATE=$(date +%Y%m%d_%H%M%S)
tar czf "technocore-backup-${DATE}.tar.gz" -C "$HOME" .technocore/global.db .technocore/config.json .technocore/recipes

# For project databases (can be large; exclude if you can rebuild)
tar czf "technocore-backup-full-${DATE}.tar.gz" -C "$HOME" .technocore
```

## Automated Backup

```bash
# ~/.local/bin/technocore-backup
#!/bin/bash
BACKUP_DIR="${TECHNOCORE_BACKUP_DIR:-$HOME/backups/technocore}"
mkdir -p "$BACKUP_DIR"

cp "$HOME/.technocore/global.db" "$BACKUP_DIR/global-$(date +%Y%m%d).db"
cp "$HOME/.technocore/config.json" "$BACKUP_DIR/config-$(date +%Y%m%d).json"

# Keep only last 7 days of backups
find "$BACKUP_DIR" -name 'global-*.db' -mtime +7 -delete
find "$BACKUP_DIR" -name 'config-*.json' -mtime +7 -delete
```

Add to cron:

```bash
0 */6 * * * ~/.local/bin/technocore-backup
```

## Restore

```bash
# Restore global database
cp technocore-backup-20240115/global.db ~/.technocore/global.db

# Restore settings
cp technocore-backup-20240115/config.json ~/.technocore/config.json

# Rebuild project caches (not backed up)
cd my-project
technocore map
technocore cache build
```

## Migrating to a New Machine

1. **Copy global.db and config.json**

   ```bash
   scp ~/.technocore/global.db newmachine:~/.technocore/
   scp ~/.technocore/config.json newmachine:~/.technocore/
   ```

2. **Reinstall Technocore binary**

3. **Re-seed recipes**

   ```bash
   technocore recipes seed
   ```

4. **Rebuild project caches**

   ```bash
   cd ~/projects/app1 && technocore map && technocore cache build
   cd ~/projects/app2 && technocore map && technocore cache build
   ```

## Selective Export / Import

### Export a Single Project

```bash
PROJECT_HASH=$(echo -n "$PWD" | sha256sum | cut -c1-16)
cp ~/.technocore/projects/$PROJECT_HASH/project.db ./technocore-project.db
```

### Import a Single Project

```bash
PROJECT_HASH=$(echo -n "$PWD" | sha256sum | cut -c1-16)
mkdir -p ~/.technocore/projects/$PROJECT_HASH
cp ./technocore-project.db ~/.technocore/projects/$PROJECT_HASH/project.db
```

### Export Brain Data Only

```bash
# Dump brain tables from global.db
sqlite3 ~/.technocore/global.db <<'EOF'
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
sqlite3 ~/.technocore/global.db < brain-export.sql
```

## Compression

Project databases can grow large. Compress old ones:

```bash
# Compress project DBs older than 30 days
find ~/.technocore/projects -name 'project.db' -mtime +30 -exec gzip {} \;

# To use a compressed DB, decompress first
gunzip ~/.technocore/projects/<hash>/project.db.gz
```

## Corruption Recovery

If a database becomes corrupted:

```bash
# Check integrity
sqlite3 ~/.technocore/global.db "PRAGMA integrity_check;"

# Dump and recreate
sqlite3 ~/.technocore/global.db ".dump" > /tmp/global_dump.sql
rm ~/.technocore/global.db
sqlite3 ~/.technocore/global.db < /tmp/global_dump.sql
```

## Size Monitoring

```bash
# Check sizes
du -sh ~/.technocore/global.db
du -sh ~/.technocore/projects/*

# Total
du -sh ~/.technocore
```

## Git-Like Versioning (Experimental)

Track changes to your brain over time:

```bash
# Initialize a git repo for your technocore data
cd ~/.technocore
git init
git add global.db config.json recipes/
git commit -m "Initial brain state"

# After significant changes
git add -A
git commit -m "Post-refactor brain update"

# Note: SQLite databases don't diff well in git. Consider using sqliterc or git-lfs for large DBs.
```
