package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"sync"
)

var mu sync.Mutex

func main() {
	// Claude server endpoint
	http.HandleFunc("/execute", func(w http.ResponseWriter, r *http.Request) {
		log.Println("Received execution request")

		var req struct {
			TaskID      string `json:"task_id"`
			Title       string `json:"title"`
			Description string `json:"description"`
			WorkerID    string `json:"worker_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Printf("Error decoding request: %v", err)
			http.Error(w, err.Error(), 400)
			return
		}

		log.Printf("Executing task %s", req.TaskID)

		// Call claude CLI
		output, err := callClaude(req.Description)
		if err != nil {
			log.Printf("Error: %v", err)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"error":   err.Error(),
			})
			return
		}

		// Send response
		json.NewEncoder(w).Encode(map[string]interface{}{
			"output":      output,
			"success":     true,
			"tokens_used": len(output) / 2,
		})

		log.Printf("Task %s completed (%d bytes)", req.TaskID, len(output))
	})

	// Health check
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "healthy",
			"service": "claude-server",
		})
	})

	log.Println("Claude server started on port 8849")
	log.Fatal(http.ListenAndServe(":8849", nil))
}

func callClaude(prompt string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	log.Println("Calling claude CLI...")

	// Pass prompt as argument instead of stdin
	cmd := exec.Command("claude", "-p", prompt)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
