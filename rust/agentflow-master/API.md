# AgentFlow Master API 文档

## 基础信息

- **Base URL**: `http://localhost:6767`
- **API 版本**: v1
- **Content-Type**: `application/json`
- **响应格式**: JSON

## 通用响应格式

所有 API 响应遵循统一格式：

```json
{
  "success": true/false,
  "data": { ... },      // 成功时的数据
  "error": "错误信息"    // 失败时的错误信息
}
```

## 错误码

| HTTP 状态码 | 说明 |
|------------|------|
| 200 | 成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

---

## 健康检查

### GET /health
### GET /api/v1/health

检查服务器健康状态。

**请求示例:**
```bash
curl http://localhost:6767/health
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "status": "healthy",
    "version": "0.2.0",
    "uptime": 3600,
    "mode": "master"
  }
}
```

**字段说明:**
- `status`: 服务器状态 (healthy/busy)
- `version`: 版本号
- `uptime`: 运行时间（秒）
- `mode`: 运行模式 (master)

---

## 任务管理 API

### 1. 创建任务

### POST /api/v1/tasks

创建新的任务。

**请求体:**
```json
{
  "title": "任务标题",
  "description": "任务描述",
  "group_name": "任务分组",
  "completion_criteria": "完成标准",
  "priority": "high|medium|low",
  "workspace_dir": "/path/to/workspace",
  "sandboxed": true,
  "allow_network": false,
  "max_memory": "1G",
  "max_cpu": 2,
  "created_by": "用户标识",
  "parent_id": 1  // 父任务 ID（可选）
}
```

**请求示例:**
```bash
curl -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "分析代码库",
    "description": "分析项目结构和依赖关系",
    "group_name": "code-analysis",
    "priority": "high"
  }'
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "parent_id": null,
    "title": "分析代码库",
    "description": "分析项目结构和依赖关系",
    "group_name": "code-analysis",
    "completion_criteria": null,
    "status": "pending",
    "priority": "High",
    "lock_holder": null,
    "lock_time": null,
    "result": null,
    "error": null,
    "workspace_dir": null,
    "sandboxed": false,
    "allow_network": false,
    "max_memory": null,
    "max_cpu": null,
    "created_by": null,
    "created_at": "2026-01-28T10:00:00Z",
    "started_at": null,
    "completed_at": null
  }
}
```

---

### 2. 获取任务详情

### GET /api/v1/tasks/:id

获取指定任务的详细信息。

**参数:**
- `id`: 任务 ID（路径参数）

**请求示例:**
```bash
curl http://localhost:6767/api/v1/tasks/1
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "task_id": "550e8400-e29b-41d4-a716-446655440000",
    "title": "分析代码库",
    "status": "completed",
    "result": "任务执行结果...",
    ...
  }
}
```

---

### 3. 列出所有任务

### GET /api/v1/tasks

获取所有任务列表（最多 100 个）。

**请求示例:**
```bash
curl http://localhost:6767/api/v1/tasks
```

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "id": 1,
      "title": "任务1",
      "status": "completed",
      "created_at": "2026-01-28T10:00:00Z"
    },
    {
      "id": 2,
      "title": "任务2",
      "status": "running",
      "created_at": "2026-01-28T10:05:00Z"
    }
  ]
}
```

---

### 4. 执行任务（SSE 流式输出）

### POST /api/v1/tasks/:id/execute

执行指定任务，使用 Server-Sent Events (SSE) 流式返回执行进度。

**参数:**
- `id`: 任务 ID（路径参数）

**请求示例:**
```bash
curl -X POST http://localhost:6767/api/v1/tasks/1/execute
```

**SSE 事件流:**
```
data: {"type":"start","task_id":1,"message":"开始执行任务"}

data: {"type":"progress","task_id":1,"message":"任务执行中"}

data: {"type":"complete","task_id":1,"result":"任务执行完成"}
```

**事件类型:**
- `start`: 任务开始执行
- `progress`: 执行进度更新
- `complete`: 任务执行成功
- `error`: 任务执行失败

---

### 5. 取消任务

### POST /api/v1/tasks/:id/cancel

取消正在运行的任务。

**参数:**
- `id`: 任务 ID（路径参数）

**请求示例:**
```bash
curl -X POST http://localhost:6767/api/v1/tasks/1/cancel
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "task_id": 1,
    "cancelled": true
  }
}
```

---

### 6. 删除任务

### DELETE /api/v1/tasks/:id

删除指定任务。

**参数:**
- `id`: 任务 ID（路径参数）

**请求示例:**
```bash
curl -X DELETE http://localhost:6767/api/v1/tasks/1
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "task_id": 1,
    "deleted": true
  }
}
```

---

## 记忆管理 API

### 1. 搜索记忆

### GET /api/v1/memory/search
### POST /api/v1/memory/search

搜索符合条件的记忆条目。

**查询参数:**
- `q`: 搜索关键词（可选）
- `category`: 记忆分类 (execution/context/result/error/checkpoint)（可选）
- `task_id`: 任务 ID（可选）
- `limit`: 返回数量限制（可选）

**请求示例:**
```bash
curl "http://localhost:6767/api/v1/memory/search?q=关键词&category=result&limit=10"
```

**响应示例:**
```json
{
  "success": true,
  "data": [
    {
      "key": "task_123_result",
      "value": {
        "output": "执行结果数据"
      },
      "expires_at": 1234567890,
      "category": "result",
      "task_id": "task_123",
      "timestamp": 1234567000
    }
  ]
}
```

---

### 2. 获取指定记忆

### GET /api/v1/memory/:key

获取指定键的记忆条目。

**参数:**
- `key`: 记忆键（路径参数）

**请求示例:**
```bash
curl http://localhost:6767/api/v1/memory/my_key
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "key": "my_key",
    "value": {
      "data": "记忆内容"
    },
    "expires_at": 1234567890,
    "category": "context",
    "task_id": null,
    "timestamp": 1234567000
  }
}
```

---

### 3. 删除记忆

### DELETE /api/v1/memory/:key

删除指定的记忆条目。

**参数:**
- `key`: 记忆键（路径参数）

**请求示例:**
```bash
curl -X DELETE http://localhost:6767/api/v1/memory/my_key
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "key": "my_key",
    "deleted": true
  }
}
```

---

### 4. 记忆统计

### GET /api/v1/memory/stats

获取记忆存储的统计信息。

**请求示例:**
```bash
curl http://localhost:6767/api/v1/memory/stats
```

**响应示例:**
```json
{
  "success": true,
  "data": {
    "total": 100,
    "active": 85,
    "expired": 15,
    "category_counts": {
      "execution": 30,
      "context": 25,
      "result": 20,
      "error": 10,
      "checkpoint": 15
    }
  }
}
```

---

## WebSocket API

### 任务实时更新

### WS /ws/task/:id

建立 WebSocket 连接以接收任务的实时更新。

**连接示例 (JavaScript):**
```javascript
const ws = new WebSocket('ws://localhost:6767/ws/task/1');

ws.onopen = () => {
  console.log('WebSocket 连接已建立');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('收到消息:', data);
};
```

**客户端消息格式:**

Ping 消息:
```json
{
  "type": "ping"
}
```

订阅更新:
```json
{
  "type": "subscribe"
}
```

**服务器消息格式:**

连接确认:
```json
{
  "type": "connected",
  "task_id": 1,
  "message": "WebSocket 连接已建立"
}
```

状态更新:
```json
{
  "type": "status",
  "task_id": 1,
  "status": "Running",
  "title": "任务标题"
}
```

进度更新:
```json
{
  "type": "progress",
  "task_id": 1,
  "progress": 20,
  "message": "执行进度: 20%"
}
```

任务完成:
```json
{
  "type": "completed",
  "task_id": 1,
  "result": "任务执行结果"
}
```

任务失败:
```json
{
  "type": "failed",
  "task_id": 1,
  "error": "错误信息"
}
```

---

## 任务状态枚举

| 状态 | 说明 |
|------|------|
| `pending` | 等待执行 |
| `running` | 执行中 |
| `completed` | 已完成 |
| `failed` | 执行失败 |
| `blocked` | 已阻塞/取消 |

## 任务优先级枚举

| 优先级 | 说明 |
|--------|------|
| `low` | 低优先级 |
| `medium` | 中等优先级（默认） |
| `high` | 高优先级 |

## 记忆分类枚举

| 分类 | 说明 |
|------|------|
| `execution` | 执行记录 |
| `context` | 上下文信息 |
| `result` | 执行结果 |
| `error` | 错误信息 |
| `checkpoint` | 检查点 |

---

## 完整使用示例

### 1. 创建并执行任务

```bash
# 1. 创建任务
TASK_RESPONSE=$(curl -s -X POST http://localhost:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -d '{
    "title": "测试任务",
    "description": "这是一个测试任务",
    "priority": "high"
  }')

# 提取任务 ID
TASK_ID=$(echo $TASK_RESPONSE | jq '.data.id')

echo "任务创建成功，ID: $TASK_ID"

# 2. 执行任务（SSE）
curl -X POST http://localhost:6767/api/v1/tasks/$TASK_ID/execute

# 3. 查询任务状态
curl http://localhost:6767/api/v1/tasks/$TASK_ID
```

### 2. WebSocket 连接示例

```python
import asyncio
import websockets
import json

async def monitor_task(task_id):
    uri = f"ws://localhost:6767/ws/task/{task_id}"

    async with websockets.connect(uri) as websocket:
        # 发送 ping
        await websocket.send(json.dumps({"type": "ping"}))

        # 接收消息
        while True:
            message = await websocket.recv()
            data = json.loads(message)

            print(f"收到消息: {data['type']}")

            if data['type'] == 'completed':
                print(f"任务完成: {data.get('result')}")
                break
            elif data['type'] == 'failed':
                print(f"任务失败: {data.get('error')}")
                break

# 运行
asyncio.run(monitor_task(1))
```

### 3. 批量操作示例

```bash
# 批量创建任务
for i in {1..5}; do
  curl -X POST http://localhost:6767/api/v1/tasks \
    -H "Content-Type: application/json" \
    -d "{
      \"title\": \"批量任务 $i\",
      \"description\": \"这是第 $i 个任务\",
      \"group_name\": \"batch\"
    }"
done

# 查看所有任务
curl http://localhost:6767/api/v1/tasks | jq '.data[] | {id, title, status}'
```

---

## 注意事项

1. **并发限制**: 服务器有最大并发任务数限制（默认 10）
2. **任务超时**: 默认任务超时时间为 300 秒
3. **SSE 连接**: 执行任务时使用 SSE，客户端需要支持 Server-Sent Events
4. **WebSocket 心跳**: 建议定期发送 ping 消息保持连接活跃
5. **错误处理**: 所有错误都会返回详细的错误信息

---

## 性能优化建议

1. **批量操作**: 对于大量任务，建议分批创建和执行
2. **连接复用**: HTTP 客户端应复用连接以提高性能
3. **监控**: 定期检查 `/health` 端点以监控服务器状态
4. **缓存**: 对于不常变化的数据，客户端可以缓存结果

---

## 版本历史

### v0.2.0 (当前)
- 初始版本
- HTTP API 支持
- WebSocket 实时更新
- SSE 流式输出
- 记忆管理功能

---

## 支持

如有问题，请查看项目文档或提交 Issue。
