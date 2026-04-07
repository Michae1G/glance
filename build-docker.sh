#!/bin/bash
# Glance 优化版 Docker 构建脚本

set -e

echo "=== 构建 Glance 优化版 Docker 镜像 ==="

# 构建镜像
docker build -t glance-optimized:latest .

echo ""
echo "=== 构建完成 ==="
docker images glance-optimized:latest

echo ""
echo "=== 镜像信息 ==="
docker inspect glance-optimized:latest --format='Size: {{.Size}} bytes ({{humanSize .Size}})'

echo ""
echo "=== 运行测试 ==="
echo "运行命令:"
echo "  docker run -d \\"
echo "    --name glance-test \\"
echo "    -p 8080:8080 \\"
echo "    -e QWEATHER_API_KEY=你的APIKey \\"
echo "    -v \$(pwd)/config:/app/config \\"
echo "    glance-optimized:latest"
