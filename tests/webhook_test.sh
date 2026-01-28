#!/bin/bash
# Webhook 接口测试脚本
# 用于测试智谱清言 Webhook 接入点

set -e

# Webhook URL
WEBHOOK_URL="${WEBHOOK_URL:-http://localhost:6767/api/v1/webhook/zhipu}"

# 颜色输出
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "========================================"
echo "AgentFlow Webhook 接口测试"
echo "========================================"
echo ""

# 测试 1: 基本请求测试
echo -e "${YELLOW}测试 1: 发送基本 Webhook 请求${NC}"
echo "请求内容: {\"text\":\"跑个测试\"}"
echo ""

response=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_secret" \
  -d '{"text":"跑个测试"}')

echo "响应内容:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"
echo ""

# 测试 2: 带 session_id 的请求
echo -e "${YELLOW}测试 2: 带 session_id 的请求${NC}"
echo "请求内容: {\"text\":\"创建测试任务\",\"session_id\":\"session-123\"}"
echo ""

response=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_secret" \
  -d '{"text":"创建测试任务","session_id":"session-123"}')

echo "响应内容:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"
echo ""

# 测试 3: 不同指令类型
echo -e "${YELLOW}测试 3: 不同指令类型测试${NC}"

commands=(
  "执行任务"
  "查询任务状态"
  "删除任务"
  "随便说点什么"
)

for cmd in "${commands[@]}"; do
  echo ""
  echo "指令: $cmd"
  response=$(curl -s -X POST "$WEBHOOK_URL" \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer test_secret" \
    -d "{\"text\":\"$cmd\"}")
  echo "$response" | jq -r '.message' 2>/dev/null || echo "$response"
done

echo ""
echo ""

# 测试 4: 鉴权测试（无 Authorization 头）
echo -e "${YELLOW}测试 4: 鉴权测试（无 Authorization 头）${NC}"
echo "预期: 请求应该成功（MVP 阶段未强制鉴权）"
echo ""

response=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -d '{"text":"无鉴权测试"}')

echo "响应内容:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"
echo ""

# 测试 5: 错误请求测试
echo -e "${YELLOW}测试 5: 错误请求测试（缺少必需字段）${NC}"
echo "预期: 返回 400 Bad Request"
echo ""

response=$(curl -s -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_secret" \
  -d '{"invalid_field":"test"}')

echo "响应状态: $?"
echo "响应内容:"
echo "$response" | jq '.' 2>/dev/null || echo "$response"
echo ""

# 测试 6: 详细输出测试（带 -v 参数）
echo -e "${YELLOW}测试 6: 详细输出测试${NC}"
echo "展示完整的 HTTP 请求和响应头"
echo ""

curl -X POST "$WEBHOOK_URL" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer test_secret" \
  -d '{"text":"详细测试"}' \
  -v

echo ""
echo ""
echo "========================================"
echo -e "${GREEN}测试完成！${NC}"
echo "========================================"
echo ""
echo "注意事项:"
echo "1. MVP 阶段鉴权未强制执行，后续会加强"
echo "2. 当前返回的是模拟数据，未实际创建任务"
echo "3. 请确保服务器正在运行: $WEBHOOK_URL"
echo ""
