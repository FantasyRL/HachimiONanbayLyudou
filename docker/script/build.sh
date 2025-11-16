# 此脚本负责构建二进制文件（由 Dockerfile 调用）

#!/usr/bin/env bash
# 该脚本负责构建二进制文件（由 Dockerfile 调用）
# Usage: ./docker/script/build.sh {SERVICE}

set -euo pipefail

RUN_NAME="${1:-}"
ROOT_DIR=$(pwd)

if [ -z "$RUN_NAME" ]; then
  echo "Error: Service name is required."
  exit 1
fi

echo "==> Building service: ${RUN_NAME}"

# 进入对应服务模块
cd "./cmd/${RUN_NAME}"

# 产物目录：/app/output/{SERVICE}
mkdir -p "${ROOT_DIR}/output/${RUN_NAME}"

# 正常构建（如需系统测试，改用 go test -c）
go build -o "${ROOT_DIR}/output/${RUN_NAME}/hachimi-${RUN_NAME}"

echo "==> Done. Binary at /app/output/${RUN_NAME}/hachimi-${RUN_NAME}"
