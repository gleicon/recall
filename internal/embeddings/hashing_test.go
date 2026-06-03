package embeddings

import (
	"math"
	"testing"
)

func TestCompute(t *testing.T) {
	vec := Compute("hello world hello")
	if len(vec) != Dim {
		t.Fatalf("expected dim %d, got %d", Dim, len(vec))
	}
	// Check normalization
	var norm float64
	for _, v := range vec {
		norm += float64(v) * float64(v)
	}
	if math.Abs(norm-1.0) > 1e-4 {
		t.Fatalf("expected normalized vector, got norm %f", norm)
	}
}

func TestCosineIdentical(t *testing.T) {
	vec := Compute("test text")
	score := Cosine(vec, vec)
	if math.Abs(score-1.0) > 1e-4 {
		t.Fatalf("expected 1.0 for identical vectors, got %f", score)
	}
}

func TestCosineDifferent(t *testing.T) {
	vec1 := Compute("hello world")
	vec2 := Compute("foo bar")
	score := Cosine(vec1, vec2)
	if score < 0 || score > 1 {
		t.Fatalf("expected score in [0,1], got %f", score)
	}
}

func TestRoundTrip(t *testing.T) {
	vec := Compute("round trip test")
	b := ToBytes(vec)
	if len(b) != Dim*4 {
		t.Fatalf("expected %d bytes, got %d", Dim*4, len(b))
	}
	recovered := FromBytes(b)
	if len(recovered) != Dim {
		t.Fatalf("expected %d dims, got %d", Dim, len(recovered))
	}
	for i := range vec {
		if vec[i] != recovered[i] {
			t.Fatalf("mismatch at index %d: %f vs %f", i, vec[i], recovered[i])
		}
	}
}
