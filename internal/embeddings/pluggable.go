package embeddings

import (
	"github.com/gleicon/recall/internal/llm"
)

func ComputeSmart(text, embedModel string) []float32 {
	client := llm.Detect()
	if client != nil {
		if embedModel != "" {
			client.Selected = embedModel
		}
		emb, err := client.GetEmbedding(text)
		if err == nil && len(emb) > 0 {
			return emb
		}
	}
	return Compute(text)
}

func ComputeSmartWithClient(text, embedModel string, client *llm.Client) []float32 {
	if client != nil {
		saved := client.Selected
		if embedModel != "" {
			client.Selected = embedModel
		}
		emb, err := client.GetEmbedding(text)
		client.Selected = saved
		if err == nil && len(emb) > 0 {
			return emb
		}
	}
	return Compute(text)
}
