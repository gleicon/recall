package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadSettingsDefaults(t *testing.T) {
	tmp := t.TempDir()
	c := &Config{ConfigFilePath: filepath.Join(tmp, "config.json")}

	s, err := c.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.QueryTimeout != 30 {
		t.Fatalf("expected default 30, got %d", s.QueryTimeout)
	}

	os.WriteFile(c.ConfigFilePath, []byte(`{"query_timeout":0}`), 0644)
	s, err = c.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.QueryTimeout != 30 {
		t.Fatalf("expected default 30, got %d", s.QueryTimeout)
	}

	os.WriteFile(c.ConfigFilePath, []byte(`{"query_timeout":500}`), 0644)
	s, err = c.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.QueryTimeout != 500 {
		t.Fatalf("expected preserved 500, got %d", s.QueryTimeout)
	}

	os.WriteFile(c.ConfigFilePath, []byte(`{"query_timeout":-1}`), 0644)
	s, err = c.LoadSettings()
	if err != nil {
		t.Fatal(err)
	}
	if s.QueryTimeout != 30 {
		t.Fatalf("expected default 30, got %d", s.QueryTimeout)
	}
}
