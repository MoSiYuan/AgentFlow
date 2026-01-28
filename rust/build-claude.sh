#!/bin/bash
set -e

export PATH="/Users/jiangxiaolong/.cargo/bin:$PATH"

cd /Users/jiangxiaolong/work/project/AgentFlow/rust

echo "=== 构建 AgentFlow Master with Claude CLI 集成 ==="
cargo build --release

echo "=== 构建完成 ==="
