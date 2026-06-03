package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatsRunsShowsModelStats(t *testing.T) {
	e := newEnv(t)

	// Seed recipes so run suggest has something to work with
	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

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

	// Run suggest to create run data with model behavior stats (pipe "y" to confirm)
	_, stderr, code = e.runWithInputInDir(projDir, "y\n", "run", "suggest", "--task", "test task", "--tokens-in", "1000", "--tokens-out", "500")
	if code != 0 {
		t.Fatalf("run suggest exited %d: stderr=%s", code, stderr)
	}

	stdout, _, code := e.run("stats", "runs")
	if code != 0 {
		t.Fatalf("stats runs exited %d", code)
	}
	if !strings.Contains(stdout, "Framework") {
		t.Fatalf("expected 'Framework' header in stats runs, got: %s", stdout)
	}
}

func TestStatsInsightsShowsRecipes(t *testing.T) {
	e := newEnv(t)

	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	// Use run suggest to increment recipe use counts
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

	_, stderr, code = e.runWithInputInDir(projDir, "y\n", "run", "suggest", "--task", "test task", "--tokens-in", "1000", "--tokens-out", "500")
	if code != 0 {
		t.Fatalf("run suggest exited %d: stderr=%s", code, stderr)
	}

	stdout, _, code := e.run("stats", "insights")
	if code != 0 {
		t.Fatalf("stats insights exited %d", code)
	}
	if !strings.Contains(stdout, "Most Useful") {
		t.Fatalf("expected 'Most Useful' in stats insights, got: %s", stdout)
	}
	if !strings.Contains(stdout, "Least Useful") {
		t.Fatalf("expected 'Least Useful' in stats insights, got: %s", stdout)
	}
}

func TestStatsGlobalShowsCrossProjectStats(t *testing.T) {
	e := newEnv(t)

	// Create two projects and map both
	for i := 0; i < 2; i++ {
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
	}

	stdout, _, code := e.run("stats", "global")
	if code != 0 {
		t.Fatalf("stats global exited %d", code)
	}
	if !strings.Contains(stdout, "Global Statistics") {
		t.Fatalf("expected 'Global Statistics' in stats global, got: %s", stdout)
	}
}
