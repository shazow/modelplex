package providers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelplex/modelplex/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAnthropicProvider(t *testing.T) {
	cfg := config.Provider{
		Name:     "anthropic",
		BaseURL:  "https://api.anthropic.com/v1",
		APIKey:   "sk-ant-test123",
		Models:   []string{"claude-3-sonnet", "claude-3-haiku"},
		Priority: 1,
	}

	provider := NewAnthropicProvider(cfg)

	assert.Equal(t, "anthropic", provider.Name())
	assert.Equal(t, "https://api.anthropic.com/v1", provider.baseURL)
	assert.Equal(t, "sk-ant-test123", provider.apiKey)
	assert.Equal(t, []string{"claude-3-sonnet", "claude-3-haiku"}, provider.ListModels())
	assert.Equal(t, 1, provider.Priority())
}

func TestAnthropicProvider_ChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/messages", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "test-key", r.Header.Get("x-api-key"))
		assert.Equal(t, "2023-06-01", r.Header.Get("anthropic-version"))

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "claude-3-sonnet", req["model"])
		assert.Equal(t, float64(4096), req["max_tokens"])
		assert.NotEmpty(t, req["messages"])

		response := map[string]interface{}{
			"id":      "msg_123",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-3-sonnet",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "Hello! How can I help you today?",
				},
			},
			"stop_reason": "end_turn",
			"usage": map[string]interface{}{
				"input_tokens":  10,
				"output_tokens": 12,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewAnthropicProvider(config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		APIKey:  "test-key",
		Models:  []string{"claude-3-sonnet"},
	})

	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	result, err := provider.ChatCompletion(context.Background(), "claude-3-sonnet", messages)
	require.NoError(t, err)
	require.NotNil(t, result)

	response, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "msg_123", response["id"])
	assert.Equal(t, "message", response["type"])
}

func TestAnthropicProvider_ChatCompletion_WithSystem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Should have system message in separate field
		assert.Equal(t, "You are a helpful assistant", req["system"])
		
		// Messages should not contain system message
		messages := req["messages"].([]interface{})
		assert.Len(t, messages, 1)
		msg := messages[0].(map[string]interface{})
		assert.Equal(t, "user", msg["role"])

		response := map[string]interface{}{
			"id":   "msg_123",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "Hello!"},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewAnthropicProvider(config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		APIKey:  "test-key",
		Models:  []string{"claude-3-sonnet"},
	})

	messages := []map[string]interface{}{
		{"role": "system", "content": "You are a helpful assistant"},
		{"role": "user", "content": "Hello"},
	}

	result, err := provider.ChatCompletion(context.Background(), "claude-3-sonnet", messages)
	require.NoError(t, err)
	require.NotNil(t, result)
}

func TestAnthropicProvider_Completion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)

		// Completion should be converted to chat format
		messages := req["messages"].([]interface{})
		assert.Len(t, messages, 1)
		msg := messages[0].(map[string]interface{})
		assert.Equal(t, "user", msg["role"])
		assert.Equal(t, "Complete this sentence", msg["content"])

		response := map[string]interface{}{
			"id":   "msg_123",
			"type": "message",
			"role": "assistant",
			"content": []map[string]interface{}{
				{"type": "text", "text": "I'll complete it for you."},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	provider := NewAnthropicProvider(config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		APIKey:  "test-key",
		Models:  []string{"claude-3-sonnet"},
	})

	result, err := provider.Completion(context.Background(), "claude-3-sonnet", "Complete this sentence")
	require.NoError(t, err)
	require.NotNil(t, result)
}