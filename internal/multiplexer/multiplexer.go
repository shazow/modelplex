package multiplexer

import (
	"context"
	"fmt"
	"sort"

	"github.com/modelplex/modelplex/internal/config"
	"github.com/modelplex/modelplex/internal/providers"
)

type ModelMultiplexer struct {
	providers []providers.Provider
	modelMap  map[string]providers.Provider
}

func New(configs []config.Provider) *ModelMultiplexer {
	m := &ModelMultiplexer{
		providers: make([]providers.Provider, 0),
		modelMap:  make(map[string]providers.Provider),
	}

	for _, cfg := range configs {
		provider := providers.NewProvider(cfg)
		if provider != nil {
			m.providers = append(m.providers, provider)
			
			for _, model := range cfg.Models {
				if _, exists := m.modelMap[model]; !exists {
					m.modelMap[model] = provider
				}
			}
		}
	}

	sort.Slice(m.providers, func(i, j int) bool {
		return m.providers[i].Priority() < m.providers[j].Priority()
	})

	return m
}

func (m *ModelMultiplexer) GetProvider(model string) (providers.Provider, error) {
	if provider, exists := m.modelMap[model]; exists {
		return provider, nil
	}
	
	if len(m.providers) > 0 {
		return m.providers[0], nil
	}
	
	return nil, fmt.Errorf("no provider available for model: %s", model)
}

func (m *ModelMultiplexer) ListModels() []string {
	models := make([]string, 0, len(m.modelMap))
	for model := range m.modelMap {
		models = append(models, model)
	}
	return models
}

func (m *ModelMultiplexer) ChatCompletion(ctx context.Context, model string, messages []map[string]interface{}) (interface{}, error) {
	provider, err := m.GetProvider(model)
	if err != nil {
		return nil, err
	}
	
	return provider.ChatCompletion(ctx, model, messages)
}

func (m *ModelMultiplexer) Completion(ctx context.Context, model string, prompt string) (interface{}, error) {
	provider, err := m.GetProvider(model)
	if err != nil {
		return nil, err
	}
	
	return provider.Completion(ctx, model, prompt)
}