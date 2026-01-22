# AgentFlow API 文档

AgentFlow 提供 RESTful API 用于任务管理和 Worker 协调。

## 基础信息

- **Base URL**: `http://localhost:8848`
- **API 版本**: v1
- **Content-Type**: `application/json`

## 健康检查

### GET /health

检查 Master 服务是否正常运行。

**请求示例**:
```bash
curl http://localhost:8848/health
```

**响应**:
```json
{
  "status": "ok"
}
```

## 任务管理

### 创建任务

### POST /api/v1/tasks

创建新任务。

**请求体**:
```json
{
  "title": "运行测试",
  "description": "shell:go test ./...",
  "group_name": "linux",
  "completion_criteria": "所有测试通过"
}
```

**字段说明**:
- `title` (string, required): 任务标题
- `description` (string, optional): 任务描述，支持 shell 命令或脚本路径
- `group_name` (string, optional): 工作组名称，默认 "default"
- `completion_criteria` (string, optional): 完成标准

**响应示例**:
```json
{
  "task_id": "1"
}
```

**状态码**: 201 Created

### 列出任务

### GET /api/v1/tasks

获取任务列表。

**查询参数**:
- `status` (string, optional): 过滤状态 (`pending`, `running`, `completed`, `failed`)
- `group` (string, optional): 过滤工作组

**请求示例**:
```bash
# 获取所有待执行任务
curl "http://localhost:8848/api/v1/tasks?status=pending"

# 获取特定组的任务
curl "http://localhost:8848/api/v1/tasks?group=docker"
```

**响应示例**:
```json
{
  "tasks": [
    {
      "id": "1",
      "title": "运行测试",
      "description": "shell:go test ./...",
      "group_name": "linux",
      "status": "pending",
      "lock_holder": "",
      "lock_time": null,
      "created_at": "2024-01-01T10:00:00Z",
      "started_at": null,
      "completed_at": null
    },
    {
      "id": "2",
      "title": "构建镜像",
      "description": "shell:docker build -t app .",
      "group_name": "docker",
      "status": "running",
      "lock_holder": "worker-123",
      "lock_time": "2024-01-01T10:05:00Z",
      "created_at": "2024-01-01T10:01:00Z",
      "started_at": "2024-01-01T10:05:00Z",
      "completed_at": null
    }
  ]
}
```

### 获取任务详情

### GET /api/v1/tasks/:id

获取单个任务的详细信息。

**请求示例**:
```bash
curl http://localhost:8848/api/v1/tasks/1
```

**响应示例**:
```json
{
  "id": "1",
  "title": "运行测试",
  "description": "shell:go test ./...",
  "group_name": "linux",
  "completion_criteria": "所有测试通过",
  "status": "completed",
  "lock_holder": "",
  "lock_time": null,
  "result": "PASS: TestMain\nPASS: TestAPI\nok",
  "error": "",
  "created_at": "2024-01-01T10:00:00Z",
  "started_at": "2024-01-01T10:05:00Z",
  "completed_at": "2024-01-01T10:05:30Z"
}
```

### 锁定任务

### POST /api/v1/tasks/:id/lock

Worker 使用此接口锁定并开始执行任务。

**请求体**:
```json
{
  "worker_id": "worker-linux-001"
}
```

**响应示例**:
```json
{
  "status": "locked"
}
```

**状态码**:
- 200 OK: 成功锁定
- 409 Conflict: 任务已被其他 worker 锁定

### 解锁任务

### POST /api/v1/tasks/:id/unlock

解锁任务（通常用于错误恢复）。

**请求体**:
```json
{
  "worker_id": "worker-linux-001"
}
```

**响应示例**:
```json
{
  "status": "unlocked"
}
```

### 完成任务

### POST /api/v1/tasks/:id/complete

标记任务为已完成。

**请求体**:
```json
{
  "worker_id": "worker-linux-001",
  "result": "命令执行成功\n输出: ..."
}
```

**响应示例**:
```json
{
  "status": "completed"
}
```

### 标记任务失败

### POST /api/v1/tasks/:id/fail

标记任务执行失败。

**请求体**:
```json
{
  "worker_id": "worker-linux-001",
  "error": "命令执行失败: exit code 1"
}
```

**响应示例**:
```json
{
  "status": "failed"
}
```

## Worker 接口

### 获取待执行任务

### GET /api/v1/tasks/pending

Worker 使用此接口轮询待执行任务。

**查询参数**:
- `group` (string, optional): 工作组名称

**请求示例**:
```bash
curl "http://localhost:8848/api/v1/tasks/pending?group=docker"
```

**响应示例**:
```json
{
  "tasks": [
    {
      "id": "5",
      "title": "构建 Docker 镜像",
      "description": "shell:docker build -t myapp:latest .",
      "group_name": "docker",
      "created_at": "2024-01-01T11:00:00Z"
    }
  ]
}
```

**状态码**:
- 200 OK: 有待执行任务
- 204 No Content: 没有待执行任务

### 发送心跳

### POST /api/v1/workers/:id/heartbeat

Worker 定期发送心跳以保持活跃状态。

**请求示例**:
```bash
curl -X POST http://localhost:8848/api/v1/workers/worker-001/heartbeat
```

**响应示例**:
```json
{
  "status": "ok"
}
```

### 列出 Workers

### GET /api/v1/workers

获取所有 Worker 信息。

**查询参数**:
- `group` (string, optional): 过滤工作组

**请求示例**:
```bash
curl "http://localhost:8848/api/v1/workers?group=linux"
```

**响应示例**:
```json
{
  "workers": [
    {
      "id": "worker-linux-001",
      "group_name": "linux",
      "type": "remote",
      "capabilities": "{\"os\":\"linux\",\"arch\":\"amd64\",\"cpu_num\":8}",
      "status": "active",
      "last_heartbeat": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T10:00:00Z"
    }
  ]
}
```

## 统计信息

### 获取统计信息

### GET /api/v1/stats

获取整体统计信息。

**请求示例**:
```bash
curl http://localhost:8848/api/v1/stats
```

**响应示例**:
```json
{
  "stats": [
    {
      "group_name": "linux",
      "total_tasks": 100,
      "pending_tasks": 10,
      "running_tasks": 5,
      "completed_tasks": 80,
      "failed_tasks": 5
    },
    {
      "group_name": "docker",
      "total_tasks": 50,
      "pending_tasks": 5,
      "running_tasks": 2,
      "completed_tasks": 43,
      "failed_tasks": 0
    }
  ]
}
```

### 获取工作组统计

### GET /api/v1/stats/groups

获取各工作组的任务统计。

**请求示例**:
```bash
curl http://localhost:8848/api/v1/stats/groups
```

**响应示例**: 同 GET /api/v1/stats

## 错误响应

所有错误响应格式统一：

```json
{
  "error": "错误描述信息"
}
```

**常见 HTTP 状态码**:
- 400 Bad Request: 请求参数错误
- 404 Not Found: 资源不存在
- 409 Conflict: 资源冲突（如任务已被锁定）
- 500 Internal Server Error: 服务器内部错误

## 使用示例

### Python 示例

```python
import requests

BASE_URL = "http://localhost:8848"

# 创建任务
response = requests.post(f"{BASE_URL}/api/v1/tasks", json={
    "title": "数据处理",
    "description": "shell:python process.py",
    "group_name": "linux"
})
task_id = response.json()["task_id"]

# 查询任务状态
response = requests.get(f"{BASE_URL}/api/v1/tasks/{task_id}")
task = response.json()
print(f"任务状态: {task['status']}")

# 列出所有任务
response = requests.get(f"{BASE_URL}/api/v1/tasks?status=pending")
tasks = response.json()["tasks"]
```

### Go 示例

```go
package main

import (
    "bytes"
    "encoding/json"
    "net/http"
)

func main() {
    baseURL := "http://localhost:8848"

    // 创建任务
    taskData := map[string]interface{}{
        "title":       "运行测试",
        "description": "shell:go test ./...",
        "group_name":  "linux",
    }
    body, _ := json.Marshal(taskData)

    resp, err := http.Post(baseURL+"/api/v1/tasks", "application/json", bytes.NewBuffer(body))
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    var result map[string]string
    json.NewDecoder(resp.Body).Decode(&result)
    println("任务 ID:", result["task_id"])
}
```

### cURL 示例

```bash
# 创建任务
curl -X POST http://localhost:8848/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "部署应用",
    "description": "shell:./deploy.sh",
    "group_name": "production"
  }'

# 查询任务
curl http://localhost:8848/api/v1/tasks/1

# 列出待执行任务
curl "http://localhost:8848/api/v1/tasks?status=pending"

# 锁定任务
curl -X POST http://localhost:8848/api/v1/tasks/1/lock \
  -H "Content-Type: application/json" \
  -d '{"worker_id": "worker-001"}'

# 完成任务
curl -X POST http://localhost:8848/api/v1/tasks/1/complete \
  -H "Content-Type: application/json" \
  -d '{"worker_id": "worker-001", "result": "部署成功"}'
```

## 任务描述格式

任务描述支持以下前缀：

### Shell 命令

```
shell:command arguments
```

示例:
- `shell:go test ./...`
- `shell:docker build -t app .`
- `shell:ls -la /tmp`

### 脚本执行

```
script:/path/to/script.sh
```

示例:
- `script:/app/deploy.sh`
- `script:C:\scripts\build.bat`

### Windows PowerShell

```
shell:Invoke-Build
```

在 Windows 环境下，自动检测并使用 PowerShell。

## 注意事项

1. **任务锁超时**: 任务锁默认 5 分钟超时，超时后自动释放
2. **心跳间隔**: Worker 应每 30 秒发送一次心跳
3. **轮询间隔**: Worker 建议每 5 秒轮询一次待执行任务
4. **幂等性**: 锁定任务接口是幂等的，同一 worker 可重复锁定
5. **错误处理**: Worker 执行任务失败时应调用 `/fail` 接口
