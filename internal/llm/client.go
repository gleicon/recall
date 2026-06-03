package llm

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client abstracts a local LLM backend with an OpenAI-compatible API.
type Client struct {
	Endpoint     string
	Type         string
	Models       []string
	Selected     string
	QueryTimeout time.Duration
}

// PreferredModel returns the best model to use.
func (c *Client) PreferredModel(preferred string) string {
	if preferred != "" {
		for _, m := range c.Models {
			if m == preferred {
				return preferred
			}
		}
	}
	return c.pickSmallestModel()
}

func (c *Client) pickSmallestModel() string {
	if len(c.Models) == 0 {
		return "default"
	}

	small := []string{"1b", "2b", "3b", "4b", "nano", "tiny", "mini", "small"}
	avoid := []string{"27b", "35b", "123b", "198b", "70b", "40b", "65b", "20b", "moe"}

	for _, m := range c.Models {
		lower := strings.ToLower(m)
		for _, kw := range small {
			if strings.Contains(lower, kw) {
				return m
			}
		}
	}

	for _, m := range c.Models {
		lower := strings.ToLower(m)
		bad := false
		for _, kw := range avoid {
			if strings.Contains(lower, kw) {
				bad = true
				break
			}
		}
		if !bad {
			return m
		}
	}

	return c.Models[0]
}

// Detect probes common local LLM endpoints and returns the first working one.
func Detect() *Client {
	if c := detectOpenAI("http://localhost:8080", "llama-app"); c != nil {
		return c
	}
	for _, port := range []string{"8000", "5000", "11434"} {
		if c := detectOpenAI("http://localhost:"+port, "generic-openai"); c != nil {
			return c
		}
	}
	return nil
}

func detectOpenAI(endpoint, clientType string) *Client {
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Get(endpoint + "/v1/models")
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
		Models []struct {
			ID string `json:"id"`
		} `json:"models"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil
	}

	var models []string
	for _, m := range result.Data {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	for _, m := range result.Models {
		if m.ID != "" {
			models = append(models, m.ID)
		}
	}
	if len(models) == 0 {
		models = append(models, "default")
	}

	return &Client{
		Endpoint:     endpoint,
		Type:         clientType,
		Models:       models,
		Selected:     models[0],
		QueryTimeout: 30 * time.Second,
	}
}

func (c *Client) SetTimeout(d time.Duration) {
	if d > 0 {
		c.QueryTimeout = d
	}
}

// Query sends a prompt to the local LLM and returns the response.
func (c *Client) Query(prompt string, systemContext string) (string, error) {
	model := c.Selected
	if model == "" {
		model = "default"
	}

	messages := []map[string]string{}
	if systemContext != "" {
		messages = append(messages, map[string]string{"role": "system", "content": systemContext})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	timeout := c.QueryTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(c.Endpoint+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return result.Choices[0].Message.Content, nil
}

// QueryStream sends a prompt to the local LLM and streams tokens via onToken callback.
func (c *Client) QueryStream(prompt string, systemContext string, onToken func(string)) error {
	model := c.Selected
	if model == "" {
		model = "default"
	}

	messages := []map[string]string{}
	if systemContext != "" {
		messages = append(messages, map[string]string{"role": "system", "content": systemContext})
	}
	messages = append(messages, map[string]string{"role": "user", "content": prompt})

	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   true,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	timeout := c.QueryTimeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	client := &http.Client{Timeout: timeout}
	resp, err := client.Post(c.Endpoint+"/v1/chat/completions", "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			onToken(chunk.Choices[0].Delta.Content)
		}
	}
	return nil
}

// GetEmbedding requests an embedding from the local model.
func (c *Client) GetEmbedding(text string) ([]float32, error) {
	model := c.Selected
	if model == "" {
		model = "default"
	}

	payload := map[string]interface{}{
		"model": model,
		"input": text,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(c.Endpoint+"/v1/embeddings", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	if len(result.Data) == 0 {
		return nil, fmt.Errorf("no embedding data in response")
	}

	out := make([]float32, len(result.Data[0].Embedding))
	for i, v := range result.Data[0].Embedding {
		out[i] = float32(v)
	}
	return out, nil
}
