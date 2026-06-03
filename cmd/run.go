package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Record and manage model run outcomes",
}

var runSuggestCmd = &cobra.Command{
	Use:   "suggest",
	Short: "Suggest recording a run and ask for confirmation",
	Run: func(cmd *cobra.Command, args []string) {
		task, _ := cmd.Flags().GetString("task")
		filesChanged, _ := cmd.Flags().GetString("files-changed")
		tokensIn, _ := cmd.Flags().GetInt("tokens-in")
		tokensOut, _ := cmd.Flags().GetInt("tokens-out")
		testsPassed, _ := cmd.Flags().GetInt("tests-passed")
		followUp, _ := cmd.Flags().GetInt("follow-up-needed")

		if task == "" {
			fmt.Println("Error: --task required")
			os.Exit(1)
		}

		reduction := 0
		if tokensIn > 0 {
			reduction = 100 - (tokensOut*100)/tokensIn
		}

		fmt.Printf("Record task '%s' (%s, %d%% token reduction)? [y/n/i] ", task, filesChanged, reduction)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if len(input) == 0 {
			input = "n"
		}

		switch input[0] {
		case 'y':
			saveRun(task, filesChanged, tokensIn, tokensOut, testsPassed, followUp, "")
		case 'i':
			fmt.Print("Insight: ")
			insight, _ := reader.ReadString('\n')
			insight = strings.TrimSpace(insight)
			saveRun(task, filesChanged, tokensIn, tokensOut, testsPassed, followUp, insight)
		default:
			fmt.Println("Skipped.")
		}
	},
}

var runRecordCmd = &cobra.Command{
	Use:   "record",
	Short: "Manually record a run without interactive gate",
	Run: func(cmd *cobra.Command, args []string) {
		task, _ := cmd.Flags().GetString("task")
		filesChanged, _ := cmd.Flags().GetString("files-changed")
		tokensIn, _ := cmd.Flags().GetInt("tokens-in")
		tokensOut, _ := cmd.Flags().GetInt("tokens-out")
		testsPassed, _ := cmd.Flags().GetInt("tests-passed")
		followUp, _ := cmd.Flags().GetInt("follow-up-needed")
		insight, _ := cmd.Flags().GetString("insight")

		if task == "" {
			fmt.Println("Error: --task required")
			os.Exit(1)
		}
		saveRun(task, filesChanged, tokensIn, tokensOut, testsPassed, followUp, insight)
	},
}

func saveRun(task, filesChanged string, tokensIn, tokensOut, testsPassed, followUp int, insight string) {
	cfg := config.NewConfig()
	m, err := cache.OpenManager(cfg, ".")
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer m.Close()

	pm, _ := m.GetMap()
	framework := ""
	if pm != nil {
		framework = pm.Framework
	}

	var fc []string
	if filesChanged != "" {
		fc = strings.Split(filesChanged, ",")
	}

	if err := m.StoreRun(task, framework, nil, fc, testsPassed, followUp, tokensIn, tokensOut, 1); err != nil {
		fmt.Println("Error saving run:", err)
		return
	}

	if insight != "" {
		if err := m.StoreMemory("insight", insight, task); err != nil {
			fmt.Println("Error saving insight:", err)
			return
		}
	}

	fmt.Println("Run recorded.")
}

func init() {
	rootCmd.AddCommand(runCmd)
	runCmd.AddCommand(runSuggestCmd)
	runCmd.AddCommand(runRecordCmd)

	for _, c := range []*cobra.Command{runSuggestCmd, runRecordCmd} {
		c.Flags().StringP("task", "t", "", "Task description")
		c.Flags().StringP("files-changed", "f", "", "Comma-separated list of changed files")
		c.Flags().IntP("tokens-in", "i", 0, "Input tokens")
		c.Flags().IntP("tokens-out", "o", 0, "Output tokens")
		c.Flags().Int("tests-passed", 1, "Tests passed (1=yes, 0=no)")
		c.Flags().Int("follow-up-needed", 0, "Follow-up needed (1=yes, 0=no)")
	}
	runRecordCmd.Flags().String("insight", "", "Optional insight text")
}
