package search

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/gleicon/technocore/internal/embeddings"
)

// Result is a single search result.
type Result struct {
	Path    string  `json:"path"`
	Score   float64 `json:"score"`
	Summary string  `json:"summary"`
	Content string  `json:"content"`
}

// Engine performs search over the project DB.
type Engine struct {
	DB *sql.DB
}

// NewEngine creates a search engine.
func NewEngine(db *sql.DB) *Engine {
	return &Engine{DB: db}
}

// Query searches the project cache using FTS5 and optional vector re-ranking.
func (e *Engine) Query(q string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}

	// Step 1: FTS5 candidate retrieval
	ftsQuery := strings.Join(strings.Fields(q), " ")
	rows, err := e.DB.Query(`
		SELECT path, content, summary FROM file_search
		WHERE file_search MATCH ?
		ORDER BY rank
		LIMIT ?
	`, ftsQuery, limit*3)
	if err != nil {
		// FTS5 may fail on empty db or syntax issues
		return nil, fmt.Errorf("fts5 query: %w", err)
	}
	defer rows.Close()

	var candidates []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Path, &r.Content, &r.Summary); err != nil {
			continue
		}
		r.Score = 1.0 // base FTS score
		candidates = append(candidates, r)
	}
	if len(candidates) == 0 {
		return []Result{}, nil
	}

	// Step 2: compute embedding for query and re-rank candidates by vector similarity
	qVec := embeddings.Compute(q)
	for i := range candidates {
		cVec := embeddings.Compute(candidates[i].Content)
		cos := embeddings.Cosine(qVec, cVec)
		// Combine FTS rank and cosine similarity
		candidates[i].Score = 0.5 + 0.5*cos
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	if len(candidates) > limit {
		candidates = candidates[:limit]
	}
	return candidates, nil
}

// ChunkQuery searches chunk-level content.
func (e *Engine) ChunkQuery(q string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}
	ftsQuery := strings.Join(strings.Fields(q), " ")
	rows, err := e.DB.Query(`
		SELECT file_id, chunk_text FROM chunk_search
		WHERE chunk_search MATCH ?
		ORDER BY rank
		LIMIT ?
	`, ftsQuery, limit*3)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type chunk struct {
		fileID int64
		text   string
	}
	var chunks []chunk
	for rows.Next() {
		var c chunk
		if err := rows.Scan(&c.fileID, &c.text); err != nil {
			continue
		}
		chunks = append(chunks, c)
	}

	qVec := embeddings.Compute(q)
	var results []Result
	for _, c := range chunks {
		cVec := embeddings.Compute(c.text)
		score := embeddings.Cosine(qVec, cVec)
		var path string
		e.DB.QueryRow(`SELECT path FROM files WHERE id=?`, c.fileID).Scan(&path)
		results = append(results, Result{
			Path:    path,
			Score:   score,
			Content: truncate(c.text, 200),
		})
	}

	sort.Slice(results, func(i, j int) bool {
		return results[i].Score > results[j].Score
	})
	if len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
