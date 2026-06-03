package summarizer

import (
	"github.com/gleicon/tldt/pkg/tldt"
)

// Summarize runs tldt summarization with defaults.
func Summarize(text string, sentences int) (string, error) {
	if sentences <= 0 {
		sentences = 3
	}
	opts := tldt.SummarizeOptions{
		Algorithm: "lexrank",
		Sentences: sentences,
	}
	res, err := tldt.Summarize(text, opts)
	if err != nil {
		return "", err
	}
	return res.Summary, nil
}

// SummarizeResult returns the full tldt result.
func SummarizeResult(text string, sentences int) (tldt.Result, error) {
	if sentences <= 0 {
		sentences = 3
	}
	opts := tldt.SummarizeOptions{
		Algorithm: "lexrank",
		Sentences: sentences,
	}
	return tldt.Summarize(text, opts)
}
