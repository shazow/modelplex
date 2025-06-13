package testutil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/modelplex/modelplex/internal/config"
	"github.com/stretchr/testify/require"
)

// CreateTestConfig creates a basic test configuration
func CreateTestConfig() *config.Config {
	return &config.Config{
		Server: config.Server{
			LogLevel:       "debug",
			MaxRequestSize: 1024 * 1024,
		},
		Providers: []config.Provider{
			{
				Name:     "test-openai",
				Type:     "openai",
				BaseURL:  "https://api.openai.com/v1",
				APIKey:   "test-key",
				Models:   []string{"gpt-4", "gpt-3.5-turbo"},
				Priority: 1,
			},
			{
				Name:     "test-anthropic",
				Type:     "anthropic",
				BaseURL:  "https://api.anthropic.com/v1",
				APIKey:   "test-anthropic-key",
				Models:   []string{"claude-3-sonnet"},
				Priority: 2,
			},
		},
		MCP: config.MCPConfig{
			Servers: []config.MCPServer{
				{
					Name:    "test-mcp",
					Command: "echo",
					Args:    []string{"test"},
				},
			},
		},
	}
}

// CreateMockHTTPServer creates a mock HTTP server for testing providers
func CreateMockHTTPServer(t *testing.T, responses map[string]func(w http.ResponseWriter, r *http.Request)) *httptest.Server {
	mux := http.NewServeMux()
	
	for path, handler := range responses {
		mux.HandleFunc(path, handler)
	}
	
	return httptest.NewServer(mux)
}

// CreateOpenAIMockResponse creates a mock OpenAI API response
func CreateOpenAIMockResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":      "chatcmpl-test123",
		"object":  "chat.completion",
		"created": 1677652288,
		"model":   "gpt-4",
		"choices": []map[string]interface{}{
			{
				"index": 0,
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": "Hello! This is a test response.",
				},
				"finish_reason": "stop",
			},
		},
		"usage": map[string]interface{}{
			"prompt_tokens":     10,
			"completion_tokens": 8,
			"total_tokens":      18,
		},
	}
}

// CreateAnthropicMockResponse creates a mock Anthropic API response
func CreateAnthropicMockResponse() map[string]interface{} {
	return map[string]interface{}{
		"id":   "msg-test123",
		"type": "message",
		"role": "assistant",
		"content": []map[string]interface{}{
			{
				"type": "text",
				"text": "Hello! This is a test response from Claude.",
			},
		},
		"model":       "claude-3-sonnet",
		"stop_reason": "end_turn",
		"usage": map[string]interface{}{
			"input_tokens":  10,
			"output_tokens": 9,
		},
	}
}

// CreateOllamaMockResponse creates a mock Ollama API response
func CreateOllamaMockResponse() map[string]interface{} {
	return map[string]interface{}{
		"model":      "llama2",
		"created_at": "2023-08-04T19:22:45.499127Z",
		"message": map[string]interface{}{
			"role":    "assistant",
			"content": "Hello! This is a test response from Llama.",
		},
		"done":                true,
		"total_duration":      4935312500,
		"load_duration":       534986708,
		"prompt_eval_count":   10,
		"prompt_eval_duration": 107345000,
		"eval_count":          9,
		"eval_duration":       4289430500,
	}
}

// AssertJSONResponse verifies that a response is valid JSON and returns the parsed data
func AssertJSONResponse(t *testing.T, resp *http.Response) map[string]interface{} {
	require.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	
	var data map[string]interface{}
	err := json.NewDecoder(resp.Body).Decode(&data)
	require.NoError(t, err)
	
	return data
}

// CreateChatMessages creates test chat messages
func CreateChatMessages() []map[string]interface{} {
	return []map[string]interface{}{
		{"role": "system", "content": "You are a helpful assistant."},
		{"role": "user", "content": "Hello, how are you?"},
	}
}

// CreateMockProviderConfig creates a test provider configuration
func CreateMockProviderConfig(name, providerType, baseURL string) config.Provider {
	return config.Provider{
		Name:     name,
		Type:     providerType,
		BaseURL:  baseURL,
		APIKey:   "test-key",
		Models:   []string{"test-model"},
		Priority: 1,
	}
}

// WriteJSONResponse writes a JSON response to an HTTP response writer
func WriteJSONResponse(t *testing.T, w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	require.NoError(t, err)
}

// WriteErrorResponse writes an error response
func WriteErrorResponse(t *testing.T, w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	response := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "test_error",
		},
	}
	WriteJSONResponse(t, w, response)
}