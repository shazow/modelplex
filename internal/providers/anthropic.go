// Package providers implements AI provider abstractions.
// AnthropicProvider provides Anthropic Claude API integration with key differences from OpenAI:
// - Uses "x-api-key" header instead of "Authorization: Bearer"
// - Requires "anthropic-version" header for API versioning
// - Transforms OpenAI message format: system messages become separate "system" field
// - Uses "/messages" endpoint instead of "/chat/completions"
// - Requires explicit max_tokens parameter (defaults to 4096)
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/modelplex/modelplex/internal/config"
)

type AnthropicProvider struct {
	name     string
	baseURL  string
	apiKey   string
	models   []string
	priority int
	client   *http.Client
}

func NewAnthropicProvider(cfg config.Provider) *AnthropicProvider {
	apiKey := cfg.APIKey
	if strings.HasPrefix(apiKey, "${") && strings.HasSuffix(apiKey, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(apiKey, "${"), "}")
		apiKey = os.Getenv(envVar)
	}

	return &AnthropicProvider{
		name:     cfg.Name,
		baseURL:  cfg.BaseURL,
		apiKey:   apiKey,
		models:   cfg.Models,
		priority: cfg.Priority,
		client:   &http.Client{},
	}
}

func (p *AnthropicProvider) Name() string {
	return p.name
}

func (p *AnthropicProvider) Priority() int {
	return p.priority
}

func (p *AnthropicProvider) ListModels() []string {
	return p.models
}

func (p *AnthropicProvider) ChatCompletion(ctx context.Context, model string, messages []map[string]interface{}) (interface{}, error) {
	anthropicMessages := make([]map[string]interface{}, 0)
	var systemMessage string

	for _, msg := range messages {
		role := msg["role"].(string)
		content := msg["content"].(string)

		if role == "system" {
			systemMessage = content
		} else {
			anthropicMessages = append(anthropicMessages, map[string]interface{}{
				"role":    role,
				"content": content,
			})
		}
	}

	payload := map[string]interface{}{
		"model":      model,
		"messages":   anthropicMessages,
		"max_tokens": 4096,
	}

	if systemMessage != "" {
		payload["system"] = systemMessage
	}

	return p.makeRequest(ctx, "/messages", payload)
}

func (p *AnthropicProvider) Completion(ctx context.Context, model string, prompt string) (interface{}, error) {
	messages := []map[string]interface{}{
		{"role": "user", "content": prompt},
	}
	return p.ChatCompletion(ctx, model, messages)
}

func (p *AnthropicProvider) makeRequest(ctx context.Context, endpoint string, payload interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var result interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return result, nil
}
