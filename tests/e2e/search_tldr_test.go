package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCacheBuildIndexesFiles(t *testing.T) {
	e := newEnv(t)

	// Create a Go project with a file
	projDir := t.TempDir()
	goMod := filepath.Join(projDir, "go.mod")
	if err := os.WriteFile(goMod, []byte("module example.com/test\n\ngo 1.22\n"), 0644); err != nil {
		t.Fatal(err)
	}
	mainGo := filepath.Join(projDir, "main.go")
	if err := os.WriteFile(mainGo, []byte("package main\nfunc main() {}\n"), 0644); err != nil {
		t.Fatal(err)
	}

	// Map then build cache
	_, stderr, code := e.runInDir(projDir, "map")
	if code != 0 {
		t.Fatalf("map exited %d: stderr=%s", code, stderr)
	}
	_, stderr, code = e.runInDir(projDir, "cache", "build")
	if code != 0 {
		t.Fatalf("cache build exited %d: stderr=%s", code, stderr)
	}

	// Verify cache inspect shows the indexed file
	inspectOut, _, code := e.runInDir(projDir, "cache", "inspect")
	if code != 0 {
		t.Fatalf("cache inspect exited %d", code)
	}
	if !strings.Contains(inspectOut, "main.go") {
		t.Fatalf("expected main.go in cache inspect, got: %s", inspectOut)
	}
}

func TestTldrSummarizesText(t *testing.T) {
	e := newEnv(t)

	input := "The quick brown fox jumps over the lazy dog. This is a classic pangram used for testing fonts and keyboards. It contains every letter of the English alphabet."
	stdout, _, code := e.runWithInput(input, "tldr")
	if code != 0 {
		t.Fatalf("tldr exited %d", code)
	}
	// TLDR should produce a shorter summary
	if len(stdout) == 0 {
		t.Fatalf("expected tldr output, got empty")
	}
	// Should contain key words from input
	if !strings.Contains(strings.ToLower(stdout), "pangram") && !strings.Contains(strings.ToLower(stdout), "alphabet") {
		t.Fatalf("expected summary to mention key concepts, got: %s", stdout)
	}
}
