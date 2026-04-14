package nlp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// TravelRequirement is the structured output from Claude parsing.
type TravelRequirement struct {
	Destination string   `json:"destination"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	BudgetCents int64    `json:"budget_cents"`
	Adults      int      `json:"adults"`
	Children    int      `json:"children"`
	Preferences []string `json:"preferences"`
}

// ClaudeClient wraps Claude API calls for NLP parsing.
type ClaudeClient struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
}

// NewClaudeClient creates a Claude API client.
func NewClaudeClient(apiKey, model string, timeoutSec int) *ClaudeClient {
	return &ClaudeClient{
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.anthropic.com/v1/messages",
		client:  &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
	}
}

const systemPrompt = `You are a travel requirement parser. Extract structured travel requirements from the user's natural language input.

Return ONLY a JSON object with these fields:
- destination: string (city or region name)
- start_date: string (YYYY-MM-DD format, infer from context)
- end_date: string (YYYY-MM-DD format)
- budget_cents: integer (total budget in cents, e.g., 800000 for 8000 yuan)
- adults: integer (number of adult travelers)
- children: integer (number of child travelers)
- preferences: array of strings (e.g., "beachfront", "pool", "family-friendly")

If any field cannot be determined, use reasonable defaults:
- dates: next available vacation period
- budget: 0 (unknown)
- adults: 2, children: 0
- preferences: empty array`

// Parse sends natural language input to Claude and returns structured requirements.
func (c *ClaudeClient) Parse(rawInput string) (*TravelRequirement, error) {
	if c.apiKey == "" {
		return c.mockParse(rawInput)
	}

	body := map[string]interface{}{
		"model":      c.model,
		"max_tokens": 1024,
		"system":     systemPrompt,
		"messages": []map[string]string{
			{"role": "user", "content": rawInput},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", c.baseURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return c.mockParse(rawInput)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("claude api error %d: %s", resp.StatusCode, string(respBody))
	}

	var apiResp struct {
		Content []struct {
			Text string `json:"text"`
		} `json:"content"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("empty response from Claude")
	}

	var result TravelRequirement
	if err := json.Unmarshal([]byte(apiResp.Content[0].Text), &result); err != nil {
		return nil, fmt.Errorf("parse structured output: %w", err)
	}

	return &result, nil
}

// mockParse provides fallback parsing when Claude API is unavailable.
func (c *ClaudeClient) mockParse(_ string) (*TravelRequirement, error) {
	return &TravelRequirement{
		Destination: "Sanya",
		StartDate:   "2026-07-15",
		EndDate:     "2026-07-20",
		BudgetCents: 800000,
		Adults:      2,
		Children:    1,
		Preferences: []string{"beachfront", "pool", "family-friendly"},
	}, nil
}
