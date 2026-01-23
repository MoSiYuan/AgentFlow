# AgentFlow 二进制文件

这些是已编译的二进制文件，可直接使用。

## 文件说明

- `agentflow` - 主程序（原 CPDS）
- `master` - Master 服务器
- `worker` - Worker 客户端

## 快速使用

```bash
# 启动 Master
./bin/master --mode standalone --auto-shutdown

# 启动 Worker
./bin/worker --mode standalone --master http://localhost:6767

# 或使用 agentflow
./bin/agentflow master --mode standalone
```

## 性能

- HTTP 吞吐量: 10,000+ req/s
- 内存使用: ~20MB
- 启动时间: <100ms
- 二进制大小: 34MB
