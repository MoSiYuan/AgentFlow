# AgentFlow Go - 实用脚本

本目录包含 AgentFlow Go 版本的各种实用脚本。

## 📋 脚本列表

### 依赖管理

#### clean-deps.sh
清理未使用的 Go 依赖。

```bash
./clean-deps.sh
```

**功能**:
- 自动运行 `go mod tidy`
- 备份原始 go.mod
- 显示变更内容
- 清理临时文件

**使用场景**:
- 提交代码前
- 重构代码后
- 发布版本前

**详见**: [依赖分析报告](../docs/DEPENDENCY_ANALYSIS.md)

### 系统管理

#### backup.sh
备份 AgentFlow 数据和配置。

```bash
./backup.sh
```

**功能**:
- 备份 SQLite 数据库
- 备份配置文件
- 生成时间戳备份

### 健康检查

#### health_check.sh
检查 AgentFlow 服务健康状态。

```bash
./health_check.sh
```

**功能**:
- 检查 Master API
- 检查 Worker 状态
- 检查数据库连接
- 显示系统信息

### 示例任务

#### example_tasks.sh
创建示例任务用于测试。

```bash
./example_tasks.sh
```

**功能**:
- 创建测试任务
- 批量任务示例
- 演示 API 使用

## 🚀 快速使用

### 1. 清理依赖

```bash
cd golang
./scripts/clean-deps.sh
```

### 2. 运行健康检查

```bash
# 确保 Master 正在运行
./bin/master --mode standalone --port 6767 &

# 运行健康检查
./scripts/health_check.sh
```

### 3. 创建测试任务

```bash
# 确保 Master 和 Worker 正在运行
./bin/master --mode standalone --port 6767 &
./bin/worker --mode standalone --master http://127.0.0.1:6767 --auto &

# 创建示例任务
./scripts/example_tasks.sh
```

## 🛠️ 脚本开发规范

### 脚本模板

```bash
#!/bin/bash
# script-name.sh - Brief description

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Functions
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Main logic
main() {
    log_info "Starting..."
    # Your code here
    log_info "Done!"
}

main "$@"
```

### 脚本要求

- ✅ 添加 shebang (`#!/bin/bash`)
- ✅ 使用 `set -e` 退出错误
- ✅ 添加功能描述注释
- ✅ 使用彩色输出提升可读性
- ✅ 提供清晰的错误信息
- ✅ 支持 `-h` 或 `--help` 参数

## 📝 添加新脚本

1. 创建脚本文件：
   ```bash
   touch scripts/your-script.sh
   chmod +x scripts/your-script.sh
   ```

2. 添加脚本说明到本 README

3. 遵循脚本开发规范

4. 测试脚本功能

5. 提交到版本控制

## 🔧 故障排查

### 脚本无法执行

```bash
# 添加执行权限
chmod +x scripts/script-name.sh
```

### 找不到命令

```bash
# 使用绝对路径
/path/to/AgentFlow/golang/scripts/script-name.sh

# 或切换到脚本目录
cd /path/to/AgentFlow/golang
./scripts/script-name.sh
```

### 权限错误

```bash
# 使用当前用户
./scripts/script-name.sh

# 或使用 sudo（如需要）
sudo ./scripts/script-name.sh
```

---

**目录**: golang/scripts/
**更新**: 2026-01-22
