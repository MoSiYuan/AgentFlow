.PHONY: all build test fmt lint clean docker docker-build docker-run docker-stop

all: fmt build

build:
	@echo "构建..."
	go build -v -o bin/agentflow ./cmd/agentflow
	go build -v -o bin/master ./cmd/master
	go build -v -o bin/worker ./cmd/worker

fmt:
	@echo "格式化..."
	goimports -w .
	gofmt -s -w .

lint:
	@echo "代码检查..."
	golangci-lint run ./...

test:
	@echo "测试..."
	go test -v ./...

clean:
	@echo "清理..."
	rm -rf bin/
	rm -f agentflow.db

# Docker targets
docker: docker-build

docker-build:
	@echo "构建 Docker 镜像..."
	docker build -t agentflow-go:latest .

docker-run:
	@echo "启动 Docker 容器..."
	docker-compose up -d

docker-stop:
	@echo "停止 Docker 容器..."
	docker-compose down

docker-logs:
	@echo "查看 Docker 日志..."
	docker-compose logs -f

# Dev targets
dev-init:
	@echo "初始化开发环境..."
	go mod download
	go mod tidy

dev-run:
	@echo "启动开发环境..."
	./bin/agentflow init agentflow.db
	./bin/agentflow master --db agentflow.db

# Test targets
test-coverage:
	@echo "运行测试覆盖率..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Release targets
release: clean fmt test
	@echo "构建发布版本..."
	GOOS=linux GOARCH=amd64 go build -o bin/agentflow-linux-amd64 ./cmd/agentflow
	GOOS=windows GOARCH=amd64 go build -o bin/agentflow-windows-amd64.exe ./cmd/agentflow
	GOOS=darwin GOARCH=amd64 go build -o bin/agentflow-darwin-amd64 ./cmd/agentflow
