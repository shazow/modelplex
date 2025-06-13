package monitoring

import (
	"bytes"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
			// Capture slog output
			var buf bytes.Buffer
			handler := slog.NewTextHandler(&buf, nil)
			oldLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			defer slog.SetDefault(oldLogger)

			logger := NewLogger(tt.enabled)
			reqLog := tt.requestLog // Create copy to avoid memory aliasing
			logger.LogRequest(&reqLog)

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "Request logged")
				assert.Contains(t, output, tt.requestLog.RequestID)
				assert.Contains(t, output, tt.requestLog.Model)
				assert.Contains(t, output, tt.requestLog.Provider)
				assert.Contains(t, output, "INFO")
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
			handler := slog.NewTextHandler(&buf, nil)
			oldLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			defer slog.SetDefault(oldLogger)

			logger := NewLogger(tt.enabled)
			logger.LogError(tt.component, tt.message, tt.err)

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "Component error")
				assert.Contains(t, output, tt.component)
				assert.Contains(t, output, tt.message)
				assert.Contains(t, output, tt.err.Error())
				assert.Contains(t, output, "ERROR")
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
			handler := slog.NewTextHandler(&buf, nil)
			oldLogger := slog.Default()
			slog.SetDefault(slog.New(handler))
			defer slog.SetDefault(oldLogger)

			logger := NewLogger(tt.enabled)
			logger.LogInfo(tt.component, tt.message, tt.metadata)

			output := buf.String()
			if tt.expectOutput {
				assert.Contains(t, output, "Component info")
				assert.Contains(t, output, tt.component)
				assert.Contains(t, output, tt.message)
				assert.Contains(t, output, "INFO")

				if tt.metadata != nil {
					for key := range tt.metadata {
						assert.Contains(t, output, key)
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
	handler := slog.NewTextHandler(&buf, nil)
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(oldLogger)

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

	logger.LogRequest(&requestLog)

	output := buf.String()
	assert.Contains(t, output, "Request logged")
	assert.Contains(t, output, requestLog.RequestID)
	assert.Contains(t, output, requestLog.Model)
	assert.Contains(t, output, requestLog.Provider)
	assert.Contains(t, output, "INFO")
	assert.Contains(t, output, "user-123")
	assert.Contains(t, output, "0.7")
}
