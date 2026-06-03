package e2e

import (
	"strings"
	"testing"
)

func TestRecipesSeedStrictSucceeds(t *testing.T) {
	e := newEnv(t)

	stdout, stderr, code := e.run("recipes", "seed", "--strict")
	if code != 0 {
		t.Fatalf("recipes seed --strict exited %d: stderr=%s", code, stderr)
	}
	if !strings.Contains(stdout, "Seeded") {
		t.Fatalf("expected 'Seeded' in stdout, got: %s", stdout)
	}
}
