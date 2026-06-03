package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCleanupCacheRemovesCacheData(t *testing.T) {
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

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	_, stderr, code = e.runInDir(projDir, "cache", "build")
	if code != 0 {
		t.Fatalf("cache build exited %d: stderr=%s", code, stderr)
	}

	// Verify cache exists
	inspectOut, _, code := e.runInDir(projDir, "cache", "inspect")
	if code != 0 {
		t.Fatalf("cache inspect exited %d", code)
	}
	if !strings.Contains(inspectOut, "=== Files ===") {
		t.Fatalf("expected cache data before cleanup, got: %s", inspectOut)
	}

	// Cleanup cache (default 30 days - should remove everything since we just created it)
	stdout, _, code := e.runInDir(projDir, "cleanup", "cache", "--days", "0")
	if code != 0 {
		t.Fatalf("cleanup cache exited %d", code)
	}
	if !strings.Contains(stdout, "Cleanup complete") {
		t.Fatalf("expected cleanup message, got: %s", stdout)
	}
}

func TestCleanupProjectCleansProjectDirectory(t *testing.T) {
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

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	stdout, _, code := e.run("cleanup", "project", projDir)
	if code != 0 {
		t.Fatalf("cleanup project exited %d", code)
	}
	if !strings.Contains(stdout, "Removed") {
		t.Fatalf("expected cleanup message, got: %s", stdout)
	}
}
