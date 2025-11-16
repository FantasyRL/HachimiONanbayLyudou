# 此脚本用于在容器内启动服务

#! /usr/bin/env bash
set -euo pipefail
# 清除代理环境变量，避免影响服务运行
unset HTTP_PROXY HTTPS_PROXY http_proxy https_proxy ALL_PROXY all_proxy NO_PROXY no_proxy

CURDIR=$(pwd)

# 日志目录
export KITEX_RUNTIME_ROOT="$CURDIR"
export KITEX_LOG_DIR="$CURDIR/log"

mkdir -p "$KITEX_LOG_DIR/app" "$KITEX_LOG_DIR/rpc"

: "${ETCD_ADDR:=localhost:2379}"
export ETCD_ADDR
: "${CONSUL_ADDR:=localhost:8500}"
export CONSUL_ADDR

# 允许通过环境变量覆盖配置目录（默认 /app/config）
: "${CONFIG_PATH:=/app/config}"

# 该环境变量由 Dockerfile/Makefile 传入
: "${SERVICE:?SERVICE not set}"

BIN="$CURDIR/output/$SERVICE/hachimi-$SERVICE"

if [ ! -x "$BIN" ]; then
  echo "Binary not found or not executable: $BIN"
  ls -l "$CURDIR/output/$SERVICE" || true
  exit 1
fi

# 保证配置文件路径存在
CFG_FILE="$CONFIG_PATH/config.yaml"
if [ ! -f "$CFG_FILE" ]; then
  echo "WARNING: Config file not found: $CFG_FILE"
fi

# 可以通过 ARGS 追加自定义启动参数：
#   docker run ... -e ARGS="--debug --some-flag=1"
: "${ARGS:=}"

echo "==> Starting $SERVICE"
exec "$BIN" -cfg "$CFG_FILE" $ARGS
