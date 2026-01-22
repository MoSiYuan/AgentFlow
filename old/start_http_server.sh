#!/bin/bash
# CPDS HTTP Server启动脚本 (简化版)

PORT=8849  # Claude server uses port 8849 (Master uses 8848)

echo "═══════════════════════════════════════════════════════════"
echo "     🚀 CPDS Claude HTTP Server"
echo "═══════════════════════════════════════════════════════════"
echo ""

# 清理旧进程
PID=$(lsof -ti :$PORT 2>/dev/null || echo "")
if [ -n "$PID" ]; then
  echo "⚠️  检测到旧服务器(PID: $PID)，正在关闭..."
  kill $PID 2>/dev/null || true
  sleep 1
fi

# 启动Master (如果还没运行)
if ! curl -s http://localhost:8848/api/health | grep -q "healthy"; then
  echo "📡 启动Master服务器..."
  ./cpds/cpds master --mode standalone --auto-shutdown --port 8848 > /tmp/cpds_master.log 2>&1 &
  sleep 3
fi

# 启动Claude HTTP server
echo "🤖 启动Claude HTTP服务器 (端口: $PORT)..."

# 使用Go内置HTTP服务器
cat > /tmp/claude_server.go <<'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
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

	log.Println("Claude server started on port", $PORT)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", $PORT), nil))
}

func callClaude(prompt string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	log.Println("Calling claude CLI...")

	cmd := exec.Command("claude", "-p")
	cmd.Stdin = strings.NewReader(prompt)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("claude failed: %w\nOutput: %s", err, string(output))
	}

	return string(output), nil
}
EOF

# 编译并启动服务器
echo "编译并启动..."
go run /tmp/claude_server.go > /tmp/cpds_claude_server.log 2>&1 &
CLAUDE_PID=$!

echo "✅ Claude HTTP服务器已启动 (PID: $CLAUDE_PID)"
echo ""
echo "═══════════════════════════════════════════════════════════"
echo "                    服务器信息"
echo "═══════════════════════════════════════════════════════════"
echo "Claude服务器: http://localhost:$PORT"
echo "Master服务器: http://localhost:8848"
echo "Claude PID:    $CLAUDE_PID"
echo ""
echo "🛑 停止服务器:"
echo "   kill $CLAUDE_PID"
echo ""
echo "📝 查看日志:"
echo "   tail -f /tmp/cpds_claude_server.log"
echo "═══════════════════════════════════════════════════════════"
