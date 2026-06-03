package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/gleicon/technocore/internal/summarizer"
	"github.com/spf13/cobra"
)

var sentenceCount *int
var fileName *string

func readFromFileOrStdin(filename string) (string, error) {
	var body []string
	var fh *os.File
	var err error
	if filename == "" {
		fh = os.Stdin
	} else {
		fh, err = os.Open(filename)
		if err != nil {
			return "", err
		}
		defer fh.Close()
	}
	scanner := bufio.NewScanner(fh)
	for scanner.Scan() {
		body = append(body, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(body, " "), nil
}

var tldrCmd = &cobra.Command{
	Use:   "tldr",
	Short: "Summarize text from stdin or a file",
	Long:  `Reads text and produces an extractive summary using tldt. No LLM calls.`,
	Run: func(cmd *cobra.Command, args []string) {
		body, err := readFromFileOrStdin(*fileName)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading:", err)
			return
		}
		res, err := summarizer.SummarizeResult(body, *sentenceCount)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error summarizing:", err)
			return
		}
		fmt.Println(res.Summary)
		fmt.Fprintf(os.Stderr, "Tokens: %d -> %d (%d%% reduction)\n", res.TokensIn, res.TokensOut, res.Reduction)
	},
}

func init() {
	rootCmd.AddCommand(tldrCmd)
	sentenceCount = tldrCmd.Flags().IntP("sentences", "s", 3, "Number of sentences")
	fileName = tldrCmd.Flags().StringP("file", "f", "", "File to read (optional)")
}
