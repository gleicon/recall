package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gleicon/recall/internal/cache"
	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/db"
	"github.com/gleicon/recall/internal/embeddings"
	"github.com/gleicon/recall/internal/recipes"
	"github.com/gleicon/recall/internal/search"
	"github.com/gleicon/recall/internal/summarizer"
	"github.com/spf13/cobra"
)

var briefCmd = &cobra.Command{
	Use:   "brief [task description]",
	Short: "Generate a task brief using global recipes and local project facts",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		settings, _ := cfg.LoadSettings()
		embedModel := ""
		if settings != nil {
			embedModel = settings.EmbedModel
		}

		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()
		m.EmbedModel = embedModel

		task := strings.Join(args, " ")
		pm, err := m.GetMap()
		if err != nil || pm == nil {
			fmt.Println("No project map found. Run 'recall map' first.")
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

		matches, err := recipes.FindMatches(gdb, task, pm.Framework, pm.Signals, 3)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: recipe search failed:", err)
		}
		for _, match := range matches {
			recipes.IncrementUseCount(gdb, match.Recipe.Name, 1.0)
		}

		var b strings.Builder
		var tokRecipes, tokSubsystems, tokFiles, tokPriorArt int

		fmt.Fprintf(&b, "# Brief: %s\n\n", task)
		fmt.Fprintf(&b, "## Project Context\n")
		fmt.Fprintf(&b, "- Language: %s\n", pm.Language)
		fmt.Fprintf(&b, "- Framework: %s\n", pm.Framework)
		fmt.Fprintf(&b, "- Package Manager: %s\n", pm.PackageManager)

		rows, err := m.ProjectDB.Query(`SELECT name, summary FROM subsystem_summaries`)
		if err == nil {
			fmt.Fprintf(&b, "\n## Subsystems\n")
			for rows.Next() {
				var name, summary string
				if err := rows.Scan(&name, &summary); err != nil {
					continue
				}
				if len(summary) > 300 {
					if short, err := summarizer.Summarize(summary, 2); err == nil {
						summary = short
					}
				}
				tokSubsystems += approxTokens(summary)
				fmt.Fprintf(&b, "- **%s**: %s\n", name, summary)
			}
			rows.Close()
		}

		if len(matches) > 0 {
			fmt.Fprintf(&b, "\n## Recipes\n")
			for _, match := range matches {
				tmpl := match.Recipe.BriefTemplate
				if len(tmpl) > 300 {
					if short, err := summarizer.Summarize(tmpl, 2); err == nil {
						tmpl = short
					}
				}
				tokRecipes += approxTokens(tmpl)
				fmt.Fprintf(&b, "- **%s** (score: %.2f)\n  %s\n", match.Recipe.Name, match.Score, indent(tmpl, 2))
			}
		} else {
			fmt.Fprintln(os.Stderr, "No relevant recipes found. Run 'recall recipes seed' to load defaults.")
		}

		eng := search.NewEngine(m.ProjectDB, embedModel)
		fileResults, err := eng.Query(task, 5)
		if err == nil && len(fileResults) > 0 {
			fmt.Fprintf(&b, "\n## Relevant Files\n")
			for _, r := range fileResults {
				summary := r.Summary
				if len(summary) > 200 {
					if short, err := summarizer.Summarize(summary, 1); err == nil {
						summary = short
					}
				}
				tokFiles += approxTokens(summary)
				fmt.Fprintf(&b, "- `%s`: %s\n", r.Path, summary)
			}
		}

		qVec := embeddings.ComputeSmartWithClient(task, embedModel, eng.LLMClient)
		if len(qVec) > 0 {
			convs, err := cache.FindSimilarConversations(gdb, qVec, 2)
			if err == nil && len(convs) > 0 {
				fmt.Fprintf(&b, "\n## Prior Art\n")
				for _, c := range convs {
					resp := c.Response
					if len(resp) > 300 {
						if short, err := summarizer.Summarize(resp, 2); err == nil {
							resp = short
						}
					}
					tokPriorArt += approxTokens(resp)
					fmt.Fprintf(&b, "- **%s** [%s]\n  %s\n", c.Task, c.ModelName, indent(resp, 2))
				}
			}
		}

		fmt.Fprintf(&b, "\n## Task\n%s\n", task)
		fmt.Println(b.String())

		total := tokRecipes + tokSubsystems + tokFiles + tokPriorArt
		fmt.Fprintf(os.Stderr, "Tokens recalled: ~%d (recipes: %d, subsystems: %d, files: %d, prior art: %d)\n",
			total, tokRecipes, tokSubsystems, tokFiles, tokPriorArt)
	},
}

func init() {
	rootCmd.AddCommand(briefCmd)
}

// approxTokens estimates BPE token count from word count (words * 1.3).
func approxTokens(s string) int {
	return int(float64(len(strings.Fields(s))) * 1.3)
}

func indent(s string, n int) string {
	prefix := strings.Repeat(" ", n)
	lines := strings.Split(s, "\n")
	for i, l := range lines {
		if l != "" {
			lines[i] = prefix + l
		}
	}
	return strings.Join(lines, "\n")
}
