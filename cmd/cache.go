package cmd

import (
	"fmt"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage the project abstraction cache",
}

var cacheBuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build or rebuild the project cache",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		sentences, _ := cmd.Flags().GetInt("sentences")
		if err := m.BuildCache(sentences); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Cache built.")
	},
}

var cacheInspectCmd = &cobra.Command{
	Use:   "inspect",
	Short: "Inspect cached items",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		out, err := m.Inspect()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Print(out)
	},
}

var cacheRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "Refresh cache for changed files only",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		sentences, _ := cmd.Flags().GetInt("sentences")
		if err := m.Refresh(sentences); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Cache refreshed.")
	},
}

var cacheInvalidateCmd = &cobra.Command{
	Use:   "invalidate",
	Short: "Invalidate stale cache entries",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		if err := m.Invalidate(); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Stale entries invalidated.")
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
	cacheCmd.AddCommand(cacheBuildCmd)
	cacheCmd.AddCommand(cacheInspectCmd)
	cacheCmd.AddCommand(cacheRefreshCmd)
	cacheCmd.AddCommand(cacheInvalidateCmd)

	cacheBuildCmd.Flags().IntP("sentences", "s", 3, "Sentences per summary")
	cacheRefreshCmd.Flags().IntP("sentences", "s", 3, "Sentences per summary")
}
