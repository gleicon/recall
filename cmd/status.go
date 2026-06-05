package cmd

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/db"
	"github.com/spf13/cobra"
)

type statusOutput struct {
	Version      string `json:"version"`
	Mapped       bool   `json:"mapped"`
	MappedAt     string `json:"mapped_at,omitempty"`
	Language     string `json:"language,omitempty"`
	Framework    string `json:"framework,omitempty"`
	FilesIndexed int    `json:"files_indexed"`
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show recall status for the current project (JSON)",
	Run: func(cmd *cobra.Command, args []string) {
		out := statusOutput{Version: Version}

		cfg := config.NewConfig()
		pdb, err := db.Open(cfg.ProjectDBPath("."))
		if err != nil {
			printStatus(out)
			return
		}
		defer pdb.Close()
		if err := db.InitProjectSchema(pdb); err != nil {
			printStatus(out)
			return
		}

		var lang, fw, mappedAt sql.NullString
		row := pdb.QueryRow(`SELECT language, framework, updated_at FROM project_map WHERE id=1`)
		if err := row.Scan(&lang, &fw, &mappedAt); err == nil {
			out.Mapped = true
			out.MappedAt = mappedAt.String
			out.Language = lang.String
			out.Framework = fw.String
		}

		pdb.QueryRow(`SELECT COUNT(*) FROM files`).Scan(&out.FilesIndexed)

		printStatus(out)
	},
}

func printStatus(out statusOutput) {
	b, _ := json.Marshal(out)
	fmt.Println(string(b))
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
