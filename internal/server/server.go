// Package server provides HTTP server functionality over Unix domain sockets.
package server

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/modelplex/modelplex/internal/config"
	"github.com/modelplex/modelplex/internal/multiplexer"
	"github.com/modelplex/modelplex/internal/proxy"
)

type Server struct {
	config     *config.Config
	socketPath string
	listener   net.Listener
	server     *http.Server
	mux        *multiplexer.ModelMultiplexer
	proxy      *proxy.OpenAIProxy
}

func New(cfg *config.Config, socketPath string) *Server {
	mux := multiplexer.New(cfg.Providers)
	proxy := proxy.New(mux)

	return &Server{
		config:     cfg,
		socketPath: socketPath,
		mux:        mux,
		proxy:      proxy,
	}
}

func (s *Server) Start() error {
	if err := os.RemoveAll(s.socketPath); err != nil {
		return err
	}

	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		return err
	}
	s.listener = listener

	router := mux.NewRouter()
	s.setupRoutes(router)

	s.server = &http.Server{
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	slog.Info("Modelplex server listening", "socket", s.socketPath)
	return s.server.Serve(listener)
}

func (s *Server) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.server.Shutdown(ctx); err != nil {
			slog.Error("Error shutting down server", "error", err)
		}
	}
	if s.listener != nil {
		if err := s.listener.Close(); err != nil {
			slog.Error("Error closing listener", "error", err)
		}
	}
	if err := os.RemoveAll(s.socketPath); err != nil {
		slog.Error("Error removing socket path", "path", s.socketPath, "error", err)
	}
}

func (s *Server) setupRoutes(router *mux.Router) {
	v1 := router.PathPrefix("/v1").Subrouter()

	// OpenAI-compatible endpoints
	v1.HandleFunc("/chat/completions", s.proxy.HandleChatCompletions).Methods("POST")
	v1.HandleFunc("/completions", s.proxy.HandleCompletions).Methods("POST")
	v1.HandleFunc("/models", s.proxy.HandleModels).Methods("GET")

	// Health check
	router.HandleFunc("/health", s.handleHealth).Methods("GET")
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write([]byte(`{"status":"ok","service":"modelplex"}`)); err != nil {
		slog.Error("Error writing health response", "error", err)
	}
}
