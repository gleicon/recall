package cmd

import (
	"fmt"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove old cache entries",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		days, _ := cmd.Flags().GetInt("days")
		if err := m.Cleanup(days); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Cleanup complete.")
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.Flags().IntP("days", "d", 30, "Delete entries older than N days")
}
