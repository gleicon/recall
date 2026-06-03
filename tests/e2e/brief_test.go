package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBriefWarnsWhenNoRecipes(t *testing.T) {
	e := newEnv(t)

	// Create a minimal Go project so map detection works
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(projDir, "main.go")
	if err := os.WriteFile(mainGo, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Build project map
	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Run brief without seeding recipes
	stdout, stderr, code := e.runInDir(projDir, "brief", "add health check")
	if code != 0 {
		t.Fatalf("brief exited %d: stderr=%s", code, stderr)
	}

	// Should still produce brief output
	if !strings.Contains(stdout, "Brief:") {
		t.Fatalf("expected brief header in stdout, got: %s", stdout)
	}

	// Should warn on stderr
	if !strings.Contains(stderr, "No relevant recipes found") {
		t.Fatalf("expected warning about missing recipes on stderr, got: %s", stderr)
	}
}

func TestBriefIncludesMatchingRecipe(t *testing.T) {
	e := newEnv(t)

	// Seed recipes
	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	// Create a minimal Go project
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(projDir, "main.go")
	if err := os.WriteFile(mainGo, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Build project map
	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Run brief for a health check in a Go project
	stdout, stderr, code := e.runInDir(projDir, "brief", "add health check")
	if code != 0 {
		t.Fatalf("brief exited %d: stderr=%s", code, stderr)
	}

	// Should include the Go healthcheck recipe
	if !strings.Contains(stdout, "health") && !strings.Contains(stdout, "healthcheck") {
		t.Fatalf("expected healthcheck recipe in brief, got stdout=%s stderr=%s", stdout, stderr)
	}
}
