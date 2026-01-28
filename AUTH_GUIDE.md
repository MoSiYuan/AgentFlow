# AgentFlow 双认证系统使用指南

AgentFlow 支持两种认证方式，分别用于不同的场景：

1. **用户 Session 认证**：用于前端用户访问
2. **API Key 认证**：用于 Master 节点之间的通信

## 认证方式对比

| 特性 | 用户 Session | API Key |
|------|------------|---------|
| **用途** | 前端用户登录 | Master 之间通信 |
| **Token 格式** | UUID | `sk_{timestamp}_{signature}` |
| **存储位置** | localStorage | 环境变量/配置文件 |
| **有效期** | 24 小时（可配置） | 5 分钟（动态生成） |
| **发送方式** | `Authorization: Bearer {session_id}` | `Authorization: Bearer {api_key}` |

## 1. 用户 Session 认证

### 前端登录流程

1. **首次访问**
   - 页面加载时检查 `localStorage` 中的 `agentflow_session_id`
   - 如果不存在，弹出原生 `prompt` 弹窗要求输入用户名和密码
   - 登录成功后存储 Session ID

2. **API 请求**
   - 所有 API 请求自动携带 `Authorization: Bearer {session_id}` 头
   - 后端验证 Session 有效性
   - Session 过期或无效时自动刷新页面重新登录

3. **登出**
   - 清除 `localStorage` 中的 Session ID
   - 刷新页面重新登录

### 登录示例

```javascript
// 自动登录（App.tsx 中自动处理）
await ensureAuthenticated();

// 手动登录
const { login, getSessionId } = await import('./utils/auth');
const result = await login();
console.log(result.session_id);

// 获取当前 Session
const sessionId = getSessionId();
```

### 环境变量配置

```bash
# 启用认证
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your_secure_password
AUTH_SESSION_TTL=86400  # 24 小时
```

## 2. API Key 认证

### API Key 格式

```
sk_{timestamp}_{signature}
```

- **`sk_`**: 固定前缀，表示 API Key
- **`timestamp`**: Unix 时间戳（秒）
- **`signature`**: HMAC-SHA256 签名（32 位十六进制）

### 签名生成算法

```rust
use hmac::{Hmac, Mac};
use sha2::Sha256;

let mut mac = Hmac::<Sha256>::new_from_slice(api_key_secret.as_bytes())?;
mac.update(timestamp.to_string().as_bytes());
let signature = mac.finalize().into_bytes();
let signature_hex = hex::encode(signature);
```

### 环境变量配置

```bash
# 设置 API Key 密钥（必须配置）
AUTH_API_KEY_SECRET=your_very_secret_key_here
```

**重要**:
- `AUTH_API_KEY_SECRET` 必须保密，不要提交到代码仓库
- 建议使用至少 32 位的随机字符串
- 定期轮换密钥

### 生成 API Key

#### 方式 1: 使用命令行工具

```bash
# 生成 API Key
cd rust
cargo run --bin generate-api-key --secret your_secret_here
```

#### 方式 2: 使用 Rust 代码

```rust
use agentflow_master::config::AuthConfig;

let config = AuthConfig {
    api_key_secret: "your_secret_here".to_string(),
    ..Default::default()
};

let api_key = config.generate_api_key()?;
println!("API Key: {}", api_key);
```

#### 方式 3: 使用在线工具

```bash
# 生成时间戳
timestamp=$(date +%s)

# 生成签名（需要 openssl）
signature=$(echo -n "$timestamp" | openssl dgst -sha256 -hmac "your_secret_here" | awk '{print $2}')

# 组合 API Key
api_key="sk_${timestamp}_${signature}"
echo "API Key: $api_key"
```

### 使用 API Key

#### Master A 调用 Master B 的 API

```bash
# 生成 API Key
API_KEY="sk_1706451200_a1b2c3d4e5f6..."

# 调用 API
curl -X POST http://master-b:6767/api/v1/tasks \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $API_KEY" \
  -d '{
    "title": "Remote Task",
    "description": "Task from Master A"
  }'
```

#### Rust 代码示例

```rust
use reqwest::Client;

let api_key = "sk_1706451200_a1b2c3d4e5f6...";
let response = Client::new()
    .post("http://master-b:6767/api/v1/tasks")
    .header("Authorization", format!("Bearer {}", api_key))
    .json(&task_data)
    .send()
    .await?;
```

### API Key 有效期

- **有效期**: 5 分钟（从生成时间戳计算）
- **原因**: 短期有效减少泄露风险
- **建议**: 每次请求前动态生成新的 API Key

### API Key 验证逻辑

1. **格式检查**: 必须以 `sk_` 开头
2. **时间戳验证**: 必须在 5 分钟内生成
3. **签名验证**: 使用 `AUTH_API_KEY_SECRET` 重新计算签名并比对

## 3. 完整配置示例

### .env 文件

```bash
# ==================== 认证配置 ====================

# 启用认证
AUTH_ENABLED=true

# 用户认证（前端登录）
AUTH_USERNAME=admin
AUTH_PASSWORD=your_secure_password
AUTH_SESSION_TTL=86400  # 24 小时

# API Key 认证（Master 之间）
AUTH_API_KEY_SECRET=your_very_secret_key_at_least_32_chars

# ==================== 服务器配置 ====================
AGENTFLOW_SERVER_PORT=6767
AGENTFLOW_LOG_LEVEL=info
```

### Systemd 服务文件

```ini
[Unit]
Description=AgentFlow Master Server
After=network.target

[Service]
Type=simple
User=agentflow
Group=agentflow
WorkingDirectory=/opt/agentflow
Environment="AUTH_ENABLED=true"
Environment="AUTH_USERNAME=admin"
Environment="AUTH_PASSWORD=your_password"
Environment="AUTH_API_KEY_SECRET=your_api_key_secret"
Environment="AUTH_SESSION_TTL=86400"
ExecStart=/opt/agentflow/agentflow-master
Restart=always

[Install]
WantedBy=multi-user.target
```

### Docker Compose

```yaml
version: '3.8'

services:
  agentflow-master:
    image: agentflow-master:latest
    environment:
      - AUTH_ENABLED=true
      - AUTH_USERNAME=admin
      - AUTH_PASSWORD=${AUTH_PASSWORD}
      - AUTH_API_KEY_SECRET=${AUTH_API_KEY_SECRET}
      - AUTH_SESSION_TTL=86400
    ports:
      - "6767:6767"
```

## 4. 安全建议

### 生产环境配置

1. **强制启用认证**
   ```bash
   AUTH_ENABLED=true
   ```

2. **使用强密码**
   - 至少 12 位
   - 包含大小写字母、数字、特殊字符
   - 定期更换

3. **保护 API Key Secret**
   - 使用环境变量或密钥管理服务
   - 不要写入代码或配置文件
   - 使用至少 32 位的随机字符串
   ```bash
   # 生成随机密钥
   openssl rand -hex 32
   ```

4. **限制访问**
   - 配置防火墙规则
   - 使用反向代理（Nginx/Apache）
   - 启用 HTTPS

5. **日志监控**
   - 记录所有认证失败尝试
   - 监控异常登录行为
   - 定期审计访问日志

### 密钥管理

```bash
# 生成强密码
openssl rand -base64 24

# 生成 API Key Secret
openssl rand -hex 32

# 安全存储（使用 systemd 凭证）
sudo systemctl edit agentflow-master

[Service]
Environment="AUTH_API_KEY_SECRET=your_secret"
```

## 5. 故障排除

### 问题 1: 401 Unauthorized

**症状**: API 请求返回 401 错误

**原因**:
1. 认证未启用但未提供 Token
2. Session 过期
3. API Key 格式错误或过期
4. 签名验证失败

**解决**:
```bash
# 检查认证是否启用
curl http://localhost:6767/health | jq '.auth_enabled'

# 检查 Session 是否有效
curl -H "Authorization: Bearer your_session" \
  http://localhost:6767/api/v1/tasks

# 检查 API Key 是否有效
curl -H "Authorization: Bearer your_api_key" \
  http://localhost:6767/api/v1/tasks
```

### 问题 2: API Key 生成失败

**症状**: 无法生成 API Key

**原因**: `AUTH_API_KEY_SECRET` 未配置

**解决**:
```bash
# 设置环境变量
export AUTH_API_KEY_SECRET="your_secret_here"

# 或在 .env 文件中配置
echo "AUTH_API_KEY_SECRET=your_secret_here" >> .env
```

### 问题 3: 前端登录弹窗不显示

**症状**: 页面直接显示内容，没有登录弹窗

**原因**: 认证未启用

**解决**:
```bash
# 检查认证配置
curl http://localhost:6767/health | jq '.auth_enabled'

# 如果返回 false，启用认证
export AUTH_ENABLED=true
./agentflow-master
```

## 6. 最佳实践

### 前端开发

1. **使用 `authenticatedFetch`**
   - 所有需要认证的 API 请求都应使用
   - 自动处理 401 错误和 Session 过期

2. **检查认证状态**
   - 在关键操作前检查 `isAuthenticated()`
   - 未认证时调用 `ensureAuthenticated()`

3. **错误处理**
   - 捕获登录失败错误
   - 提供用户友好的错误提示

### Master 间通信

1. **动态生成 API Key**
   - 每次请求前生成新的 API Key
   - 不要硬编码 API Key

2. **使用安全的存储**
   - API Key Secret 存储在环境变量
   - 不要写入日志或错误消息

3. **重试机制**
   - API Key 过期时重新生成
   - 实现指数退避重试

### 部署建议

1. **开发环境**
   - 可以不启用认证（`AUTH_ENABLED=false`）
   - 简化本地开发和测试

2. **生产环境**
   - 必须启用认证
   - 使用强密码
   - 配置 API Key Secret
   - 启用 HTTPS

3. **密钥轮换**
   - 定期更换 `AUTH_API_KEY_SECRET`
   - 通知所有 Master 节点更新密钥
   - 使用滚动更新避免服务中断

## 7. 代码示例

### 前端完整示例

```typescript
import { authenticatedFetch, ensureAuthenticated } from './utils/auth';

async function createTask() {
  // 确保已登录
  await ensureAuthenticated();

  // 创建任务（自动携带认证头）
  const response = await authenticatedFetch('/api/v1/tasks', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      title: 'New Task',
      description: 'Task description'
    }),
  });

  const data = await response.json();
  console.log('Task created:', data);
}
```

### Master 间通信完整示例

```rust
use reqwest::Client;
use agentflow_master::config::AuthConfig;

async fn call_remote_master(
    remote_url: &str,
    auth_config: &AuthConfig,
) -> Result<(), Box<dyn std::error::Error>> {
    // 生成 API Key
    let api_key = auth_config.generate_api_key()?;

    // 调用远程 Master API
    let client = Client::new();
    let response = client
        .post(&format!("{}/api/v1/tasks", remote_url))
        .header("Authorization", format!("Bearer {}", api_key))
        .json(&task_data)
        .send()
        .await?;

    println!("Response: {:?}", response);

    Ok(())
}
```

## 8. 测试脚本

### 测试用户登录

```bash
# 1. 登录获取 Session
SESSION=$(curl -s -X POST http://localhost:6767/api/v1/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin"}' \
  | jq -r '.session_id')

echo "Session ID: $SESSION"

# 2. 使用 Session 访问 API
curl http://localhost:6767/api/v1/tasks \
  -H "Authorization: Bearer $SESSION"
```

### 测试 API Key

```bash
# 1. 生成 API Key
TIMESTAMP=$(date +%s)
SIGNATURE=$(echo -n "$TIMESTAMP" | openssl dgst -sha256 -hmac "your_secret" | awk '{print $2}')
API_KEY="sk_${TIMESTAMP}_${SIGNATURE}"

echo "API Key: $API_KEY"

# 2. 使用 API Key 访问 API
curl http://localhost:6767/api/v1/tasks \
  -H "Authorization: Bearer $API_KEY"
```

## 总结

AgentFlow 的双认证系统提供了：

✅ **用户认证**: 简单的用户名密码登录，使用原生弹窗
✅ **API 认证**: 基于 HMAC-SHA256 的 API Key，用于 Master 间通信
✅ **灵活配置**: 通过环境变量轻松配置
✅ **安全性**: Session 过期机制，API Key 短期有效
✅ **易用性**: 前端自动处理认证，后端统一验证

---

**最后更新**: 2026-01-28
**文档版本**: 1.0.0
