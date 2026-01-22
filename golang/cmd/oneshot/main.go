package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jiangxiaolong/agentflow-go/internal/config"
	"github.com/jiangxiaolong/agentflow-go/internal/worker"
)

func main() {
	// Command line flags
	configFile := flag.String("config", "", "Configuration file path")
	masterURL := flag.String("master", "", "Master URL (overrides config)")
	dbPath := flag.String("db", "", "Database path (overrides config)")
	groupName := flag.String("group", "", "Worker group name (overrides config)")
	workerID := flag.String("worker-id", "", "Worker ID (overrides config)")
	timeout := flag.Duration("timeout", 5*time.Minute, "Maximum time to wait for a task")
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
	if *masterURL != "" {
		cfg.Worker.MasterURL = *masterURL
	}
	if *dbPath != "" {
		cfg.Worker.DBPath = *dbPath
	}
	if *groupName != "" {
		cfg.Worker.GroupName = *groupName
	}
	if *workerID != "" {
		cfg.Worker.ID = *workerID
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Create worker
	w, err := worker.New(&worker.Config{
		ID:        cfg.Worker.ID,
		MasterURL: cfg.Worker.MasterURL,
		DBPath:    cfg.Worker.DBPath,
		GroupName: cfg.Worker.GroupName,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create worker: %v\n", err)
		os.Exit(1)
	}

	// Run in one-shot mode
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()

	result, err := w.RunOneShot(ctx)
	if err != nil {
		fmt.Fprintf(os.Stderr, "One-shot execution failed: %v\n", err)
		os.Exit(1)
	}

	if result == "" {
		fmt.Println("No tasks to execute")
		os.Exit(0)
	}

	fmt.Println("Task completed successfully:")
	fmt.Println(result)
}
