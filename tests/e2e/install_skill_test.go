package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInstallSkillClaudeCreatesSkillFile(t *testing.T) {
	e := newEnv(t)

	stdout, stderr, code := e.run("install-skill", "--target", "claude")
	if code != 0 {
		t.Fatalf("install-skill exited %d: stdout=%s stderr=%s", code, stdout, stderr)
	}

	// Verify skill file exists
	skillPath := filepath.Join(e.HomeDir, ".claude", "skills", "technocore", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("expected skill file at %s, not found", skillPath)
	}

	// Verify it references the correct commands
	content, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "technocore brief") {
		t.Fatalf("expected skill to reference 'technocore brief', got: %s", string(content))
	}
	if !strings.Contains(string(content), "technocore run suggest") {
		t.Fatalf("expected skill to reference 'technocore run suggest', got: %s", string(content))
	}
}
