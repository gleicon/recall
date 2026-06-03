package recipes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/gleicon/technocore/internal/embeddings"
)

// Recipe is a reusable task recipe.
type Recipe struct {
	ID              int64    `json:"id"`
	Name            string   `json:"name"`
	Language        string   `json:"language"`
	Framework       string   `json:"framework"`
	Signals         []string `json:"signals"`
	ContextNeeded   []string `json:"context_needed"`
	Avoid           []string `json:"avoid"`
	BriefTemplate   string   `json:"brief_template"`
	Source          string   `json:"source"`
	Tags            []string `json:"tags"`
}

// Store inserts or updates a recipe, computing its embedding.
func Store(db *sql.DB, r *Recipe) error {
	text := r.Name + " " + r.BriefTemplate + " " + join(r.Tags)
	emb := embeddings.Compute(text)
	embBytes := embeddings.ToBytes(emb)

	signals, _ := json.Marshal(r.Signals)
	ctxNeeded, _ := json.Marshal(r.ContextNeeded)
	avoid, _ := json.Marshal(r.Avoid)
	tags, _ := json.Marshal(r.Tags)

	_, err := db.Exec(`
		INSERT INTO task_recipes (name, language, framework, signals, context_needed, avoid, brief_template, embedding, source, tags)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(name) DO UPDATE SET
			language=excluded.language,
			framework=excluded.framework,
			signals=excluded.signals,
			context_needed=excluded.context_needed,
			avoid=excluded.avoid,
			brief_template=excluded.brief_template,
			embedding=excluded.embedding,
			source=excluded.source,
			tags=excluded.tags,
			updated_at=CURRENT_TIMESTAMP
	`, r.Name, r.Language, r.Framework, signals, ctxNeeded, avoid, r.BriefTemplate, embBytes, r.Source, tags)
	return err
}

// LoadFromFile reads and validates a recipe JSON file.
func LoadFromFile(path string) (*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read recipe file: %w", err)
	}
	var r Recipe
	if err := json.Unmarshal(data, &r); err != nil {
		return nil, fmt.Errorf("parse recipe JSON: %w", err)
	}
	if r.Name == "" {
		return nil, fmt.Errorf("recipe missing required field: name")
	}
	if r.BriefTemplate == "" {
		return nil, fmt.Errorf("recipe %s missing required field: brief_template", r.Name)
	}
	return &r, nil
}

// LoadAllFromDir reads every `.json` file in a directory.
func LoadAllFromDir(dir string) ([]*Recipe, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var out []*Recipe
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		r, err := LoadFromFile(filepath.Join(dir, e.Name()))
		if err != nil {
			return nil, fmt.Errorf("%s: %w", e.Name(), err)
		}
		out = append(out, r)
	}
	return out, nil
}

// RecipeMatch holds a recipe and its relevance score.
type RecipeMatch struct {
	Recipe *Recipe
	Score  float64
}

// FindMatches retrieves recipes by vector similarity, then re-ranks by framework/signal overlap.
func FindMatches(db *sql.DB, query string, projectFramework string, projectSignals []string, k int) ([]RecipeMatch, error) {
	qVec := embeddings.Compute(query)

	// Optimization: if framework is known, pre-filter to reduce vector scan
	var rows *sql.Rows
	var err error
	if projectFramework != "" {
		rows, err = db.Query(`SELECT name, language, framework, signals, context_needed, avoid, brief_template, source, tags, embedding FROM task_recipes WHERE framework = ? OR framework = ''`, projectFramework)
	} else {
		rows, err = db.Query(`SELECT name, language, framework, signals, context_needed, avoid, brief_template, source, tags, embedding FROM task_recipes`)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var matches []RecipeMatch
	for rows.Next() {
		var r Recipe
		var signalsJSON, ctxNeededJSON, avoidJSON, tagsJSON string
		var embBytes []byte
		if err := rows.Scan(&r.Name, &r.Language, &r.Framework, &signalsJSON, &ctxNeededJSON, &avoidJSON, &r.BriefTemplate, &r.Source, &tagsJSON, &embBytes); err != nil {
			continue
		}
		json.Unmarshal([]byte(signalsJSON), &r.Signals)
		json.Unmarshal([]byte(ctxNeededJSON), &r.ContextNeeded)
		json.Unmarshal([]byte(avoidJSON), &r.Avoid)
		json.Unmarshal([]byte(tagsJSON), &r.Tags)

		var score float64
		if len(embBytes) > 0 {
			rVec := embeddings.FromBytes(embBytes)
			score = embeddings.Cosine(qVec, rVec)
		}

		// Boost framework match
		if projectFramework != "" && r.Framework == projectFramework {
			score += 0.3
		}

		// Boost signal overlap
		matched := 0
		for _, ps := range projectSignals {
			for _, rs := range r.Signals {
				if ps == rs {
					matched++
				}
			}
		}
		score += float64(matched) * 0.1

		if score > 1.0 {
			score = 1.0
		}

		matches = append(matches, RecipeMatch{Recipe: &r, Score: score})
	}

	// If framework filter yielded too few results, fallback to all recipes
	if projectFramework != "" && len(matches) < k {
		remaining, err := FindMatches(db, query, "", projectSignals, k)
		if err != nil {
			return nil, err
		}
		// Merge without duplicates
		seen := make(map[string]bool)
		for _, m := range matches {
			seen[m.Recipe.Name] = true
		}
		for _, m := range remaining {
			if !seen[m.Recipe.Name] {
				matches = append(matches, m)
			}
		}
		sort.Slice(matches, func(i, j int) bool {
			return matches[i].Score > matches[j].Score
		})
		if len(matches) > k {
			matches = matches[:k]
		}
	}

	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Score > matches[j].Score
	})

	if len(matches) > k {
		matches = matches[:k]
	}
	return matches, nil
}

// List returns all recipes.
func List(db *sql.DB) ([]*Recipe, error) {
	rows, err := db.Query(`SELECT name, language, framework, signals, context_needed, avoid, brief_template, source, tags FROM task_recipes ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*Recipe
	for rows.Next() {
		var r Recipe
		var signalsJSON, ctxNeededJSON, avoidJSON, tagsJSON string
		if err := rows.Scan(&r.Name, &r.Language, &r.Framework, &signalsJSON, &ctxNeededJSON, &avoidJSON, &r.BriefTemplate, &r.Source, &tagsJSON); err != nil {
			continue
		}
		json.Unmarshal([]byte(signalsJSON), &r.Signals)
		json.Unmarshal([]byte(ctxNeededJSON), &r.ContextNeeded)
		json.Unmarshal([]byte(avoidJSON), &r.Avoid)
		json.Unmarshal([]byte(tagsJSON), &r.Tags)
		out = append(out, &r)
	}
	return out, nil
}

func join(ss []string) string {
	return " " + fmt.Sprintf("%v", ss)
}
