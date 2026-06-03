package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRecipesAddInvalidJSONExitsNonZero(t *testing.T) {
	e := newEnv(t)

	// Create an invalid recipe JSON (missing required brief_template)
	badJSON := `{"name": "bad_recipe"}`
	badPath := filepath.Join(e.HomeDir, "bad.json")
	if err := os.WriteFile(badPath, []byte(badJSON), 0644); err != nil {
		t.Fatal(err)
	}

	stdout, stderr, code := e.run("recipes", "add", "--from-file", badPath)
	if code == 0 {
		t.Fatalf("expected non-zero exit for bad JSON, got 0. stdout=%s stderr=%s", stdout, stderr)
	}
	combined := stdout + stderr
	if !strings.Contains(combined, "brief_template") && !strings.Contains(combined, "missing") && !strings.Contains(combined, "Error") {
		t.Fatalf("expected validation error mentioning missing field, got stdout=%s stderr=%s", stdout, stderr)
	}
}
