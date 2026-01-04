# WebDAV 功能使用说明

## 功能概述

MyObj 现已支持 WebDAV 协议，允许用户将网盘挂载为本地网络驱动器，像操作本地文件夹一样访问云端文件。

## 配置启用

### 1. 修改配置文件

编辑 `config.toml`，找到 `[webdav]` 配置段：

```toml
[webdav]
# 是否启用 WebDAV 服务
enable = true           # 改为 true 启用
# 监听地址（127.0.0.1 仅本地访问，0.0.0.0 允许外部访问）
host = "127.0.0.1"      # 内网使用建议 127.0.0.1
# 监听端口
port = 8081             # 与主服务不同的端口
# 路径前缀
prefix = "/dav"         # WebDAV 访问路径
```

### 2. 添加权限（首次使用）

执行 SQL 脚本添加 WebDAV 访问权限：

```bash
# SQLite
sqlite3 libs/my_obj.db < sql/add_webdav_permission.sql

# 或手动执行
sqlite3 libs/my_obj.db
sqlite> INSERT INTO power (name, characteristic, description, created_at, updated_at) 
        VALUES ('WebDAV访问', 'webdav:access', '允许通过WebDAV协议访问文件系统', datetime('now'), datetime('now'));
sqlite> INSERT INTO group_power (group_id, power_id) 
        SELECT 1, id FROM power WHERE characteristic = 'webdav:access';
```

### 3. 创建 API Key

用户需要先创建 API Key 作为 WebDAV 密码：

1. 登录网盘系统
2. 进入 API Key 管理页面
3. 创建新的 API Key（建议备注为"WebDAV专用"）
4. 保存生成的 Key（格式如：`myo_abc123def456...`）

## 客户端连接

### Windows 系统

#### 方法 1: 映射网络驱动器（推荐）

```powershell
# 使用命令行
net use Z: http://localhost:8081/dav /user:你的用户名 你的API_Key

# 示例
net use Z: http://localhost:8081/dav /user:admin myo_abc123def456
```

#### 方法 2: 图形界面

1. 打开"此电脑"
2. 点击"映射网络驱动器"
3. 输入地址：`http://localhost:8081/dav`
4. 勾选"使用其他凭据连接"
5. 输入：
   - 用户名：你的用户名
   - 密码：你的 API Key

### macOS 系统

1. 打开 Finder
2. 按 `Cmd + K` 或菜单 "前往" → "连接服务器"
3. 输入：`http://localhost:8081/dav`
4. 点击"连接"
5. 输入用户名和 API Key

### Linux 系统

#### 使用 davfs2

```bash
# 安装 davfs2
sudo apt install davfs2  # Debian/Ubuntu
sudo yum install davfs2  # CentOS/RHEL

# 创建挂载点
sudo mkdir -p /mnt/webdav

# 挂载
sudo mount -t davfs http://localhost:8081/dav /mnt/webdav
# 输入用户名和 API Key

# 卸载
sudo umount /mnt/webdav
```

#### 自动挂载（可选）

```bash
# 编辑 /etc/fstab
http://localhost:8081/dav /mnt/webdav davfs user,noauto 0 0

# 配置凭据 ~/.davfs2/secrets
http://localhost:8081/dav 用户名 API_Key
```

## 安全建议

### 1. 使用 HTTPS（生产环境强烈推荐）

HTTP 传输未加密，API Key 可能被截获。生产环境请配置 SSL：

```toml
[server]
ssl = true
ssl_cert = "/path/to/cert.pem"
ssl_key = "/path/to/key.pem"
```

### 2. 限制访问范围

```toml
[webdav]
# 仅监听本地（更安全）
host = "127.0.0.1"

# 如需局域网访问
host = "0.0.0.0"  # 配合防火墙规则使用
```

### 3. API Key 管理

- 定期轮换 API Key
- 为 WebDAV 创建专用 Key（便于单独撤销）
- 设置合理的过期时间
- 不要与他人共享 Key

## 功能特性

### ✅ 已支持

- **文件浏览**：查看目录和文件列表
- **文件下载**：读取文件内容
- **目录创建**：创建新文件夹
- **文件/目录删除**：删除文件和文件夹（移入回收站）
- **文件/目录重命名**：修改名称或移动位置
- **权限控制**：基于用户权限系统
- **多用户隔离**：每个用户只能访问自己的文件
- **磁盘空间显示**：正确显示用户存储配额和已用空间

### ⚠️ 限制

- **文件上传**：暂不支持（建议通过网页端上传）
- **大文件**：建议通过网页端上传大文件（支持秒传）
- **实时同步**：不支持自动同步，需手动刷新
- **Windows 系统文件**：desktop.ini 等系统文件会被自动过滤

## 故障排查

### 1. 连接失败

```bash
# 检查服务是否启动
curl -v http://localhost:8081/dav

# 检查端口是否被占用
netstat -ano | findstr 8081  # Windows
lsof -i :8081                # Linux/macOS
```

### 2. 认证失败

- 确认用户名正确
- 确认 API Key 有效且未过期
- 检查是否有 `webdav:access` 权限
- 查看服务器日志 `logs/` 目录

### 3. 文件不可见

- 确认文件在 `virtual_path` 和 `user_files` 表中存在
- 检查文件权限
- **刷新客户端**（重新打开文件夹）
- 查看后台日志确认是否成功读取列表

### 4. desktop.ini 错误

Windows 会自动查询 desktop.ini 文件，这是正常现象，已被自动过滤，不影响使用。

## 日志查看

WebDAV 操作会记录到日志文件：

```bash
# 查看实时日志
tail -f logs/app.log

# 搜索 WebDAV 相关日志
grep "WebDAV" logs/app.log
```

日志包含：
- 认证成功/失败
- 文件操作记录
- 错误信息

## 性能优化

### 1. 局域网使用

```toml
[webdav]
host = "192.168.1.100"  # 使用内网IP
```

### 2. 缓存配置

确保 Redis 缓存已启用以提升性能：

```toml
[cache]
type = "redis"
host = "127.0.0.1"
port = 6379
```

## 技术架构

- **认证方式**：HTTP Basic Auth + API Key
- **权限系统**：复用现有权限表（`webdav:access`）
- **文件系统**：虚拟文件系统适配器
- **数据隔离**：基于 `user_id` 的数据隔离
- **独立服务**：独立端口运行，不影响主服务

## 常见问题

**Q: 为什么要用 API Key 而不是密码？**  
A: API Key 可以单独创建和撤销，不影响账号密码，更安全。

**Q: 可以多个客户端同时连接吗？**  
A: 可以，但不支持实时同步，建议避免并发修改同一文件。

**Q: 上传的文件会秒传吗？**  
A: 目前 WebDAV 上传功能有限，建议大文件通过网页端上传以利用秒传。

**Q: 内网访问需要配置什么？**  
A: 修改 `host = "0.0.0.0"` 并确保防火墙允许 8081 端口。

---

**提示**：首次使用建议先在本地测试，确认正常后再开放到生产环境。
