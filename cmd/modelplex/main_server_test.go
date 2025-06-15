package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/modelplex/modelplex/internal/config"
	"github.com/modelplex/modelplex/internal/server"
)

func TestHTTPServerByDefault(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:     "test",
				Type:     "openai",
				BaseURL:  "http://localhost:8080",
				APIKey:   "test-key",
				Models:   []string{"test-model"},
				Priority: 1,
			},
		},
		Server: config.Server{
			LogLevel:       "info",
			MaxRequestSize: 1024,
		},
	}

	// Test HTTP server creation
	srv := server.NewWithHTTP(cfg, "127.0.0.1", 0) // Use port 0 to get a random available port
	
	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start HTTP server: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Stop server
	srv.Stop()
}

func TestSocketServerWhenSpecified(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:     "test",
				Type:     "openai",
				BaseURL:  "http://localhost:8080",
				APIKey:   "test-key",
				Models:   []string{"test-model"},
				Priority: 1,
			},
		},
		Server: config.Server{
			LogLevel:       "info",
			MaxRequestSize: 1024,
		},
	}

	// Test socket server creation
	socketPath := "/tmp/test-modelplex.socket"
	srv := server.NewWithSocket(cfg, socketPath)
	
	// Start server in goroutine
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start socket server: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(100 * time.Millisecond)
	
	// Check if socket file exists
	if _, err := os.Stat(socketPath); os.IsNotExist(err) {
		t.Errorf("Socket file was not created: %s", socketPath)
	}
	
	// Stop server
	srv.Stop()
	
	// Check if socket file was cleaned up
	if _, err := os.Stat(socketPath); !os.IsNotExist(err) {
		t.Errorf("Socket file was not cleaned up: %s", socketPath)
	}
}

func TestEndpointStructure(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:     "test",
				Type:     "openai",
				BaseURL:  "http://localhost:8080",
				APIKey:   "test-key",
				Models:   []string{"test-model"},
				Priority: 1,
			},
		},
		Server: config.Server{
			LogLevel:       "info",
			MaxRequestSize: 1024,
		},
	}

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Start HTTP server
	srv := server.NewWithHTTP(cfg, "127.0.0.1", port)
	
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start HTTP server: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(200 * time.Millisecond)
	
	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)
	
	// Test endpoints
	tests := []struct {
		name     string
		endpoint string
		expected int
	}{
		{"Health check", "/health", http.StatusOK},
		{"OpenAI models endpoint", "/models/v1/models", http.StatusOK},
		{"MCP tools endpoint", "/mcp/v1/tools", http.StatusOK},
		{"Internal status endpoint", "/_internal/status", http.StatusOK},
		{"Internal config endpoint", "/_internal/config", http.StatusOK},
		{"Internal metrics endpoint", "/_internal/metrics", http.StatusOK},
		{"Backward compatibility - old models endpoint", "/v1/models", http.StatusOK},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			
			req, err := http.NewRequestWithContext(ctx, "GET", baseURL+test.endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			client := &http.Client{Timeout: 5 * time.Second}
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", test.endpoint, err)
			}
			defer resp.Body.Close()
			
			if resp.StatusCode != test.expected {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status %d for %s, got %d. Body: %s", 
					test.expected, test.endpoint, resp.StatusCode, string(body))
			}
		})
	}
	
	// Stop server
	srv.Stop()
}

func TestInternalEndpointsNotAvailableOnSocket(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Providers: []config.Provider{
			{
				Name:     "test",
				Type:     "openai",
				BaseURL:  "http://localhost:8080",
				APIKey:   "test-key",
				Models:   []string{"test-model"},
				Priority: 1,
			},
		},
		Server: config.Server{
			LogLevel:       "info",
			MaxRequestSize: 1024,
		},
	}

	// Test that socket server doesn't expose internal endpoints
	socketPath := "/tmp/test-modelplex-internal.socket"
	srv := server.NewWithSocket(cfg, socketPath)
	
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start socket server: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(200 * time.Millisecond)
	
	// Create HTTP client for Unix socket
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.Dial("unix", socketPath)
			},
		},
		Timeout: 5 * time.Second,
	}
	
	// Test that internal endpoints return 404 on socket
	internalEndpoints := []string{
		"/_internal/status",
		"/_internal/config", 
		"/_internal/metrics",
	}
	
	for _, endpoint := range internalEndpoints {
		t.Run(fmt.Sprintf("Internal endpoint %s not available on socket", endpoint), func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://unix"+endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", endpoint, err)
			}
			defer resp.Body.Close()
			
			// Should return 404 since internal endpoints are not available on socket
			if resp.StatusCode != http.StatusNotFound {
				t.Errorf("Expected status 404 for %s on socket, got %d", endpoint, resp.StatusCode)
			}
		})
	}
	
	// Test that regular endpoints work on socket
	regularEndpoints := []string{
		"/health",
		"/models/v1/models",
		"/mcp/v1/tools",
	}
	
	for _, endpoint := range regularEndpoints {
		t.Run(fmt.Sprintf("Regular endpoint %s available on socket", endpoint), func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://unix"+endpoint, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			
			resp, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to make request to %s: %v", endpoint, err)
			}
			defer resp.Body.Close()
			
			// Should return 200 OK
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				t.Errorf("Expected status 200 for %s on socket, got %d. Body: %s", 
					endpoint, resp.StatusCode, string(body))
			}
		})
	}
	
	// Stop server
	srv.Stop()
}

func TestInternalStatusEndpoint(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		Providers: []config.Provider{
			{Name: "test1", Type: "openai"},
			{Name: "test2", Type: "anthropic"},
		},
		MCP: config.MCPConfig{
			Servers: []config.MCPServer{
				{Name: "server1", Command: "test"},
			},
		},
		Server: config.Server{
			LogLevel:       "info",
			MaxRequestSize: 1024,
		},
	}

	// Find an available port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("Failed to get available port: %v", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	listener.Close()

	// Start HTTP server
	srv := server.NewWithHTTP(cfg, "127.0.0.1", port)
	
	go func() {
		if err := srv.Start(); err != nil && err != http.ErrServerClosed {
			t.Errorf("Failed to start HTTP server: %v", err)
		}
	}()
	
	// Give server time to start
	time.Sleep(200 * time.Millisecond)
	
	// Test internal status endpoint
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/_internal/status", port))
	if err != nil {
		t.Fatalf("Failed to get status: %v", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status 200, got %d", resp.StatusCode)
	}
	
	var status map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		t.Fatalf("Failed to decode status response: %v", err)
	}
	
	// Verify status content
	expectedFields := []string{"service", "status", "mode", "host", "port", "providers", "mcp_servers"}
	for _, field := range expectedFields {
		if _, exists := status[field]; !exists {
			t.Errorf("Expected field %s in status response", field)
		}
	}
	
	if status["service"] != "modelplex" {
		t.Errorf("Expected service=modelplex, got %v", status["service"])
	}
	
	if status["providers"] != float64(2) { // JSON numbers are float64
		t.Errorf("Expected 2 providers, got %v", status["providers"])
	}
	
	if status["mcp_servers"] != float64(1) {
		t.Errorf("Expected 1 MCP server, got %v", status["mcp_servers"])
	}
	
	// Stop server
	srv.Stop()
}
