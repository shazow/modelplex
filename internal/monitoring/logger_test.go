package monitoring

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name    string
		enabled bool
	}{
		{"enabled logger", true},
		{"disabled logger", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := NewLogger(tt.enabled)
			assert.Equal(t, tt.enabled, logger.enabled)
		})
	}
}

func TestLogger_LogRequest(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		requestLog   RequestLog
		expectOutput bool
	}{
		{
			name:    "enabled logger logs request",
			enabled: true,
			requestLog: RequestLog{
				RequestID:  "req-123",
				Model:      "gpt-4",
				Provider:   "openai",
				Method:     "chat.completions",
				TokensUsed: 150,
				Duration:   500 * time.Millisecond,
				Success:    true,
			},
			expectOutput: true,
		},
		{
			name:    "disabled logger does not log",
			enabled: false,
			requestLog: RequestLog{
				RequestID: "req-456",
				Model:     "claude-3-sonnet",
				Provider:  "anthropic",
				Method:    "chat.completions",
				Success:   false,
				Error:     "Rate limit exceeded",
			},
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture log output
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			logger := NewLogger(tt.enabled)
			logger.LogRequest(tt.requestLog)

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "REQUEST_LOG:")
				assert.Contains(t, output, tt.requestLog.RequestID)
				assert.Contains(t, output, tt.requestLog.Model)
				assert.Contains(t, output, tt.requestLog.Provider)

				// Parse and verify JSON structure
				jsonStart := bytes.Index(buf.Bytes(), []byte("{"))
				if jsonStart != -1 {
					var logData RequestLog
					err := json.Unmarshal(buf.Bytes()[jsonStart:], &logData)
					require.NoError(t, err)
					assert.Equal(t, tt.requestLog.RequestID, logData.RequestID)
					assert.Equal(t, tt.requestLog.Model, logData.Model)
					assert.Equal(t, tt.requestLog.Success, logData.Success)
					assert.NotZero(t, logData.Timestamp)
				}
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_LogError(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		component    string
		message      string
		err          error
		expectOutput bool
	}{
		{
			name:         "enabled logger logs error",
			enabled:      true,
			component:    "multiplexer",
			message:      "Failed to route request",
			err:          errors.New("no providers available"),
			expectOutput: true,
		},
		{
			name:         "disabled logger does not log error",
			enabled:      false,
			component:    "proxy",
			message:      "Request failed",
			err:          errors.New("timeout"),
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			logger := NewLogger(tt.enabled)
			logger.LogError(tt.component, tt.message, tt.err)

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "ERROR_LOG:")
				assert.Contains(t, output, tt.component)
				assert.Contains(t, output, tt.message)
				assert.Contains(t, output, tt.err.Error())

				// Parse and verify JSON structure
				jsonStart := bytes.Index(buf.Bytes(), []byte("{"))
				if jsonStart != -1 {
					var logData map[string]interface{}
					err := json.Unmarshal(buf.Bytes()[jsonStart:], &logData)
					require.NoError(t, err)
					assert.Equal(t, tt.component, logData["component"])
					assert.Equal(t, tt.message, logData["message"])
					assert.Equal(t, tt.err.Error(), logData["error"])
					assert.NotNil(t, logData["timestamp"])
				}
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_LogInfo(t *testing.T) {
	tests := []struct {
		name         string
		enabled      bool
		component    string
		message      string
		metadata     map[string]interface{}
		expectOutput bool
	}{
		{
			name:      "enabled logger logs info",
			enabled:   true,
			component: "server",
			message:   "Server started",
			metadata: map[string]interface{}{
				"port":        float64(8080),
				"socket_path": "/tmp/modelplex.socket",
			},
			expectOutput: true,
		},
		{
			name:         "disabled logger does not log info",
			enabled:      false,
			component:    "mcp",
			message:      "MCP server connected",
			metadata:     nil,
			expectOutput: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log.SetOutput(&buf)
			defer log.SetOutput(nil)

			logger := NewLogger(tt.enabled)
			logger.LogInfo(tt.component, tt.message, tt.metadata)

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "INFO_LOG:")
				assert.Contains(t, output, tt.component)
				assert.Contains(t, output, tt.message)

				if tt.metadata != nil {
					for key := range tt.metadata {
						assert.Contains(t, output, key)
					}
				}

				// Parse and verify JSON structure
				jsonStart := bytes.Index(buf.Bytes(), []byte("{"))
				if jsonStart != -1 {
					var logData map[string]interface{}
					err := json.Unmarshal(buf.Bytes()[jsonStart:], &logData)
					require.NoError(t, err)
					assert.Equal(t, tt.component, logData["component"])
					assert.Equal(t, tt.message, logData["message"])
					assert.NotNil(t, logData["timestamp"])
					
					if tt.metadata != nil {
						metadata := logData["metadata"].(map[string]interface{})
						for key, value := range tt.metadata {
							assert.Equal(t, value, metadata[key])
						}
					}
				}
			} else {
				assert.Empty(t, output)
			}
		})
	}
}

func TestLogger_LogRequest_WithCompleteData(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer log.SetOutput(nil)

	logger := NewLogger(true)
	
	requestLog := RequestLog{
		RequestID:  "req-full-test",
		Model:      "gpt-4",
		Provider:   "openai",
		Method:     "chat.completions",
		TokensUsed: 245,
		Duration:   750 * time.Millisecond,
		Success:    true,
		Metadata: map[string]interface{}{
			"user_id":     "user-123",
			"temperature": 0.7,
		},
	}

	logger.LogRequest(requestLog)

	output := buf.String()
	assert.Contains(t, output, "REQUEST_LOG:")

	// Parse the JSON and verify all fields
	jsonStart := bytes.Index(buf.Bytes(), []byte("{"))
	require.Greater(t, jsonStart, -1)

	var logData RequestLog
	err := json.Unmarshal(buf.Bytes()[jsonStart:], &logData)
	require.NoError(t, err)

	assert.Equal(t, requestLog.RequestID, logData.RequestID)
	assert.Equal(t, requestLog.Model, logData.Model)
	assert.Equal(t, requestLog.Provider, logData.Provider)
	assert.Equal(t, requestLog.Method, logData.Method)
	assert.Equal(t, requestLog.TokensUsed, logData.TokensUsed)
	assert.Equal(t, requestLog.Duration, logData.Duration)
	assert.Equal(t, requestLog.Success, logData.Success)
	assert.NotZero(t, logData.Timestamp)
	
	// Verify metadata
	require.NotNil(t, logData.Metadata)
	assert.Equal(t, "user-123", logData.Metadata["user_id"])
	assert.Equal(t, 0.7, logData.Metadata["temperature"])
}