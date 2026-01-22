package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jiangxiaolong/agentflow-go/internal/master"
)

func main() {
	// Command line flags
	host := flag.String("host", "0.0.0.0", "Host to bind to")
	port := flag.Int("port", 8848, "Port to listen on")
	dbPath := flag.String("db", ".claude/cpds-manager/agentflow.db", "Database path")
	flag.Parse()

	// Create master
	m, err := master.New(&master.Config{
		DBPath: *dbPath,
		Host:   *host,
		Port:   *port,
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create master: %v\n", err)
		os.Exit(1)
	}

	// Start master
	addr := fmt.Sprintf("%s:%d", *host, *port)
	if err := m.Run(addr); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start master: %v\n", err)
		os.Exit(1)
	}
}
