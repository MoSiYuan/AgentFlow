package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/jiangxiaolong/agentflow-go/internal/worker"
)

func main() {
	// Command line flags
	masterURL := flag.String("master", "http://localhost:8848", "Master URL")
	dbPath := flag.String("db", ".claude/cpds-manager/agentflow.db", "Database path")
	groupName := flag.String("group", "default", "Worker group name")
	workerID := flag.String("worker-id", "", "Worker ID (auto-generated if empty)")
	timeout := flag.Duration("timeout", 5*time.Minute, "Maximum time to wait for a task")
	flag.Parse()

	// Create worker
	w, err := worker.New(&worker.Config{
		ID:        *workerID,
		MasterURL: *masterURL,
		DBPath:    *dbPath,
		GroupName: *groupName,
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
