package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cfgFile string
var dataDir string

var rootCmd = &cobra.Command{
	Use:   "recall",
	Short: "recall — memory, context, and RAG for your projects",
	Long: `recall is a local and global context storage system.
It caches project abstractions, indexes files, provides vector search,
and generates task briefs to save tokens and avoid unnecessary model roundtrips.

Local cache = "what is true in this repo?"
Global cache = "what patterns keep being true across repos?"`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	defaultDir := filepath.Join(home, ".recall")
	if _, err := os.Stat(defaultDir); os.IsNotExist(err) {
		os.MkdirAll(defaultDir, 0755)
	}
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", filepath.Join(defaultDir, ".recall.yaml"), "config file")
	rootCmd.PersistentFlags().StringVar(&dataDir, "datadir", defaultDir, "data directory")
}
