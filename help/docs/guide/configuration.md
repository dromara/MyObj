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
大文件分片配置用于兼容部分格式硬盘（例如FAT32最大支持单文件4G）可根据磁盘格式进行配置，防止单文件过大导致上传失败
```toml
[file]
thumbnail = true            # 是否生成缩略图
big_file_threshold = 3      # 大文件分片阈值（GB）
big_chunk_size = 3          # 大文件分片大小（GB）
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

### 跨域配置
跨域配置可用于开启跨域访问
```toml
[cors]
# 跨域开启
enable = true
# 跨域域名配置 , 多个用,隔开
allow_origin = "*"
# 跨域请求方法 , 多个用,隔开
allow_methods = "*"
# 跨域请求头 , 多个用,隔开
allow_headers = "*"
# 允许发送凭证(cookies)
allow_credentials = true
# 跨域响应头 , 多个用,隔开
expose_headers = "*"
```

### 缓存配置
系统缓存支持两种形式，redis和系统内存缓存，以下示例为redis示例，如需使用系统内存缓存，则type设置为`local`
```toml
[cache]
type = "redis"
host = "127.0.0.1"
port = 6379
password = ""
db = 0
pool_size = 10
```

### S3服务配置
```toml
# S3 服务配置
[s3]
# 是否启用 S3 服务
enable = true
# 区域名称
region = "us-east-1"
# 是否与主服务共用端口（true: 共用 8080 端口，false: 使用独立端口,强烈建议使用独立端口部署）
share_port = false
# 独立端口（当 share_port = false 时生效）
port = 9000
# S3 API 路径前缀（仅在 share_port=true 时生效，留空表示根路径 /）
# 注意：使用路径前缀会导致与标准S3客户端SDK不兼容，推荐使用独立端口
path_prefix = ""
# 加密主密钥（用于服务端加密，32字节，支持环境变量 S3_ENCRYPTION_MASTER_KEY）
# 如果未配置，将使用默认密钥（生产环境请务必配置）
encryption_master_key = ""
# 操作超时时间（秒），默认30秒，用于控制数据库操作和文件操作的超时
operation_timeout = 30
```

## 安全建议

1. **修改 JWT Secret**：使用强随机字符串作为 JWT 密钥
2. **启用 SSL**：生产环境建议启用 HTTPS
3. **限制访问**：使用防火墙限制服务器访问
4. **定期备份**：定期备份数据库和配置文件
