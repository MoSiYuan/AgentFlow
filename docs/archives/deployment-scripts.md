# AgentFlow 部署脚本使用指南

## 📋 概述

AgentFlow 提供三个平台的自动化部署脚本，用于快速搭建 Claude 开发环境：

| 平台 | 脚本文件 | 支持命令 |
|------|---------|---------|
| **macOS** | `scripts/deploy-macos.sh` | check, install, build, deploy, start, stop, status |
| **Linux** | `scripts/deploy-linux.sh` | check, install, build, deploy, start, stop, status |
| **Windows** | `scripts/deploy-windows.bat` | check, install, build, deploy, start, stop, status |

---

## 🍎 macOS 部署

### 前置要求

- macOS 10.15 (Catalina) 或更高版本
- 管理员权限（用于安装 Homebrew 和系统服务）
- Xcode Command Line Tools

### 使用步骤

#### 1. 检查环境

```bash
./scripts/deploy-macos.sh check
```

**输出示例**:
```
✅ macOS Version: 14.0
✅ Homebrew: Installed
✅ Python: 3.11.5
✅ Go: 1.21.3
⚠️  Claude CLI: Not installed
```

#### 2. 安装依赖

```bash
./scripts/deploy-macos.sh install
```

**自动安装内容**:
- Homebrew（如果未安装）
- Python 3.x 和 pip
- Go 1.21+
- Claude CLI（通过 npm）
- Python 依赖包（Flask, requests 等）

#### 3. 编译 Go 版本

```bash
./scripts/deploy-macos.sh build
```

**编译产物**:
```
golang/bin/master       - Master 服务器
golang/bin/worker       - Worker（持续模式）
golang/bin/oneshot      - Worker（单次模式）
```

#### 4. 部署配置

```bash
./scripts/deploy-macos.sh deploy
```

**生成文件**:
- `.agentflow/boundaries.json` - 文件边界配置
- `golang/config.yaml` - Go 版本配置文件
- `.claude/cpds-manager/` - 数据库目录

#### 5. 启动服务

```bash
# 启动 Master 和 Worker
./scripts/deploy-macos.sh start

# 只启动 Master
./scripts/deploy-macos.sh start master

# 只启动 Worker
./scripts/deploy-macos.sh start worker
```

**输出示例**:
```
✅ Starting Master on port 8848...
✅ Master started (PID: 12345)
✅ Starting Worker in auto mode...
✅ Worker started (PID: 12346)
```

#### 6. 查看状态

```bash
./scripts/deploy-macos.sh status
```

**输出示例**:
```
📊 AgentFlow Status:
├─ Master: Running (PID: 12345, Port: 8848)
├─ Worker: Running (PID: 12346, Mode: auto)
└─ Database: .claude/cpds-manager/agentflow.db
```

#### 7. 停止服务

```bash
./scripts/deploy-macos.sh stop
```

---

## 🐧 Linux 部署

### 前置要求

- 支持的发行版：Ubuntu 18.04+, Debian 10+, Fedora 31+, Arch Linux, CentOS/RHEL 8+
- sudo 权限
- systemd 支持（用于服务管理）

### 使用步骤

#### 1. 检查环境

```bash
./scripts/deploy-linux.sh check
```

**输出示例**:
```
✅ Distribution: Ubuntu 22.04
✅ Kernel: 5.15.0
✅ Python: 3.10.12
✅ Go: Not installed
⚠️  Claude CLI: Not installed
```

#### 2. 安装依赖

```bash
sudo ./scripts/deploy-linux.sh install
```

**支持的包管理器**:
- Ubuntu/Debian: `apt`
- Fedora/RHEL: `dnf` / `yum`
- Arch Linux: `pacman`

**自动安装内容**:
- Python 3.x 和 pip
- Go 1.21+
- Claude CLI（通过 npm）
- Python 依赖包

#### 3. 编译 Go 版本

```bash
./scripts/deploy-linux.sh build
```

#### 4. 部署配置和服务

```bash
sudo ./scripts/deploy-linux.sh deploy
```

**创建 systemd 服务**:
- `/etc/systemd/system/agentflow-master.service` - Master 服务
- `/etc/systemd/system/agentflow-worker.service` - Worker 服务

#### 5. 启动服务

```bash
# 启动并设为开机自启
sudo ./scripts/deploy-linux.sh start

# 手动使用 systemctl
sudo systemctl start agentflow-master agentflow-worker
sudo systemctl enable agentflow-master agentflow-worker
```

#### 6. 查看状态

```bash
./scripts/deploy-linux.sh status
```

**输出示例**:
```
📊 AgentFlow Status:
├─ Master: Active (PID: 1234, Port: 8848)
├─ Worker: Active (PID: 1235, Mode: auto)
├─ Services:
│  ├─ agentflow-master.service: enabled
│  └─ agentflow-worker.service: enabled
└─ Database: /var/lib/agentflow/agentflow.db
```

#### 7. 查看日志

```bash
# 查看 Master 日志
sudo journalctl -u agentflow-master -f

# 查看 Worker 日志
sudo journalctl -u agentflow-worker -f
```

#### 8. 停止服务

```bash
sudo ./scripts/deploy-linux.sh stop

# 或使用 systemctl
sudo systemctl stop agentflow-master agentflow-worker
```

---

## 🪟 Windows 部署

### 前置要求

- Windows 10/11（64 位）
- PowerShell 5.1 或更高版本
- 管理员权限（以管理员身份运行 PowerShell）

### 使用步骤

#### 1. 检查环境

```powershell
scripts\deploy-windows.bat check
```

**输出示例**:
```
✅ Windows Version: 10.0.19045
✅ Architecture: AMD64
⚠️  Python: Not installed
⚠️  Go: Not installed
⚠️  Claude CLI: Not installed
```

#### 2. 安装依赖

```powershell
scripts\deploy-windows.bat install
```

**自动安装内容**:
- Python 3.x（从 python.org）
- Go 1.21+（从 golang.org）
- Claude CLI（通过 npm）
- Python 依赖包

**安装方式**:
- 如果检测到 Chocolatey，优先使用 Chocolatey
- 否则使用官方安装程序

#### 3. 编译 Go 版本

```powershell
scripts\deploy-windows.bat build
```

**编译产物**:
```
golang\bin\master.exe       - Master 服务器
golang\bin\worker.exe       - Worker（持续模式）
golang\bin\oneshot.exe      - Worker（单次模式）
```

#### 4. 部署配置

```powershell
scripts\deploy-windows.bat deploy
```

**生成文件**:
- `.agentflow\boundaries.json` - 文件边界配置
- `golang\config.yaml` - Go 版本配置文件
- `.claude\cpds-manager\` - 数据库目录

#### 5. 启动服务

```powershell
# 启动 Master 和 Worker（在新窗口中）
scripts\deploy-windows.bat start

# 只启动 Master
scripts\deploy-windows.bat start master

# 只启动 Worker
scripts\deploy-windows.bat start worker
```

**输出示例**:
```
✅ Starting Master in new window...
✅ Starting Worker in new window...
```

服务会在新的命令行窗口中运行，保持窗口打开可查看日志。

#### 6. 查看状态

```powershell
scripts\deploy-windows.bat status
```

**输出示例**:
```
📊 AgentFlow Status:
├─ Master: Running (Port: 8848)
├─ Worker: Running (Mode: auto)
└─ Database: .claude\cpds-manager\agentflow.db
```

#### 7. 停止服务

```powershell
scripts\deploy-windows.bat stop
```

**停止方式**:
- 查找并终止 `master.exe` 和 `worker.exe` 进程
- 或直接关闭服务窗口

---

## 🔧 高级配置

### 自定义安装路径

编辑脚本中的变量：

**macOS/Linux**:
```bash
# 编辑 deploy-macos.sh 或 deploy-linux.sh
PROJECT_ROOT="/path/to/agentflow"  # 项目根目录
PYTHON_VERSION="3.11"              # Python 版本
GO_VERSION="1.21.3"                # Go 版本
```

**Windows**:
```batch
REM 编辑 deploy-windows.bat
set PROJECT_ROOT=C:\path\to\agentflow
set PYTHON_VERSION=3.11
set GO_VERSION=1.21.3
```

### 使用自定义配置文件

```bash
# macOS/Linux
./golang/bin/master -config /path/to/custom-config.yaml
./golang/bin/worker -config /path/to/custom-config.yaml

# Windows
golang\bin\master.exe -config C:\path\to\custom-config.yaml
golang\bin\worker.exe -config C:\path\to\custom-config.yaml
```

### 配置防火墙

**macOS/Linux**:
```bash
# 允许 8848 端口
sudo ufw allow 8848/tcp  # Ubuntu/Debian
sudo firewall-cmd --add-port=8848/tcp --permanent  # Fedora/RHEL
```

**Windows PowerShell**:
```powershell
New-NetFirewallRule -DisplayName "AgentFlow Master" -Direction Inbound -LocalPort 8848 -Protocol TCP -Action Allow
```

---

## 🐛 故障排除

### 问题 1: Claude CLI 安装失败

**错误**: `npm ERR! code EACCES`

**解决**:
```bash
# 使用 sudo 安装（Linux/macOS）
sudo npm install -g @anthropic-ai/claude-code

# 或配置 npm 全局目录
mkdir ~/.npm-global
npm config set prefix '~/.npm-global'
echo 'export PATH=~/.npm-global/bin:$PATH' >> ~/.bashrc
source ~/.bashrc
npm install -g @anthropic-ai/claude-code
```

### 问题 2: Go 编译失败

**错误**: `go: cannot find main module`

**解决**:
```bash
# 确保在项目根目录
cd /Users/jiangxiaolong/work/project/AgentFlow
go mod tidy
./scripts/deploy-macos.sh build
```

### 问题 3: 权限错误

**错误**: `Permission denied`

**解决**:
```bash
# macOS/Linux: 添加执行权限
chmod +x scripts/deploy-macos.sh
chmod +x scripts/deploy-linux.sh

# Windows: 以管理员身份运行 PowerShell
右键点击 PowerShell -> 以管理员身份运行
```

### 问题 4: 端口被占用

**错误**: `bind: address already in use`

**解决**:
```bash
# macOS/Linux: 查找并终止占用端口的进程
lsof -ti:8848 | xargs kill -9  # macOS
fuser -k 8848/tcp              # Linux

# Windows: 查找并终止进程
netstat -ano | findstr :8848
taskkill /PID <PID> /F
```

### 问题 5: systemd 服务启动失败（Linux）

**错误**: `Failed to start agentflow-master.service`

**解决**:
```bash
# 查看详细日志
sudo journalctl -u agentflow-master -n 50

# 检查配置文件路径
sudo systemctl cat agentflow-master

# 手动测试服务
sudo -u agentflow /path/to/golang/bin/master -config /path/to/config.yaml
```

---

## 📚 相关文档

- [README.md](../README.md) - 项目整体说明
- [Git 集成指南](git-integration-guide.md) - 文件边界和权限控制
- [Agent 任务升级机制](agent-upgrade-mechanism.md) - 冲突解决机制
- [Python 版本 README](../python/README.md) - Python 版本详细文档
- [Go 版本 README](../golang/README.md) - Go 版本详细文档

---

## 🎯 下一步

部署完成后，您可以：

1. **创建任务** - 通过 REST API 或 CLI
   ```bash
   curl -X POST http://127.0.0.1:8848/api/v1/tasks \
     -H "Content-Type: application/json" \
     -d '{"title": "测试任务", "description": "echo Hello AgentFlow"}'
   ```

2. **配置 Git 集成** - 启用多 Agent 协作
   ```bash
   # 编辑 .agentflow/boundaries.json
   # 参考 docs/git-integration-guide.md
   ```

3. **部署到云端** - 使用 Docker/Kubernetes
   ```bash
   # 参考 golang/deployments/
   kubectl apply -f golang/deployments/
   ```

---

**版本**: v1.0.0
**更新**: 2026-01-22
**作者**: AgentFlow Team
