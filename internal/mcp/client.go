// Package mcp provides Model Context Protocol (MCP) client implementation.
package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"sync"

	"github.com/modelplex/modelplex/internal/config"
)

type MCPClient struct {
	servers map[string]*MCPServer
	mu      sync.RWMutex
}

type MCPServer struct {
	name   string
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout io.ReadCloser
	stderr io.ReadCloser
	tools  []Tool
	mu     sync.RWMutex
}

type Tool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	InputSchema map[string]interface{} `json:"inputSchema"`
}

type MCPRequest struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Method  string      `json:"method"`
	Params  interface{} `json:"params,omitempty"`
}

type MCPResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      int         `json:"id"`
	Result  interface{} `json:"result,omitempty"`
	Error   *MCPError   `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewMCPClient(configs []config.MCPServer) *MCPClient {
	client := &MCPClient{
		servers: make(map[string]*MCPServer),
	}

	for _, cfg := range configs {
		if err := client.StartServer(cfg); err != nil {
			slog.Error("Failed to start MCP server", "server", cfg.Name, "error", err)
		}
	}

	return client
}

func (c *MCPClient) StartServer(cfg config.MCPServer) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd := exec.Command(cfg.Command, cfg.Args...) // #nosec G204 -- MCP command execution is intentional from trusted config

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	stdout, err2 := cmd.StdoutPipe()
	if err2 != nil {
		return err2
	}

	stderr, err3 := cmd.StderrPipe()
	if err3 != nil {
		return err3
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	server := &MCPServer{
		name:   cfg.Name,
		cmd:    cmd,
		stdin:  stdin,
		stdout: stdout,
		stderr: stderr,
		tools:  make([]Tool, 0),
	}

	c.servers[cfg.Name] = server

	go server.handleOutput()
	go server.handleErrors()

	if err := server.initialize(); err != nil {
		return err
	}

	return nil
}

func (s *MCPServer) initialize() error {
	initReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities": map[string]interface{}{
				"tools": map[string]interface{}{},
			},
			"clientInfo": map[string]interface{}{
				"name":    "modelplex",
				"version": "0.1.0",
			},
		},
	}

	if err := s.sendRequest(initReq); err != nil {
		return err
	}

	listToolsReq := MCPRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	return s.sendRequest(listToolsReq)
}

func (s *MCPServer) sendRequest(req MCPRequest) error {
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	_, err = s.stdin.Write(append(data, '\n'))
	return err
}

func (s *MCPServer) handleOutput() {
	scanner := bufio.NewScanner(s.stdout)
	for scanner.Scan() {
		line := scanner.Text()

		var resp MCPResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			slog.Error("Failed to parse MCP response", "server", s.name, "error", err)
			continue
		}

		s.handleResponse(resp)
	}
}

func (s *MCPServer) handleErrors() {
	scanner := bufio.NewScanner(s.stderr)
	for scanner.Scan() {
		slog.Warn("MCP server stderr", "server", s.name, "message", scanner.Text())
	}
}

func (s *MCPServer) handleResponse(resp MCPResponse) {
	if resp.Error != nil {
		slog.Error("MCP server error", "server", s.name, "message", resp.Error.Message)
		return
	}

	if resp.ID == 2 {
		if toolsData, ok := resp.Result.(map[string]interface{}); ok {
			if tools, ok := toolsData["tools"].([]interface{}); ok {
				s.mu.Lock()
				s.tools = make([]Tool, 0, len(tools))
				for _, toolData := range tools {
					if toolMap, ok := toolData.(map[string]interface{}); ok {
						tool := Tool{
							Name:        getString(toolMap, "name"),
							Description: getString(toolMap, "description"),
						}
						if schema, ok := toolMap["inputSchema"].(map[string]interface{}); ok {
							tool.InputSchema = schema
						}
						s.tools = append(s.tools, tool)
					}
				}
				s.mu.Unlock()
				slog.Info("MCP server loaded tools", "server", s.name, "count", len(s.tools))
			}
		}
	}
}

func (c *MCPClient) ListTools() []Tool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var allTools []Tool
	for _, server := range c.servers {
		server.mu.RLock()
		allTools = append(allTools, server.tools...)
		server.mu.RUnlock()
	}

	return allTools
}

func (c *MCPClient) CallTool(ctx context.Context, name string, args map[string]interface{}) (interface{}, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	for _, server := range c.servers {
		server.mu.RLock()
		found := false
		for _, tool := range server.tools {
			if tool.Name == name {
				found = true
				break
			}
		}
		server.mu.RUnlock()

		if found {
			return server.callTool(name, args)
		}
	}

	return nil, fmt.Errorf("tool not found: %s", name)
}

func (s *MCPServer) callTool(name string, args map[string]interface{}) (interface{}, error) {
	req := MCPRequest{
		JSONRPC: "2.0",
		ID:      99,
		Method:  "tools/call",
		Params: map[string]interface{}{
			"name":      name,
			"arguments": args,
		},
	}

	if err := s.sendRequest(req); err != nil {
		return nil, err
	}

	return nil, nil
}

func (c *MCPClient) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, server := range c.servers {
		if err := server.stdin.Close(); err != nil {
			slog.Error("Error closing MCP server stdin", "error", err)
		}
		if err := server.cmd.Process.Kill(); err != nil {
			slog.Error("Error killing MCP server process", "error", err)
		}
		if err := server.cmd.Wait(); err != nil {
			slog.Error("Error waiting for MCP server process", "error", err)
		}
	}
}

func getString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}
