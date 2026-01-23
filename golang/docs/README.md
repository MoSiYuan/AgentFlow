# Go 版本 - 开发文档

## 📚 文档目录

- [依赖分析报告](DEPENDENCY_ANALYSIS.md) - 当前依赖使用情况分析
- [清理脚本说明](../scripts/README.md#clean-depssh) - 依赖清理脚本使用

## 🔧 常用命令

### 依赖管理

```bash
# 清理未使用的依赖
go mod tidy

# 下载依赖
go mod download

# 验证依赖
go mod verify

# 查看依赖关系
go list -m all

# 查看特定依赖
go mod why <package>
```

### 构建和测试

```bash
# 编译所有包
go build ./...

# 编译 Master
go build -o ../bin/master ./cmd/master

# 编译 Worker
go build -o ../bin/worker ./cmd/worker

# 运行测试
go test ./...

# 运行测试并查看覆盖率
go test -cover ./...
```

### 代码质量

```bash
# 格式化代码
go fmt ./...

# 静态检查
go vet ./...

# 使用 golangci-lint（需安装）
golangci-lint run
```

## 🛠️ 脚本工具

项目根目录提供了一些实用脚本：

- `scripts/clean-deps.sh` - 清理未使用的 Go 依赖

使用方法：

```bash
cd golang
./scripts/clean-deps.sh
```

## 📦 当前依赖状态

### 直接依赖（4个）

- ✅ `github.com/gin-gonic/gin` - HTTP 框架
- ✅ `github.com/google/uuid` - UUID 生成
- ✅ `github.com/mattn/go-sqlite3` - SQLite 驱动
- ✅ `github.com/sirupsen/logrus` - 日志库

### 可能未使用（1个）

- ⚠️ `github.com/stretchr/testify` - 测试框架（项目无测试文件）

**建议**: 运行 `go mod tidy` 移除未使用的依赖

详见：[依赖分析报告](DEPENDENCY_ANALYSIS.md)

## 🚀 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 编译

```bash
# 编译 Master
go build -o bin/master cmd/master/main.go

# 编译 Worker
go build -o bin/worker cmd/worker/main.go
```

### 3. 运行

```bash
# 启动 Master
./bin/master --mode standalone --port 6767

# 启动 Worker
./bin/worker --mode standalone --master http://127.0.0.1:6767 --auto
```

## 📝 开发规范

### 代码风格

- 遵循 [Effective Go](https://golang.org/doc/effective_go) 指南
- 使用 `gofmt` 格式化代码
- 添加必要的注释和文档
- 导出函数和类型添加文档注释

### 提交前检查

```bash
# 1. 格式化代码
go fmt ./...

# 2. 静态检查
go vet ./...

# 3. 清理依赖
go mod tidy

# 4. 运行测试（如果有）
go test ./...

# 5. 编译检查
go build ./...
```

## 🔍 故障排查

### 依赖下载失败

```bash
# 设置代理（中国大陆）
go env -w GOPROXY=https://goproxy.cn,direct

# 重试下载
go mod download
```

### 编译错误

```bash
# 清理缓存
go clean -cache

# 重新下载依赖
go mod download

# 重新编译
go build ./...
```

### go.mod 不同步

```bash
# 同步 go.mod 和 go.sum
go mod tidy

# 验证依赖
go mod verify
```

---

**更新**: 2026-01-22
**Go 版本**: 1.21
