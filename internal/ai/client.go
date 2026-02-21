package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/nikmd1306/cwai/internal/prompt"
)

type ModelFamily int

const (
	ModelFamilyStandard  ModelFamily = iota
	ModelFamilyReasoning
	ModelFamilyGPT5
)

type Params struct {
	APIKey           string
	APIURL           string
	Model            string
	MaxTokensOutput  int
	Temperature      float64
	HasTemperature   bool
	ReasoningEffort  string
	Verbosity        string
	StructuredOutput string
}

type Client struct {
	params Params
	http   *http.Client
}

func NewClient(p Params) *Client {
	p.APIURL = strings.TrimRight(p.APIURL, "/")
	return &Client{
		params: p,
		http: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func DetectModelFamily(model string) ModelFamily {
	m := strings.ToLower(model)

	if strings.HasPrefix(m, "gpt-5") {
		return ModelFamilyGPT5
	}

	if strings.HasPrefix(m, "o1") ||
		strings.HasPrefix(m, "o3") ||
		strings.HasPrefix(m, "o4") {
		return ModelFamilyReasoning
	}

	return ModelFamilyStandard
}

func tokenParamName(apiURL string) string {
	u := strings.ToLower(apiURL)
	if strings.Contains(u, "openrouter.ai") || strings.Contains(u, "deepseek.com") {
		return "max_tokens"
	}
	return "max_completion_tokens"
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *Client) IsStructuredOutput() bool {
	return SupportsStructuredOutput(c.params.APIURL, c.params.StructuredOutput)
}

func (c *Client) GenerateCommitMessage(messages []prompt.Message) (string, error) {
	body := c.buildRequestBody(messages)

	data, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.params.APIURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.params.APIKey)

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("API request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("API returned no choices")
	}

	content := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	if c.IsStructuredOutput() {
		parsed, err := ParseCommitMessageJSON(content)
		if err == nil {
			return AssembleCommitMessage(parsed), nil
		}
	}

	return content, nil
}

func (c *Client) buildRequestBody(messages []prompt.Message) map[string]any {
	family := DetectModelFamily(c.params.Model)
	body := map[string]any{
		"model":    c.params.Model,
		"messages": messages,
	}

	tokenParam := tokenParamName(c.params.APIURL)
	body[tokenParam] = c.params.MaxTokensOutput

	switch family {
	case ModelFamilyStandard:
		if c.params.HasTemperature {
			body["temperature"] = c.params.Temperature
		} else {
			body["temperature"] = 0.7
		}

	case ModelFamilyReasoning:
		effort := c.params.ReasoningEffort
		if effort == "" {
			effort = "low"
		}
		body["reasoning_effort"] = effort

	case ModelFamilyGPT5:
		effort := c.params.ReasoningEffort
		if effort == "" {
			effort = "low"
		}
		body["reasoning_effort"] = effort

		verbosity := c.params.Verbosity
		if verbosity == "" {
			verbosity = "low"
		}
		body["verbosity"] = verbosity
	}

	if c.IsStructuredOutput() {
		body["response_format"] = BuildResponseFormat()
	}

	return body
}
