package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gleicon/recall/internal/config"
	"github.com/gleicon/recall/internal/db"
	"github.com/gleicon/recall/internal/recipes"
	"github.com/spf13/cobra"
)

var recipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "Manage global task recipes",
}

var recipesSeedCmd = &cobra.Command{
	Use:   "seed",
	Short: "Load recipes from ~/.recall/recipes/ into global.db",
	Run: func(cmd *cobra.Command, args []string) {
		cfg := config.NewConfig()
		if err := cfg.EnsureDirs(); err != nil {
			fmt.Println("Error:", err)
			return
		}

		// Copy defaults on first run, sync new recipes on subsequent runs
		userRecipesDir := filepath.Join(cfg.HomeDir, "recipes")
		exe, _ := os.Executable()
		exeDir := filepath.Dir(exe)
		defaultDir := filepath.Join(exeDir, "..", "recipes")
		if _, err := os.Stat(defaultDir); os.IsNotExist(err) {
			// Fallback to CWD/recipes for dev builds
			defaultDir = "recipes"
		}
		if _, err := os.Stat(userRecipesDir); os.IsNotExist(err) {
			copyDir(defaultDir, userRecipesDir)
		} else {
			// Sync any new default recipes into user dir
			copyDir(defaultDir, userRecipesDir)
		}

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

		list, err := recipes.LoadAllFromDir(userRecipesDir)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		strict, _ := cmd.Flags().GetBool("strict")
		inserted := 0
		for _, r := range list {
			if err := recipes.Store(gdb, r); err != nil {
				if strict {
					fmt.Printf("Error: failed to store %s: %v\n", r.Name, err)
					os.Exit(1)
				}
				fmt.Printf("Warning: failed to store %s: %v\n", r.Name, err)
				continue
			}
			inserted++
		}
		fmt.Printf("Seeded %d recipes into global.db\n", inserted)
	},
}

var recipesAddCmd = &cobra.Command{
	Use:   "add --from-file <path>",
	Short: "Add a single recipe JSON file to global.db",
	Run: func(cmd *cobra.Command, args []string) {
		path, _ := cmd.Flags().GetString("from-file")
		if path == "" {
			fmt.Println("Error: --from-file required")
			os.Exit(1)
		}
		r, err := recipes.LoadFromFile(path)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
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
		if err := recipes.Store(gdb, r); err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		fmt.Println("Added recipe:", r.Name)
	},
}

var recipesListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all recipes in global.db",
	Run: func(cmd *cobra.Command, args []string) {
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
		list, err := recipes.List(gdb)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, r := range list {
			fmt.Printf("- %s (%s/%s) [%s]\n", r.Name, r.Language, r.Framework, r.Source)
		}
		fmt.Printf("Total: %d\n", len(list))
	},
}

func init() {
	rootCmd.AddCommand(recipesCmd)
	recipesCmd.AddCommand(recipesSeedCmd)
	recipesCmd.AddCommand(recipesAddCmd)
	recipesCmd.AddCommand(recipesListCmd)
	recipesAddCmd.Flags().StringP("from-file", "f", "", "Path to recipe JSON file")
	recipesSeedCmd.Flags().Bool("strict", false, "Fail on first invalid recipe")
}

func copyDir(src, dst string) error {
	if err := os.MkdirAll(dst, 0755); err != nil {
		return err
	}
	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(src, e.Name()))
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dst, e.Name()), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
