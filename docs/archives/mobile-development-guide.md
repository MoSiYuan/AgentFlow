# AgentFlow - 手机端快速开发指南

## 🎯 场景说明

使用手机（通过个人 AI）控制云端 Master，调度本地 Windows/macOS 机器进行跨平台编译和开发。

## 📱 架构优势

### 为什么这个架构强大？

1. **手机作为控制中心**
   - 随时随地提交开发任务
   - 无需本地开发环境
   - 利用个人 AI 智能调度

2. **云端 Master (Go)**
   - 24/7 运行，公网可访问
   - 任务调度和分发中心
   - 高并发处理能力

3. **本地 Workers (Python)**
   - Windows 组：处理 Windows 特定任务
   - macOS 组：处理 macOS 特定任务
   - 可利用 GUI 操作（VSCode、编译器等）

---

## 🛠️ 部署步骤

### 步骤 1: 部署云端 Master（Go 版本）

#### 方案 A: 使用云服务器（推荐）

```bash
# 1. 在云服务器上（如阿里云、腾讯云、AWS）
ssh user@your-cloud-server

# 2. 下载 AgentFlow Go 版本
git clone https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow/golang

# 3. 编译（或直接使用预编译二进制）
go build -o bin/master cmd/master/main.go

# 4. 准备配置文件
cat > config.yaml <<EOF
master:
  host: "0.0.0.0"  # 监听所有网络接口
  port: 8848
  db_path: "/var/lib/agentflow/agentflow.db"
  auto_start: false
EOF

# 5. 创建数据库目录
mkdir -p /var/lib/agentflow

# 6. 启动 Master（使用 systemd 或 supervisor）
nohup ./bin/master -config config.yaml > /var/log/agentflow.log 2>&1 &

# 7. 配置防火墙（开放 8848 端口）
# Ubuntu/Debian:
sudo ufw allow 8848/tcp
# CentOS/RHEL:
sudo firewall-cmd --permanent --add-port=8848/tcp
sudo firewall-cmd --reload

# 8. 配置域名（可选，推荐使用 nginx 反向代理）
# 使用 nginx 添加 SSL
```

#### 方案 B: 使用 Kubernetes（企业级）

```bash
# 1. 创建部署文件
cat > k8s-master-deployment.yaml <<EOF
apiVersion: apps/v1
kind: Deployment
metadata:
  name: agentflow-master
spec:
  replicas: 1
  selector:
    matchLabels:
      app: agentflow-master
  template:
    metadata:
      labels:
        app: agentflow-master
    spec:
      containers:
      - name: master
        image: your-registry/agentflow-master:latest
        ports:
        - containerPort: 8848
        volumeMounts:
        - name: data
          mountPath: /var/lib/agentflow
        env:
        - name: MASTER_HOST
          value: "0.0.0.0"
        - name: MASTER_PORT
          value: "8848"
      volumes:
      - name: data
        persistentVolumeClaim:
          claimName: agentflow-pvc
---
apiVersion: v1
kind: Service
metadata:
  name: agentflow-master
spec:
  type: LoadBalancer
  ports:
  - port: 8848
    targetPort: 8848
  selector:
    app: agentflow-master
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: agentflow-pvc
spec:
  accessModes:
    - ReadWriteOnce
  resources:
    requests:
      storage: 10Gi
EOF

# 2. 部署
kubectl apply -f k8s-master-deployment.yaml

# 3. 获取服务地址
kubectl get svc agentflow-master
```

---

### 步骤 2: 部署本地 Workers（Python 版本）

#### Windows 机器

```batch
:: 1. 安装 Python (3.8+)
:: 下载: https://www.python.org/downloads/

:: 2. 克隆代码
git clone https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow/python

:: 3. 安装依赖
pip install -r requirements.txt

:: 4. 启动 Worker（Windows 组）
python -m agentflow.cli worker ^
  --master https://your-cloud-server:8848 ^
  --group windows ^
  --auto
```

创建 Windows 服务（开机自启）：

```python
# create_windows_service.py
import win32serviceutil
import win32service
import win32event
import servicemanager
import subprocess
import sys

class AgentFlowWorkerService(win32serviceutil.ServiceFramework):
    _svc_name_ = "AgentFlowWorker"
    _svc_display_name_ = "AgentFlow Worker (Windows)"
    _svc_description_ = "AgentFlow Worker for Windows group"

    def __init__(self, args):
        win32serviceutil.ServiceFramework.__init__(self, args)
        self.hWaitStop = win32event.CreateEvent(None, 0, 0, None)
        self.process = None

    def SvcStop(self):
        self.ReportServiceStatus(win32service.SERVICE_STOP_PENDING)
        if self.process:
            self.process.terminate()
        win32event.SetEvent(self.hWaitStop)

    def SvcDoRun(self):
        servicemanager.LogMsg(
            servicemanager.EVENTLOG_INFORMATION_TYPE,
            servicemanager.PYS_SERVICE_STARTED,
            (self._svc_name_, '')
        )
        self.main()

    def main(self):
        # 启动 Worker
        self.process = subprocess.Popen([
            sys.executable, '-m', 'agentflow.cli',
            'worker',
            '--master', 'https://your-cloud-server:8848',
            '--group', 'windows',
            '--auto'
        ])
        win32event.WaitForSingleObject(self.hWaitStop, win32event.INFINITE)

if __name__ == '__main__':
    win32serviceutil.HandleCommandLine(AgentFlowWorkerService)
```

安装服务：

```batch
pip install pywin32
python create_windows_service.py install
python create_windows_service.py start
```

#### macOS 机器

```bash
# 1. 安装 Python 3.8+
brew install python@3.11

# 2. 克隆代码
git clone https://github.com/MoSiYuan/AgentFlow.git
cd AgentFlow/python

# 3. 安装依赖
pip3 install -r requirements.txt

# 4. 启动 Worker（macOS 组）
python3 -m agentflow.cli \
  --master https://your-cloud-server:8848 \
  --group macos \
  --auto
```

创建 macOS LaunchAgent（开机自启）：

```plist
<!-- ~/Library/LaunchAgents/com.agentflow.worker.plist -->
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.agentflow.worker</string>
    <key>ProgramArguments</key>
    <array>
        <string>/usr/local/bin/python3</string>
        <string>-m</string>
        <string>agentflow.cli</string>
        <string>worker</string>
        <string>--master</string>
        <string>https://your-cloud-server:8848</string>
        <string>--group</string>
        <string>macos</string>
        <string>--auto</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/agentflow-worker.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/agentflow-worker.error</string>
</dict>
</plist>
```

加载服务：

```bash
launchctl load ~/Library/LaunchAgents/com.agentflow.worker.plist
launchctl start com.agentflow.worker
```

---

### 步骤 3: 手机端集成（个人 AI）

#### 方案 A: 使用 Claude Code / ChatGPT

```markdown
## 提示词模板

你是一个软件开发助手，可以通过 AgentFlow API 控制云端 Master。

**云端 Master 地址**: https://your-cloud-server:8848

**可用 Worker 组**:
- windows: Windows 编译和 GUI 操作
- macos: macOS 编译和 GUI 操作
- cloud: 云端批量处理

当你需要执行开发任务时，请：

1. 解析任务需求
2. 确定需要哪个 Worker 组
3. 调用 AgentFlow API 创建任务
4. 等待执行结果
5. 返回结果给用户

**API 示例**:
```bash
curl -X POST https://your-cloud-server:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "编译 Windows 版本",
    "description": "shell:cd /path/to/project && go build -o bin/app.exe",
    "group_name": "windows"
  }'
```
```

#### 方案 B: 手机 App 直接集成

```python
# mobile_client.py - 手机端 Python 客户端示例

import requests
import json

class AgentFlowMobileClient:
    """AgentFlow 手机端客户端"""

    def __init__(self, master_url, api_key=None):
        self.master_url = master_url
        self.api_key = api_key
        self.session = requests.Session()

        if api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {api_key}'
            })

    def create_task(self, title, description, group_name):
        """创建任务"""
        url = f"{self.master_url}/api/v1/tasks"
        payload = {
            'title': title,
            'description': description,
            'group_name': group_name
        }

        response = self.session.post(url, json=payload)
        response.raise_for_status()
        return response.json()

    def get_task_status(self, task_id):
        """获取任务状态"""
        url = f"{self.master_url}/api/v1/tasks/{task_id}"
        response = self.session.get(url)
        response.raise_for_status()
        return response.json()

    def compile_windows(self, project_path, binary_name="app.exe"):
        """编译 Windows 版本"""
        description = f"""
编译 Windows 可执行文件

1. 进入项目目录: {project_path}
2. 设置编译环境
3. 执行编译命令
4. 输出到 bin/{binary_name}

编译命令:
cd {project_path}
go build -o bin/{binary_name}

验证编译结果:
if [ -f bin/{binary_name} ]; then
    echo "编译成功: {binary_name}"
    ls -lh bin/{binary_name}
else
    echo "编译失败"
    exit 1
fi
        """.strip()

        return self.create_task(
            title=f"编译 {binary_name}",
            description=description,
            group_name="windows"
        )

    def compile_macos(self, project_path, binary_name="app"):
        """编译 macOS 版本"""
        description = f"""
编译 macOS 可执行文件

1. 进入项目目录: {project_path}
2. 设置编译环境
3. 执行编译命令
4. 输出到 bin/{binary_name}

编译命令:
cd {project_path}
go build -o bin/{binary_name}

验证编译结果:
if [ -f bin/{binary_name} ]; then
    echo "编译成功: {binary_name}"
    ls -lh bin/{binary_name}
else
    echo "编译失败"
    exit 1
fi
        """.strip()

        return self.create_task(
            title=f"编译 {binary_name}",
            description=description,
            group_name="macos"
        )

    def cross_compile(self, project_path):
        """交叉编译所有平台"""
        tasks = []

        # Windows
        tasks.append(self.create_task(
            title="编译 Windows amd64 版本",
            description=f"cd {project_path} && GOOS=windows GOARCH=amd64 go build -o bin/app.exe",
            group_name="windows"
        ))

        # macOS
        tasks.append(self.create_task(
            title="编译 macOS amd64 版本",
            description=f"cd {project_path} && GOOS=darwin GOARCH=amd64 go build -o bin/app-mac-amd64",
            group_name="macos"
        ))

        tasks.append(self.create_task(
            title="编译 macOS arm64 版本",
            description=f"cd {project_path} && GOOS=darwin GOARCH=arm64 go build -o bin/app-mac-arm64",
            group_name="macos"
        ))

        return tasks

# 使用示例
def mobile_quick_compile():
    """手机端快速编译示例"""

    # 初始化客户端（云端 Master 地址）
    client = AgentFlowMobileClient(
        master_url="https://your-cloud-server:8848",
        api_key="your-api-key"  # 可选
    )

    # 快速编译 Windows 版本
    task = client.compile_windows(
        project_path="/path/to/your/project",
        binary_name="myapp.exe"
    )

    task_id = task['task_id']
    print(f"任务已创建: {task_id}")

    # 轮询任务状态
    import time
    while True:
        status = client.get_task_status(task_id)
        state = status['status']

        if state == 'completed':
            print("✅ 编译成功!")
            print(f"结果: {status['result']}")
            break
        elif state == 'failed':
            print("❌ 编译失败!")
            print(f"错误: {status['error']}")
            break
        else:
            print(f"⏳ 正在编译... ({state})")
            time.sleep(5)

    return status

if __name__ == '__main__':
    mobile_quick_compile()
```

---

## 🎯 实际使用场景

### 场景 1: 手机提交跨平台编译任务

```python
# 从手机执行
client = AgentFlowMobileClient("https://your-cloud-server:8848")

# 一次性提交多个平台编译
tasks = client.cross_compile("/path/to/project")

print(f"已提交 {len(tasks)} 个编译任务:")
for task in tasks:
    print(f"  - {task['title']}: {task['task_id']}")
```

**输出**:
```
已提交 3 个编译任务:
  - 编译 Windows amd64 版本: TASK-001
  - 编译 macOS amd64 版本: TASK-002
  - 编译 macOS arm64 版本: TASK-003
```

### 场景 2: 通过 Claude Code 提交任务

在手机上打开 Claude Code / ChatGPT:

```
用户: 帮我编译一下我的 Go 项目，需要 Windows 和 macOS 版本

AI: 好的，我来为你提交编译任务。

[调用 AgentFlow API]

✅ 已创建以下任务：
1. Windows 版本编译 - 分配到 windows 组
2. macOS amd64 版本编译 - 分配到 macos 组
3. macOS arm64 版本编译 - 分配到 macos 组

预计完成时间: 2-3 分钟

[2分钟后]

✅ 所有编译任务已完成！

Windows 版本: bin/app.exe (15MB)
macOS amd64: bin/app-mac-amd64 (14MB)
macOS arm64: bin/app-mac-arm64 (13MB)

所有二进制文件已上传到云存储，下载链接:
https://your-storage.com/compilation/xxx
```

### 场景 3: 复杂的多步骤开发任务

```python
# 手机端提交完整开发流程
client = AgentFlowMobileClient("https://your-cloud-server:8848")

# 1. 运行测试（Windows）
task1 = client.create_task(
    title="运行 Windows 测试",
    description="""
cd /path/to/project
go test ./... -v -cover
""",
    group_name="windows"
)

# 2. 编译 macOS 版本
task2 = client.compile_macos(
    project_path="/path/to/project"
)

# 3. 打包发布
task3 = client.create_task(
    title="打包发布版本",
    description="""
cd /path/to/project
mkdir -p release
cp bin/*.exe release/
cp bin/app-mac-* release/
tar czf release.tar.gz release/
""",
    group_name="macos"
)

print(f"已创建 3 个任务: {task1['task_id']}, {task2['task_id']}, {task3['task_id']}")
```

---

## 📊 任务监控

### 查看所有 Workers 状态

```bash
curl https://your-cloud-server:8848/api/v1/workers
```

**响应示例**:
```json
{
  "workers": [
    {
      "id": "windows-001",
      "group_name": "windows",
      "type": "local",
      "status": "online",
      "last_heartbeat": "2026-01-22T10:30:00Z"
    },
    {
      "id": "macos-001",
      "group_name": "macos",
      "type": "local",
      "status": "online",
      "last_heartbeat": "2026-01-22T10:30:05Z"
    },
    {
      "id": "cloud-worker-1",
      "group_name": "cloud",
      "type": "remote",
      "status": "online",
      "last_heartbeat": "2026-01-22T10:29:55Z"
    }
  ]
}
```

### 查看统计信息

```bash
curl https://your-cloud-server:8848/api/v1/stats
```

**响应示例**:
```json
{
  "stats": {
    "total_tasks": 1234,
    "pending_tasks": 5,
    "running_tasks": 2,
    "completed_tasks": 1220,
    "failed_tasks": 7
  }
}
```

---

## 🔒 安全建议

### 1. 使用 HTTPS + API Key

```nginx
# nginx 反向代理配置
server {
    listen 443 ssl http2;
    server_name agentflow.yourdomain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://localhost:8848;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    # API Key 认证（通过 nginx 实现）
    # 或在应用层实现
}
```

### 2. VPN / 内网穿透

如果云服务器没有公网 IP:

```bash
# 使用 frp 内网穿透
# 在云服务器上运行 frps
./frps -c frps.ini

# 在本地机器上运行 frpc
./frpc -c frpc.ini
```

### 3. API 限流

在 Master 代码中添加限流中间件（可选）。

---

## 🎉 总结

### 这个架构的优势

✅ **随时随地开发** - 手机即可控制
✅ **充分利用本地资源** - GUI 操作、特定平台编译
✅ **24/7 可用** - 云端 Master 持续运行
✅ **智能调度** - 自动分配到合适的 Worker 组
✅ **API 统一** - Python/Go 版本 API 100% 兼容

### 扩展性

- 🔹 可以添加更多平台组（linux、android、ios）
- 🔹 可以添加专用 Workers（gpu-worker、storage-worker）
- 🔹 可以集成 CI/CD 流程
- 🔹 可以添加 Web Dashboard 监控

---

**文档版本**: v1.0.0
**更新日期**: 2026-01-22
