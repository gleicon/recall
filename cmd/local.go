package cmd

import (
	"fmt"
	"os"

	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/llm"
	"github.com/spf13/cobra"
)

var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Configure local LLM",
}

var localStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show local LLM status",
	Run: func(cmd *cobra.Command, args []string) {
		client := llm.Detect()
		if client == nil {
			fmt.Println("No local LLM detected.")
			return
		}

		cfg := config.NewConfig()
		settings, err := cfg.LoadSettings()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading settings:", err)
			return
		}
		preferred := client.PreferredModel(settings.LocalModel)

		fmt.Printf("%s (%s)\n", client.Type, client.Endpoint)
		for _, m := range client.Models {
			marker := ""
			if m == preferred {
				marker = " *"
			}
			fmt.Printf("  %s%s\n", m, marker)
		}
		if settings.LocalModel != "" {
			fmt.Printf("Preference: %s\n", settings.LocalModel)
		}
	},
}

var localModelsCmd = &cobra.Command{
	Use:   "models",
	Short: "List local models",
	Run: func(cmd *cobra.Command, args []string) {
		client := llm.Detect()
		if client == nil {
			fmt.Println("No local LLM detected.")
			return
		}
		cfg := config.NewConfig()
		settings, err := cfg.LoadSettings()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading settings:", err)
			return
		}
		preferred := client.PreferredModel(settings.LocalModel)
		for _, m := range client.Models {
			marker := ""
			if m == preferred {
				marker = " *"
			}
			fmt.Printf("  %s%s\n", m, marker)
		}
	},
}

var localUseCmd = &cobra.Command{
	Use:   "use <model-name>",
	Short: "Set preferred local model",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		desired := args[0]
		client := llm.Detect()
		if client == nil {
			fmt.Println("No local LLM detected. Start your local server first.")
			return
		}

		found := false
		for _, m := range client.Models {
			if m == desired {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Model '%s' not found.\n", desired)
			os.Exit(1)
		}

		cfg := config.NewConfig()
		settings, err := cfg.LoadSettings()
		if err != nil {
			fmt.Println("Error loading settings:", err)
			return
		}
		settings.LocalModel = desired
		if err := cfg.SaveSettings(settings); err != nil {
			fmt.Println("Error saving settings:", err)
			return
		}
		fmt.Printf("Local model set to: %s\n", desired)
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
	localCmd.AddCommand(localStatusCmd)
	localCmd.AddCommand(localModelsCmd)
	localCmd.AddCommand(localUseCmd)
}
