package search

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"

	"github.com/gleicon/recall/internal/embeddings"
	"github.com/gleicon/recall/internal/llm"
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
	DB         *sql.DB
	EmbedModel string
	LLMClient  *llm.Client
}

func NewEngine(db *sql.DB, embedModel string) *Engine {
	return &Engine{
		DB:         db,
		EmbedModel: embedModel,
		LLMClient:  llm.Detect(),
	}
}

func (e *Engine) Query(q string, limit int) ([]Result, error) {
	if limit <= 0 {
		limit = 10
	}

	ftsQuery := strings.Join(strings.Fields(q), " ")
	rows, err := e.DB.Query(`
		SELECT path, content, summary FROM file_search
		WHERE file_search MATCH ?
		ORDER BY rank
		LIMIT ?
	`, ftsQuery, limit*3)
	if err != nil {
		return nil, fmt.Errorf("fts5 query: %w", err)
	}
	defer rows.Close()

	var candidates []Result
	for rows.Next() {
		var r Result
		if err := rows.Scan(&r.Path, &r.Content, &r.Summary); err != nil {
			continue
		}
		r.Score = 1.0 // base FTS rank
		candidates = append(candidates, r)
	}
	if len(candidates) == 0 {
		return []Result{}, nil
	}

	qVec := embeddings.ComputeSmartWithClient(q, e.EmbedModel, e.LLMClient)
	if len(qVec) > embeddings.Dim {
		reranked := false
		for i := range candidates {
			var embBytes []byte
			e.DB.QueryRow(`SELECT embedding FROM files WHERE path=?`, candidates[i].Path).Scan(&embBytes)
			if len(embBytes) == 0 {
				continue
			}
			cVec := embeddings.FromBytes(embBytes)
			if len(cVec) != len(qVec) {
				continue // dimension mismatch — stored with different model
			}
			candidates[i].Score = embeddings.Cosine(qVec, cVec)
			reranked = true
		}
		if reranked {
			sort.Slice(candidates, func(i, j int) bool {
				return candidates[i].Score > candidates[j].Score
			})
		}
	}

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

	qVec := embeddings.ComputeSmartWithClient(q, e.EmbedModel, e.LLMClient)
	var results []Result
	for _, c := range chunks {
		var embBytes []byte
		e.DB.QueryRow(`SELECT embedding FROM chunks WHERE file_id=? AND chunk_text=?`, c.fileID, c.text).Scan(&embBytes)

		var score float64
		if len(embBytes) > 0 && len(qVec) > embeddings.Dim {
			cVec := embeddings.FromBytes(embBytes)
			if len(cVec) == len(qVec) {
				score = embeddings.Cosine(qVec, cVec)
			}
		}
		if score == 0 {
			score = 0.5 // FTS match but no vector score
		}

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
