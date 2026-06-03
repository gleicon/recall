package cache

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
)

// GlobalStats holds aggregated statistics across all projects.
type GlobalStats struct {
	TotalRuns          int
	TotalProjects      int
	AvgTokenReduction  float64
	TopFrameworks      []FrameworkStat
	TopRecipes         []RecipeStat
}

// FrameworkStat holds per-framework aggregated data.
type FrameworkStat struct {
	Framework       string
	Runs            int
	AvgTokenReduction float64
	SuccessRate     float64
}

// RecipeStat holds per-recipe aggregated data.
type RecipeStat struct {
	Name      string
	UseCount  int
	AvgScore  float64
}

// AggregateGlobalStats walks all project DBs and global DB to compute cross-project stats.
func AggregateGlobalStats(cfg *config.Config) (*GlobalStats, error) {
	gs := &GlobalStats{}

	// Open global DB
	gdb, err := db.Open(cfg.GlobalDBPath)
	if err != nil {
		return nil, err
	}
	defer gdb.Close()
	if err := db.InitGlobalSchema(gdb); err != nil {
		return nil, err
	}

	// Count total runs from global stats
	var totalRuns int
	gdb.QueryRow(`SELECT COUNT(*) FROM model_behavior_stats`).Scan(&totalRuns)
	gs.TotalRuns = totalRuns

	// Count projects
	projects, err := os.ReadDir(cfg.ProjectsDir)
	if err != nil {
		return nil, err
	}
	gs.TotalProjects = len(projects)

	// Aggregate per-framework stats
	rows, err := gdb.Query(`
		SELECT framework,
			COUNT(*) as runs,
			AVG(CASE WHEN input_tokens > 0 THEN (input_tokens - output_tokens)*100.0/input_tokens ELSE 0 END) as avg_reduction,
			AVG(CASE WHEN tests_passed = 1 THEN 1.0 ELSE 0.0 END) as success_rate
		FROM model_behavior_stats
		GROUP BY framework
		ORDER BY runs DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var fw string
		var runs int
		var avgRed, success sql.NullFloat64
		if err := rows.Scan(&fw, &runs, &avgRed, &success); err != nil {
			continue
		}
		gs.TopFrameworks = append(gs.TopFrameworks, FrameworkStat{
			Framework:         fw,
			Runs:              runs,
			AvgTokenReduction: nullFloat(avgRed),
			SuccessRate:       nullFloat(success) * 100,
		})
	}

	// Aggregate recipe stats
	rrows, err := gdb.Query(`
		SELECT name, use_count, avg_score
		FROM task_recipes
		ORDER BY use_count DESC, avg_score DESC
		LIMIT 20
	`)
	if err != nil {
		return nil, err
	}
	defer rrows.Close()
	for rrows.Next() {
		var name string
		var useCount int
		var avgScore sql.NullFloat64
		if err := rrows.Scan(&name, &useCount, &avgScore); err != nil {
			continue
		}
		gs.TopRecipes = append(gs.TopRecipes, RecipeStat{
			Name:     name,
			UseCount: useCount,
			AvgScore: nullFloat(avgScore),
		})
	}

	// Aggregate token reduction across all projects
	var avgRed sql.NullFloat64
	gdb.QueryRow(`
		SELECT AVG(CASE WHEN input_tokens > 0 THEN (input_tokens - output_tokens)*100.0/input_tokens ELSE 0 END)
		FROM model_behavior_stats
	`).Scan(&avgRed)
	gs.AvgTokenReduction = nullFloat(avgRed)

	return gs, nil
}

func nullFloat(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}

func (gs *GlobalStats) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, "=== Global Statistics ===\n")
	fmt.Fprintf(&b, "Projects tracked: %d\n", gs.TotalProjects)
	fmt.Fprintf(&b, "Total runs recorded: %d\n", gs.TotalRuns)
	fmt.Fprintf(&b, "Avg token reduction: %.1f%%\n\n", gs.AvgTokenReduction)

	fmt.Fprintf(&b, "=== Per-Framework ===\n")
	fmt.Fprintf(&b, "%-15s | Runs | Reduction %% | Success %%\n", "Framework")
	fmt.Fprintln(&b, strings.Repeat("-", 50))
	for _, fw := range gs.TopFrameworks {
		fmt.Fprintf(&b, "%-15s | %4d | %11.1f | %9.1f\n", fw.Framework, fw.Runs, fw.AvgTokenReduction, fw.SuccessRate)
	}

	fmt.Fprintf(&b, "\n=== Top Recipes ===\n")
	fmt.Fprintf(&b, "%-30s | Uses | Avg Score\n", "Recipe")
	fmt.Fprintln(&b, strings.Repeat("-", 50))
	for _, r := range gs.TopRecipes {
		fmt.Fprintf(&b, "%-30s | %4d | %9.2f\n", r.Name, r.UseCount, r.AvgScore)
	}

	return b.String()
}
