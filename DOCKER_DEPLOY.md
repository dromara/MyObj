# Docker 部署指南

本文档介绍如何使用 Docker 和 Docker Compose 部署 MyObj 文件存储系统。

## 📋 前置要求

- Docker 20.10+
- Docker Compose 2.0+
- 至少 2GB 可用磁盘空间

## 🏗️ 构建镜像

### 方式一：使用 Docker Compose（推荐）

```bash
# 构建并启动服务
docker-compose up -d

# 查看日志
docker-compose logs -f

# 停止服务
docker-compose down

# 停止并删除数据卷
docker-compose down -v
```

### 方式二：手动构建 Docker 镜像

```bash
# 构建镜像
docker build -t myobj:latest .

# 运行容器
docker run -d \
  --name myobj-server \
  -p 8080:8080 \
  -p 8081:8081 \
  -v $(pwd)/config.toml:/app/config.toml:ro \
  -v $(pwd)/logs:/app/logs \
  -v $(pwd)/libs:/app/libs \
  -v $(pwd)/obj_data:/app/obj_data \
  -v $(pwd)/obj_temp:/app/obj_temp \
  -e TZ=Asia/Shanghai \
  myobj:latest
```

## 📁 目录挂载说明

Docker Compose 配置了以下挂载点：

| 容器内路径 | 宿主机路径 | 说明 | 权限 |
|-----------|-----------|------|------|
| `/app/config.toml` | `./config.toml` | 配置文件 | 只读 |
| `/app/logs` | `./logs` | 日志目录 | 读写 |
| `/app/libs` | `./libs` | 数据库文件目录 | 读写 |
| `/app/obj_data` | `./obj_data` | 文件存储目录 | 读写 |
| `/app/obj_temp` | `./obj_temp` | 临时文件目录 | 读写 |

## 🔌 端口映射

| 容器端口 | 宿主机端口 | 服务 |
|---------|-----------|------|
| 8080 | 8080 | HTTP 主服务 |
| 8081 | 8081 | WebDAV 服务 |
| 6379 | 6379 | Redis 缓存 |

## ⚙️ 配置文件

在启动容器前，请确保 `config.toml` 已正确配置：

```toml
[server]
host = "0.0.0.0"
port = 8080

[database]
type = "sqlite"
host = "./libs/my_obj.db"

[cache]
type = "redis"
host = "redis"  # Docker Compose 中使用服务名
port = 6379

[webdav]
enable = true
host = "0.0.0.0"
port = 8081
```

**重要提示**：
- 在 Docker 环境中，Redis 的 host 应该设置为 `redis`（服务名）而不是 `127.0.0.1`
- 数据库路径使用相对路径 `./libs/my_obj.db`
- 文件存储路径 `obj_data` 和 `obj_temp` 使用默认相对路径

## 🚀 快速启动

1. **克隆项目并进入目录**
   ```bash
   cd myobj
   ```

2. **检查配置文件**
   ```bash
   # 确保 config.toml 存在并已正确配置
   cat config.toml
   ```

3. **创建必要的目录**
   ```bash
   mkdir -p logs libs obj_data obj_temp
   ```

4. **启动服务**
   ```bash
   docker-compose up -d
   ```

5. **查看启动日志**
   ```bash
   docker-compose logs -f myobj
   ```

6. **访问服务**
   - 主服务：http://localhost:8080
   - WebDAV：http://localhost:8081/dav
   - Swagger 文档：http://localhost:8080/swagger/index.html

## 🔍 常用命令

### 查看服务状态
```bash
docker-compose ps
```

### 查看日志
```bash
# 查看所有服务日志
docker-compose logs

# 查看特定服务日志
docker-compose logs myobj
docker-compose logs redis

# 实时查看日志
docker-compose logs -f --tail=100
```

### 重启服务
```bash
# 重启所有服务
docker-compose restart

# 重启特定服务
docker-compose restart myobj
```

### 进入容器
```bash
# 进入应用容器
docker-compose exec myobj sh

# 进入 Redis 容器
docker-compose exec redis sh
```

### 更新镜像
```bash
# 重新构建镜像
docker-compose build

# 重新构建并启动
docker-compose up -d --build
```

### 清理资源
```bash
# 停止并删除容器
docker-compose down

# 停止并删除容器、网络、数据卷
docker-compose down -v

# 清理未使用的镜像
docker image prune -f
```

## 🔧 故障排查

### 1. 容器无法启动

```bash
# 查看详细错误日志
docker-compose logs myobj

# 检查配置文件是否存在
ls -la config.toml

# 检查端口是否被占用
netstat -tulpn | grep -E '8080|8081|6379'
```

### 2. 无法连接 Redis

确保 `config.toml` 中的 Redis 配置正确：
```toml
[cache]
type = "redis"
host = "redis"  # 使用 Docker Compose 服务名
port = 6379
```

### 3. 数据库文件权限问题

```bash
# 确保目录权限正确
chmod -R 755 libs logs obj_data obj_temp
```

### 4. 前端页面无法访问

检查 `webview/dist` 目录是否存在：
```bash
# 如果不存在，需要先构建前端
cd webview
npm install
npm run build
```

## 📊 性能优化

### 调整资源限制

编辑 `docker-compose.yml` 添加资源限制：

```yaml
services:
  myobj:
    # ...
    deploy:
      resources:
        limits:
          cpus: '2'
          memory: 2G
        reservations:
          cpus: '1'
          memory: 1G
```

### 使用 tmpfs 加速临时文件

```yaml
services:
  myobj:
    # ...
    tmpfs:
      - /app/obj_temp:size=1G
```

## 🔒 安全建议

1. **修改默认密钥**
   ```toml
   [auth]
   secret = "your-random-secret-key-here"
   ```

2. **限制端口暴露**
   ```yaml
   ports:
     - "127.0.0.1:8080:8080"  # 仅本地访问
   ```

3. **使用环境变量**
   ```yaml
   environment:
     - DB_PASSWORD=${DB_PASSWORD}
     - REDIS_PASSWORD=${REDIS_PASSWORD}
   ```

4. **定期备份**
   ```bash
   # 备份数据库
   docker cp myobj-server:/app/libs/my_obj.db ./backup/
   
   # 备份文件数据
   tar -czf obj_data_backup.tar.gz obj_data/
   ```

## 🔄 数据迁移

### 从非 Docker 环境迁移到 Docker

1. 备份数据
   ```bash
   cp -r libs libs_backup
   cp -r obj_data obj_data_backup
   ```

2. 停止原服务

3. 启动 Docker 服务
   ```bash
   docker-compose up -d
   ```

### 从 Docker 迁移到非 Docker 环境

1. 停止容器
   ```bash
   docker-compose down
   ```

2. 复制数据
   ```bash
   # 数据已在宿主机目录中，直接使用即可
   ```

## 📝 环境变量

可以通过环境变量覆盖配置：

```bash
# 创建 .env 文件
cat > .env << EOF
TZ=Asia/Shanghai
SERVER_PORT=8080
WEBDAV_PORT=8081
REDIS_HOST=redis
REDIS_PORT=6379
EOF
```

## 🌐 生产环境部署

### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }

    location /dav {
        proxy_pass http://localhost:8081;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

### 配置 HTTPS

```bash
# 使用 Let's Encrypt
certbot --nginx -d your-domain.com
```

## 📞 技术支持

如遇问题，请：
1. 查看日志：`docker-compose logs -f`
2. 检查容器状态：`docker-compose ps`
3. 查看项目文档：[README.md](README.md)
4. 提交 Issue：https://gitee.com/MR-wind/my-obj.git/issues

## 📄 许可证

本项目采用 Apache-2.0 许可证。
