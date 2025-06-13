// Package monitoring provides structured logging and metrics collection.
package monitoring

import (
	"log/slog"
	"time"
)

type RequestLog struct {
	Timestamp  time.Time              `json:"timestamp"`
	RequestID  string                 `json:"request_id"`
	Model      string                 `json:"model"`
	Provider   string                 `json:"provider"`
	Method     string                 `json:"method"`
	TokensUsed int                    `json:"tokens_used,omitempty"`
	Duration   time.Duration          `json:"duration"`
	Success    bool                   `json:"success"`
	Error      string                 `json:"error,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

type Logger struct {
	enabled bool
}

func NewLogger(enabled bool) *Logger {
	return &Logger{enabled: enabled}
}

func (l *Logger) LogRequest(reqLog *RequestLog) {
	if !l.enabled {
		return
	}

	reqLog.Timestamp = time.Now()

	slog.Info("Request logged",
		"timestamp", reqLog.Timestamp,
		"request_id", reqLog.RequestID,
		"model", reqLog.Model,
		"provider", reqLog.Provider,
		"method", reqLog.Method,
		"tokens_used", reqLog.TokensUsed,
		"duration", reqLog.Duration,
		"success", reqLog.Success,
		"error", reqLog.Error,
		"metadata", reqLog.Metadata)
}

func (l *Logger) LogError(component, message string, err error) {
	if !l.enabled {
		return
	}

	slog.Error("Component error",
		"timestamp", time.Now(),
		"component", component,
		"message", message,
		"error", err.Error())
}

func (l *Logger) LogInfo(component, message string, metadata map[string]interface{}) {
	if !l.enabled {
		return
	}

	slog.Info("Component info",
		"timestamp", time.Now(),
		"component", component,
		"message", message,
		"metadata", metadata)
}
