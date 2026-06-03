package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestQueryTimeoutFlag(t *testing.T) {
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

	// Query with --timeout flag (should not crash, even if no local model)
	stdout, _, code := e.runInDir(projDir, "query", "--delegate", "--timeout", "5", "add health check")
	if code != 0 {
		t.Fatalf("query exited %d", code)
	}
	if !strings.Contains(stdout, "DELEGATE") {
		t.Fatalf("expected DELEGATE with --delegate flag, got: %s", stdout)
	}
}

func TestBrainFrameworksEmpty(t *testing.T) {
	e := newEnv(t)

	stdout, _, code := e.run("brain", "frameworks")
	if code != 0 {
		t.Fatalf("brain frameworks exited %d", code)
	}
	if !strings.Contains(stdout, "No framework data yet") {
		t.Fatalf("expected empty state message, got: %s", stdout)
	}
}

func TestLocalModelConfig(t *testing.T) {
	e := newEnv(t)

	stdout, _, code := e.run("local", "models")
	if code != 0 {
		t.Fatalf("local models exited %d", code)
	}
	hasNoLLM := strings.Contains(stdout, "No local LLM detected")
	if hasNoLLM {
		stdout, _, code = e.run("local", "use", "some-model")
		if code == 0 {
			t.Fatal("expected local use to fail when no LLM is running")
		}
	} else {
		stdout, _, code = e.run("local", "use", "nonexistent-model-12345")
		if code == 0 {
			t.Fatal("expected local use to fail for nonexistent model")
		}
		if !strings.Contains(stdout, "not found") {
			t.Fatalf("expected 'not found' error, got: %s", stdout)
		}
	}
}
