package cache

import (
	"strings"
	"testing"
)

func TestExtractSnippets(t *testing.T) {
	text := `
Here is how to add health check:

` + "```go\nfunc HealthCheck(w http.ResponseWriter, r *http.Request) {\n    w.WriteHeader(http.StatusOK)\n}\n```" + `

And another snippet:

` + "```go\nfunc main() {\n    http.HandleFunc(\"/health\", HealthCheck)\n}\n```"

	snippets := ExtractSnippets(text, "go", "go")
	if len(snippets) != 2 {
		t.Fatalf("expected 2 snippets, got %d", len(snippets))
	}
	if !strings.Contains(snippets[0].Code, "HealthCheck") {
		t.Fatalf("expected HealthCheck in first snippet, got: %s", snippets[0].Code)
	}
	if snippets[0].Language != "go" {
		t.Fatalf("expected language go, got %s", snippets[0].Language)
	}
}
