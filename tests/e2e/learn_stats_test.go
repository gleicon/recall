package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLearnStoresInsight(t *testing.T) {
	e := newEnv(t)

	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	stdout, _, code := e.runInDir(projDir, "learn", "always use context for cancellation")
	if code != 0 {
		t.Fatalf("learn exited %d", code)
	}
	if !strings.Contains(stdout, "Insight stored") {
		t.Fatalf("expected 'saved' in stdout, got: %s", stdout)
	}

	// Verify via cache inspect
	inspectOut, _, inspectCode := e.runInDir(projDir, "cache", "inspect")
	if inspectCode != 0 {
		t.Fatalf("cache inspect exited %d", inspectCode)
	}
	if !strings.Contains(inspectOut, "always use context") {
		t.Fatalf("expected insight in cache inspect, got: %s", inspectOut)
	}
}

func TestStatsCacheShowsProjectStats(t *testing.T) {
	e := newEnv(t)

	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(projDir, "main.go")
	if err := os.WriteFile(mainGo, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Build map and cache
	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}
	_, stderr, code = e.runInDir(projDir, "cache", "build")
	if code != 0 {
		t.Fatalf("cache build exited %d: stderr=%s", code, stderr)
	}

	stdout, _, code := e.runInDir(projDir, "stats", "cache")
	if code != 0 {
		t.Fatalf("stats cache exited %d", code)
	}
	if !strings.Contains(stdout, "Files:") {
		t.Fatalf("expected 'Files:' in stats cache, got: %s", stdout)
	}
}
