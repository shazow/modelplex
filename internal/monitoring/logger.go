package monitoring

import (
	"encoding/json"
	"log"
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

func (l *Logger) LogRequest(reqLog RequestLog) {
	if !l.enabled {
		return
	}

	reqLog.Timestamp = time.Now()

	data, err := json.Marshal(reqLog)
	if err != nil {
		log.Printf("Failed to marshal request log: %v", err)
		return
	}

	log.Printf("REQUEST_LOG: %s", string(data))
}

func (l *Logger) LogError(component, message string, err error) {
	if !l.enabled {
		return
	}

	errorLog := map[string]interface{}{
		"timestamp": time.Now(),
		"component": component,
		"message":   message,
		"error":     err.Error(),
	}

	data, jsonErr := json.Marshal(errorLog)
	if jsonErr != nil {
		log.Printf("ERROR in %s: %s - %v", component, message, err)
		return
	}

	log.Printf("ERROR_LOG: %s", string(data))
}

func (l *Logger) LogInfo(component, message string, metadata map[string]interface{}) {
	if !l.enabled {
		return
	}

	infoLog := map[string]interface{}{
		"timestamp": time.Now(),
		"component": component,
		"message":   message,
		"metadata":  metadata,
	}

	data, err := json.Marshal(infoLog)
	if err != nil {
		log.Printf("INFO %s: %s", component, message)
		return
	}

	log.Printf("INFO_LOG: %s", string(data))
}
