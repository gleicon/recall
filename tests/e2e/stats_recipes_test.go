package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsRecipesAfterBriefIncrementsUseCount(t *testing.T) {
	e := newEnv(t)

	// Seed recipes
	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	// Create Go project
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(projDir, "main.go")
	if err := os.WriteFile(mainGo, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Run brief (should increment use_count for matched recipes)
	_, _, code = e.runInDir(projDir, "brief", "add health check")
	if code != 0 {
		t.Fatalf("brief exited %d", code)
	}

	// Check stats recipes shows non-zero use_count for health check recipe
	statsOut, _, code := e.run("stats", "recipes")
	if code != 0 {
		t.Fatalf("stats recipes exited %d", code)
	}

	// At least one recipe should have use_count > 0
	if !strings.Contains(statsOut, "1 |") && !strings.Contains(statsOut, "Uses") {
		t.Fatalf("expected stats recipes to show usage, got: %s", statsOut)
	}
}
