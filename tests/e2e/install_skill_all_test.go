package e2e

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInstallSkillOpencodeCreatesSkillFile(t *testing.T) {
	e := newEnv(t)
	_, _, code := e.run("install-skill", "--target", "opencode")
	if code != 0 {
		t.Fatalf("install-skill opencode exited %d", code)
	}
	skillPath := filepath.Join(e.HomeDir, ".opencode", "skills", "technocore", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("expected skill file at %s", skillPath)
	}
}

func TestInstallSkillCursorCreatesSkillFile(t *testing.T) {
	e := newEnv(t)
	_, _, code := e.run("install-skill", "--target", "cursor")
	if code != 0 {
		t.Fatalf("install-skill cursor exited %d", code)
	}
	skillPath := filepath.Join(e.HomeDir, ".cursor", "skills", "technocore", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("expected skill file at %s", skillPath)
	}
}

func TestInstallSkillCodexCreatesSkillFile(t *testing.T) {
	e := newEnv(t)
	_, _, code := e.run("install-skill", "--target", "codex")
	if code != 0 {
		t.Fatalf("install-skill codex exited %d", code)
	}
	skillPath := filepath.Join(e.HomeDir, ".codex", "skills", "technocore", "SKILL.md")
	if _, err := os.Stat(skillPath); os.IsNotExist(err) {
		t.Fatalf("expected skill file at %s", skillPath)
	}
}
