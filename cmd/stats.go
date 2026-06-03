package cmd

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache and run statistics",
}

var statsCacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Show project cache statistics",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		out, err := m.Stats()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Print(out)
	},
}

var statsRecipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "Show recipe usage statistics",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		gdb, err := db.Open(cfg.GlobalDBPath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer gdb.Close()
		if err := db.InitGlobalSchema(gdb); err != nil {
			fmt.Println("Error:", err)
			return
		}

		rows, err := gdb.Query(`SELECT name, use_count, avg_score, framework FROM task_recipes ORDER BY use_count DESC`)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer rows.Close()

		fmt.Println("Recipe               | Uses | Avg Score | Framework")
		fmt.Println(strings.Repeat("-", 55))
		for rows.Next() {
			var name, fw string
			var useCount int
			var avgScore sql.NullFloat64
			if err := rows.Scan(&name, &useCount, &avgScore, &fw); err != nil {
				continue
			}
			score := 0.0
			if avgScore.Valid {
				score = avgScore.Float64
			}
			fmt.Printf("%-20s | %4d | %9.2f | %s\n", name, useCount, score, fw)
		}
	},
}

var statsRunsCmd = &cobra.Command{
	Use:   "runs",
	Short: "Show run statistics",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		gdb, err := db.Open(cfg.GlobalDBPath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer gdb.Close()
		if err := db.InitGlobalSchema(gdb); err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Per-framework aggregation
		rows, err := gdb.Query(`
			SELECT framework,
				COUNT(*) as runs,
				AVG(input_tokens) as avg_in,
				AVG(output_tokens) as avg_out,
				AVG(CASE WHEN input_tokens > 0 THEN (input_tokens - output_tokens)*100.0/input_tokens ELSE 0 END) as avg_reduction
			FROM model_behavior_stats
			GROUP BY framework
		`)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer rows.Close()

		fmt.Println("Framework      | Runs | Avg In | Avg Out | Reduction %")
		fmt.Println(strings.Repeat("-", 60))
		for rows.Next() {
			var fw string
			var runs int
			var avgIn, avgOut, avgRed sql.NullFloat64
			if err := rows.Scan(&fw, &runs, &avgIn, &avgOut, &avgRed); err != nil {
				continue
			}
			fmt.Printf("%-14s | %4d | %6.0f | %7.0f | %11.1f\n", fw, runs, nullFloat(avgIn), nullFloat(avgOut), nullFloat(avgRed))
		}
	},
}

func nullFloat(n sql.NullFloat64) float64 {
	if n.Valid {
		return n.Float64
	}
	return 0
}

var statsGlobalCmd = &cobra.Command{
	Use:   "global",
	Short: "Show cross-project global statistics",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		gs, err := cache.AggregateGlobalStats(cfg)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Print(gs.String())
	},
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.AddCommand(statsCacheCmd)
	statsCmd.AddCommand(statsRecipesCmd)
	statsCmd.AddCommand(statsRunsCmd)
	statsCmd.AddCommand(statsGlobalCmd)
}
