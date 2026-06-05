package cmd

import (
	"fmt"

	"github.com/gleicon/recall/internal/cache"
	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/db"
	"github.com/spf13/cobra"
)

var feedbackCmd = &cobra.Command{
	Use:   "feedback",
	Short: "Rate the last query answer (--good or --bad)",
	Run: func(cmd *cobra.Command, args []string) {
		good, _ := cmd.Flags().GetBool("good")
		bad, _ := cmd.Flags().GetBool("bad")
		note, _ := cmd.Flags().GetString("note")

		if !good && !bad {
			fmt.Println("Use --good or --bad")
			return
		}
		if good && bad {
			fmt.Println("Use either --good or --bad, not both")
			return
		}

		cfg := config.NewConfig()
		gdb, err := db.Open(cfg.GlobalDBPath)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer gdb.Close()
		if err := db.InitGlobalSchema(gdb); err != nil {
			fmt.Println("Error:", err)
			return
		}

		id, err := cache.GetLastConversationID(gdb)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		if id == 0 {
			fmt.Println("No conversations recorded yet.")
			return
		}

		if err := cache.MarkConversationFeedback(gdb, id, good, note); err != nil {
			fmt.Println("Error storing feedback:", err)
			return
		}

		label := "bad"
		if good {
			label = "good"
		}
		fmt.Printf("Feedback recorded: conversation #%d → %s\n", id, label)
		if note != "" {
			fmt.Printf("Note: %s\n", note)
		}
	},
}

func init() {
	rootCmd.AddCommand(feedbackCmd)
	feedbackCmd.Flags().Bool("good", false, "Mark last answer as accepted")
	feedbackCmd.Flags().Bool("bad", false, "Mark last answer as rejected")
	feedbackCmd.Flags().String("note", "", "Optional note explaining the rating")
}
