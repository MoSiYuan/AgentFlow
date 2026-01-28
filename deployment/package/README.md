# AgentFlow 部署包

## 编译产物位置

### Rust Master 服务器
- **路径**: `../../rust/target/release/agentflow-master`
- **大小**: 8.1 MB
- **用途**: 主服务器，提供 HTTP API 和 WebSocket 服务

### React 前端
- **路径**: `../../dashboard/dist/`
- **构建**: `cd dashboard && npm run build`
- **用途**: Web 管理界面

## 快速部署

### 方式 1: 直接使用二进制文件

```bash
# 1. 复制 Master 二进制文件
cp rust/target/release/agentflow-master /usr/local/bin/

# 2. 构建前端
cd dashboard
npm install
npm run build

# 3. 配置环境变量
export AGENTFLOW_PORT=6767
export AUTH_ENABLED=true
export AUTH_USERNAME=admin
export AUTH_PASSWORD=admin

# 4. 启动服务
agentflow-master
```

### 方式 2: 使用安装脚本（推荐）

```bash
npm install -g @agentflow/installer
agentflow-install
```

### 方式 3: Docker 部署

```bash
docker-compose up -d
```

## 文件清单

```
deployment/
├── package/                    # 部署包
│   ├── README.md              # 本文件
│   ├── install.sh             # 安装脚本
│   └── agentflow-helper       # AgentFlow CLI Helper
├── docker-compose.yml         # Docker 编排
├── Dockerfile                 # Docker 镜像
└── .env.example              # 环境变量示例
```
