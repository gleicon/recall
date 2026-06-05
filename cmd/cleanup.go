package cmd

import (
	"fmt"
	"os"

	"github.com/gleicon/recall/internal/cache"
	"github.com/gleicon/recall/internal/config"
	"github.com/spf13/cobra"
)

var cleanupCmd = &cobra.Command{
	Use:   "cleanup",
	Short: "Remove old cache entries or project data",
}

var cleanupCacheCmd = &cobra.Command{
	Use:   "cache",
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

var cleanupProjectCmd = &cobra.Command{
	Use:   "project <dir>",
	Short: "Remove a project's database and data directory",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectDir := args[0]
		cfg := config.NewConfig()
		pd := cfg.ProjectDir(projectDir)

		if _, err := os.Stat(pd); os.IsNotExist(err) {
			fmt.Printf("No data found for project %s\n", projectDir)
			return
		}

		if err := os.RemoveAll(pd); err != nil {
			fmt.Println("Error removing project data:", err)
			os.Exit(1)
		}
		fmt.Printf("Removed project data for %s\n", projectDir)
	},
}

func init() {
	rootCmd.AddCommand(cleanupCmd)
	cleanupCmd.AddCommand(cleanupCacheCmd)
	cleanupCmd.AddCommand(cleanupProjectCmd)
	cleanupCacheCmd.Flags().IntP("days", "d", 30, "Delete entries older than N days")
}
