package cmd

import (
	"fmt"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/spf13/cobra"
)

var indexCmd = &cobra.Command{
	Use:   "index [directory]",
	Short: "Index a directory into the project cache",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, args[0])
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		sentences, _ := cmd.Flags().GetInt("sentences")
		fmt.Printf("Indexing %s ...\n", args[0])
		if err := m.BuildCache(sentences); err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println("Done.")
	},
}

func init() {
	rootCmd.AddCommand(indexCmd)
	indexCmd.Flags().IntP("sentences", "s", 3, "Sentences per summary")
}
