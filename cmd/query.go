package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
	"github.com/gleicon/technocore/internal/llm"
	"github.com/gleicon/technocore/internal/recipes"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query [prompt]",
	Short: "Query local model or build enriched prompt for big model",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		prompt := strings.Join(args, " ")
		forceLocal, _ := cmd.Flags().GetBool("local-only")
		forceDelegate, _ := cmd.Flags().GetBool("delegate")

		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()

		pm, err := m.GetMap()
		if err != nil || pm == nil {
			fmt.Println("No project map found. Run 'technocore map' first.")
			return
		}

		gdb, err := db.Open(cfg.GlobalDBPath)
		if err != nil {
			fmt.Println("Error opening global db:", err)
			return
		}
		defer gdb.Close()
		if err := db.InitGlobalSchema(gdb); err != nil {
			fmt.Println("Error init global schema:", err)
			return
		}

		matches, err := recipes.FindMatches(gdb, prompt, pm.Framework, pm.Signals, 3)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: recipe search failed:", err)
		}
		for _, match := range matches {
			recipes.IncrementUseCount(gdb, match.Recipe.Name, 1.0)
		}

		var b strings.Builder
		fmt.Fprintf(&b, "# Context for: %s\n\n", prompt)
		fmt.Fprintf(&b, "## Project\n- Language: %s\n- Framework: %s\n", pm.Language, pm.Framework)
		if len(matches) > 0 {
			fmt.Fprintf(&b, "\n## Recipes\n")
			for _, m := range matches {
				fmt.Fprintf(&b, "- **%s**: %s\n", m.Recipe.Name, m.Recipe.BriefTemplate)
			}
		}
		brief := b.String()

		client := llm.Detect()
		if forceDelegate || client == nil {
			modelName := ""
			if client != nil && len(client.Models) > 0 {
				modelName = client.Models[0]
			}
			conv := &cache.Conversation{
				Task:             prompt,
				Prompt:           prompt,
				Response:         "DELEGATE: forced or no local model",
				ModelName:        modelName,
				InputTokens:      0,
				OutputTokens:     0,
				Delegated:        true,
				DelegationReason: "forced or no local model",
				ProjectHash:      config.ProjectHash("."),
				Framework:        pm.Framework,
			}
			if saveErr := cache.StoreConversation(gdb, conv); saveErr != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to save conversation:", saveErr)
			}
			lessonPattern := fmt.Sprintf("delegate to big model for %s", pm.Framework)
			if saveErr := cache.StoreLesson(gdb, lessonPattern, pm.Framework, modelName, prompt, 0.5); saveErr != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to save lesson:", saveErr)
			}
			fmt.Println(brief)
			fmt.Println("\n---")
			fmt.Println("DELEGATE: no local model")
			return
		}

		settings, err := cfg.LoadSettings()
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error loading settings:", err)
			return
		}
		client.Selected = client.PreferredModel(settings.LocalModel)
		timeoutFlag, _ := cmd.Flags().GetInt("timeout")
		if timeoutFlag > 0 {
			client.SetTimeout(time.Duration(timeoutFlag) * time.Second)
		} else if settings.QueryTimeout > 0 {
			client.SetTimeout(time.Duration(settings.QueryTimeout) * time.Second)
		}

		stream, _ := cmd.Flags().GetBool("stream")

		if forceLocal {
			fmt.Println("Local model:", client.Type)
			fmt.Println("Models:", client.Models)
			fmt.Println("Selected:", client.Selected)
			fmt.Println("Timeout:", client.QueryTimeout)
			return
		}

		systemPrompt := fmt.Sprintf(
			"You are a helpful coding assistant. Use the following context to answer the user's question.\n\n%s\n\nIf you can answer fully from the context, provide the answer. If you need more reasoning or code generation beyond the context, respond with DELEGATE and a brief summary of what's needed.",
			brief,
		)

		var resp string
		var queryErr error
		if stream {
			fmt.Print("# Answer (local model, streaming)\n\n")
			var sb strings.Builder
			queryErr = client.QueryStream(prompt, systemPrompt, func(token string) {
				fmt.Print(token)
				sb.WriteString(token)
			})
			fmt.Println()
			fmt.Println()
			resp = sb.String()
		} else {
			resp, queryErr = client.Query(prompt, systemPrompt)
		}
		if queryErr != nil {
			fmt.Fprintln(os.Stderr, "Local model error:", queryErr)
			fmt.Println(brief)
			fmt.Println("\n---")
			fmt.Println("DELEGATE: local model failed")
			return
		}

		upperResp := strings.ToUpper(resp)
		delegated := strings.Contains(upperResp, "DELEGATE")
		delegationReason := ""
		if delegated {
			idx := strings.Index(upperResp, "DELEGATE")
			if idx != -1 {
				delegationReason = strings.TrimSpace(resp[idx+len("DELEGATE"):])
			}
			if delegationReason == "" {
				delegationReason = "needs larger model"
			}
		}
		conv := &cache.Conversation{
			Task:             prompt,
			Prompt:           prompt,
			Response:         resp,
			ModelName:        client.Models[0],
			InputTokens:      0,
			OutputTokens:     0,
			Delegated:        delegated,
			DelegationReason: delegationReason,
			ProjectHash:      config.ProjectHash("."),
			Framework:        pm.Framework,
		}
		if saveErr := cache.StoreConversation(gdb, conv); saveErr != nil {
			fmt.Fprintln(os.Stderr, "Warning: failed to save conversation:", saveErr)
		}

		lessonPattern := fmt.Sprintf("local model %s for %s tasks", client.Models[0], pm.Framework)
		if delegated {
			lessonPattern = fmt.Sprintf("delegate %s to big model for %s", client.Models[0], pm.Framework)
		}
		if saveErr := cache.StoreLesson(gdb, lessonPattern, pm.Framework, client.Models[0], prompt, 0.5); saveErr != nil {
			fmt.Fprintln(os.Stderr, "Warning: failed to save lesson:", saveErr)
		}

		if delegated {
			fmt.Println(brief)
			fmt.Println("\n---")
			fmt.Println("DELEGATE: needs larger model")
			if delegationReason != "" {
				fmt.Println("Reason:", delegationReason)
			}
			return
		}

		snippets := cache.ExtractSnippets(resp, pm.Language, pm.Framework)
		for _, s := range snippets {
			if saveErr := cache.StoreSnippet(gdb, &s); saveErr != nil {
				fmt.Fprintln(os.Stderr, "Warning: failed to save snippet:", saveErr)
			}
		}
		if len(snippets) > 0 {
			fmt.Fprintf(os.Stderr, "Extracted %d snippets\n", len(snippets))
		}

		if !stream {
			fmt.Println("# Answer (local model)")
			fmt.Println()
			fmt.Println(resp)
			fmt.Println()
		}
		fmt.Println("---")
		fmt.Printf("Answered by %s (%s)\n", client.Type, client.Models[0])
	},
}

func init() {
	rootCmd.AddCommand(queryCmd)
	queryCmd.Flags().Bool("local-only", false, "Only show local model info, do not delegate")
	queryCmd.Flags().Bool("delegate", false, "Force delegation to big model")
	queryCmd.Flags().Int("timeout", 0, "Query timeout in seconds (0 = use default 30s or config)")
	queryCmd.Flags().Bool("stream", false, "Stream response tokens as they arrive")
}
