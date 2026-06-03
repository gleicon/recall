package embeddings

import (
	"testing"
)

func TestComputeSmartFallback(t *testing.T) {
	vec := ComputeSmart("test text for embedding")
	if len(vec) == 0 {
		t.Fatal("expected non-empty vector")
	}
	var sum float64
	for _, v := range vec {
		sum += float64(v) * float64(v)
	}
	if sum == 0 {
		t.Fatal("expected non-zero normalized vector")
	}
}
