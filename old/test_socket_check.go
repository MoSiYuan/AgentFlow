package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	socketPath := "/tmp/cpds-claude.sock"

	fmt.Println("Testing socket connection...")

	conn, err := net.DialTimeout("unix", socketPath, 1*time.Second)
	if err != nil {
		fmt.Printf("❌ Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Println("✅ Socket connection successful!")

	// Test write
	_, err = conn.Write([]byte("REQUEST\ntest\nEND_REQUEST\n"))
	if err != nil {
		fmt.Printf("❌ Failed to write: %v\n", err)
		return
	}

	fmt.Println("✅ Write successful!")

	// Test read
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Printf("❌ Failed to read: %v\n", err)
		return
	}

	fmt.Printf("✅ Read successful: %d bytes\n", n)
	fmt.Printf("Response: %s\n", string(buf[:n]))
}
