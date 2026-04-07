# GitHub Actions 自动构建 Docker 镜像

## 功能

每次推送到 `main` 分支或打标签时，GitHub Actions 会自动：
1. 构建 Docker 镜像（支持 AMD64 和 ARM64）
2. 推送到 GitHub Container Registry (ghcr.io)
3. 生成多版本标签

## 镜像地址

构建完成后，镜像地址为：
```
ghcr.io/michae1g/glance:latest
```

## 在 NAS 上使用

### 1. 登录 GitHub Container Registry
```bash
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

或者使用 Personal Access Token (PAT)：
```bash
docker login ghcr.io -u Michae1G
# 输入你的 GitHub Personal Access Token
```

### 2. 拉取并运行
```bash
# 拉取镜像
docker pull ghcr.io/michae1g/glance:latest

# 运行
docker run -d \
  --name glance \
  -p 8080:8080 \
  -e QWEATHER_API_KEY=你的APIKey \
  -v $(pwd)/config:/app/config \
  ghcr.io/michae1g/glance:latest
```

### 3. Docker Compose
```yaml
version: '3.8'

services:
  glance:
    image: ghcr.io/michae1g/glance:latest
    container_name: glance
    ports:
      - "8080:8080"
    environment:
      - QWEATHER_API_KEY=${QWEATHER_API_KEY}
    volumes:
      - ./config:/app/config
    restart: unless-stopped
```

## 配置 GitHub Actions

### 启用权限
1. 进入你的 fork 仓库
2. 点击 **Settings** → **Actions** → **General**
3. 确保 **Workflow permissions** 设置为 **Read and write permissions**
4. 勾选 **Allow GitHub Actions to create and approve pull requests**（可选）

### 启用 GitHub Packages
1. 点击 **Settings** → **Packages**
2. 确保 **Inherit access from source repository** 已启用

## 触发构建

### 自动触发
- 推送到 `main` 分支 → 构建 `latest` 标签
- 推送标签 `v1.0.0` → 构建 `1.0.0`、`1.0`、`1` 标签

### 手动触发
1. 进入仓库 **Actions** 标签
2. 选择 **Build and Push Docker Image**
3. 点击 **Run workflow**

## 查看镜像

构建完成后，在 GitHub 上查看：
1. 点击仓库主页右侧的 **Packages**
2. 或者访问: `https://github.com/Michae1G?tab=packages`

## 标签说明

| 标签 | 说明 |
|------|------|
| `latest` | 最新 main 分支构建 |
| `main` | main 分支最新提交 |
| `v1.0.0` | 具体版本号 |
| `1.0` | 次要版本 |
| `1` | 主版本 |
| `sha-xxxx` | 提交哈希 |

## 故障排查

### 权限错误
确保 GitHub Token 有 `packages:write` 权限。

### 镜像拉取失败
检查镜像是否公开：
1. 进入 **Packages**
2. 点击镜像名称
3. 点击 **Package settings**
4. 在 **Danger Zone** 中设置为 **Public**

## 优势

- ✅ 无需本地构建，节省 NAS 资源
- ✅ 自动多架构支持（AMD64/ARM64）
- ✅ 版本标签管理
- ✅ 构建缓存加速
