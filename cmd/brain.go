package cmd

import (
	"fmt"
	"strings"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
	"github.com/gleicon/technocore/internal/embeddings"
	"github.com/spf13/cobra"
)

var brainCmd = &cobra.Command{
	Use:   "brain",
	Short: "Search the global brain",
}

var brainConversationsCmd = &cobra.Command{
	Use:   "conversations",
	Short: "List recent conversations",
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

		limit, _ := cmd.Flags().GetInt("limit")
		convs, err := cache.ListConversations(gdb, limit)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if len(convs) == 0 {
			fmt.Println("No conversations recorded yet.")
			return
		}
		for _, c := range convs {
			status := "answered"
			if c.Delegated {
				status = "delegated"
			}
			fmt.Printf("[%s] %s | %s | %s\n", c.CreatedAt, status, c.ModelName, c.Task)
			if c.Response != "" && !c.Delegated {
				fmt.Printf("  → %.80s\n", strings.ReplaceAll(c.Response, "\n", " "))
			}
		}
	},
}

var brainSnippetsCmd = &cobra.Command{
	Use:   "snippets",
	Short: "List code snippets",
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

		limit, _ := cmd.Flags().GetInt("limit")
		snippets, err := cache.ListSnippets(gdb, limit)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if len(snippets) == 0 {
			fmt.Println("No snippets recorded yet.")
			return
		}
		for _, s := range snippets {
			fmt.Printf("[%s/%s] %s (used %d)\n", s.Language, s.Framework, s.Name, s.UseCount)
			fmt.Printf("  Context: %s\n", s.Context)
			fmt.Printf("  Code:\n%s\n", indent(s.Code, 4))
			fmt.Println()
		}
	},
}

var brainLessonsCmd = &cobra.Command{
	Use:   "lessons",
	Short: "Show agent lessons",
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

		limit, _ := cmd.Flags().GetInt("limit")
		lessons, err := cache.ListLessons(gdb, limit)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if len(lessons) == 0 {
			fmt.Println("No lessons recorded yet.")
			return
		}
		for _, l := range lessons {
			fmt.Printf("[%s | %.0f%%] %s | %s\n", l.Framework, l.SuccessRate*100, l.Pattern, l.ModelName)
			if l.Context != "" {
				fmt.Printf("  Context: %s\n", l.Context)
			}
		}
	},
}

var brainSearchCmd = &cobra.Command{
	Use:   "search [keywords]",
	Short: "Search brain",
	Args:  cobra.MinimumNArgs(1),
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

		keyword := strings.Join(args, " ")
		limit, _ := cmd.Flags().GetInt("limit")
		vector, _ := cmd.Flags().GetBool("vector")

		if vector {
			qVec := embeddings.ComputeSmart(keyword)
			snippets, err := cache.SearchSnippetsByVector(gdb, qVec, limit)
			if err != nil {
				fmt.Println("Error:", err)
				return
			}
			fmt.Printf("=== Vector Snippet Matches for '%s' ===\n", keyword)
			for _, s := range snippets {
				fmt.Printf("[%s/%s] %s\n", s.Language, s.Framework, s.Name)
				fmt.Printf("  %s\n", indent(s.Code, 2))
			}
			return
		}

		result, err := cache.SearchBrain(gdb, keyword, limit)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Printf("=== Brain Search: '%s' ===\n\n", keyword)

		if len(result.Conversations) > 0 {
			fmt.Printf("--- Conversations (%d) ---\n", len(result.Conversations))
			for _, c := range result.Conversations {
				status := "answered"
				if c.Delegated {
					status = "delegated"
				}
				fmt.Printf("  [%s] %s | %s\n", status, c.ModelName, c.Task)
			}
			fmt.Println()
		}

		if len(result.Snippets) > 0 {
			fmt.Printf("--- Snippets (%d) ---\n", len(result.Snippets))
			for _, s := range result.Snippets {
				fmt.Printf("  [%s/%s] %s (used %d)\n", s.Language, s.Framework, s.Name, s.UseCount)
			}
			fmt.Println()
		}

		if len(result.Lessons) > 0 {
			fmt.Printf("--- Lessons (%d) ---\n", len(result.Lessons))
			for _, l := range result.Lessons {
				fmt.Printf("  [%s | %.0f%%] %s\n", l.Framework, l.SuccessRate*100, l.Pattern)
			}
			fmt.Println()
		}

		if len(result.Conversations) == 0 && len(result.Snippets) == 0 && len(result.Lessons) == 0 {
			fmt.Println("No matches found.")
		}
	},
}

var brainStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show brain stats",
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

		stats, err := cache.GetBrainStats(gdb)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("=== Brain Statistics ===")
		fmt.Printf("Total conversations: %d\n", stats.TotalConversations)
		fmt.Printf("  - Answered by local model: %d\n", stats.Answered)
		fmt.Printf("  - Delegated to big model:  %d\n", stats.Delegated)
		if stats.TotalConversations > 0 {
			fmt.Printf("Success rate: %.1f%%\n", stats.SuccessRate*100)
			fmt.Printf("Estimated tokens saved: %d\n", stats.AvgTokensSaved)
		}
		fmt.Println()

		if len(stats.TopSnippets) > 0 {
			fmt.Printf("--- Top Snippets (%d) ---\n", len(stats.TopSnippets))
			for _, s := range stats.TopSnippets {
				fmt.Printf("  [%s/%s] %s (used %d)\n", s.Language, s.Framework, s.Name, s.UseCount)
			}
			fmt.Println()
		}

		if len(stats.TopLessons) > 0 {
			fmt.Printf("--- Top Lessons (%d) ---\n", len(stats.TopLessons))
			for _, l := range stats.TopLessons {
				fmt.Printf("  [%s | %.0f%%] %s | %s\n", l.Framework, l.SuccessRate*100, l.Pattern, l.ModelName)
			}
			fmt.Println()
		}

		if len(stats.FrameworkStats) > 0 {
			fmt.Printf("--- Per-Framework Performance ---\n")
			fmt.Printf("%-12s | Total | Answered | Delegated | Success | Top Reason\n", "Framework")
			fmt.Println(strings.Repeat("-", 75))
			for _, fw := range stats.FrameworkStats {
				fmt.Printf("%-12s | %5d | %8d | %9d | %6.0f%% | %s\n",
					fw.Framework, fw.Conversations, fw.Answered, fw.Delegated,
					fw.SuccessRate*100, fw.TopReason)
			}
			fmt.Println()
		}
	},
}

var brainFrameworksCmd = &cobra.Command{
	Use:   "frameworks",
	Short: "Show framework stats",
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

		stats, err := cache.GetBrainStats(gdb)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		if len(stats.FrameworkStats) == 0 {
			fmt.Println("No framework data yet.")
			return
		}

		fmt.Println("=== Per-Framework Brain Performance ===")
		fmt.Printf("%-12s | Total | Answered | Delegated | Success | Top Delegation Reason\n", "Framework")
		fmt.Println(strings.Repeat("-", 85))
		for _, fw := range stats.FrameworkStats {
			fmt.Printf("%-12s | %5d | %8d | %9d | %6.0f%% | %s\n",
				fw.Framework, fw.Conversations, fw.Answered, fw.Delegated,
				fw.SuccessRate*100, fw.TopReason)
		}
	},
}

func init() {
	rootCmd.AddCommand(brainCmd)
	brainCmd.AddCommand(brainConversationsCmd)
	brainCmd.AddCommand(brainSnippetsCmd)
	brainCmd.AddCommand(brainLessonsCmd)
	brainCmd.AddCommand(brainSearchCmd)
	brainCmd.AddCommand(brainStatsCmd)
	brainCmd.AddCommand(brainFrameworksCmd)

	for _, c := range []*cobra.Command{brainConversationsCmd, brainSnippetsCmd, brainLessonsCmd, brainSearchCmd} {
		c.Flags().IntP("limit", "n", 20, "Max results")
	}
	brainSearchCmd.Flags().BoolP("vector", "v", false, "Use vector search on snippets")
}
