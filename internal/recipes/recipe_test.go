package recipes

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gleicon/recall/internal/db"
)

func TestLoadFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.json")
	data := []byte(`{
		"name": "test_recipe",
		"framework": "testfw",
		"language": "go",
		"signals": ["go.mod"],
		"context_needed": ["main.go"],
		"avoid": ["vendor/"],
		"brief_template": "Test template",
		"source": "test",
		"tags": ["test"]
	}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	r, err := LoadFromFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "test_recipe" {
		t.Fatalf("expected name 'test_recipe', got %s", r.Name)
	}
	if r.Framework != "testfw" {
		t.Fatalf("expected framework 'testfw', got %s", r.Framework)
	}
}

func TestLoadFromFileMissingBriefTemplate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	data := []byte(`{"name": "bad_recipe"}`)
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}

	_, err := LoadFromFile(path)
	if err == nil {
		t.Fatal("expected error for missing brief_template")
	}
}

func TestStoreAndFind(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	dbConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	if err := db.InitGlobalSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	r := &Recipe{
		Name:          "add_healthcheck_go",
		Framework:     "go",
		Language:      "go",
		Signals:       []string{"main.go"},
		BriefTemplate: "Add /healthz endpoint",
		Source:        "test",
		Tags:          []string{"health"},
	}
	if err := Store(dbConn, r); err != nil {
		t.Fatal(err)
	}

	matches, err := FindMatches(dbConn, "add health check", "go", []string{"main.go"}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Recipe.Name != "add_healthcheck_go" {
		t.Fatalf("expected 'add_healthcheck_go', got %s", matches[0].Recipe.Name)
	}
}

func TestFindMatchesFrameworkBoost(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	dbConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	if err := db.InitGlobalSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	// Store two recipes with similar text but different frameworks
	goRecipe := &Recipe{
		Name:          "go_task",
		Framework:     "go",
		BriefTemplate: "Do something with auth",
		Source:        "test",
	}
	nodeRecipe := &Recipe{
		Name:          "node_task",
		Framework:     "nextjs",
		BriefTemplate: "Do something with auth",
		Source:        "test",
	}
	Store(dbConn, goRecipe)
	Store(dbConn, nodeRecipe)

	// Query for "auth" with framework="go"
	matches, err := FindMatches(dbConn, "auth", "go", []string{}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) < 2 {
		t.Fatal("expected at least 2 matches")
	}
	// go_task should rank higher due to framework match
	if matches[0].Recipe.Name != "go_task" {
		t.Fatalf("expected 'go_task' first, got %s", matches[0].Recipe.Name)
	}
	if matches[0].Score < 0.3 {
		t.Fatalf("expected framework-boosted score > 0.3, got %f", matches[0].Score)
	}
}

func TestDuplicateStore(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	dbConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	if err := db.InitGlobalSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	r := &Recipe{Name: "dup", BriefTemplate: "T", Source: "test"}
	if err := Store(dbConn, r); err != nil {
		t.Fatal(err)
	}
	if err := Store(dbConn, r); err != nil {
		t.Fatal(err)
	}

	list, err := List(dbConn)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 recipe, got %d", len(list))
	}
}

func TestIncrementUseCount(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "test.db")
	dbConn, err := db.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbConn.Close()
	if err := db.InitGlobalSchema(dbConn); err != nil {
		t.Fatal(err)
	}

	r := &Recipe{Name: "popular", BriefTemplate: "T", Source: "test"}
	if err := Store(dbConn, r); err != nil {
		t.Fatal(err)
	}

	if err := IncrementUseCount(dbConn, "popular", 0.8); err != nil {
		t.Fatal(err)
	}
	if err := IncrementUseCount(dbConn, "popular", 1.0); err != nil {
		t.Fatal(err)
	}

	var useCount int
	var avgScore float64
	row := dbConn.QueryRow(`SELECT use_count, avg_score FROM task_recipes WHERE name = ?`, "popular")
	if err := row.Scan(&useCount, &avgScore); err != nil {
		t.Fatal(err)
	}
	if useCount != 2 {
		t.Fatalf("expected use_count 2, got %d", useCount)
	}
	// Running average: (0.8 + 1.0) / 2 = 0.9
	if avgScore < 0.89 || avgScore > 0.91 {
		t.Fatalf("expected avg_score ~0.9, got %f", avgScore)
	}
}
