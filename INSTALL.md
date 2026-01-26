# AgentFlow - AI Agent Task Collaboration System

**一键安装，开箱即用**

---

## ⚡ 快速安装（推荐）

### 方式 1: 独立可执行文件（最简单）⭐

```bash
# 下载并运行（无需 Node.js）
curl -fsSL https://raw.githubusercontent.com/jiangxiaolong/AgentFlow/main/nodejs/dist/agentflow-master -o agentflow-master
chmod +x agentflow-master
./agentflow-master
```

### 方式 2: Docker（一键部署）🐳

```bash
docker run -d \
  --name agentflow \
  -p 6767:6767 \
  -v ~/.claude:/root/.claude \
  -v agentflow-data:/data \
  jiangxiaolong/agentflow:latest
```

### 方式 3: 从源码构建

```bash
# 克隆仓库
git clone https://github.com/jiangxiaolong/AgentFlow.git
cd AgentFlow/nodejs

# 打包（生成独立可执行文件）
chmod +x package.sh
./package.sh

# 运行
./dist/agentflow-master
```

---

## 🎯 Claude Skill 安装

### 自动安装（推荐）

让 Claude AI 直接安装：

```
请帮我安装 AgentFlow skill：
1. 下载 https://github.com/jiangxiaolong/AgentFlow
2. 将 agentflow.md 复制到 ~/.claude/skills/agentflow/
3. 启动服务
```

### 手动安装

```bash
# 1. 创建 skill 目录
mkdir -p ~/.claude/skills/agentflow

# 2. 下载 skill 定义
curl -fsSL https://raw.githubusercontent.com/jiangxiaolong/AgentFlow/main/.claude/skills/agentflow.md \
  -o ~/.claude/skills/agentflow/agentflow.md

# 3. 下载可执行文件
curl -fsSL https://github.com/jiangxiaolong/AgentFlow/releases/download/v1.0.0/agentflow-master \
  -o ~/.claude/skills/agentflow/agentflow-master
chmod +x ~/.claude/skills/agentflow/agentflow-master

# 4. 启动服务
~/.claude/skills/agentflow/agentflow-master
```

---

## ✅ 验证安装

```bash
# 健康检查
curl http://localhost:6767/health

# 预期输出
{
  "status": "ok",
  "version": "1.0.0",
  "uptime": 123.456
}
```

---

## 📚 快速开始

### 创建第一个任务

```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Hello World",
    "description": "echo Hello from AgentFlow",
    "priority": "high"
  }'
```

### 使用 Claude CLI

```bash
# 安装 Claude CLI（如果还没有）
npm install -g @anthropic-ai/claude-code

# 使用 AgentFlow skill
claude "使用 AgentFlow 创建一个测试任务"
```

---

## 🔧 配置

### 默认配置

- **端口**: 6767 (Node.js) / 8848 (Go)
- **数据库**: `~/.claude/skills/agentflow/agentflow.db`
- **日志**: 控制台输出

### 自定义配置

```bash
# 环境变量
export AGENTFLOW_DB_PATH="/custom/path/agentflow.db"
export AGENTFLOW_PORT=8080

# 命令行参数
agentflow-master --port 8080 --db /data/agentflow.db
```

---

## 📦 系统要求

### 独立可执行文件
- ✅ **无需 Node.js**
- ✅ **无需依赖**
- ✅ 支持 macOS, Linux, Windows

### Docker
- ✅ Docker 20.10+
- ✅ Docker Compose (可选)

### 从源码构建
- Node.js 20 LTS
- pnpm 10+

---

## 🌟 特性

- ✅ **任务管理**: 创建、执行、监控任务
- ✅ **Claude 集成**: 与 Claude AI 深度集成
- ✅ **任务链**: 支持串行、并行、树形任务链
- ✅ **Git 集成**: 自动分支创建和管理
- ✅ **状态同步**: AgentFlow ↔ Claude 双向同步
- ✅ **Worker 支持**: 分布式任务执行

---

## 📚 文档

- [完整文档](https://github.com/jiangxiaolong/AgentFlow/wiki)
- [API 文档](docs/API.md)
- [开发指南](docs/NODEJS_GUIDE.md)
- [打包指南](nodejs/PACKAGING_GUIDE.md)

---

## 🆘 问题排查

### 问题 1: 端口被占用

```bash
# 更换端口
agentflow-master --port 8080
```

### 问题 2: 数据库权限错误

```bash
# 创建目录
mkdir -p ~/.claude/skills/agentflow
chmod 755 ~/.claude/skills/agentflow
```

### 问题 3: Docker 容器无法启动

```bash
# 查看日志
docker logs agentflow

# 重新构建
docker build -t agentflow:latest .
```

---

## 🤝 贡献

欢迎贡献！请查看 [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 📄 许可证

MIT License - 详见 [LICENSE](LICENSE)

---

**🎉 开始使用 AgentFlow，让你的 AI 任务管理更高效！**

---

**快速链接**:
- GitHub: https://github.com/jiangxiaolong/AgentFlow
- 文档: https://github.com/jiangxiaolong/AgentFlow/wiki
- 问题反馈: https://github.com/jiangxiaolong/AgentFlow/issues
