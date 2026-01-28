# 构建阶段
FROM rust:1.75-alpine AS builder

# 安装构建依赖
RUN apk add --no-cache musl-dev sqlite-dev gcc

# 设置工作目录
WORKDIR /build

# 复制 Cargo 配置
COPY Cargo.toml Cargo.lock ./
COPY agentflow-core ./agentflow-core
COPY agentflow-master ./agentflow-master

# 构建（使用预编译依赖加速）
RUN cargo build --release && \
    mv target/release/agentflow-master /tmp/ && \
    cargo clean

# 运行阶段
FROM alpine:latest

# 安装运行时依赖
RUN apk add --no-cache \
    sqlite \
    ca-certificates \
    && addgroup -g 1000 agentflow && \
    adduser -D -u 1000 -G agentflow agentflow

# 创建目录
RUN mkdir -p /var/lib/agentflow /var/log/agentflow && \
    chown -R agentflow:agentflow /var/lib/agentflow /var/log/agentflow

# 复制二进制
COPY --from=builder /tmp/agentflow-master /opt/agentflow/bin/

# 创建非 root 用户
USER agentflow

# 工作目录
WORKDIR /opt/agentflow

# 暴露端口
EXPOSE 6767 8849

# 健康检查
HEALTHCHECK --interval=30s --timeout=3s \
  CMD wget --no-verbose --tries=1 --spider http://localhost:6767/health || exit 1

# 启动服务
CMD ["/opt/agentflow/bin/agentflow-master"]
