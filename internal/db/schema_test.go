package db

import (
	"path/filepath"
	"testing"
)

func TestInitGlobalSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "global.db")
	dbConn, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()

	if err := InitGlobalSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	var colCount int
	err = dbConn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('task_recipes') WHERE name = 'embedding'`).Scan(&colCount)
	if err != nil {
		t.Fatal(err)
	}
	if colCount != 1 {
		t.Fatal("expected embedding column in task_recipes")
	}

	for _, tbl := range []string{"conversations", "snippets", "agent_lessons"} {
		var count int
		err = dbConn.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?`, tbl).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Fatalf("expected %s table to exist", tbl)
		}
	}
}

func TestInitProjectSchema(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "project.db")
	dbConn, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()

	if err := InitProjectSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	var tblCount int
	err = dbConn.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='files'`).Scan(&tblCount)
	if err != nil {
		t.Fatal(err)
	}
	if tblCount != 1 {
		t.Fatal("expected files table")
	}
}

func TestSchemaBackwardCompat(t *testing.T) {
	// Create a DB with old schema (no embedding column)
	dbPath := filepath.Join(t.TempDir(), "compat.db")
	dbConn, err := Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	// Create table without new columns
	_, err = dbConn.Exec(`CREATE TABLE task_recipes (id INTEGER PRIMARY KEY, name TEXT)`)
	if err != nil {
		t.Fatal(err)
	}
	dbConn.Close()

	// Reopen and run InitGlobalSchema (should add columns via ALTER)
	dbConn, err = Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	if err := InitGlobalSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	// Verify embedding column exists now
	var colCount int
	err = dbConn.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('task_recipes') WHERE name = 'embedding'`).Scan(&colCount)
	if err != nil {
		t.Fatal(err)
	}
	if colCount != 1 {
		t.Fatal("expected embedding column added via backward-compat ALTER")
	}
}
