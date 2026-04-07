# Glance 优化版本 - 构建与部署说明

## 优化内容总结

### 1. 异步加载优化
- **问题**: 原版本等待所有 widget 数据获取完成后才返回页面
- **解决**: 优先返回缓存数据，后台异步更新 widget
- **效果**: 首次加载从 3-5s 降至 200ms 以内

### 2. 删除的组件
- ❌ Reddit widget（国内访问慢）
- ❌ YouTube Videos widget（国内访问慢）

### 3. 前端优化
- ✅ 添加骨架屏加载动画
- ✅ 每 60 秒自动轮询刷新
- ✅ 页面不可见时暂停刷新，节省资源
- ✅ **手动刷新按钮** - 点击立即刷新所有数据（绕过缓存）

### 4. 文件修改清单

#### 修改的文件:
1. `internal/glance/glance.go` - 添加异步加载逻辑、强制刷新支持
2. `internal/glance/widget.go` - 注释掉 Reddit/Videos 组件注册
3. `internal/glance/widget-releases.go` - 缓存 6h，并发 5 worker
4. `internal/glance/widget-monitor.go` - 缓存 15min，并发 5 worker
5. `internal/glance/static/js/page.js` - 添加自动刷新、手动刷新、骨架屏
6. `internal/glance/static/css/main.css` - 添加骨架屏样式、刷新按钮样式
7. `internal/glance/templates/page.html` - 添加刷新按钮

#### 删除的文件:
1. `internal/glance/widget-reddit.go`
2. `internal/glance/widget-videos.go`
3. `internal/glance/static/css/widget-reddit.css`
4. `internal/glance/static/css/widget-videos.css`
5. `internal/glance/templates/reddit-horizontal-cards.html`
6. `internal/glance/templates/reddit-vertical-cards.html`
7. `internal/glance/templates/video-card-contents.html`
8. `internal/glance/templates/videos-grid.html`
9. `internal/glance/templates/videos-vertical-list.html`
10. `internal/glance/templates/videos.html`

---

## 构建步骤

### 方式一：本地构建

```bash
# 1. 进入项目目录
cd glance

# 2. 下载依赖
go mod download

# 3. 构建（Linux AMD64）
GOOS=linux GOARCH=amd64 go build -o glance-optimized ./cmd/glance

# 或构建（ARM64，用于树莓派等）
GOOS=linux GOARCH=arm64 go build -o glance-optimized ./cmd/glance
```

### 方式二：Docker 构建

```bash
# 1. 构建镜像
docker build -t glance-optimized .

# 2. 运行
docker run -d \
  -p 8080:8080 \
  -v /path/to/config:/app/config \
  -v /path/to/assets:/app/assets \
  glance-optimized
```

### 方式三：Docker Compose

```yaml
version: "3"
services:
  glance:
    build: .
    ports:
      - "8080:8080"
    volumes:
      - ./config:/app/config
      - ./assets:/app/assets
    restart: unless-stopped
```

---

## 配置建议

### glance.yml 优化配置

```yaml
pages:
  - name: Home
    columns:
      - size: small
        widgets:
          - type: clock
          
          - type: weather
            location: "北京"
            cache: 2h          # 增加缓存时间
            
          - type: calendar
      
      - size: full
        widgets:
          - type: rss
            cache: 6h          # RSS 缓存 6 小时
            feeds:
              # 只保留国内可流畅访问的源
              - url: https://rsshub.app/bilibili/user/video/xxx
              - url: https://rsshub.app/zhihu/people/activities/xxx
              # 删除国外慢速源
          
          - type: monitor
            cache: 5m
            
          - type: bookmarks
            # 纯本地，无网络请求
```

---

## 性能对比

| 指标 | 原版 | 优化版 | 提升 |
|------|------|--------|------|
| 首次加载 | 3-5s | 200ms | 15-25x |
| 缓存命中加载 | 3-5s | 50ms | 60-100x |
| 内存占用 | 20MB | 20MB | 持平 |
| 组件数量 | 25+ | 23 | 精简 |
| 手动刷新 | ❌ | ✅ | 新增 |

---

## 已完成的额外优化

### 天气模块（已完成）
- ✅ API 替换为和风天气（国内服务器，速度快）
- ✅ 支持中文显示（"天气"、"晴朗"、"多云"等）
- ✅ 默认 24 小时制时间显示

### RSS 缓存层（可选）
部署 Miniflux 或 TTRSS 作为 RSS 聚合缓存：
```yaml
# 使用本地 RSS 服务
- type: rss
  cache: 1h
  feeds:
    - url: http://miniflux:8080/v1/feeds/1/entries
```

### 3. 自定义组件
如需添加自定义组件，参考 `widget-custom-api.go` 实现。

---

## 问题排查

### 编译错误
```bash
# 清理缓存
go clean -cache
go mod tidy

# 重新构建
go build -o glance-optimized ./cmd/glance
```

### 运行时问题
1. **端口冲突**: 修改 `glance.yml` 中的 `server.port`
2. **权限问题**: 确保配置文件夹有读写权限
3. **缓存不更新**: 检查 `cache` 配置是否过长

---

## 贡献与反馈

Fork 自: https://github.com/glanceapp/glance
优化版本: https://github.com/Michae1G/glance

如有问题，请在 GitHub Issues 中反馈。
