package providers

import (
	"context"

	"github.com/modelplex/modelplex/internal/config"
)

type Provider interface {
	Name() string
	Priority() int
	ChatCompletion(ctx context.Context, model string, messages []map[string]interface{}) (interface{}, error)
	Completion(ctx context.Context, model string, prompt string) (interface{}, error)
	ListModels() []string
}

func NewProvider(cfg config.Provider) Provider {
	switch cfg.Type {
	case "openai":
		return NewOpenAIProvider(cfg)
	case "anthropic":
		return NewAnthropicProvider(cfg)
	case "ollama":
		return NewOllamaProvider(cfg)
	default:
		return nil
	}
}
