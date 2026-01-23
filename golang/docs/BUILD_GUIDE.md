# AgentFlow Golang 版本 - 编译指南

## 📋 前置要求

### 必需
- Go 1.21+ (`go version`)
- Make (可选) 或 Go 编译器

### 可选
- China Go Proxy（如果在国内）

## 🚀 快速编译

### 方式 1: 使用 Make（推荐）

```bash
cd /path/to/AgentFlow/golang

# 编译所有二进制
make build

# 只编译 Master
make master

# 只编译 Worker
make worker
```

### 方式 2: 使用 go build

```bash
cd /path/to/AgentFlow/golang

# 编译 Master
go build -o bin/master cmd/master/main.go

# 编译 Worker
go build -o bin/worker cmd/worker/main.go
```

### 方式 3: 设置国内代理后编译

```bash
# 1. 设置代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 2. 下载依赖
go mod download

# 3. 编译
go build -o bin/master cmd/master/main.go
go build -o bin/worker cmd/worker/main.go
```

## 🔧 编译脚本

### 自动编译脚本

```bash
#!/bin/bash
# build.sh - AgentFlow 编译脚本

set -e

echo "🔨 Building AgentFlow..."
echo ""

# 设置代理（中国）
export GOPROXY=https://goproxy.cn,direct
export GOSUMDB=off

# 编译 Master
echo "Building Master..."
go build -o bin/master cmd/master/main.go

# 编译 Worker
echo "Building Worker..."
go build -o bin/worker cmd/worker/main.go

# 验证
echo ""
echo "✅ Build complete!"
ls -lh bin/master bin/worker
```

使用方法：
```bash
chmod +x build.sh
./build.sh
```

## ✅ 验证编译结果

```bash
# 检查二进制文件
ls -lh bin/
# 应该看到:
# master (约 10-20 MB)
# worker (约 10-20 MB)

# 测试运行
./bin/master --help
./bin/worker --help
```

## 📦 编译产物

编译成功后，`bin/` 目录应包含：

```
bin/
├── master     # Master 服务器
└── worker     # Worker 进程
```

## 🐳 Docker 编译

如果需要在 Docker 中编译：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .

# 设置代理
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

# 下载依赖
RUN go mod download

# 编译
RUN go build -o bin/master cmd/master/main.go
RUN go build -o bin/worker cmd/worker/main.go

# 最终镜像
FROM alpine:latest
COPY --from=builder /app/bin/ /usr/local/bin/
```

## 🔍 故障排查

### 问题 1: 依赖下载失败

**错误**: `go: github.com/xxx: unreachable`

**解决方案**:
```bash
# 设置中国代理
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off

# 重试下载
go mod download
```

### 问题 2: 编译错误

**错误**: `cannot find package`

**解决方案**:
```bash
# 清理缓存
go clean -cache
go mod tidy

# 重新下载
go mod download

# 重新编译
go build -o bin/master cmd/master/main.go
```

### 问题 3: CGO 错误（SQLite）

**错误**: `gcc: command not found`

**解决方案**:

**macOS**:
```bash
xcode-select --install
```

**Ubuntu/Debian**:
```bash
sudo apt-get install build-essential
```

**Alpine Linux**:
```bash
apk add gcc musl-dev
```

### 问题 4: 平台特定编译

**交叉编译到其他平台**:

```bash
# Linux (amd64)
GOOS=linux GOARCH=amd64 go build -o bin/master-linux cmd/master/main.go

# Windows (amd64)
GOOS=windows GOARCH=amd64 go build -o bin/master.exe cmd/master/main.go

# macOS (Apple Silicon)
GOOS=darwin GOARCH=arm64 go build -o bin/master-arm64 cmd/master/main.go
```

## 📝 编译选项

### 减小二进制大小

```bash
# 去除调试信息
go build -ldflags="-s -w" -o bin/master cmd/master/main.go

# 使用 upx 压缩（需要安装 upx）
upx --best --lzma bin/master
```

### 添加版本信息

```bash
VERSION=$(git describe --tags --always)
go build \
  -ldflags="-X main.version=$VERSION" \
  -o bin/master cmd/master/main.go
```

## 🎯 Makefile（推荐）

创建 `Makefile`:

```makefile
.PHONY: all master worker clean

all: master worker

master:
	go build -o bin/master cmd/master/main.go

worker:
	go build -o bin/worker cmd/worker/main.go

clean:
	rm -f bin/master bin/worker

install: all
	install -m 0755 bin/master /usr/local/bin/
	install -m 0755 bin/worker /usr/local/bin/

test:
	go test ./...
```

使用：
```bash
make         # 编译所有
make master  # 只编译 Master
make worker  # 只编译 Worker
make clean   # 清理
```

## ✅ 编译检查清单

- [ ] Go 版本 >= 1.21
- [ ] 依赖已下载 (`go mod download`)
- [ ] 编译无错误
- [ ] 二进制文件可执行
- [ ] `--help` 参数正常
- [ ] 文件大小合理 (10-20 MB)

---

**更新**: 2026-01-22
**Go 版本**: 1.21
