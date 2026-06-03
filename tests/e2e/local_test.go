package e2e

import (
	"strings"
	"testing"
)

func TestLocalStatusCommandWorks(t *testing.T) {
	e := newEnv(t)

	stdout, _, code := e.run("local", "status")
	if code != 0 {
		t.Fatalf("local status exited %d", code)
	}
	hasLLM := strings.Contains(stdout, "http://localhost:")
	hasNoLLM := strings.Contains(stdout, "No local LLM detected")
	if !hasLLM && !hasNoLLM {
		t.Fatalf("expected LLM status output, got: %s", stdout)
	}
}
