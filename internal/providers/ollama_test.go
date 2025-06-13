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

func TestNewOllamaProvider(t *testing.T) {
	cfg := config.Provider{
		Name:     "local",
		BaseURL:  "http://localhost:11434",
		Models:   []string{"llama2", "codellama"},
		Priority: 3,
	}

	provider := NewOllamaProvider(&cfg)

	assert.Equal(t, "local", provider.Name())
	assert.Equal(t, "http://localhost:11434", provider.baseURL)
	assert.Equal(t, []string{"llama2", "codellama"}, provider.ListModels())
	assert.Equal(t, 3, provider.Priority())
}

func TestOllamaProvider_ChatCompletion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/chat", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "llama2", req["model"])
		assert.Equal(t, false, req["stream"])
		assert.NotEmpty(t, req["messages"])

		response := map[string]interface{}{
			"model":      "llama2",
			"created_at": "2023-08-04T19:22:45.499127Z",
			"message": map[string]interface{}{
				"role":    "assistant",
				"content": "Hello! How can I help you today?",
			},
			"done":                 true,
			"total_duration":       4935312500,
			"load_duration":        534986708,
			"prompt_eval_count":    26,
			"prompt_eval_duration": 107345000,
			"eval_count":           298,
			"eval_duration":        4289430500,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	provider := NewOllamaProvider(&config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Models:  []string{"llama2"},
	})

	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	result, err := provider.ChatCompletion(context.Background(), "llama2", messages)
	require.NoError(t, err)
	require.NotNil(t, result)

	response, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "llama2", response["model"])
	assert.Equal(t, true, response["done"])

	message := response["message"].(map[string]interface{})
	assert.Equal(t, "assistant", message["role"])
	assert.Contains(t, message["content"], "Hello")
}

func TestOllamaProvider_Completion(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/generate", r.URL.Path)

		var req map[string]interface{}
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		assert.Equal(t, "codellama", req["model"])
		assert.Equal(t, "def fibonacci(n):", req["prompt"])
		assert.Equal(t, false, req["stream"])

		response := map[string]interface{}{
			"model":                "codellama",
			"created_at":           "2023-08-04T19:22:45.499127Z",
			"response":             "\n    if n <= 1:\n        return n\n    return fibonacci(n-1) + fibonacci(n-2)",
			"done":                 true,
			"context":              []int{1, 2, 3},
			"total_duration":       4935312500,
			"load_duration":        534986708,
			"prompt_eval_count":    26,
			"prompt_eval_duration": 107345000,
			"eval_count":           298,
			"eval_duration":        4289430500,
		}

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("Failed to encode response: %v", err)
		}
	}))
	defer server.Close()

	provider := NewOllamaProvider(&config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Models:  []string{"codellama"},
	})

	result, err := provider.Completion(context.Background(), "codellama", "def fibonacci(n):")
	require.NoError(t, err)
	require.NotNil(t, result)

	response, ok := result.(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "codellama", response["model"])
	assert.Contains(t, response["response"], "fibonacci")
}

func TestOllamaProvider_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		if _, err := w.Write([]byte(`{"error": "model not found"}`)); err != nil {
			t.Errorf("Failed to write error response: %v", err)
		}
	}))
	defer server.Close()

	provider := NewOllamaProvider(&config.Provider{
		Name:    "test",
		BaseURL: server.URL,
		Models:  []string{"nonexistent"},
	})

	messages := []map[string]interface{}{
		{"role": "user", "content": "Hello"},
	}

	result, err := provider.ChatCompletion(context.Background(), "nonexistent", messages)
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "404")
}
