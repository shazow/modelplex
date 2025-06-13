package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

type Config struct {
	Providers []Provider `toml:"providers"`
	MCP       MCPConfig  `toml:"mcp"`
	Server    Server     `toml:"server"`
}

type Provider struct {
	Name     string `toml:"name"`
	Type     string `toml:"type"`
	BaseURL  string `toml:"base_url"`
	APIKey   string `toml:"api_key"`
	Models   []string `toml:"models"`
	Priority int    `toml:"priority"`
}

type MCPConfig struct {
	Servers []MCPServer `toml:"servers"`
}

type MCPServer struct {
	Name    string `toml:"name"`
	Command string `toml:"command"`
	Args    []string `toml:"args"`
}

type Server struct {
	LogLevel string `toml:"log_level"`
	MaxRequestSize int64 `toml:"max_request_size"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}