# 配置说明

MyObj 使用 TOML 格式的配置文件 `config.toml`。

## 配置文件位置

配置文件默认位于程序运行目录下的 `config.toml`。

## 配置项说明

### 服务器配置

```toml
[server]
host = "0.0.0.0"    # 监听地址（0.0.0.0 允许外部访问）
port = 8080         # 监听端口
ssl = false         # 是否启用 SSL
ssl_cert = ""       # SSL 证书路径
ssl_key = ""        # SSL 私钥路径
swagger = true      # 是否启用 Swagger API 文档
```

### 数据库配置

```toml
[database]
type = "sqlite"     # 数据库类型: mysql 或 sqlite
host = "./libs/my_obj.db"  # SQLite 数据库文件路径

# MySQL 配置示例：
# type = "mysql"
# host = "localhost"
# port = 3306
# user = "root"
# password = "your-password"
# db_name = "my_obj"
```

### 认证配置

```toml
[auth]
secret = "your-secret-key"  # JWT 密钥（请修改为随机字符串）
api_key = true              # 是否启用 API Key
jwt_expire = 2              # Token 有效期（小时）
```

### 文件存储配置

```toml
[file]
thumbnail = true            # 是否生成缩略图
big_file_threshold = 1      # 大文件分片阈值（GB）
big_chunk_size = 1          # 大文件分片大小（GB）
data_dir = "obj_data"       # 文件存储目录
temp_dir = "obj_temp"       # 临时文件目录
```

### WebDAV 配置

```toml
[webdav]
enable = true               # 是否启用 WebDAV 服务
host = "0.0.0.0"           # 监听地址
port = 8081                 # 监听端口
prefix = "/dav"             # 路径前缀
```

### 日志配置

```toml
[log]
level = "debug"             # 日志级别: debug, info, warn, error
log_path = "./logs/"        # 日志路径
max_size = 10               # 日志文件最大大小（MB）
max_age = 7                 # 日志保留天数
```

## 安全建议

1. **修改 JWT Secret**：使用强随机字符串作为 JWT 密钥
2. **启用 SSL**：生产环境建议启用 HTTPS
3. **限制访问**：使用防火墙限制服务器访问
4. **定期备份**：定期备份数据库和配置文件
