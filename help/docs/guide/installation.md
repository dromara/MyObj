# 安装部署

## 系统要求

### 运行环境

- **操作系统**: Windows 7+, macOS, Linux
- **CPU**: 2 核或更高
- **内存**: 2GB RAM 或更高
- **磁盘**: 根据存储需求而定

### 开发环境

- **Go**: 1.25 或更高版本（从源码构建时需要）
- **Node.js**: 18.0 或更高版本（从源码构建时需要）
- **数据库**: MySQL 5.7+ 或 SQLite 3

## 安装方式

### 方式一：预编译版本

1. 从 [Releases](https://github.com/dromara/MyObj/releases) 下载对应平台的二进制文件
2. 解压到目标目录
3. 配置 `config.toml` 文件
4. 运行可执行文件

### 方式二：Docker 部署

```bash
docker pull myobj/myobj:latest
docker run -d --name myobj -p 8080:8080 myobj/myobj:latest
```

### 方式三：从源码构建

详见 [README.md](https://github.com/dromara/MyObj/blob/master/README.md#从源码构建) 中的构建说明。

## 数据库配置

### SQLite（推荐用于小型部署）

```toml
[database]
type = "sqlite"
host = "./libs/my_obj.db"
```

### MySQL（推荐用于生产环境）

```toml
[database]
type = "mysql"
host = "localhost"
port = 3306
user = "root"
password = "your-password"
db_name = "my_obj"
```

## 首次启动

1. 系统会自动创建数据库表结构
2. 默认创建管理员账户（用户名：`admin`）
3. 首次登录密码会在控制台输出，请妥善保存
4. 建议首次登录后立即修改密码

## 验证安装

访问 `http://localhost:8080`，如果能看到登录页面，说明安装成功。
