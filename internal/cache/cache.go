package cache

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
	"github.com/gleicon/technocore/internal/embeddings"
	"github.com/gleicon/technocore/internal/project"
	"github.com/gleicon/technocore/internal/summarizer"
)

// Manager handles cache operations for a project.
type Manager struct {
	GlobalDB   *sql.DB
	ProjectDB  *sql.DB
	ProjectDir string
}

// OpenManager opens both global and project DBs.
func OpenManager(cfg *config.Config, dir string) (*Manager, error) {
	if err := cfg.EnsureDirs(); err != nil {
		return nil, err
	}
	if err := cfg.EnsureProjectDir(dir); err != nil {
		return nil, err
	}

	gdb, err := db.Open(cfg.GlobalDBPath)
	if err != nil {
		return nil, err
	}
	if err := db.InitGlobalSchema(gdb); err != nil {
		return nil, err
	}

	pdbPath := cfg.ProjectDBPath(dir)
	pdb, err := db.Open(pdbPath)
	if err != nil {
		return nil, err
	}
	if err := db.InitProjectSchema(pdb); err != nil {
		return nil, err
	}

	return &Manager{
		GlobalDB:   gdb,
		ProjectDB:  pdb,
		ProjectDir: dir,
	}, nil
}

// Close closes both DBs.
func (m *Manager) Close() {
	if m.GlobalDB != nil {
		m.GlobalDB.Close()
	}
	if m.ProjectDB != nil {
		m.ProjectDB.Close()
	}
}

// BuildMap detects the project and stores the map.
func (m *Manager) BuildMap() (*project.Map, error) {
	pm, err := project.Detect(m.ProjectDir)
	if err != nil {
		return nil, err
	}

	entry, _ := json.Marshal(pm.Entrypoints)
	mods, _ := json.Marshal(pm.ModuleBoundaries)
	imps, _ := json.Marshal(pm.ImportantDirs)
	ign, _ := json.Marshal(pm.IgnoredAreas)

	_, err = m.ProjectDB.Exec(`
		INSERT INTO project_map (id, language, framework, package_manager, entrypoints, module_boundaries, important_dirs, ignored_areas)
		VALUES (1, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			language=excluded.language,
			framework=excluded.framework,
			package_manager=excluded.package_manager,
			entrypoints=excluded.entrypoints,
			module_boundaries=excluded.module_boundaries,
			important_dirs=excluded.important_dirs,
			ignored_areas=excluded.ignored_areas,
			updated_at=CURRENT_TIMESTAMP
	`, pm.Language, pm.Framework, pm.PackageManager, entry, mods, imps, ign)
	if err != nil {
		return nil, fmt.Errorf("storing project map: %w", err)
	}
	return pm, nil
}

// GetMap returns the stored project map.
func (m *Manager) GetMap() (*project.Map, error) {
	row := m.ProjectDB.QueryRow(`SELECT language, framework, package_manager, entrypoints, module_boundaries, important_dirs, ignored_areas FROM project_map WHERE id=1`)
	var lang, fw, pm string
	var entry, mods, imps, ign string
	if err := row.Scan(&lang, &fw, &pm, &entry, &mods, &imps, &ign); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	p := &project.Map{Language: lang, Framework: fw, PackageManager: pm}
	json.Unmarshal([]byte(entry), &p.Entrypoints)
	json.Unmarshal([]byte(mods), &p.ModuleBoundaries)
	json.Unmarshal([]byte(imps), &p.ImportantDirs)
	json.Unmarshal([]byte(ign), &p.IgnoredAreas)
	return p, nil
}

// BuildCache indexes the project, creates summaries and embeddings.
func (m *Manager) BuildCache(sentences int) error {
	pm, err := m.GetMap()
	if err != nil {
		return err
	}
	if pm == nil {
		pm, err = m.BuildMap()
		if err != nil {
			return err
		}
	}

	ignored := make(map[string]bool)
	for _, d := range pm.IgnoredAreas {
		ignored[d] = true
	}

	// Walk and index files
	err = filepath.Walk(m.ProjectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if path == m.ProjectDir {
				return nil
			}
			if ignored[info.Name()] || strings.HasPrefix(info.Name(), ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !shouldIndex(path) {
			return nil
		}
		if err := m.indexFile(path, sentences); err != nil {
			// Log and continue
		}
		return nil
	})
	if err != nil {
		return err
	}

	// Build subsystem summaries heuristically
	return m.buildSubsystems()
}

// Refresh updates only changed files.
func (m *Manager) Refresh(sentences int) error {
	pm, err := m.GetMap()
	if err != nil {
		return err
	}
	if pm == nil {
		return m.BuildCache(sentences)
	}

	ignored := make(map[string]bool)
	for _, d := range pm.IgnoredAreas {
		ignored[d] = true
	}

	return filepath.Walk(m.ProjectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			if info != nil && info.IsDir() {
				if path == m.ProjectDir {
					return nil
				}
				if ignored[info.Name()] || strings.HasPrefix(info.Name(), ".") {
					return filepath.SkipDir
				}
			}
			return nil
		}
		if !shouldIndex(path) {
			return nil
		}
		// Check hash
		fh, err := fileHash(path)
		if err != nil {
			return nil
		}
		var storedHash string
		row := m.ProjectDB.QueryRow(`SELECT hash FROM files WHERE path=?`, path)
		if row.Scan(&storedHash) == nil && storedHash == fh {
			return nil // unchanged
		}
		return m.indexFile(path, sentences)
	})
}

// Invalidate removes summaries for changed files.
func (m *Manager) Invalidate() error {
	rows, err := m.ProjectDB.Query(`SELECT path, hash FROM files`)
	if err != nil {
		return err
	}
	defer rows.Close()

	var toDelete []string
	for rows.Next() {
		var path, hash string
		if err := rows.Scan(&path, &hash); err != nil {
			continue
		}
		cur, err := fileHash(path)
		if err != nil || cur != hash {
			toDelete = append(toDelete, path)
		}
	}

	for _, p := range toDelete {
		m.ProjectDB.Exec(`DELETE FROM files WHERE path=?`, p)
		m.ProjectDB.Exec(`DELETE FROM file_search WHERE path=?`, p)
	}
	return nil
}

// indexFile reads, summarizes, chunks, embeds, and stores a file.
func (m *Manager) indexFile(path string, sentences int) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	body, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	content := string(body)
	if len(content) == 0 {
		return nil
	}
	fh, _ := fileHash(path)

	summary, err := summarizer.Summarize(content, sentences)
	if err != nil {
		summary = ""
	}

	var fileID int64
	row := m.ProjectDB.QueryRow(`SELECT id FROM files WHERE path=?`, path)
	if err := row.Scan(&fileID); err == sql.ErrNoRows {
		res, err := m.ProjectDB.Exec(`INSERT INTO files(path, hash, content, summary) VALUES (?,?,?,?)`, path, fh, content, summary)
		if err != nil {
			return err
		}
		fileID, _ = res.LastInsertId()
	} else if err != nil {
		return err
	} else {
		_, err = m.ProjectDB.Exec(`UPDATE files SET hash=?, content=?, summary=?, updated_at=CURRENT_TIMESTAMP WHERE id=?`, fh, content, summary, fileID)
		if err != nil {
			return err
		}
	}

	// FTS5 sync
	m.ProjectDB.Exec(`DELETE FROM file_search WHERE path=?`, path)
	m.ProjectDB.Exec(`INSERT INTO file_search(path, content, summary) VALUES (?,?,?)`, path, content, summary)

	// file summary
	_, err = m.ProjectDB.Exec(`INSERT INTO file_summaries(file_id, summary, patterns) VALUES (?, ?, ?) ON CONFLICT(file_id) DO UPDATE SET summary=excluded.summary, updated_at=CURRENT_TIMESTAMP`, fileID, summary, "[]")
	if err != nil {
		return err
	}

	// Chunking and embeddings
	m.ProjectDB.Exec(`DELETE FROM chunks WHERE file_id=?`, fileID)
	chunks := chunkText(content, 1500)
	for _, c := range chunks {
		emb := embeddings.Compute(c)
		_, err := m.ProjectDB.Exec(`INSERT INTO chunks(file_id, chunk_text, embedding) VALUES (?,?,?)`, fileID, c, embeddings.ToBytes(emb))
		if err != nil {
			return err
		}
	}
	return nil
}

// buildSubsystems heuristically groups files into subsystems.
func (m *Manager) buildSubsystems() error {
	pm, err := m.GetMap()
	if err != nil || pm == nil {
		return nil
	}

	subsystems := make(map[string][]string)
	rows, err := m.ProjectDB.Query(`SELECT path FROM files`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			continue
		}
		// Simple heuristic: group by top-level directory after project root
		rel, _ := filepath.Rel(m.ProjectDir, path)
		parts := strings.Split(rel, string(filepath.Separator))
		if len(parts) > 1 {
			sub := parts[0]
			if sub == "src" && len(parts) > 2 {
				sub = parts[1]
			}
			subsystems[sub] = append(subsystems[sub], path)
		}
	}

	for name, files := range subsystems {
		if len(files) < 2 {
			continue
		}
		scope, _ := json.Marshal(files)
		_, err := m.ProjectDB.Exec(`
			INSERT INTO subsystem_summaries(name, scope_files, summary, patterns)
			VALUES (?, ?, ?, ?)
			ON CONFLICT(name) DO UPDATE SET
				scope_files=excluded.scope_files,
				updated_at=CURRENT_TIMESTAMP
		`, name, scope, fmt.Sprintf("%s subsystem with %d files", name, len(files)), "[]")
		if err != nil {
			return err
		}
	}
	return nil
}

// Inspect prints cache contents.
func (m *Manager) Inspect() (string, error) {
	var out strings.Builder
	out.WriteString("=== Project Map ===\n")
	pm, err := m.GetMap()
	if err != nil || pm == nil {
		out.WriteString("not built yet\n")
	} else {
		out.WriteString(pm.String())
	}

	out.WriteString("\n=== Files ===\n")
	rows, err := m.ProjectDB.Query(`SELECT path, summary FROM files`)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		var p, s string
		if err := rows.Scan(&p, &s); err != nil {
			continue
		}
		out.WriteString(fmt.Sprintf("- %s | %s\n", p, truncate(s, 60)))
		count++
	}
	out.WriteString(fmt.Sprintf("Total: %d\n", count))

	out.WriteString("\n=== Subsystems ===\n")
	srows, err := m.ProjectDB.Query(`SELECT name, summary FROM subsystem_summaries`)
	if err != nil {
		return "", err
	}
	defer srows.Close()
	for srows.Next() {
		var n, s string
		if err := srows.Scan(&n, &s); err != nil {
			continue
		}
		out.WriteString(fmt.Sprintf("- %s | %s\n", n, truncate(s, 60)))
	}

	out.WriteString("\n=== Memories ===\n")
	mrows, err := m.ProjectDB.Query(`SELECT kind, content FROM memories ORDER BY created_at DESC LIMIT 10`)
	if err != nil {
		return "", err
	}
	defer mrows.Close()
	for mrows.Next() {
		var k, c string
		if err := mrows.Scan(&k, &c); err != nil {
			continue
		}
		out.WriteString(fmt.Sprintf("[%s] %s\n", k, truncate(c, 80)))
	}

	return out.String(), nil
}

// StoreMemory saves an insight or memory.
func (m *Manager) StoreMemory(kind, content, context string) error {
	_, err := m.ProjectDB.Exec(`INSERT INTO memories(kind, content, context) VALUES (?,?,?)`, kind, content, context)
	return err
}

// StoreRun records a model run for learning.
func (m *Manager) StoreRun(taskType, framework, modelName string, filesIncluded, filesChanged []string, testsPassed, followUp, inputTokens, outputTokens, accepted int) error {
	fi, _ := json.Marshal(filesIncluded)
	fc, _ := json.Marshal(filesChanged)
	_, err := m.ProjectDB.Exec(`INSERT INTO runs(task_type, framework, files_included, files_changed, tests_passed, follow_up_needed, input_tokens, output_tokens, accepted) VALUES (?,?,?,?,?,?,?,?,?)`, taskType, framework, fi, fc, testsPassed, followUp, inputTokens, outputTokens, accepted)
	if err != nil {
		return err
	}
	// Update global stats
	_, err = m.GlobalDB.Exec(`INSERT INTO model_behavior_stats(model_name, task_type, framework, files_included, files_changed, tests_passed, follow_up_needed, input_tokens, output_tokens, accepted) VALUES (?,?,?,?,?,?,?,?,?,?)`, modelName, taskType, framework, fi, fc, testsPassed, followUp, inputTokens, outputTokens, accepted)
	return err
}

// Stats returns cache stats.
func (m *Manager) Stats() (string, error) {
	var out strings.Builder
	var fcount, ccount, mcount int
	m.ProjectDB.QueryRow(`SELECT COUNT(*) FROM files`).Scan(&fcount)
	m.ProjectDB.QueryRow(`SELECT COUNT(*) FROM chunks`).Scan(&ccount)
	m.ProjectDB.QueryRow(`SELECT COUNT(*) FROM memories`).Scan(&mcount)
	out.WriteString(fmt.Sprintf("Files: %d\nChunks: %d\nMemories: %d\n", fcount, ccount, mcount))

	var gcount int
	m.GlobalDB.QueryRow(`SELECT COUNT(*) FROM model_behavior_stats`).Scan(&gcount)
	out.WriteString(fmt.Sprintf("Global run stats: %d\n", gcount))
	return out.String(), nil
}

// Cleanup removes old entries.
func (m *Manager) Cleanup(days int) error {
	if days <= 0 {
		days = 30
	}
	_, err := m.ProjectDB.Exec(`DELETE FROM memories WHERE created_at < datetime('now', '-`+fmt.Sprint(days)+` days')`)
	if err != nil {
		return err
	}
	_, err = m.ProjectDB.Exec(`DELETE FROM chunks WHERE file_id NOT IN (SELECT id FROM files)`)
	return err
}

// shouldIndex decides if a file path is worth indexing.
func shouldIndex(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".go", ".ts", ".tsx", ".js", ".jsx", ".py", ".rs", ".java", ".rb", ".php", ".c", ".cpp", ".h", ".md", ".txt", ".sql", ".prisma", ".yaml", ".yml", ".json", ".toml", ".mod", ".sum":
		return true
	}
	name := filepath.Base(path)
	if name == "Dockerfile" || name == "Makefile" || name == ".gitignore" || name == "README" || strings.HasPrefix(name, "README") {
		return true
	}
	return false
}

func fileHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func chunkText(text string, size int) []string {
	var chunks []string
	var cur strings.Builder
	for _, line := range strings.Split(text, "\n") {
		if cur.Len()+len(line) > size && cur.Len() > 0 {
			chunks = append(chunks, cur.String())
			cur.Reset()
		}
		cur.WriteString(line)
		cur.WriteByte('\n')
	}
	if cur.Len() > 0 {
		chunks = append(chunks, cur.String())
	}
	return chunks
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
