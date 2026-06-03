package cmd

import (
	"fmt"
	"time"

	"github.com/gleicon/technocore/internal/config"
	"github.com/gleicon/technocore/internal/db"
	"github.com/gleicon/technocore/internal/embeddings"
	"github.com/gleicon/technocore/internal/recipes"
	"github.com/gleicon/technocore/internal/search"
	"github.com/spf13/cobra"
)

var benchCmd = &cobra.Command{
	Use:   "bench",
	Short: "Run performance benchmarks",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("=== Pipecamp Benchmarks ===")
	fmt.Println()

		// Benchmark 1: Recipe vector search
		fmt.Println("--- Recipe Vector Search ---")
		benchRecipeSearch(10)
		benchRecipeSearch(100)
		benchRecipeSearch(500)

		// Benchmark 2: File FTS5 + vector search
		fmt.Println("\n--- File Search ---")
		benchFileSearch()

		// Benchmark 3: Embedding computation
		fmt.Println("\n--- Embedding Computation ---")
		benchEmbeddings()
	},
}

func benchRecipeSearch(n int) {
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

	// Count actual recipes
	var count int
	gdb.QueryRow(`SELECT COUNT(*) FROM task_recipes`).Scan(&count)
	if count == 0 {
		fmt.Printf("  (skip: no recipes in global.db)\n")
		return
	}

	start := time.Now()
	_, err = recipes.FindMatches(gdb, "add health check", "go", []string{}, n)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	elapsed := time.Since(start)
	fmt.Printf("  %d recipes: %v (top-%d)\n", count, elapsed, n)
}

func benchFileSearch() {
	cfg := config.NewConfig()
	gdb, err := db.Open(cfg.GlobalDBPath)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer gdb.Close()

	// Open current project DB
	pdb, err := db.Open(cfg.ProjectDBPath("."))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer pdb.Close()

	var fcount int
	pdb.QueryRow(`SELECT COUNT(*) FROM files`).Scan(&fcount)
	if fcount == 0 {
		fmt.Printf("  (skip: no indexed files)\n")
		return
	}

	engine := search.NewEngine(pdb)
	start := time.Now()
	_, err = engine.Query("cache build", 10)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	elapsed := time.Since(start)
	fmt.Printf("  %d files: %v (top-10)\n", fcount, elapsed)
}

func benchEmbeddings() {
	text := "This is a sample text for embedding benchmark. It contains multiple sentences to test the feature hashing implementation."

	iterations := 1000
	start := time.Now()
	for i := 0; i < iterations; i++ {
		embeddings.Compute(text)
	}
	elapsed := time.Since(start)
	perOp := elapsed / time.Duration(iterations)
	fmt.Printf("  %d embeddings: %v total, %v/op\n", iterations, elapsed, perOp)
}

func init() {
	rootCmd.AddCommand(benchCmd)
}
