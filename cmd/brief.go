package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/gleicon/technocore/internal/cache"
	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
	"github.com/gleicon/technocore/internal/recipes"
	"github.com/spf13/cobra"
)

var briefCmd = &cobra.Command{
	Use:   "brief [task description]",
	Short: "Generate a task brief using global recipes and local project facts",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		m, err := cache.OpenManager(cfg, ".")
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		defer m.Close()

		task := strings.Join(args, " ")
		pm, err := m.GetMap()
		if err != nil || pm == nil {
			fmt.Println("No project map found. Run 'technocore map' first.")
			return
		}

		// Fetch matching recipes via vector RAG
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

		var projectSignals []string
		for _, s := range pm.Signals {
			projectSignals = append(projectSignals, s)
		}
		matches, err := recipes.FindMatches(gdb, task, pm.Framework, projectSignals, 3)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Warning: recipe search failed:", err)
		}

		var b strings.Builder
		fmt.Fprintf(&b, "# Brief: %s\n\n", task)
		fmt.Fprintf(&b, "## Project Context\n")
		fmt.Fprintf(&b, "- Language: %s\n", pm.Language)
		fmt.Fprintf(&b, "- Framework: %s\n", pm.Framework)
		fmt.Fprintf(&b, "- Package Manager: %s\n", pm.PackageManager)
		fmt.Fprintf(&b, "- Signals: %v\n", pm.Signals)
		fmt.Fprintf(&b, "- Entrypoints: %v\n", pm.Entrypoints)

		// Subsystems
		rows, err := m.ProjectDB.Query(`SELECT name, summary FROM subsystem_summaries`)
		if err == nil {
			defer rows.Close()
			fmt.Fprintf(&b, "\n## Subsystems\n")
			for rows.Next() {
				var name, summary string
				if err := rows.Scan(&name, &summary); err != nil {
					continue
				}
				fmt.Fprintf(&b, "- **%s**: %s\n", name, summary)
			}
		}

		// Recipes
		if len(matches) > 0 {
			fmt.Fprintf(&b, "\n## Recipes\n")
			for _, m := range matches {
				fmt.Fprintf(&b, "- **%s** (score: %.2f)\n", m.Recipe.Name, m.Score)
				fmt.Fprintf(&b, "  %s\n", indent(m.Recipe.BriefTemplate, 2))
			}
		} else {
			fmt.Fprintln(os.Stderr, "No relevant recipes found. Run 'technocore recipes seed' to load defaults.")
		}

		// Relevant files
		like := "%" + task + "%"
		frows, err := m.ProjectDB.Query(`SELECT path, summary FROM files WHERE summary LIKE ? OR path LIKE ? LIMIT 10`, like, like)
		if err == nil {
			defer frows.Close()
			fmt.Fprintf(&b, "\n## Relevant Files\n")
			for frows.Next() {
				var p, s string
				if err := frows.Scan(&p, &s); err != nil {
					continue
				}
				fmt.Fprintf(&b, "- `%s`: %s\n", p, s)
			}
		}

		fmt.Fprintf(&b, "\n## Task\n%s\n", task)
		fmt.Println(b.String())
	},
}

func init() {
	rootCmd.AddCommand(briefCmd)
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
