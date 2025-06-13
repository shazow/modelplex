package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelplex/modelplex/internal/config"
	"github.com/modelplex/modelplex/internal/server"
)

func main() {
	var (
		configPath = flag.String("config", "config.toml", "Path to configuration file")
		socketPath = flag.String("socket", "./modelplex.socket", "Path to Unix socket")
	)
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	srv := server.New(cfg, *socketPath)
	
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Println("Shutting down...")
	srv.Stop()
}