# Webhook Server 快速开始指南

## 目录
- [快速启动](#快速启动)
- [测试接口](#测试接口)
- [开发指南](#开发指南)
- [常见问题](#常见问题)

## 快速启动

### 1. 设置环境变量

```bash
# 设置 Webhook 密钥（可选，默认为 default_secret）
export WEBHOOK_SECRET="your_secure_secret"
```

### 2. 启动服务器

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master

# 编译并运行
cargo run
```

服务器将在 `http://0.0.0.0:6767` 启动。

### 3. 验证服务状态

```bash
curl http://localhost:6767/health
```

## 测试接口

### 使用测试脚本（推荐）

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow/tests

# 运行完整测试套件
./webhook_test.sh
```

### 手动测试

#### 基本请求

```bash
curl -X POST "http://localhost:6767/api/v1/webhook/zhipu" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_secret" \
  -d '{"text":"创建测试任务"}'
```

#### 带会话 ID 的请求

```bash
curl -X POST "http://localhost:6767/api/v1/webhook/zhipu" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_secret" \
  -d '{
    "text":"执行任务",
    "session_id":"session-123"
  }'
```

## 开发指南

### 添加新的 Webhook 接入点

#### 步骤 1: 创建模块文件

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master/src/webhook

# 创建新的接入点模块
touch chatgpt.rs
```

#### 步骤 2: 实现处理函数

```rust
// chatgpt.rs
use axum::{
    extract::State,
    http::StatusCode,
    Json,
};
use serde::{Deserialize, Serialize};
use crate::AppState;

#[derive(Debug, Deserialize)]
pub struct ChatGPTWebhookRequest {
    pub query: String,
}

#[derive(Debug, Serialize)]
pub struct ChatGPTWebhookResponse {
    pub status: String,
    pub message: String,
}

pub async fn handle_webhook(
    State(_executor): State<AppState>,
    Json(payload): Json<ChatGPTWebhookRequest>,
) -> Result<Json<ChatGPTWebhookResponse>, StatusCode> {
    // 处理逻辑
    Ok(Json(ChatGPTWebhookResponse {
        status: "success".to_string(),
        message: format!("收到查询: {}", payload.query),
    }))
}
```

#### 步骤 3: 注册路由

编辑 `mod.rs`:

```rust
pub mod auth;
pub mod zhipu;
pub mod chatgpt;  // 添加新模块

pub fn create_routes() -> Router<AppState> {
    Router::new()
        .route("/api/v1/webhook/zhipu", post(zhipu::handle_webhook))
        .route("/api/v1/webhook/chatgpt", post(chatgpt::handle_webhook))  // 添加新路由
}
```

#### 步骤 4: 添加单元测试

```rust
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_request_deserialize() {
        let data = serde_json::json!({"query": "test"});
        let req: ChatGPTWebhookRequest = serde_json::from_value(data).unwrap();
        assert_eq!(req.query, "test");
    }
}
```

### 使用鉴权中间件

在路由中添加鉴权层：

```rust
use axum::middleware;

pub fn create_routes() -> Router<AppState> {
    Router::new()
        .route("/api/v1/webhook/zhipu", post(zhipu::handle_webhook))
        .route_layer(middleware::from_fn(auth::webhook_auth_middleware))
}
```

### 运行测试

```bash
cd /Users/jiangxiaolong/work/project/AgentFlow/rust/agentflow-master

# 运行所有测试
cargo test

# 只运行 webhook 模块测试
cargo test webhook

# 显示测试输出
cargo test -- --nocapture

# 运行特定测试
cargo test test_verify_success
```

## API 参考

### 智谱清言 Webhook

**端点**: `POST /api/v1/webhook/zhipu`

**请求头**:
```
Content-Type: application/json
Authorization: Bearer <secret>
```

**请求体**:
```json
{
  "text": "用户指令",
  "session_id": "会话ID（可选）"
}
```

**响应**:
```json
{
  "task_id": "task-uuid",
  "status": "dispatched",
  "message": "收到指令: 用户指令"
}
```

**支持的意图类型**:
- `create_task` - 创建任务（关键词：创建、新建）
- `execute_task` - 执行任务（关键词：执行、运行）
- `query_task` - 查询任务（关键词：查询、查看）
- `delete_task` - 删除任务（关键词：删除、取消）
- `unknown` - 未知指令

## 常见问题

### Q: 服务器启动失败？

**A**: 检查端口是否被占用：
```bash
lsof -i :6767
```

### Q: 请求返回 401 Unauthorized？

**A**: 检查 Authorization 头是否正确：
```bash
# 正确格式
Authorization: Bearer your_secret

# 错误格式
Authorization: your_secret
```

### Q: 如何查看详细日志？

**A**: 设置日志级别：
```bash
export RUST_LOG=debug
cargo run
```

### Q: 如何连接真实数据库？

**A**: 编辑 `.env` 文件：
```
DATABASE_URL=sqlite:///path/to/database.db
```

### Q: 添加新的依赖？

**A**: 编辑 `Cargo.toml`:
```toml
[dependencies]
new_dependency = "1.0"
```

然后运行：
```bash
cargo build
```

## 文件结构

```
rust/agentflow-master/src/
├── webhook/
│   ├── mod.rs          # 模块入口
│   ├── auth.rs         # 鉴权中间件
│   ├── zhipu.rs        # 智谱清言处理
│   └── README.md       # 详细文档
├── lib.rs              # 库入口
└── main.rs             # 主程序

tests/
└── webhook_test.sh     # 集成测试脚本

docs/
├── v0.3.0-phase1-webhook-implementation.md  # 实施总结
└── WEBHOOK_QUICK_START.md                   # 本文档
```

## 下一步

1. 阅读 [完整文档](rust/agentflow-master/src/webhook/README.md)
2. 查看 [实施总结](docs/v0.3.0-phase1-webhook-implementation.md)
3. 运行测试脚本验证功能
4. 开始添加自定义 Webhook 接入点

## 获取帮助

- 提交 Issue: [GitHub Issues](https://github.com/your-repo/issues)
- 查看文档: [docs/](docs/)
- API 文档: 运行 `cargo doc --open`

---

**最后更新**: 2026-01-28
**版本**: v0.3.0 Phase 1
