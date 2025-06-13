// Package config provides TOML configuration loading and parsing.
package config

import (
	"os"

	"github.com/pelletier/go-toml/v2"
)

// Config represents the main configuration structure for modelplex.
type Config struct {
	Providers []Provider `toml:"providers"`
	MCP       MCPConfig  `toml:"mcp"`
	Server    Server     `toml:"server"`
}

// Provider represents configuration for an AI provider.
type Provider struct {
	Name     string   `toml:"name"`
	Type     string   `toml:"type"`
	BaseURL  string   `toml:"base_url"`
	APIKey   string   `toml:"api_key"`
	Models   []string `toml:"models"`
	Priority int      `toml:"priority"`
}

// MCPConfig represents MCP (Model Context Protocol) configuration.
type MCPConfig struct {
	Servers []MCPServer `toml:"servers"`
}

// MCPServer represents configuration for a single MCP server.
type MCPServer struct {
	Name    string   `toml:"name"`
	Command string   `toml:"command"`
	Args    []string `toml:"args"`
}

// Server represents HTTP server configuration.
type Server struct {
	LogLevel       string `toml:"log_level"`
	MaxRequestSize int64  `toml:"max_request_size"`
}

// Load reads and parses a TOML configuration file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- config file path is provided by user via CLI flag
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
