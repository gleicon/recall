package cmd

import (
	"fmt"
	"strings"

	"github.com/gleicon/recall/internal/cache"
	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/search"
	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search [terms]",
	Short: "Search indexed project content",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()

		q := strings.Join(args, " ")
		limit, _ := cmd.Flags().GetInt("limit")
		chunks, _ := cmd.Flags().GetBool("chunks")

		settings, _ := cfg.LoadSettings()
		embedModel := ""
		if settings != nil {
			embedModel = settings.EmbedModel
		}
		engine := search.NewEngine(m.ProjectDB, embedModel)
		var results []search.Result
		if chunks {
			results, err = engine.ChunkQuery(q, limit)
		} else {
			results, err = engine.Query(q, limit)
		}
		if err != nil {
			fmt.Println("Search error:", err)
			return
		}

		for _, r := range results {
			fmt.Printf("[%.3f] %s\n  %s\n\n", r.Score, r.Path, r.Content)
		}
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
	searchCmd.Flags().IntP("limit", "n", 10, "Max results")
	searchCmd.Flags().BoolP("chunks", "c", false, "Search at chunk level")
}
