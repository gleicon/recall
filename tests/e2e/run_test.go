package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunSuggestRecordsOnYes(t *testing.T) {
	e := newEnv(t)

	// Create a minimal Go project so map detection works
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Build project map (needed for framework detection in saveRun)
	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Pipe 'y' to run suggest
	stdout, stderr, code := e.runWithInputInDir(projDir, "y\n",
		"run", "suggest",
		"--task", "add health check",
		"--files-changed", "main.go",
		"--tokens-in", "4000",
		"--tokens-out", "800",
	)
	if code != 0 {
		t.Fatalf("run suggest exited %d: stdout=%s stderr=%s", code, stdout, stderr)
	}
	if !strings.Contains(stdout, "Run recorded.") {
		t.Fatalf("expected 'Run recorded.' in stdout, got: %s", stdout)
	}

	// Verify via stats runs
	statsOut, statsErr, statsCode := e.run("stats", "runs")
	if statsCode != 0 {
		t.Fatalf("stats runs exited %d: stderr=%s", statsCode, statsErr)
	}
	if !strings.Contains(statsOut, "go") {
		t.Fatalf("expected 'go' framework in stats runs, got: %s", statsOut)
	}
}

func TestRunSuggestInsightRecordsWithMemory(t *testing.T) {
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

	// Pipe 'i' then insight text
	stdout, _, code := e.runWithInputInDir(projDir, "i\nuse middleware for auth\n",
		"run", "suggest",
		"--task", "add OAuth",
		"--files-changed", "src/lib/auth.ts",
		"--tokens-in", "4000",
		"--tokens-out", "800",
	)
	if code != 0 {
		t.Fatalf("run suggest exited %d", code)
	}
	if !strings.Contains(stdout, "Run recorded.") {
		t.Fatalf("expected 'Run recorded.' in stdout, got: %s", stdout)
	}

	// Verify via cache inspect that memory was stored
	inspectOut, _, inspectCode := e.runInDir(projDir, "cache", "inspect")
	if inspectCode != 0 {
		t.Fatalf("cache inspect exited %d", inspectCode)
	}
	if !strings.Contains(inspectOut, "use middleware for auth") {
		t.Fatalf("expected insight in cache inspect, got: %s", inspectOut)
	}
}

func TestRunRecordNonInteractive(t *testing.T) {
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

	stdout, _, code := e.runInDir(projDir, "run", "record",
		"--task", "add health check",
		"--files-changed", "main.go",
		"--tokens-in", "4000",
		"--tokens-out", "800",
		"--insight", "use stdlib http",
	)
	if code != 0 {
		t.Fatalf("run record exited %d", code)
	}
	if !strings.Contains(stdout, "Run recorded.") {
		t.Fatalf("expected 'Run recorded.' in stdout, got: %s", stdout)
	}

	// Verify stats show the run
	statsOut, _, statsCode := e.run("stats", "runs")
	if statsCode != 0 {
		t.Fatalf("stats runs exited %d", statsCode)
	}
	if !strings.Contains(statsOut, "go") {
		t.Fatalf("expected 'go' framework in stats, got: %s", statsOut)
	}
}

func TestRunSuggestSkipsOnNo(t *testing.T) {
	e := newEnv(t)

	// Create a minimal Go project
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}

	// Pipe 'n' to run suggest
	stdout, _, code := e.runWithInputInDir(projDir, "n\n",
		"run", "suggest",
		"--task", "add health check",
		"--files-changed", "main.go",
	)
	if code != 0 {
		t.Fatalf("run suggest exited %d", code)
	}
	if !strings.Contains(stdout, "Skipped.") {
		t.Fatalf("expected 'Skipped.' in stdout, got: %s", stdout)
	}

	// Verify no runs recorded
	statsOut, _, statsCode := e.run("stats", "runs")
	if statsCode != 0 {
		t.Fatalf("stats runs exited %d", statsCode)
	}
	// Should show headers but no data rows
	if strings.Contains(statsOut, "go") && strings.Contains(statsOut, "Runs") {
		// It might show the framework if there are runs; let's check more carefully
		// The stats table has a data row with numbers. If no runs, only headers appear.
		lines := strings.Split(strings.TrimSpace(statsOut), "\n")
		// Last line should be a separator or header; no data rows
		if len(lines) > 3 {
			t.Fatalf("expected no runs recorded, but stats shows data: %s", statsOut)
		}
	}
}
