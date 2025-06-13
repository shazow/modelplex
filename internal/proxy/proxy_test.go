package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockMultiplexer implements the multiplexer interface for testing
type MockMultiplexer struct {
	mock.Mock
}

func (m *MockMultiplexer) ChatCompletion(ctx context.Context, model string, messages []map[string]interface{}) (interface{}, error) {
	args := m.Called(ctx, model, messages)
	return args.Get(0), args.Error(1)
}

func (m *MockMultiplexer) Completion(ctx context.Context, model string, prompt string) (interface{}, error) {
	args := m.Called(ctx, model, prompt)
	return args.Get(0), args.Error(1)
}

func (m *MockMultiplexer) ListModels() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestOpenAIProxy_HandleChatCompletions(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		mockResponse   interface{}
		mockError      error
		expectedStatus int
		expectedModel  string
	}{
		{
			name: "successful request",
			requestBody: map[string]interface{}{
				"model": "gpt-4",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			mockResponse: map[string]interface{}{
				"id":      "chatcmpl-123",
				"object":  "chat.completion",
				"created": float64(1677652288),
				"choices": []interface{}{
					map[string]interface{}{
						"message": map[string]interface{}{
							"role":    "assistant",
							"content": "Hello! How can I help you?",
						},
					},
				},
			},
			expectedStatus: http.StatusOK,
			expectedModel:  "gpt-4",
		},
		{
			name: "modelplex prefix stripped",
			requestBody: map[string]interface{}{
				"model": "modelplex-claude-3-sonnet",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			mockResponse: map[string]interface{}{
				"id": "msg-123",
			},
			expectedStatus: http.StatusOK,
			expectedModel:  "claude-3-sonnet",
		},
		{
			name: "provider error",
			requestBody: map[string]interface{}{
				"model": "gpt-4",
				"messages": []map[string]interface{}{
					{"role": "user", "content": "Hello"},
				},
			},
			mockError:      errors.New("provider unavailable"),
			expectedStatus: http.StatusInternalServerError,
			expectedModel:  "gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMux := &MockMultiplexer{}
			proxy := New(mockMux)

			// Set up mock expectations
			if tt.mockError != nil {
				mockMux.On("ChatCompletion", mock.Anything, tt.expectedModel, mock.Anything).Return(nil, tt.mockError)
			} else {
				mockMux.On("ChatCompletion", mock.Anything, tt.expectedModel, mock.Anything).Return(tt.mockResponse, nil)
			}

			// Create request
			reqBody, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Execute
			proxy.HandleChatCompletions(w, req)

			// Verify
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.NewDecoder(w.Body).Decode(&response)
				require.NoError(t, err)
				assert.Equal(t, tt.mockResponse, response)
			}

			mockMux.AssertExpectations(t)
		})
	}
}

func TestOpenAIProxy_HandleChatCompletions_InvalidJSON(t *testing.T) {
	mockMux := &MockMultiplexer{}
	proxy := New(mockMux)

	req := httptest.NewRequest("POST", "/v1/chat/completions", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proxy.HandleChatCompletions(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid JSON")
}

func TestOpenAIProxy_HandleCompletions(t *testing.T) {
	mockMux := &MockMultiplexer{}
	proxy := New(mockMux)

	requestBody := map[string]interface{}{
		"model":  "gpt-3.5-turbo-instruct",
		"prompt": "Complete this sentence",
	}

	mockResponse := map[string]interface{}{
		"id":      "cmpl-123",
		"object":  "text_completion",
		"created": float64(1677652288),
		"choices": []interface{}{
			map[string]interface{}{
				"text":  " with something interesting.",
				"index": float64(0),
			},
		},
	}

	mockMux.On("Completion", mock.Anything, "gpt-3.5-turbo-instruct", "Complete this sentence").Return(mockResponse, nil)

	reqBody, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "/v1/completions", bytes.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	proxy.HandleCompletions(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, mockResponse, response)

	mockMux.AssertExpectations(t)
}

func TestOpenAIProxy_HandleModels(t *testing.T) {
	mockMux := &MockMultiplexer{}
	proxy := New(mockMux)

	mockModels := []string{"gpt-4", "gpt-3.5-turbo", "claude-3-sonnet"}
	mockMux.On("ListModels").Return(mockModels)

	req := httptest.NewRequest("GET", "/v1/models", http.NoBody)
	w := httptest.NewRecorder()

	proxy.HandleModels(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response ModelsResponse
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "list", response.Object)
	assert.Len(t, response.Data, 3)

	for i, model := range mockModels {
		assert.Equal(t, model, response.Data[i].ID)
		assert.Equal(t, "model", response.Data[i].Object)
		assert.Equal(t, "modelplex", response.Data[i].OwnedBy)
		assert.Equal(t, int64(1677610602), response.Data[i].Created)
	}

	mockMux.AssertExpectations(t)
}

func TestNormalizeModel(t *testing.T) {
	proxy := &OpenAIProxy{}

	tests := []struct {
		input    string
		expected string
	}{
		{"gpt-4", "gpt-4"},
		{"modelplex-gpt-4", "gpt-4"},
		{"modelplex-claude-3-sonnet", "claude-3-sonnet"},
		{"claude-3-sonnet", "claude-3-sonnet"},
		{"modelplex-", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := proxy.normalizeModel(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()

	writeError(w, http.StatusBadRequest, "Test error message")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, "application/json", w.Header().Get("Content-Type"))

	var response map[string]interface{}
	err := json.NewDecoder(w.Body).Decode(&response)
	require.NoError(t, err)

	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "Test error message", errorObj["message"])
	assert.Equal(t, "invalid_request_error", errorObj["type"])
}
