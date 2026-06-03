package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryBuildsBrief(t *testing.T) {
	e := newEnv(t)

	// Seed recipes for context
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

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Query should always produce output (either brief + DELEGATE or local model answer)
	stdout, _, code := e.runInDir(projDir, "query", "add health check")
	if code != 0 {
		t.Fatalf("query exited %d", code)
	}
	// Should either show enriched context or local model answer
	hasBrief := strings.Contains(stdout, "Context for:") || strings.Contains(stdout, "Project")
	hasAnswer := strings.Contains(stdout, "Answer")
	hasDelegate := strings.Contains(stdout, "DELEGATE")
	if !hasBrief && !hasAnswer {
		t.Fatalf("expected enriched brief or answer, got: %s", stdout)
	}
	// Should either delegate or answer - both are valid outcomes
	if !hasDelegate && !hasAnswer {
		t.Fatalf("expected either DELEGATE or Answer, got: %s", stdout)
	}
}

func TestQueryForceDelegateFlag(t *testing.T) {
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

	stdout, _, code := e.runInDir(projDir, "query", "--delegate", "add health check")
	if code != 0 {
		t.Fatalf("query exited %d", code)
	}
	if !strings.Contains(stdout, "DELEGATE") {
		t.Fatalf("expected DELEGATE with --delegate flag, got: %s", stdout)
	}
}
