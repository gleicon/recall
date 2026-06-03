package project

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGo(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example\n\ngo 1.22\n"), 0644)

	m, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Language != "go" {
		t.Fatalf("expected language 'go', got %s", m.Language)
	}
	if m.PackageManager != "go" {
		t.Fatalf("expected package manager 'go', got %s", m.PackageManager)
	}
}

func TestDetectNode(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "package.json"), []byte(`{"dependencies":{"next":"14"}}`), 0644)

	m, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Language != "typescript" {
		t.Fatalf("expected language 'typescript', got %s", m.Language)
	}
	if m.Framework != "nextjs" {
		t.Fatalf("expected framework 'nextjs', got %s", m.Framework)
	}
}

func TestDetectPython(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "requirements.txt"), []byte("fastapi\n"), 0644)

	m, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Language != "python" {
		t.Fatalf("expected language 'python', got %s", m.Language)
	}
}

func TestDetectRust(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "Cargo.toml"), []byte("[package]\nname = \"test\"\n"), 0644)

	m, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Language != "rust" {
		t.Fatalf("expected language 'rust', got %s", m.Language)
	}
	if m.PackageManager != "cargo" {
		t.Fatalf("expected package manager 'cargo', got %s", m.PackageManager)
	}
}

func TestDetectEmpty(t *testing.T) {
	dir := t.TempDir()
	m, err := Detect(dir)
	if err != nil {
		t.Fatal(err)
	}
	if m.Language != "unknown" {
		t.Fatalf("expected language 'unknown', got %s", m.Language)
	}
}
