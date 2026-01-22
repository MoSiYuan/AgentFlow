# Go 依赖清理说明

## 🎯 目标

清理 `go.mod` 和 `go.sum` 中未使用的依赖。

## 📊 当前状态

### 依赖分析

基于代码分析，当前使用情况：

**✅ 正在使用的依赖（4个）**:
- `github.com/gin-gonic/gin` - HTTP 框架
- `github.com/google/uuid` - UUID 生成
- `github.com/mattn/go-sqlite3` - SQLite 驱动
- `github.com/sirupsen/logrus` - 日志库

**❌ 可能未使用的依赖（1个）**:
- `github.com/stretchr/testify` - 测试框架
  - 说明：项目中没有 `*_test.go` 文件
  - 建议：可以移除，或在未来添加测试时使用

## 🔧 清理方法

### 方法 1: 使用清理脚本（推荐）

```bash
cd golang
./scripts/clean-deps.sh
```

**优点**:
- 自动备份 go.mod
- 显示变更内容
- 安全可靠

### 方法 2: 手动运行 go mod tidy

```bash
cd golang
go mod tidy
```

**优点**:
- Go 官方推荐方法
- 自动处理所有依赖
- 更新 go.sum

### 方法 3: 移除特定依赖

如果只想移除 testify：

```bash
cd golang
go get github.com/stretchr/testify@none
go mod tidy
```

## ⚠️ 注意事项

### 网络要求

`go mod tidy` 需要网络连接来：
- 下载最新的依赖元数据
- 验证依赖完整性
- 更新 go.sum 文件

### 代理设置（中国大陆）

如果网络受限，设置 Go 代理：

```bash
go env -w GOPROXY=https://goproxy.cn,direct
go env -w GOSUMDB=off
```

### 执行清理

```bash
cd golang

# 设置代理（如需要）
export GOPROXY=https://goproxy.cn,direct

# 运行清理
./scripts/clean-deps.sh

# 或手动
go mod tidy
```

## ✅ 验证结果

清理后验证：

```bash
# 1. 查看依赖
cat go.mod

# 2. 检查代码是否能编译
go build ./...

# 3. 查看依赖关系
go list -m all

# 4. 验证 go.sum
go mod verify
```

## 📝 后续维护

### 定期清理

建议在以下情况运行 `go mod tidy`：

- ✅ 提交代码前
- ✅ 发布版本前
- ✅ 重构代码后
- ✅ 删除功能后

### 自动化

添加到 pre-commit hook：

```bash
# .git/hooks/pre-commit
#!/bin/bash
cd golang
go mod tidy
go build ./...
```

## 📚 相关文档

- [依赖分析报告](DEPENDENCY_ANALYSIS.md) - 详细依赖使用情况
- [开发文档](README.md) - Go 版本开发指南
- [清理脚本](../scripts/README.md#clean-depssh) - 脚本使用说明

## 🔄 当前限制

由于网络限制，无法在当前环境自动运行 `go mod tidy`。

**解决方案**:
1. 在能联网的环境中运行 `./scripts/clean-deps.sh`
2. 或设置 Go 代理后重试

---

**创建日期**: 2026-01-22
**Go 版本**: 1.21
**模块**: github.com/jiangxiaolong/agentflow-go
