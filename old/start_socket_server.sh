#!/bin/bash
# CPDS Socket Server启动脚本

SOCKET_SERVER="/Users/jiangxiaolong/work/project/game/AdStar/cpds-go/bin/socket-server"
SOCKET_PATH="/tmp/cpds-claude.sock"
LOG_FILE="/tmp/cpds_socket_server.log"

echo "═══════════════════════════════════════════════════════════"
echo "     🚀 CPDS Claude Socket Server 启动脚本"
echo "═══════════════════════════════════════════════════════════"
echo ""

# 检查是否已有服务器在运行
if [ -S "$SOCKET_PATH" ]; then
  echo "⚠️  检测到socket文件已存在: $SOCKET_PATH"
  echo "   正在尝试关闭旧服务器..."
  rm -f "$SOCKET_PATH"
  sleep 1
fi

# 检查端口占用
PID=$(lsof -t "$SOCKET_PATH" 2>/dev/null || echo "")
if [ -n "$PID" ]; then
  echo "⚠️  检测到旧服务器进程(PID: $PID)"
  echo "   正在关闭..."
  kill $PID 2>/dev/null || true
  sleep 1
fi

echo "📡 启动Socket服务器..."
echo "   Socket路径: $SOCKET_PATH"
echo "   日志文件: $LOG_FILE"
echo ""

# 启动服务器（后台运行）
nohup "$SOCKET_SERVER" > "$LOG_FILE" 2>&1 &
SERVER_PID=$!

echo "✅ Socket服务器已启动 (PID: $SERVER_PID)"
echo ""

# 等待服务器启动
echo "⏳ 等待服务器初始化..."
for i in {1..10}; do
  if [ -S "$SOCKET_PATH" ]; then
    echo "✅ Socket已创建: $SOCKET_PATH"
    break
  fi
  sleep 0.5
done

# 验证服务器是否正常运行
sleep 1
if ps -p $SERVER_PID > /dev/null 2>&1; then
  echo "✅ 服务器运行正常"
  echo ""
  echo "═══════════════════════════════════════════════════════════"
  echo "                  服务器信息"
  echo "═══════════════════════════════════════════════════════════"
  echo "PID:      $SERVER_PID"
  echo "Socket:   $SOCKET_PATH"
  echo "Log:      $LOG_FILE"
  echo ""
  echo "📝 查看日志:"
  echo "   tail -f $LOG_FILE"
  echo ""
  echo "🛑 停止服务器:"
  echo "   kill $SERVER_PID"
  echo "   或"
  echo "   rm -f $SOCKET_PATH"
  echo "═══════════════════════════════════════════════════════════"
else
  echo "❌ 服务器启动失败"
  echo ""
  echo "查看日志:"
  cat "$LOG_FILE"
  exit 1
fi
