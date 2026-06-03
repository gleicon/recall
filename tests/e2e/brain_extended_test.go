package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrainStatsWithData(t *testing.T) {
	e := newEnv(t)

	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Query with delegate to generate brain data
	_, _, code = e.runInDir(projDir, "query", "--delegate", "add health check")
	if code != 0 {
		t.Fatalf("query exited %d", code)
	}

	stdout, _, code := e.run("brain", "stats")
	if code != 0 {
		t.Fatalf("brain stats exited %d", code)
	}
	if !strings.Contains(stdout, "conversations") {
		t.Fatalf("expected 'conversations' in brain stats, got: %s", stdout)
	}
}

func TestBrainFrameworksWithData(t *testing.T) {
	e := newEnv(t)

	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Query to generate framework-specific brain data
	_, _, code = e.runInDir(projDir, "query", "--delegate", "add health check")
	if code != 0 {
		t.Fatalf("query exited %d", code)
	}

	stdout, _, code := e.run("brain", "frameworks")
	if code != 0 {
		t.Fatalf("brain frameworks exited %d", code)
	}
	if !strings.Contains(stdout, "Framework") {
		t.Fatalf("expected 'Framework' in brain frameworks, got: %s", stdout)
	}
}
