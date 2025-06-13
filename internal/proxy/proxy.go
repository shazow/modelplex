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
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	model := p.normalizeModel(req.Model)
	
	result, err := p.mux.ChatCompletion(r.Context(), model, req.Messages)
	if err != nil {
		log.Printf("Chat completion error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (p *OpenAIProxy) HandleCompletions(w http.ResponseWriter, r *http.Request) {
	var req CompletionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	model := p.normalizeModel(req.Model)
	
	result, err := p.mux.Completion(r.Context(), model, req.Prompt)
	if err != nil {
		log.Printf("Completion error: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
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
	
	json.NewEncoder(w).Encode(errorResp)
}