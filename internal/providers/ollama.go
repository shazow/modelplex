// Package providers implements AI provider abstractions.
// OllamaProvider provides local Ollama API integration with key differences from OpenAI:
// - No authentication required (local server)
// - Uses "/api/chat" and "/api/generate" endpoints instead of "/chat/completions" and "/completions"
// - Requires explicit "stream": false parameter to disable streaming
// - Typically runs on localhost:11434 by default
// - Supports local LLM models without external API dependencies
package providers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/modelplex/modelplex/internal/config"
)

// OllamaProvider implements the Provider interface for Ollama local API.
type OllamaProvider struct {
	name     string
	baseURL  string
	models   []string
	priority int
	client   *http.Client
}

// NewOllamaProvider creates a new Ollama provider instance.
func NewOllamaProvider(cfg *config.Provider) *OllamaProvider {
	return &OllamaProvider{
		name:     cfg.Name,
		baseURL:  cfg.BaseURL,
		models:   cfg.Models,
		priority: cfg.Priority,
		client:   &http.Client{},
	}
}

// Name returns the provider name.
func (p *OllamaProvider) Name() string {
	return p.name
}

// Priority returns the provider priority for model routing.
func (p *OllamaProvider) Priority() int {
	return p.priority
}

// ListModels returns the list of available models for this provider.
func (p *OllamaProvider) ListModels() []string {
	return p.models
}

// ChatCompletion performs a chat completion request with Ollama-specific parameters.
func (p *OllamaProvider) ChatCompletion(
	ctx context.Context, model string, messages []map[string]interface{},
) (interface{}, error) {
	payload := map[string]interface{}{
		"model":    model,
		"messages": messages,
		"stream":   false,
	}

	return p.makeRequest(ctx, "/api/chat", payload)
}

// Completion performs a completion request using Ollama's generate endpoint.
func (p *OllamaProvider) Completion(ctx context.Context, model, prompt string) (interface{}, error) {
	payload := map[string]interface{}{
		"model":  model,
		"prompt": prompt,
		"stream": false,
	}

	return p.makeRequest(ctx, "/api/generate", payload)
}

func (p *OllamaProvider) makeRequest(ctx context.Context, endpoint string, payload interface{}) (interface{}, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+endpoint, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

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
