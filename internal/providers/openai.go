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

type OpenAIProvider struct {
	name     string
	baseURL  string
	apiKey   string
	models   []string
	priority int
	client   *http.Client
}

func NewOpenAIProvider(cfg config.Provider) *OpenAIProvider {
	apiKey := cfg.APIKey
	if strings.HasPrefix(apiKey, "${") && strings.HasSuffix(apiKey, "}") {
		envVar := strings.TrimSuffix(strings.TrimPrefix(apiKey, "${"), "}")
		apiKey = os.Getenv(envVar)
	}

	return &OpenAIProvider{
		name:     cfg.Name,
		baseURL:  cfg.BaseURL,
		apiKey:   apiKey,
		models:   cfg.Models,
		priority: cfg.Priority,
		client:   &http.Client{},
	}
}

func (p *OpenAIProvider) Name() string {
	return p.name
}

func (p *OpenAIProvider) Priority() int {
	return p.priority
}

func (p *OpenAIProvider) ListModels() []string {
	return p.models
}

func (p *OpenAIProvider) ChatCompletion(ctx context.Context, model string, messages []map[string]interface{}) (interface{}, error) {
	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
	}

	return p.makeRequest(ctx, "/chat/completions", payload)
}

func (p *OpenAIProvider) Completion(ctx context.Context, model string, prompt string) (interface{}, error) {
	payload := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
	}

	return p.makeRequest(ctx, "/completions", payload)
}

func (p *OpenAIProvider) makeRequest(ctx context.Context, endpoint string, payload interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
