# Zhipu AI Integration Guide

This guide explains how to integrate AgentFlow with Zhipu AI (智谱清言) using webhooks for AI-powered task orchestration.

## Table of Contents

- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [Configuration](#configuration)
- [Setting up Zhipu Webhook](#setting-up-zhipu-webhook)
- [Request/Response Examples](#requestresponse-examples)
- [Security Configuration](#security-configuration)
- [Troubleshooting](#troubleshooting)
- [Advanced Usage](#advanced-usage)

---

## Overview

AgentFlow can be integrated with Zhipu AI to create an AI-powered task orchestration system. When users send messages to Zhipu AI, the platform can forward these requests to AgentFlow via webhooks, enabling:

- Natural language task creation
- Automated task planning and execution
- Real-time status updates back to Zhipu AI
- Multi-agent collaboration

### Architecture

```
┌─────────────┐         ┌─────────────┐         ┌──────────────┐
│ Zhipu AI    │────────>│ AgentFlow   │────────>│ Claude CLI   │
│ Platform    │ Webhook │ Server      │ Execute  │              │
└─────────────┘         └─────────────┘         └──────────────┘
                              │
                              ↓
                        ┌─────────────┐
                        │ Tasks &     │
                        │ Memory      │
                        └─────────────┘
```

---

## Prerequisites

1. **AgentFlow Server Running**: Ensure AgentFlow is running and accessible from the internet
2. **Public URL**: You need a public URL for Zhipu AI to send webhooks to (use ngrok, localtunnel, or deploy to a server)
3. **Zhipu AI Account**: Register at [https://open.bigmodel.cn/](https://open.bigmodel.cn/)
4. **API Key**: Get your Zhipu AI API key from the console

---

## Configuration

### 1. AgentFlow Configuration

Edit `~/.agentflow/config.toml`:

```toml
[server]
host = "0.0.0.0"
port = 6767

[webhook]
# Enable webhook support
enabled = true
# Webhook path (default: /api/v1/webhook)
path = "/api/v1/webhook"
# Optional: Verify webhook signature
secret = "your-webhook-secret-key"

[zhipu]
# Zhipu AI integration settings
enabled = true
# API key for Zhipu AI
api_key = "your-zhipu-api-key"
# Model to use (e.g., "glm-4", "glm-3-turbo")
model = "glm-4"
```

### 2. Environment Variables

Alternatively, you can use environment variables:

```bash
export AGENTFLOW_WEBHOOK_ENABLED=true
export AGENTFLOW_WEBHOOK_SECRET="your-webhook-secret-key"
export AGENTFLOW_ZHIPU_API_KEY="your-zhipu-api-key"
export AGENTFLOW_ZHIPU_MODEL="glm-4"
```

### 3. Start AgentFlow Server

```bash
agentflow server cloud
```

The server will start on `http://localhost:6767` and expose the webhook endpoint at:
```
http://your-public-url:6767/api/v1/webhook
```

---

## Setting up Zhipu Webhook

### Option 1: Using Zhipu AI Console

1. **Login to Zhipu AI Console**
   - Visit [https://open.bigmodel.cn/console](https://open.bigmodel.cn/console)
   - Navigate to "应用管理" (Application Management)

2. **Create or Select Application**
   - Create a new application or select an existing one
   - Go to "回调设置" (Webhook Settings)

3. **Configure Webhook URL**
   ```
   URL: https://your-public-domain.com/api/v1/webhook
   Method: POST
   Content-Type: application/json
   ```

4. **Set Headers (Optional but recommended)**
   ```
   X-Webhook-Secret: your-webhook-secret-key
   ```

5. **Enable Events**
   - ✅ Message received
   - ✅ User query
   - ✅ Task request

### Option 2: Using Zhipu AI API

```bash
curl -X POST https://open.bigmodel.cn/api/paas/v4/apps/{app_id}/webhook \
  -H "Authorization: Bearer $ZHIPU_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://your-public-domain.com/api/v1/webhook",
    "events": ["message.received", "user.query"],
    "secret": "your-webhook-secret-key"
  }'
```

### Getting a Public URL

For testing, you can use tunneling services:

#### Using ngrok

```bash
# Install ngrok
brew install ngrok  # macOS
# or download from https://ngrok.com/

# Start tunnel
ngrok http 6767

# Your public URL will be: https://xxxx-xx-xx-xx-xx.ngrok.io
```

#### Using localtunnel

```bash
# Install localtunnel
npm install -g localtunnel

# Start tunnel
lt --port 6767

# Your public URL will be: https://xxxx-xxx.loca.lt
```

---

## Request/Response Examples

### Webhook Request from Zhipu AI

#### Example 1: Simple Task Creation

```json
{
  "event": "message.received",
  "timestamp": "2026-01-28T10:30:00Z",
  "data": {
    "message_id": "msg_123456789",
    "user_id": "user_abc123",
    "conversation_id": "conv_xyz789",
    "content": "帮我创建一个任务，分析这个项目的代码结构",
    "metadata": {
      "source": "zhipu",
      "model": "glm-4",
      "language": "zh-CN"
    }
  }
}
```

#### Example 2: Complex Task with Parameters

```json
{
  "event": "user.query",
  "timestamp": "2026-01-28T10:35:00Z",
  "data": {
    "message_id": "msg_987654321",
    "user_id": "user_def456",
    "conversation_id": "conv_uvz123",
    "content": "执行以下任务：\n1. 阅读README.md\n2. 总结项目架构\n3. 生成架构图\n\n优先级：高\n分组：文档",
    "parameters": {
      "priority": "high",
      "group": "documentation",
      "timeout": 600
    },
    "metadata": {
      "source": "zhipu",
      "model": "glm-4",
      "reply_to_webhook": "https://api.zhipu.ai/v1/webhook/callback"
    }
  }
}
```

### AgentFlow Response

#### Success Response

```json
{
  "success": true,
  "message": "Task created successfully",
  "data": {
    "task_id": "task_abc123",
    "status": "pending",
    "title": "分析项目代码结构",
    "priority": "high",
    "group": "documentation",
    "estimated_duration": 300,
    "webhook_status_url": "https://your-domain.com/api/v1/tasks/task_abc123/status"
  }
}
```

#### Error Response

```json
{
  "success": false,
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Missing required field: content",
    "details": {
      "field": "content",
      "constraint": "required"
    }
  }
}
```

### Task Status Update Callback

When the task is completed, AgentFlow can send a callback to Zhipu AI:

```json
{
  "event": "task.completed",
  "timestamp": "2026-01-28T10:40:00Z",
  "data": {
    "task_id": "task_abc123",
    "status": "completed",
    "result": "任务完成！\n\n项目架构分析：\n1. 前端：React + TypeScript\n2. 后端：Rust (Axum)\n3. 数据库：SQLite\n...",
    "duration": 45.2,
    "tokens_used": 1250
  }
}
```

---

## Security Configuration

### 1. Webhook Signature Verification

AgentFlow supports HMAC signature verification for webhook requests:

```toml
[webhook]
# Shared secret for signature verification
secret = "your-webhook-secret-key"

# Algorithm to use (sha256, sha512)
algorithm = "sha256"

# Header name containing signature
signature_header = "X-Webhook-Signature"
```

### 2. IP Whitelist

Restrict webhook requests to Zhipu AI's IP addresses:

```toml
[webhook]
# Enable IP whitelist
ip_whitelist_enabled = true
# Zhipu AI IP ranges (example)
ip_whitelist = ["203.119.0.0/16", "43.243.0.0/16"]
```

### 3. Rate Limiting

Protect against webhook floods:

```toml
[webhook]
# Maximum requests per minute
rate_limit = 100
# Burst size
rate_limit_burst = 20
```

### 4. Authentication Headers

Require authentication in webhook requests:

```toml
[webhook]
# Require API key in requests
require_auth = true
# Header name for API key
auth_header = "X-API-Key"
# Expected API key
auth_key = "your-expected-api-key"
```

---

## Troubleshooting

### Issue 1: Webhook Not Receiving Requests

**Symptoms**: No webhook requests are being received by AgentFlow

**Solutions**:
1. Check AgentFlow logs: `agentflow server cloud --log-level debug`
2. Verify public URL is accessible: `curl https://your-public-url:6767/health`
3. Check firewall settings: Ensure port 6767 is open
4. Verify webhook URL in Zhipu AI console

### Issue 2: Signature Verification Failed

**Symptoms**: `403 Forbidden - Invalid signature`

**Solutions**:
1. Verify webhook secret matches in both systems
2. Check signature algorithm matches
3. Ensure timestamp is within tolerance (default: 5 minutes)
4. Check header name for signature

### Issue 3: Task Creation Failed

**Symptoms**: Webhook received but task not created

**Solutions**:
1. Check request body format matches expected schema
2. Verify all required fields are present
3. Check database connection
4. Review logs for detailed error messages

### Issue 4: Callback Not Sent

**Symptoms**: Task completed but Zhipu AI not notified

**Solutions**:
1. Verify callback URL is correct in webhook configuration
2. Check network connectivity from AgentFlow to Zhipu AI
3. Ensure callback URL accepts POST requests
4. Check for firewall rules blocking outbound requests

### Debug Mode

Enable debug logging for troubleshooting:

```bash
agentflow server cloud --log-level debug
```

Check logs for webhook-related messages:
```
[DEBUG] Webhook request received from 203.119.xx.xx
[DEBUG] Request body: {"event":"message.received",...}
[DEBUG] Signature verification: PASSED
[INFO] Task created: task_abc123
```

---

## Advanced Usage

### 1. Custom Webhook Handlers

Create custom webhook handlers by extending AgentFlow:

```rust
// Custom webhook handler in Rust
async fn custom_zhipu_handler(
    payload: ZhipuWebhookPayload,
) -> Result<WebhookResponse> {
    match payload.event.as_str() {
        "message.received" => {
            // Handle incoming message
            let task = create_task_from_message(payload.data).await?;
            Ok(WebhookResponse::success(task))
        },
        "task.query" => {
            // Handle task status query
            let status = get_task_status(payload.data.task_id).await?;
            Ok(WebhookResponse::success(status))
        },
        _ => Ok(WebhookResponse::error("Unknown event"))
    }
}
```

### 2. Multi-Model Support

Configure AgentFlow to work with multiple AI models:

```toml
[zhipu.models]
# Default model
default = "glm-4"

# Model-specific configurations
[zhipu.models.glm-4]
max_tokens = 8192
temperature = 0.7
timeout = 300

[zhipu.models.glm-3-turbo]
max_tokens = 4096
temperature = 0.5
timeout = 180
```

### 3. Conversation Context

Maintain conversation context across multiple webhook calls:

```toml
[memory]
# Enable persistent memory
backend = "sqlite"
database_url = "sqlite://agentflow_memory.db"

# Context retention
context_ttl = 86400  # 24 hours
max_context_length = 10000
```

### 4. Webhook Retry Strategy

Configure retry behavior for failed webhook deliveries:

```toml
[webhook.retry]
# Maximum retry attempts
max_attempts = 3

# Initial retry delay (seconds)
initial_delay = 1

# Backoff multiplier
multiplier = 2

# Maximum delay (seconds)
max_delay = 60
```

---

## Example: Complete Integration Workflow

### Step 1: Start AgentFlow with Webhook Support

```bash
# Set environment variables
export AGENTFLOW_WEBHOOK_ENABLED=true
export AGENTFLOW_WEBHOOK_SECRET="my-secret-key"
export AGENTFLOW_ZHIPU_API_KEY="zhipu-api-key-xxx"

# Start server
agentflow server cloud
```

### Step 2: Setup Tunnel (for testing)

```bash
ngrok http 6767
# Output: Forwarding https://abc123.ngrok.io -> http://localhost:6767
```

### Step 3: Configure Zhipu AI Webhook

Using Zhipu AI console, set webhook URL to:
```
https://abc123.ngrok.io/api/v1/webhook
```

### Step 4: Test Integration

Send a test message through Zhipu AI:

```bash
curl -X POST http://localhost:6767/api/v1/webhook \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Secret: my-secret-key" \
  -d '{
    "event": "message.received",
    "timestamp": "2026-01-28T10:30:00Z",
    "data": {
      "message_id": "test_msg_001",
      "user_id": "test_user",
      "content": "帮我分析这个Go项目的代码结构",
      "metadata": {
        "source": "zhipu"
      }
    }
  }'
```

### Step 5: Monitor Execution

```bash
# Check task status
curl http://localhost:6767/api/v1/tasks/task_xxx

# View logs
tail -f /var/log/agentflow/webhook.log
```

---

## Best Practices

1. **Always use HTTPS** for production webhook URLs
2. **Implement signature verification** to prevent unauthorized requests
3. **Set appropriate timeouts** to prevent hanging requests
4. **Monitor webhook health** with status endpoints
5. **Log all webhook events** for debugging and auditing
6. **Use idempotency keys** to handle duplicate requests
7. **Implement rate limiting** to protect against abuse
8. **Test thoroughly** in staging before production deployment

---

## Additional Resources

- **[AgentFlow Documentation](README.md)**: Main documentation
- **[Configuration Guide](CONFIGURATION.md)**: Complete configuration reference
- **[API Documentation](rust/agentflow-master/API.md)**: REST API reference
- **[Zhipu AI Documentation](https://open.bigmodel.cn/dev/api)**: Zhipu AI API docs

---

## Support

For issues and questions:
- GitHub Issues: [https://github.com/MoSiYuan/AgentFlow/issues](https://github.com/MoSiYuan/AgentFlow/issues)
- Zhipu AI Support: [https://open.bigmodel.cn/support](https://open.bigmodel.cn/support)
