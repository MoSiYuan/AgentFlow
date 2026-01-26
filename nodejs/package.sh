#!/bin/bash

# AgentFlow 打包脚本
# 将所有依赖打包，生成独立可执行文件

set -e

echo "📦 AgentFlow 打包工具"
echo "===================="

# 检查 Node.js 版本
NODE_VERSION=$(node -v | cut -d'v' -f2 | cut -d'.' -f1)
if [ "$NODE_VERSION" -ne 20 ]; then
  echo "❌ 需要 Node.js 20 LTS"
  echo "当前版本: $(node -v)"
  exit 1
fi

echo "✅ Node.js 版本检查通过: $(node -v)"

# 检查是否安装了依赖
if [ ! -d "node_modules" ]; then
  echo "📥 安装依赖..."
  pnpm install
  npm rebuild better-sqlite3
fi

# 构建项目
echo "🔨 构建项目..."
pnpm run build

# 创建输出目录
DIST_DIR="dist/standalone"
mkdir -p "$DIST_DIR"

# 方案1: 使用 pkg 打包
echo ""
echo "📦 方案 1: 使用 pkg 打包独立可执行文件..."

# 检查 pkg 是否安装
if ! command -v pkg &> /dev/null; then
  echo "安装 pkg..."
  npm install -g pkg
fi

# 安装 pkg 作为开发依赖（如果还没有）
if ! grep -q '"pkg"' package.json; then
  npm install --save-dev pkg
fi

# 创建 pkg 配置
cat > pkg.config.json <<EOF
{
  "name": "agentflow-master",
  "version": "1.0.0",
  "main": "packages/master/dist/index.js",
  "bin": "dist/standalone/agentflow-master",
  "pkg": {
    "scripts": [
      "packages/master/dist/**/*.js",
      "packages/worker/dist/**/*.js",
      "packages/database/dist/**/*.js",
      "packages/shared/dist/**/*.js",
      "packages/git-integration/dist/**/*.js",
      "packages/sync/dist/**/*.js",
      "packages/query/dist/**/*.js",
      "node_modules/better-sqlite3/**/*"
    ],
    "assets": [
      "packages/database/src/schema.sql",
      "node_modules/better-sqlite3/build/Release/**/*"
    ],
    "targets": [
      "node20-linux-x64",
      "node20-macos-x64",
      "node20-win-x64"
    ],
    "outputPath": "dist/standalone"
  }
}
EOF

# 打包 Master
echo "  打包 Master..."
pkg packages/master/dist/index.js \
  --targets node20-macos-x64,node20-linux-x64 \
  --output "$DIST_DIR/agentflow-master" \
  -C Brotli \
  --compress GZip

# 打包 Worker
echo "  打包 Worker..."
pkg packages/worker/dist/index.js \
  --targets node20-macos-x64,node20-linux-x64 \
  --output "$DIST_DIR/agentflow-worker" \
  -C Brotli \
  --compress GZip

echo "✅ 独立可执行文件已生成: $DIST_DIR/"

# 方案2: 创建包含所有依赖的 tarball
echo ""
echo "📦 方案 2: 创建完整依赖包..."

BUNDLE_DIR="dist/bundle"
mkdir -p "$BUNDLE_DIR"

# 复制必要的文件
echo "  复制文件..."
cp -r packages/master/dist "$BUNDLE_DIR/master"
cp -r packages/worker/dist "$BUNDLE_DIR/worker"
cp -r packages/database/dist "$BUNDLE_DIR/database"
cp -r packages/shared/dist "$BUNDLE_DIR/shared"
cp -r packages/git-integration/dist "$BUNDLE_DIR/git-integration"
cp -r packages/sync/dist "$BUNDLE_DIR/sync"
cp -r packages/query/dist "$BUNDLE_DIR/query"
cp -r packages/cli/dist "$BUNDLE_DIR/cli"

# 复制 node_modules 中运行时需要的依赖
echo "  复制依赖..."
mkdir -p "$BUNDLE_DIR/node_modules"

# 复制 better-sqlite3（运行时必需）
cp -r node_modules/better-sqlite3 "$BUNDLE_DIR/node_modules/"

# 复制其他运行时依赖
for pkg in express ws; do
  if [ -d "node_modules/$pkg" ]; then
    cp -r "node_modules/$pkg" "$BUNDLE_DIR/node_modules/"
  fi
done

# 创建启动脚本
cat > "$BUNDLE_DIR/start-master.sh" <<'SCRIPT'
#!/bin/bash
cd "$(dirname "$0")"
export NODE_PATH="$(pwd)/node_modules"
node master/dist/index.js "$@"
SCRIPT

cat > "$BUNDLE_DIR/start-worker.sh" <<'SCRIPT'
#!/bin/bash
cd "$(dirname "$0")"
export NODE_PATH="$(pwd)/node_modules"
node worker/dist/index.js "$@"
SCRIPT

chmod +x "$BUNDLE_DIR/start-master.sh"
chmod +x "$BUNDLE_DIR/start-worker.sh"

# 创建 tarball
echo "  创建压缩包..."
cd dist
tar -czf agentflow-bundle-$(uname -s)-$(uname -m).tar.gz bundle/
cd ..

echo "✅ 完整依赖包已生成: $BUNDLE_DIR/"

# 方案3: Docker 镜像
echo ""
echo "📦 方案 3: Docker 镜像..."
echo "  运行: npm run docker:build"
echo "  或查看 deployment/docker/ 目录"

# 完成
echo ""
echo "===================="
echo "✅ 打包完成！"
echo ""
echo "生成的文件:"
echo "  1. 独立可执行文件: $DIST_DIR/agentflow-master"
echo "  2. 完整依赖包: $BUNDLE_DIR/"
echo "  3. Docker 镜像: deployment/docker/"
echo ""
echo "使用方法:"
echo "  # 方案1: 独立可执行文件（推荐）"
echo "  ./dist/standalone/agentflow-master"
echo ""
echo "  # 方案2: 完整依赖包"
echo "  cd dist/bundle"
echo "  ./start-master.sh"
echo ""
echo "  # 方案3: Docker"
echo "  docker run agentflow:latest"
