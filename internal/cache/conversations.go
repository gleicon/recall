package cache

import (
	"database/sql"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/gleicon/recall/internal/embeddings"
)

type Conversation struct {
	ID               int64
	Task             string
	Prompt           string
	Response         string
	ModelName        string
	InputTokens      int
	OutputTokens     int
	Delegated        bool
	DelegationReason string
	ProjectHash      string
	Framework        string
	Embedding        []float32
}

type Snippet struct {
	ID        int64
	Name      string
	Language  string
	Framework string
	Code      string
	Context   string
	Embedding []float32
	UseCount  int
	Source    string
}

func StoreConversation(gdb *sql.DB, c *Conversation) error {
	var embBytes []byte
	if len(c.Embedding) > 0 {
		embBytes = embeddings.ToBytes(c.Embedding)
	} else {
		// Hash fallback so conversations are always searchable by vector
		emb := embeddings.Compute(c.Task + " " + c.Prompt)
		embBytes = embeddings.ToBytes(emb)
	}

	_, err := gdb.Exec(`
		INSERT INTO conversations (task, prompt, response, model_name, input_tokens, output_tokens, delegated, delegation_reason, project_hash, framework, embedding)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, c.Task, c.Prompt, c.Response, c.ModelName, c.InputTokens, c.OutputTokens, boolToInt(c.Delegated), c.DelegationReason, c.ProjectHash, c.Framework, embBytes)
	return err
}

func FindSimilarConversations(gdb *sql.DB, queryVec []float32, limit int) ([]BrainConversation, error) {
	if limit <= 0 {
		limit = 5
	}

	rows, err := gdb.Query(`
		SELECT task, prompt, response, model_name, delegated, delegation_reason, created_at, embedding
		FROM conversations
		WHERE delegated = 0 AND embedding IS NOT NULL
		ORDER BY created_at DESC
		LIMIT 200
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		BrainConversation
		score float64
	}
	var candidates []scored

	for rows.Next() {
		var c BrainConversation
		var del int
		var embBytes []byte
		if err := rows.Scan(&c.Task, &c.Prompt, &c.Response, &c.ModelName, &del, &c.DelegationReason, &c.CreatedAt, &embBytes); err != nil {
			continue
		}
		c.Delegated = del == 1
		if len(embBytes) == 0 || len(queryVec) == 0 {
			continue
		}
		emb := embeddings.FromBytes(embBytes)
		if len(emb) != len(queryVec) {
			continue
		}
		candidates = append(candidates, scored{BrainConversation: c, score: embeddings.Cosine(queryVec, emb)})
	}

	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].score > candidates[j].score
	})

	var out []BrainConversation
	for i := 0; i < len(candidates) && i < limit; i++ {
		out = append(out, candidates[i].BrainConversation)
	}
	return out, nil
}

func StoreSnippet(gdb *sql.DB, s *Snippet) error {
	var emb []byte
	if len(s.Embedding) > 0 {
		emb = embeddings.ToBytes(s.Embedding)
	} else {
		emb = embeddings.ToBytes(embeddings.Compute(s.Code))
	}

	_, err := gdb.Exec(`
		INSERT INTO snippets (name, language, framework, code, context, embedding, use_count, source)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, s.Name, s.Language, s.Framework, s.Code, s.Context, emb, s.UseCount, s.Source)
	return err
}

func ExtractSnippets(text, language, framework string) []Snippet {
	re := regexp.MustCompile("(?s)```(?:(\\w+)\\n)?(.*?)```")
	matches := re.FindAllStringSubmatch(text, -1)

	var snippets []Snippet
	for i, m := range matches {
		lang := language
		if len(m) > 1 && m[1] != "" {
			lang = m[1]
		}
		code := strings.TrimSpace(m[2])
		if len(code) < 10 {
			continue
		}
		snippets = append(snippets, Snippet{
			Name:      fmt.Sprintf("snippet_%d", i),
			Language:  lang,
			Framework: framework,
			Code:      code,
			Context:   text[:min(len(text), 200)],
		})
	}
	return snippets
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func GetLastConversationID(gdb *sql.DB) (int64, error) {
	var id int64
	err := gdb.QueryRow(`SELECT id FROM conversations ORDER BY id DESC LIMIT 1`).Scan(&id)
	if err == sql.ErrNoRows {
		return 0, nil
	}
	return id, err
}

func MarkConversationFeedback(gdb *sql.DB, id int64, good bool, note string) error {
	accepted := 0
	if good {
		accepted = 1
	}
	_, err := gdb.Exec(`UPDATE conversations SET accepted=?, feedback_note=? WHERE id=?`, accepted, note, id)
	return err
}

func StoreLesson(gdb *sql.DB, pattern, framework, modelName, context string, successRate float64) error {
	_, err := gdb.Exec(`
		INSERT INTO agent_lessons (pattern, framework, model_name, success_rate, context)
		VALUES (?, ?, ?, ?, ?)
	`, pattern, framework, modelName, successRate, context)
	return err
}
