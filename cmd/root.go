package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/llm"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags: -X github.com/gleicon/recall/cmd.Version=v1.2.3
var Version = "dev"

var cfgFile string
var dataDir string
var endpointFlag string

var rootCmd = &cobra.Command{
	Use:   "recall",
	Short: "recall — memory, context, and RAG for your projects",
	Long: `recall is a local and global context storage system.
It caches project abstractions, indexes files, provides vector search,
and generates task briefs to save tokens and avoid unnecessary model roundtrips.

Local cache = "what is true in this repo?"
Global cache = "what patterns keep being true across repos?"`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		if settings, err := cfg.LoadSettings(); err == nil {
			if settings.LocalEndpoint != "" {
				llm.PreferredEndpoint = settings.LocalEndpoint
			}
			if settings.DetectTimeout > 0 {
				llm.DetectTimeout = time.Duration(settings.DetectTimeout) * time.Second
			}
		}
		if endpointFlag != "" {
			llm.PreferredEndpoint = endpointFlag
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	if Version == "dev" {
		if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
			Version = info.Main.Version
		}
	}

	home, err := os.UserHomeDir()
	if err != nil {
		home = "."
	}
	defaultDir := filepath.Join(home, ".recall")
	if _, err := os.Stat(defaultDir); os.IsNotExist(err) {
		os.MkdirAll(defaultDir, 0755)
	}
	rootCmd.Version = Version
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", filepath.Join(defaultDir, ".recall.yaml"), "config file")
	rootCmd.PersistentFlags().StringVar(&dataDir, "datadir", defaultDir, "data directory")
	rootCmd.PersistentFlags().StringVar(&endpointFlag, "endpoint", "", "local LLM endpoint (e.g. http://localhost:1234)")
}
