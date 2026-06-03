package cache

import (
	"database/sql"
	"fmt"
	"regexp"
	"strings"

	"github.com/gleicon/technocore/internal/embeddings"
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
	_, err := gdb.Exec(`
		INSERT INTO conversations (task, prompt, response, model_name, input_tokens, output_tokens, delegated, delegation_reason, project_hash, framework)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, c.Task, c.Prompt, c.Response, c.ModelName, c.InputTokens, c.OutputTokens, boolToInt(c.Delegated), c.DelegationReason, c.ProjectHash, c.Framework)
	return err
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

func StoreLesson(gdb *sql.DB, pattern, framework, modelName, context string, successRate float64) error {
	_, err := gdb.Exec(`
		INSERT INTO agent_lessons (pattern, framework, model_name, success_rate, context)
		VALUES (?, ?, ?, ?, ?)
	`, pattern, framework, modelName, successRate, context)
	return err
}


