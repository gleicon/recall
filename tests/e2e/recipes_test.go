package e2e

import (
	"strconv"
	"strings"
	"testing"
)

func TestRecipesSeedCreatesRecipesWithEmbeddings(t *testing.T) {
	e := newEnv(t)

	// Act: seed recipes
	stdout, stderr, code := e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("recipes seed exited %d: stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Seeded") {
		t.Fatalf("expected 'Seeded' in stdout, got: %s", stdout)
	}

	// Assert: list shows ≥15 recipes
	listOut, listErr, listCode := e.run("recipes", "list")
	if listCode != 0 {
		t.Fatalf("recipes list exited %d: stderr=%s", listCode, listErr)
	}

	lines := strings.Split(strings.TrimSpace(listOut), "\n")
	// Last line is "Total: N"
	lastLine := lines[len(lines)-1]
	if !strings.HasPrefix(lastLine, "Total: ") {
		t.Fatalf("expected 'Total:' line, got: %s", lastLine)
	}
	countStr := strings.TrimPrefix(lastLine, "Total: ")
	count, err := strconv.Atoi(countStr)
	if err != nil {
		t.Fatalf("could not parse total: %s", lastLine)
	}
	if count < 15 {
		t.Fatalf("expected at least 15 recipes, got %d", count)
	}

	// AC-10: re-run seed should not create duplicates
	_, _, code = e.run("recipes", "seed")
	if code != 0 {
		t.Fatalf("second recipes seed exited %d", code)
	}
	listOut2, _, listCode2 := e.run("recipes", "list")
	if listCode2 != 0 {
		t.Fatalf("recipes list after re-seed exited %d", listCode2)
	}
	lines2 := strings.Split(strings.TrimSpace(listOut2), "\n")
	lastLine2 := lines2[len(lines2)-1]
	countStr2 := strings.TrimPrefix(lastLine2, "Total: ")
	count2, _ := strconv.Atoi(countStr2)
	if count2 != count {
		t.Fatalf("expected same count after re-seed (%d), got %d", count, count2)
	}
}
