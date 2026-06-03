package project

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Map describes a detected project.
type Map struct {
	Language         string   `json:"language"`
	Framework        string   `json:"framework"`
	PackageManager   string   `json:"package_manager"`
	Entrypoints      []string `json:"entrypoints"`
	ModuleBoundaries []string `json:"module_boundaries"`
	ImportantDirs    []string `json:"important_dirs"`
	IgnoredAreas     []string `json:"ignored_areas"`
	Signals          []string `json:"signals"`
}

// Detect analyzes a directory and returns project metadata.
func Detect(dir string) (*Map, error) {
	m := &Map{
		Entrypoints:      []string{},
		ModuleBoundaries: []string{},
		ImportantDirs:    []string{},
		IgnoredAreas:     []string{"node_modules", ".git", ".next", "dist", "build", "vendor", "target", "__pycache__", ".venv"},
		Signals:          []string{},
	}

	// Node / TypeScript / Next.js
	if exists(dir, "package.json") {
		m.Language = "typescript"
		m.PackageManager = detectPackageManager(dir)
		m.Signals = append(m.Signals, "package.json")
		data, _ := os.ReadFile(filepath.Join(dir, "package.json"))
		if data != nil {
			var pkg map[string]interface{}
			if err := json.Unmarshal(data, &pkg); err == nil {
				deps := mergeMaps(pkg["dependencies"], pkg["devDependencies"])
				if hasKey(deps, "next") {
					m.Framework = "nextjs"
					m.Signals = append(m.Signals, "app/", "middleware.ts", "prisma/schema.prisma")
					m.Entrypoints = append(m.Entrypoints, "src/app/page.tsx", "src/middleware.ts")
					m.ImportantDirs = append(m.ImportantDirs, "src/app", "src/lib", "src/components")
				} else if hasKey(deps, "react") {
					m.Framework = "react"
					m.Entrypoints = append(m.Entrypoints, "src/index.tsx", "src/App.tsx")
				} else if hasKey(deps, "express") {
					m.Framework = "express"
					m.Entrypoints = append(m.Entrypoints, "src/index.ts", "src/server.ts")
				}
			}
		}
		if m.Framework == "" {
			m.Framework = "nodejs"
		}
	}

	// Go
	if exists(dir, "go.mod") {
		m.Language = "go"
		m.PackageManager = "go"
		m.Signals = append(m.Signals, "go.mod")
		m.Entrypoints = append(m.Entrypoints, "main.go", "cmd/")
		m.ModuleBoundaries = append(m.ModuleBoundaries, "internal/", "pkg/")
		m.ImportantDirs = append(m.ImportantDirs, "cmd", "internal", "pkg")
		if exists(dir, "go.work") {
			m.Signals = append(m.Signals, "go.work")
			m.Framework = "go-workspace"
		}
	}

	// Rust
	if exists(dir, "Cargo.toml") {
		m.Language = "rust"
		m.PackageManager = "cargo"
		m.Signals = append(m.Signals, "Cargo.toml")
		m.Entrypoints = append(m.Entrypoints, "src/main.rs", "src/lib.rs")
		m.ModuleBoundaries = append(m.ModuleBoundaries, "src/")
		m.ImportantDirs = append(m.ImportantDirs, "src", "tests")
		if exists(dir, "Cargo.lock") {
			m.Signals = append(m.Signals, "Cargo.lock")
		}
	}

	// Python
	if exists(dir, "pyproject.toml") || exists(dir, "requirements.txt") || exists(dir, "setup.py") {
		m.Language = "python"
		m.Signals = append(m.Signals, "pyproject.toml", "requirements.txt")
		m.Entrypoints = append(m.Entrypoints, "main.py", "app.py")
		m.ImportantDirs = append(m.ImportantDirs, "src", "tests")
		if exists(dir, "pyproject.toml") {
			data, _ := os.ReadFile(filepath.Join(dir, "pyproject.toml"))
			if data != nil && strings.Contains(string(data), "fastapi") {
				m.Framework = "fastapi"
			}
		}
		if m.PackageManager == "" {
			if exists(dir, "poetry.lock") {
				m.PackageManager = "poetry"
			} else if exists(dir, "Pipfile") {
				m.PackageManager = "pipenv"
			} else if exists(dir, "uv.lock") {
				m.PackageManager = "uv"
			} else {
				m.PackageManager = "pip"
			}
		}
	}

	// Prisma
	if exists(dir, "prisma/schema.prisma") {
		m.Signals = append(m.Signals, "prisma/schema.prisma")
		m.ImportantDirs = append(m.ImportantDirs, "prisma")
	}

	if m.Language == "" {
		m.Language = "unknown"
	}

	return m, nil
}

func exists(dir, name string) bool {
	_, err := os.Stat(filepath.Join(dir, name))
	return err == nil
}

func detectPackageManager(dir string) string {
	if exists(dir, "pnpm-lock.yaml") {
		return "pnpm"
	}
	if exists(dir, "yarn.lock") {
		return "yarn"
	}
	if exists(dir, "package-lock.json") {
		return "npm"
	}
	if exists(dir, "bun.lockb") || exists(dir, "bun.lock") {
		return "bun"
	}
	return "npm"
}

func mergeMaps(a, b interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	for _, v := range []interface{}{a, b} {
		if m, ok := v.(map[string]interface{}); ok {
			for k, val := range m {
				out[k] = val
			}
		}
	}
	return out
}

func hasKey(m map[string]interface{}, key string) bool {
	_, ok := m[key]
	return ok
}

// String returns a human-readable project description.
func (m *Map) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "language: %s\n", m.Language)
	fmt.Fprintf(&b, "framework: %s\n", m.Framework)
	fmt.Fprintf(&b, "package_manager: %s\n", m.PackageManager)
	fmt.Fprintf(&b, "signals: %v\n", m.Signals)
	fmt.Fprintf(&b, "entrypoints: %v\n", m.Entrypoints)
	fmt.Fprintf(&b, "important_dirs: %v\n", m.ImportantDirs)
	fmt.Fprintf(&b, "ignored_areas: %v\n", m.IgnoredAreas)
	return b.String()
}
