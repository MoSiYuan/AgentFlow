package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jiangxiaolong/agentflow-go/internal/worker"
)

func main() {
	// Command line flags
	masterURL := flag.String("master", "http://localhost:8848", "Master URL")
	dbPath := flag.String("db", ".claude/cpds-manager/agentflow.db", "Database path")
	groupName := flag.String("group", "default", "Worker group name")
	flag.Parse()

	// Create worker
	w, err := worker.New(&worker.Config{
		MasterURL: *masterURL,
		DBPath:    *dbPath,
		GroupName: *groupName,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create worker: %v\n", err)
		os.Exit(1)
	}

	// Run worker
	ctx := context.Background()
	if err := w.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Worker error: %v\n", err)
		os.Exit(1)
	}
}
