package config

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
)

// Config holds technocore paths.
type Config struct {
	HomeDir        string
	GlobalDBPath   string
	ProjectsDir    string
}

// NewConfig creates config from environment.
func NewConfig() *Config {
	home, _ := os.UserHomeDir()
	base := filepath.Join(home, ".technocore")
	return &Config{
		HomeDir:        base,
		GlobalDBPath:   filepath.Join(base, "global.db"),
		ProjectsDir:    filepath.Join(base, "projects"),
	}
}

// EnsureDirs creates the base directories if they don't exist.
func (c *Config) EnsureDirs() error {
	for _, d := range []string{c.HomeDir, c.ProjectsDir} {
		if err := os.MkdirAll(d, 0755); err != nil {
			return err
		}
	}
	return nil
}

// ProjectHash returns a stable hash for a project directory.
func ProjectHash(dir string) string {
	abs, _ := filepath.Abs(dir)
	h := sha256.Sum256([]byte(abs))
	return hex.EncodeToString(h[:])[:16]
}

// ProjectDBPath returns the project DB path for a directory.
func (c *Config) ProjectDBPath(dir string) string {
	hash := ProjectHash(dir)
	return filepath.Join(c.ProjectsDir, hash, "project.db")
}

// ProjectDir returns the project data directory for a directory.
func (c *Config) ProjectDir(dir string) string {
	hash := ProjectHash(dir)
	return filepath.Join(c.ProjectsDir, hash)
}

// EnsureProjectDir creates the project directory for a given project root.
func (c *Config) EnsureProjectDir(dir string) error {
	pd := c.ProjectDir(dir)
	if err := os.MkdirAll(pd, 0755); err != nil {
		return fmt.Errorf("creating project dir: %w", err)
	}
	return nil
}
