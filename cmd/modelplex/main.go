package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jessevdk/go-flags"
	"github.com/modelplex/modelplex/internal/config"
	"github.com/modelplex/modelplex/internal/server"
)

// Options defines command line options
type Options struct {
	Config  string `short:"c" long:"config" default:"config.toml" description:"Path to configuration file"`
	Socket  string `short:"s" long:"socket" default:"./modelplex.socket" description:"Path to Unix socket"`
	Verbose bool   `short:"v" long:"verbose" description:"Enable verbose logging"`
	Version bool   `long:"version" description:"Show version information"`
}

var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	var opts Options
	parser := flags.NewParser(&opts, flags.Default)
	parser.Name = "modelplex"
	parser.Usage = "[OPTIONS]"

	_, err := parser.Parse()
	if err != nil {
		if flagsErr, ok := err.(*flags.Error); ok && flagsErr.Type == flags.ErrHelp {
			os.Exit(0)
		}
		os.Exit(1)
	}

	if opts.Version {
		fmt.Printf("modelplex %s\n", version)
		fmt.Printf("commit: %s\n", commit)
		fmt.Printf("built: %s\n", date)
		os.Exit(0)
	}

	if opts.Verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.Println("Verbose logging enabled")
	}

	cfg, err := config.Load(opts.Config)
	if err != nil {
		log.Fatalf("Failed to load config from %s: %v", opts.Config, err)
	}

	if opts.Verbose {
		log.Printf("Loaded configuration from %s", opts.Config)
		log.Printf("Starting server on socket: %s", opts.Socket)
	}

	srv := server.New(cfg, opts.Socket)
	
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