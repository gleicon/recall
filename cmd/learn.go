package cmd

import (
	"fmt"
	"strings"

	"github.com/gleicon/recall/internal/cache"
	"github.com/gleicon/recall/internal/config"
	"github.com/spf13/cobra"
)

var learnCmd = &cobra.Command{
	Use:   "learn [insight]",
	Short: "Store an insight or memory into the local project cache",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		content := strings.Join(args, " ")
		kind, _ := cmd.Flags().GetString("kind")
		ctx, _ := cmd.Flags().GetString("context")
		if err := m.StoreMemory(kind, content, ctx); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Insight stored.")
	},
}

func init() {
	rootCmd.AddCommand(learnCmd)
	learnCmd.Flags().StringP("kind", "k", "insight", "Kind of memory (insight, pattern, mistake, recipe)")
	learnCmd.Flags().StringP("context", "c", "", "Additional context")
}
