package proxy

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type OpenAIProxy struct {
	mux Multiplexer
}

func New(mux Multiplexer) *OpenAIProxy {
	return &OpenAIProxy{mux: mux}
}

type ChatCompletionRequest struct {
	Model    string                   `json:"model"`
	Messages []map[string]interface{} `json:"messages"`
}

type CompletionRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ModelsResponse struct {
	Object string      `json:"object"`
	Data   []ModelInfo `json:"data"`
}

type ModelInfo struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

func (p *OpenAIProxy) HandleChatCompletions(w http.ResponseWriter, r *http.Request) {
	var req ChatCompletionRequest
	if err := p.decodeJSONRequest(r, &req, w); err != nil {
		return
	}

	model := p.normalizeModel(req.Model)
	result, err := p.mux.ChatCompletion(r.Context(), model, req.Messages)
	p.handleResponse(w, result, err, "chat completion")
}

func (p *OpenAIProxy) HandleCompletions(w http.ResponseWriter, r *http.Request) {
	var req CompletionRequest
	if err := p.decodeJSONRequest(r, &req, w); err != nil {
		return
	}

	model := p.normalizeModel(req.Model)
	result, err := p.mux.Completion(r.Context(), model, req.Prompt)
	p.handleResponse(w, result, err, "completion")
}

func (p *OpenAIProxy) HandleModels(w http.ResponseWriter, r *http.Request) {
	models := p.mux.ListModels()

	data := make([]ModelInfo, len(models))
	for i, model := range models {
		data[i] = ModelInfo{
			ID:      model,
			Object:  "model",
			Created: 1677610602,
			OwnedBy: "modelplex",
		}
	}

	response := ModelsResponse{
		Object: "list",
		Data:   data,
	}

	p.writeJSONResponse(w, response, "models")
}

func (p *OpenAIProxy) decodeJSONRequest(r *http.Request, req interface{}, w http.ResponseWriter) error {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return err
	}
	return nil
}

func (p *OpenAIProxy) handleResponse(w http.ResponseWriter, result interface{}, err error, operation string) {
	if err != nil {
		log.Printf("%s error: %v", operation, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	p.writeJSONResponse(w, result, operation)
}

func (p *OpenAIProxy) writeJSONResponse(w http.ResponseWriter, data interface{}, responseType string) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Failed to encode %s response: %v", responseType, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (p *OpenAIProxy) normalizeModel(model string) string {
	if strings.HasPrefix(model, "modelplex-") {
		return strings.TrimPrefix(model, "modelplex-")
	}
	return model
}

func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	errorResp := map[string]interface{}{
		"error": map[string]interface{}{
			"message": message,
			"type":    "invalid_request_error",
		},
	}

	if err := json.NewEncoder(w).Encode(errorResp); err != nil {
		log.Printf("Failed to encode error response: %v", err)
	}
}
