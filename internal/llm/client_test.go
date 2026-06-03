package llm

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDetectOpenAI(t *testing.T) {
	// Mock OpenAI-compatible server (like llama.app / llama.cpp)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/models":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]string{
					{"id": "llama-3.2-1b"},
					{"id": "qwen3.6-27b"},
				},
			})
		case "/v1/chat/completions":
			var req struct {
				Model    string              `json:"model"`
				Messages []map[string]string `json:"messages"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatal(err)
			}
			if req.Model == "" {
				t.Fatal("expected model name")
			}
			if len(req.Messages) == 0 {
				t.Fatal("expected messages")
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"choices": []map[string]interface{}{
					{
						"message": map[string]string{
							"content": "Use middleware for auth.",
							"role":    "assistant",
						},
					},
				},
			})
		case "/v1/embeddings":
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"embedding": []float64{0.1, 0.2, 0.3, 0.4}},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	// Test detection by temporarily overriding endpoints would need DI,
	// but we can test the query/embedding methods directly.
	c := &Client{
		Endpoint: server.URL,
		Type:     "llama-app",
		Models:   []string{"llama-3.2-1b", "qwen3.6-27b"},
	}

	resp, err := c.Query("How do I add auth?", "You are a helpful coding assistant.")
	if err != nil {
		t.Fatal(err)
	}
	if resp != "Use middleware for auth." {
		t.Fatalf("unexpected response: %s", resp)
	}
}

func TestClientGetEmbedding(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"embedding": []float64{0.1, 0.2, 0.3, 0.4, 0.5, 0.6}},
			},
		})
	}))
	defer server.Close()

	c := &Client{
		Endpoint: server.URL,
		Type:     "llama-app",
		Models:   []string{"nomic-embed-text"},
	}

	emb, err := c.GetEmbedding("test text")
	if err != nil {
		t.Fatal(err)
	}
	if len(emb) != 6 {
		t.Fatalf("expected 6 dims, got %d", len(emb))
	}
	if emb[0] != 0.1 {
		t.Fatalf("expected 0.1, got %f", emb[0])
	}
}

func TestClientQueryNoSystemContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Messages []map[string]string `json:"messages"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatal(err)
		}
		// Without system context, should only have user message
		if len(req.Messages) != 1 {
			t.Fatalf("expected 1 message (user only), got %d", len(req.Messages))
		}
		if req.Messages[0]["role"] != "user" {
			t.Fatalf("expected user role, got %s", req.Messages[0]["role"])
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"choices": []map[string]interface{}{
				{"message": map[string]string{"content": "Hello!", "role": "assistant"}},
			},
		})
	}))
	defer server.Close()

	c := &Client{
		Endpoint: server.URL,
		Type:     "llama-app",
		Models:   []string{"default"},
	}

	resp, err := c.Query("Hello?", "")
	if err != nil {
		t.Fatal(err)
	}
	if resp != "Hello!" {
		t.Fatalf("unexpected response: %s", resp)
	}
}

func TestGetEmbeddingError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("model not loaded"))
	}))
	defer server.Close()

	c := &Client{
		Endpoint: server.URL,
		Type:     "llama-app",
		Models:   []string{"default"},
	}
	_, err := c.GetEmbedding("test")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
