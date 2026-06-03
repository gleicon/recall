package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBrainShowsEmptyState(t *testing.T) {
	e := newEnv(t)

	// Seed so DB exists
	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	for _, cmd := range []string{"brain conversations", "brain snippets", "brain lessons"} {
		parts := strings.Fields(cmd)
		stdout, _, code := e.run(parts...)
		if code != 0 {
			t.Fatalf("%s exited %d", cmd, code)
		}
		// Should show empty state message
		if !strings.Contains(stdout, "No") && !strings.Contains(stdout, "yet") {
			t.Fatalf("expected empty state message for %s, got: %s", cmd, stdout)
		}
	}
}

func TestBrainSearch(t *testing.T) {
	e := newEnv(t)

	_, _, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d", code)
	}

	// Create a Go project and query it to generate brain data
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Query generates a conversation
	_, _, code = e.runInDir(projDir, "query", "--delegate", "add health check")
	if code != 0 {
		t.Fatalf("query exited %d", code)
	}

	// Brain search should find the conversation
	stdout, _, code := e.run("brain", "search", "health")
	if code != 0 {
		t.Fatalf("brain search exited %d", code)
	}
	if !strings.Contains(stdout, "Conversations") && !strings.Contains(stdout, "No matches") {
		t.Fatalf("expected Conversations section or empty, got: %s", stdout)
	}
}
