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
		if err == nil {
			defer pdb.Close()
			if db.InitProjectSchema(pdb) == nil {
				var lang, fw, mappedAt sql.NullString
				if pdb.QueryRow(`SELECT language, framework, updated_at FROM project_map WHERE id=1`).Scan(&lang, &fw, &mappedAt) == nil {
					out.Mapped = true
					out.MappedAt = mappedAt.String
					out.Language = lang.String
					out.Framework = fw.String
				}
				pdb.QueryRow(`SELECT COUNT(*) FROM files`).Scan(&out.FilesIndexed)
			}
		}

		b, _ := json.Marshal(out)
		fmt.Println(string(b))
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
