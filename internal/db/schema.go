package db

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
)

// Open opens a SQLite database, creating it if necessary.
func Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path+"?_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("open db %s: %w", path, err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db %s: %w", path, err)
	}
	return db, nil
}

// InitGlobalSchema creates the global cache tables.
func InitGlobalSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS patterns (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			language TEXT,
			framework TEXT,
			signals TEXT,
			context_needed TEXT,
			avoid TEXT,
			brief_template TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS task_recipes (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			language TEXT,
			framework TEXT,
			signals TEXT,
			context_needed TEXT,
			avoid TEXT,
			brief_template TEXT,
			embedding BLOB,
			use_count INTEGER DEFAULT 0,
			avg_score REAL DEFAULT 0,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS prompt_templates (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			template TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS framework_fingerprints (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			framework TEXT NOT NULL,
			signals TEXT,
			files TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS model_behavior_stats (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			model_name TEXT,
			task_type TEXT,
			framework TEXT,
			files_included TEXT,
			files_changed TEXT,
			tests_passed INTEGER,
			follow_up_needed INTEGER,
			input_tokens INTEGER,
			output_tokens INTEGER,
			accepted INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS tool_profiles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tool_name TEXT NOT NULL,
			behavior_note TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task TEXT,
			prompt TEXT,
			response TEXT,
			model_name TEXT,
			input_tokens INTEGER,
			output_tokens INTEGER,
			delegated INTEGER DEFAULT 0,
			delegation_reason TEXT,
			project_hash TEXT,
			framework TEXT,
			embedding BLOB,
			accepted INTEGER DEFAULT NULL,
			feedback_note TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS snippets (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			language TEXT,
			framework TEXT,
			code TEXT,
			context TEXT,
			embedding BLOB,
			use_count INTEGER DEFAULT 0,
			source TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS agent_lessons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			pattern TEXT,
			framework TEXT,
			model_name TEXT,
			success_rate REAL DEFAULT 0,
			context TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("global schema: %w", err)
		}
	}
	// Backward-compat: add columns that may be missing in older DBs
	alterStmts := []string{
		`ALTER TABLE task_recipes ADD COLUMN embedding BLOB`,
		`ALTER TABLE task_recipes ADD COLUMN use_count INTEGER DEFAULT 0`,
		`ALTER TABLE task_recipes ADD COLUMN avg_score REAL DEFAULT 0`,
		`ALTER TABLE task_recipes ADD COLUMN source TEXT`,
		`ALTER TABLE task_recipes ADD COLUMN tags TEXT`,
		`ALTER TABLE conversations ADD COLUMN delegation_reason TEXT`,
		`ALTER TABLE conversations ADD COLUMN embedding BLOB`,
		`ALTER TABLE conversations ADD COLUMN accepted INTEGER DEFAULT NULL`,
		`ALTER TABLE conversations ADD COLUMN feedback_note TEXT`,
	}
	for _, s := range alterStmts {
		_, _ = db.Exec(s) // ignore "duplicate column name" errors
	}
	return nil
}

// InitProjectSchema creates the per-project cache tables.
func InitProjectSchema(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS files (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			path TEXT NOT NULL UNIQUE,
			hash TEXT,
			content TEXT,
			summary TEXT,
			embedding BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS file_search USING fts5(path, content, summary)`,
		`CREATE TABLE IF NOT EXISTS chunks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL,
			chunk_text TEXT,
			embedding BLOB,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(file_id) REFERENCES files(id)
		)`,
		`CREATE VIRTUAL TABLE IF NOT EXISTS chunk_search USING fts5(chunk_text)`,
		`CREATE TABLE IF NOT EXISTS file_summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			file_id INTEGER NOT NULL UNIQUE,
			summary TEXT,
			patterns TEXT,
			last_verified_commit TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(file_id) REFERENCES files(id)
		)`,
		`CREATE TABLE IF NOT EXISTS subsystem_summaries (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			scope_files TEXT,
			summary TEXT,
			patterns TEXT,
			last_verified_commit TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS runs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			task_type TEXT,
			framework TEXT,
			files_included TEXT,
			files_changed TEXT,
			tests_passed INTEGER,
			follow_up_needed INTEGER,
			input_tokens INTEGER,
			output_tokens INTEGER,
			accepted INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS memories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			kind TEXT,
			content TEXT,
			context TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS project_map (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			language TEXT,
			framework TEXT,
			package_manager TEXT,
			entrypoints TEXT,
			module_boundaries TEXT,
			important_dirs TEXT,
			ignored_areas TEXT,
			embed_model TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return fmt.Errorf("project schema: %w", err)
		}
	}
	// Backward-compat for project schema
	projectAlterStmts := []string{
		`ALTER TABLE files ADD COLUMN embedding BLOB`,
		`ALTER TABLE project_map ADD COLUMN embed_model TEXT`,
	}
	for _, s := range projectAlterStmts {
		_, _ = db.Exec(s)
	}
	return nil
}


