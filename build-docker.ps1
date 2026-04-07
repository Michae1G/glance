# Glance 优化版 Docker 构建脚本 (PowerShell)

Write-Host "=== 构建 Glance 优化版 Docker 镜像 ===" -ForegroundColor Green

# 构建镜像
docker build -t glance-optimized:latest .

if ($LASTEXITCODE -ne 0) {
    Write-Host "构建失败！" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=== 构建完成 ===" -ForegroundColor Green

# 显示镜像信息
Write-Host ""
Write-Host "=== 镜像信息 ===" -ForegroundColor Cyan
$image = docker images glance-optimized:latest --format "table {{.Repository}}\t{{.Tag}}\t{{.Size}}\t{{.CreatedAt}}"
Write-Host $image

# 获取镜像 ID
$imageId = docker images glance-optimized:latest --format "{{.ID}}"
Write-Host ""
Write-Host "镜像 ID: $imageId"

Write-Host ""
Write-Host "=== 运行测试命令 ===" -ForegroundColor Yellow
Write-Host @"
docker run -d `
  --name glance-test `
  -p 8080:8080 `
  -e QWEATHER_API_KEY=你的APIKey `
  -v "`$(pwd)/config:/app/config" `
  glance-optimized:latest
"@

Write-Host ""
Write-Host "=== 验证步骤 ===" -ForegroundColor Green
Write-Host "1. 检查容器运行状态: docker ps"
Write-Host "2. 查看日志: docker logs glance-test"
Write-Host "3. 访问测试: http://localhost:8080"
Write-Host "4. 停止容器: docker stop glance-test && docker rm glance-test"
