package embeddings

import (
	"github.com/gleicon/technocore/internal/llm"
)

func ComputeSmart(text string) []float32 {
	client := llm.Detect()
	if client != nil {
		emb, err := client.GetEmbedding(text)
		if err == nil && len(emb) > 0 {
			return emb
		}
	}
	return Compute(text)
}
