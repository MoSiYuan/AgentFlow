#!/bin/bash
set -e

echo "编译 AgentFlow Go 多平台版本..."

# macOS ARM64 (当前平台)
echo "编译: darwin-arm64"
GOOS=darwin GOARCH=arm64 go build -o ../build-release/agentflow-master-darwin-arm64 ./cmd/master/
GOOS=darwin GOARCH=arm64 go build -o ../build-release/agentflow-worker-darwin-arm64 ./cmd/worker/

# macOS AMD64
echo "编译: darwin-amd64"
GOOS=darwin GOARCH=amd64 go build -o ../build-release/agentflow-master-darwin-amd64 ./cmd/master/
GOOS=darwin GOARCH=amd64 go build -o ../build-release/agentflow-worker-darwin-amd64 ./cmd/worker/

# Linux AMD64
echo "编译: linux-amd64"
GOOS=linux GOARCH=amd64 go build -o ../build-release/agentflow-master-linux-amd64 ./cmd/master/
GOOS=linux GOARCH=amd64 go build -o ../build-release/agentflow-worker-linux-amd64 ./cmd/worker/

# Linux ARM64
echo "编译: linux-arm64"
GOOS=linux GOARCH=arm64 go build -o ../build-release/agentflow-master-linux-arm64 ./cmd/master/
GOOS=linux GOARCH=arm64 go build -o ../build-release/agentflow-worker-linux-arm64 ./cmd/worker/

# Windows AMD64
echo "编译: windows-amd64"
GOOS=windows GOARCH=amd64 go build -o ../build-release/agentflow-master-windows-amd64.exe ./cmd/master/
GOOS=windows GOARCH=amd64 go build -o ../build-release/agentflow-worker-windows-amd64.exe ./cmd/worker/

echo ""
echo "编译完成！"
ls -lh ../build-release/
