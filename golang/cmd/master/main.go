package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jiangxiaolong/agentflow-go/internal/config"
	"github.com/jiangxiaolong/agentflow-go/internal/master"
)

func main() {
	// Command line flags
	configFile := flag.String("config", "", "Configuration file path")
	host := flag.String("host", "", "Host to bind to (overrides config)")
	port := flag.Int("port", 0, "Port to listen on (overrides config)")
	dbPath := flag.String("db", "", "Database path (overrides config)")
	flag.Parse()

	// Load configuration
	var cfg *config.Config
	var err error

	if *configFile != "" {
		cfg, err = config.Load(*configFile)
	} else {
		cfg = config.DefaultConfig()
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Override with command line flags
	if *host != "" {
		cfg.Master.Host = *host
	}
	if *port != 0 {
		cfg.Master.Port = *port
	}
	if *dbPath != "" {
		cfg.Master.DBPath = *dbPath
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create master
	m, err := master.New(&master.Config{
		DBPath:    cfg.Master.DBPath,
		Host:      cfg.Master.Host,
		Port:      cfg.Master.Port,
		AutoStart: cfg.Master.AutoStart,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create master: %v\n", err)
		os.Exit(1)
	}

	// Start master
	addr := fmt.Sprintf("%s:%d", cfg.Master.Host, cfg.Master.Port)
	if err := m.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start master: %v\n", err)
		os.Exit(1)
	}
}
