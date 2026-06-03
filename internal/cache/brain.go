package cache

import (
	"database/sql"
	"sort"

	"github.com/gleicon/technocore/internal/embeddings"
)

type BrainSearchResult struct {
	Conversations []BrainConversation
	Snippets      []BrainSnippet
	Lessons       []BrainLesson
}

type BrainConversation struct {
	Task             string
	Prompt           string
	Response         string
	ModelName        string
	Delegated        bool
	DelegationReason string
	CreatedAt        string
}

type BrainSnippet struct {
	Name      string
	Language  string
	Framework string
	Code      string
	Context   string
	UseCount  int
}

type BrainLesson struct {
	Pattern     string
	Framework   string
	ModelName   string
	SuccessRate float64
	Context     string
}

type BrainFrameworkStat struct {
	Framework     string
	Conversations int
	Answered      int
	Delegated     int
	SuccessRate   float64
	TopReason     string
}

type BrainStats struct {
	TotalConversations int
	Answered           int
	Delegated          int
	TopSnippets        []BrainSnippet
	TopLessons         []BrainLesson
	SuccessRate        float64
	AvgTokensSaved     int
	FrameworkStats     []BrainFrameworkStat
}

func SearchBrain(gdb *sql.DB, keyword string, limit int) (*BrainSearchResult, error) {
	if limit <= 0 {
		limit = 10
	}
	like := "%" + keyword + "%"

	result := &BrainSearchResult{}

	convRows, err := gdb.Query(`
		SELECT task, prompt, response, model_name, delegated, delegation_reason, created_at
		FROM conversations
		WHERE task LIKE ? OR prompt LIKE ? OR response LIKE ?
		ORDER BY created_at DESC
		LIMIT ?
	`, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer convRows.Close()
	for convRows.Next() {
		var c BrainConversation
		var del int
		if err := convRows.Scan(&c.Task, &c.Prompt, &c.Response, &c.ModelName, &del, &c.DelegationReason, &c.CreatedAt); err != nil {
			continue
		}
		c.Delegated = del == 1
		result.Conversations = append(result.Conversations, c)
	}

	snipRows, err := gdb.Query(`
		SELECT name, language, framework, code, context, use_count
		FROM snippets
		WHERE name LIKE ? OR code LIKE ? OR context LIKE ?
		ORDER BY use_count DESC
		LIMIT ?
	`, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer snipRows.Close()
	for snipRows.Next() {
		var s BrainSnippet
		if err := snipRows.Scan(&s.Name, &s.Language, &s.Framework, &s.Code, &s.Context, &s.UseCount); err != nil {
			continue
		}
		result.Snippets = append(result.Snippets, s)
	}

	lessonRows, err := gdb.Query(`
		SELECT pattern, framework, model_name, success_rate, context
		FROM agent_lessons
		WHERE pattern LIKE ? OR framework LIKE ? OR context LIKE ?
		ORDER BY success_rate DESC
		LIMIT ?
	`, like, like, like, limit)
	if err != nil {
		return nil, err
	}
	defer lessonRows.Close()
	for lessonRows.Next() {
		var l BrainLesson
		if err := lessonRows.Scan(&l.Pattern, &l.Framework, &l.ModelName, &l.SuccessRate, &l.Context); err != nil {
			continue
		}
		result.Lessons = append(result.Lessons, l)
	}

	return result, nil
}

func ListConversations(gdb *sql.DB, limit int) ([]BrainConversation, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := gdb.Query(`
		SELECT task, prompt, response, model_name, delegated, delegation_reason, created_at
		FROM conversations
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BrainConversation
	for rows.Next() {
		var c BrainConversation
		var del int
		if err := rows.Scan(&c.Task, &c.Prompt, &c.Response, &c.ModelName, &del, &c.DelegationReason, &c.CreatedAt); err != nil {
			continue
		}
		c.Delegated = del == 1
		out = append(out, c)
	}
	return out, nil
}

func ListSnippets(gdb *sql.DB, limit int) ([]BrainSnippet, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := gdb.Query(`
		SELECT name, language, framework, code, context, use_count
		FROM snippets
		ORDER BY use_count DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BrainSnippet
	for rows.Next() {
		var s BrainSnippet
		if err := rows.Scan(&s.Name, &s.Language, &s.Framework, &s.Code, &s.Context, &s.UseCount); err != nil {
			continue
		}
		out = append(out, s)
	}
	return out, nil
}

func ListLessons(gdb *sql.DB, limit int) ([]BrainLesson, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := gdb.Query(`
		SELECT pattern, framework, model_name, success_rate, context
		FROM agent_lessons
		ORDER BY success_rate DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []BrainLesson
	for rows.Next() {
		var l BrainLesson
		if err := rows.Scan(&l.Pattern, &l.Framework, &l.ModelName, &l.SuccessRate, &l.Context); err != nil {
			continue
		}
		out = append(out, l)
	}
	return out, nil
}

func GetBrainStats(gdb *sql.DB) (*BrainStats, error) {
	stats := &BrainStats{}

	var total, answered, delegated int
	row := gdb.QueryRow(`SELECT COUNT(*), SUM(CASE WHEN delegated=0 THEN 1 ELSE 0 END), SUM(delegated) FROM conversations`)
	if err := row.Scan(&total, &answered, &delegated); err == nil {
		stats.TotalConversations = total
		stats.Answered = answered
		stats.Delegated = delegated
		if total > 0 {
			stats.SuccessRate = float64(answered) / float64(total)
		}
		stats.AvgTokensSaved = answered * 2000
	}

	fwRows, err := gdb.Query(`
		SELECT
			COALESCE(framework, 'unknown') as fw,
			COUNT(*) as total,
			SUM(CASE WHEN delegated=0 THEN 1 ELSE 0 END) as answered,
			SUM(delegated) as delegated
		FROM conversations
		GROUP BY framework
		ORDER BY total DESC
	`)
	if err == nil {
		defer fwRows.Close()
		for fwRows.Next() {
			var fs BrainFrameworkStat
			if err := fwRows.Scan(&fs.Framework, &fs.Conversations, &fs.Answered, &fs.Delegated); err != nil {
				continue
			}
			if fs.Conversations > 0 {
				fs.SuccessRate = float64(fs.Answered) / float64(fs.Conversations)
			}
			var reason string
			reasonRow := gdb.QueryRow(`
				SELECT delegation_reason
				FROM conversations
				WHERE framework = ? AND delegated = 1 AND delegation_reason IS NOT NULL
				GROUP BY delegation_reason
				ORDER BY COUNT(*) DESC
				LIMIT 1
			`, fs.Framework)
			if err := reasonRow.Scan(&reason); err == nil {
				fs.TopReason = reason
			}
			stats.FrameworkStats = append(stats.FrameworkStats, fs)
		}
	}

	snippets, _ := ListSnippets(gdb, 5)
	stats.TopSnippets = snippets

	lessons, _ := ListLessons(gdb, 5)
	stats.TopLessons = lessons

	return stats, nil
}

func SearchSnippetsByVector(gdb *sql.DB, queryVec []float32, limit int) ([]BrainSnippet, error) {
	if limit <= 0 {
		limit = 5
	}
	rows, err := gdb.Query(`SELECT name, language, framework, code, context, embedding FROM snippets`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scoredSnippet struct {
		BrainSnippet
		score float64
	}
	var scored []scoredSnippet

	for rows.Next() {
		var s BrainSnippet
		var embBytes []byte
		if err := rows.Scan(&s.Name, &s.Language, &s.Framework, &s.Code, &s.Context, &embBytes); err != nil {
			continue
		}
		if len(embBytes) > 0 && len(queryVec) > 0 {
			emb := embeddings.FromBytes(embBytes)
			scored = append(scored, scoredSnippet{BrainSnippet: s, score: embeddings.Cosine(queryVec, emb)})
		}
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	var out []BrainSnippet
	for i := 0; i < len(scored) && i < limit; i++ {
		out = append(out, scored[i].BrainSnippet)
	}
	return out, nil
}
