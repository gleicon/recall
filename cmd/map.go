package cmd

import (
	"fmt"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/spf13/cobra"
)

var mapCmd = &cobra.Command{
	Use:   "map",
	Short: "Detect and display the project map for the current directory",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		pm, err := m.BuildMap()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		fmt.Println(pm.String())
	},
}

func init() {
	rootCmd.AddCommand(mapCmd)
}
