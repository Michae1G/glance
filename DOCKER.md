# Glance 优化版 Docker 使用指南

## 构建镜像

### Linux/macOS
```bash
chmod +x build-docker.sh
./build-docker.sh
```

### Windows PowerShell
```powershell
.\build-docker.ps1
```

### 手动构建
```bash
docker build -t glance-optimized:latest .
```

## 运行容器

### 基本运行
```bash
docker run -d \
  --name glance \
  -p 8080:8080 \
  -v $(pwd)/config:/app/config \
  glance-optimized:latest
```

### 带环境变量（推荐）
```bash
docker run -d \
  --name glance \
  -p 8080:8080 \
  -e QWEATHER_API_KEY=你的APIKey \
  -v $(pwd)/config:/app/config \
  glance-optimized:latest
```

### Windows PowerShell
```powershell
docker run -d `
  --name glance `
  -p 8080:8080 `
  -e QWEATHER_API_KEY="你的APIKey" `
  -v "${PWD}/config:/app/config" `
  glance-optimized:latest
```

## 验证步骤

### 1. 检查镜像
```bash
docker images glance-optimized:latest
```

### 2. 检查容器运行状态
```bash
docker ps
```

### 3. 查看日志
```bash
docker logs glance
```

### 4. 访问测试
打开浏览器访问: http://localhost:8080

### 5. 检查容器详情
```bash
docker inspect glance
```

## 常用命令

```bash
# 停止容器
docker stop glance

# 删除容器
docker rm glance

# 重启容器
docker restart glance

# 进入容器
docker exec -it glance sh

# 查看资源使用
docker stats glance
```

## Docker Compose（推荐）

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  glance:
    image: glance-optimized:latest
    container_name: glance
    ports:
      - "8080:8080"
    environment:
      - QWEATHER_API_KEY=${QWEATHER_API_KEY}
    volumes:
      - ./config:/app/config
    restart: unless-stopped
```

运行:
```bash
# 创建 .env 文件存放 API Key
echo "QWEATHER_API_KEY=你的APIKey" > .env

# 启动
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止
docker-compose down
```

## 故障排查

### 容器无法启动
```bash
# 查看详细日志
docker logs glance 2>&1

# 检查配置文件是否存在
ls -la config/glance.yml
```

### 端口被占用
```bash
# 更换端口映射
docker run -d -p 8081:8080 ...
```

### 权限问题
```bash
# 确保配置文件可读
chmod 644 config/glance.yml
```
