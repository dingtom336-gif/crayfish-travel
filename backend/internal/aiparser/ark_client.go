package aiparser

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ARKClient wraps Volcengine ARK API calls for NLP parsing (OpenAI-compatible).
type ARKClient struct {
	apiKey      string
	baseURL     string
	model       string
	temperature float64
	maxTokens   int
	client      *http.Client
}

// NewARKClient creates a Volcengine ARK API client.
func NewARKClient(apiKey, baseURL, model string, temperature float64, maxTokens, timeout int) *ARKClient {
	return &ARKClient{
		apiKey:      apiKey,
		baseURL:     baseURL,
		model:       model,
		temperature: temperature,
		maxTokens:   maxTokens,
		client:      &http.Client{Timeout: time.Duration(timeout) * time.Second},
	}
}

const arkSystemPrompt = `你是旅行需求解析助手。从用户输入中提取结构化旅行参数，返回JSON。
字段：destination(目的地), start_date(YYYY-MM-DD), end_date(YYYY-MM-DD), budget_cents(预算，单位分), adults(成人数), children(儿童数), preferences(偏好数组)
如果某字段无法确定，使用合理默认值。budget用分表示(如8000元=800000)。
只返回JSON，不要其他文字。`

// arkChatRequest is the OpenAI-compatible request body.
type arkChatRequest struct {
	Model       string       `json:"model"`
	Messages    []arkMessage `json:"messages"`
	Temperature float64      `json:"temperature"`
	MaxTokens   int          `json:"max_tokens"`
}

type arkMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// arkChatResponse is the OpenAI-compatible response body.
type arkChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

// Parse sends natural language input to ARK and returns structured requirements.
// Retries once on failure. Returns error (NOT mock data) if both attempts fail.
func (a *ARKClient) Parse(rawInput string) (*TravelRequirement, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		result, err := a.doRequest(rawInput)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("ark parse failed after 2 attempts: %w", lastErr)
}

func (a *ARKClient) doRequest(rawInput string) (*TravelRequirement, error) {
	reqBody := arkChatRequest{
		Model: a.model,
		Messages: []arkMessage{
			{Role: "system", Content: arkSystemPrompt},
			{Role: "user", Content: rawInput},
		},
		Temperature: a.temperature,
		MaxTokens:   a.maxTokens,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal ark request: %w", err)
	}

	url := strings.TrimRight(a.baseURL, "/") + "/chat/completions"
	req, err := http.NewRequest("POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create ark request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+a.apiKey)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ark api call: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ark api error %d: %s", resp.StatusCode, string(respBody))
	}

	var chatResp arkChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode ark response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("empty response from ark")
	}

	text := chatResp.Choices[0].Message.Content
	// Strip ```json fences if present
	text = stripJSONFences(text)

	var result TravelRequirement
	if err := json.Unmarshal([]byte(text), &result); err != nil {
		return nil, fmt.Errorf("parse ark structured output: %w", err)
	}

	return &result, nil
}

// stripJSONFences removes markdown code fences from LLM output.
func stripJSONFences(s string) string {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "```json") {
		s = strings.TrimPrefix(s, "```json")
	} else if strings.HasPrefix(s, "```") {
		s = strings.TrimPrefix(s, "```")
	}
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}
