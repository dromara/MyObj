
<h1 align="center">MyObj · 一个功能强大、安全可靠的私有云存储解决方案</h1>

## 📖 项目简介

MyObj 是一个基于 Go 语言开发的开源家庭网盘系统，旨在为个人和家庭用户提供一个安全、高效、易用的私有云存储服务。系统支持大文件上传、断点续传、秒传、文件分享、权限管理等丰富功能，可作为企业级对象存储服务的轻量化替代方案。

## ✨ 核心特性

### 📤 文件上传与下载
- **大文件分片上传** - 支持大文件切片传输，提升上传稳定性
- **断点续传** - 网络中断后可从断点继续上传，避免重复传输
- **秒传功能** - 基于文件哈希值的智能去重，相同文件无需重复上传
- **多文件批量上传** - 支持一次性上传多个文件
- **多文件批量下载** - 支持批量文件打包下载

### 👥 用户与权限管理
- **用户存储空间配额** - 为每个用户设置独立的存储空间限制
- **细粒度权限控制** - 支持文件级别的访问权限设置
- **组权限管理** - 基于用户组的系统权限策略控制
- **文件公开/私有** - 灵活设置文件的访问范围

### 🗂️ 文件管理
- **虚拟目录结构** - 用户界面采用虚拟目录，支持自由组织文件
- **文件重命名** - 支持对文件进行重命名操作
- **限时分享链接** - 生成带有效期的文件分享链接
- **文件预览** - 支持图片、视频等多种格式的在线预览
- **自动生成缩略图** - 为图片和视频自动生成预览缩略图
- **支持用户文件分享** - 一键将文件分享给其他用户

### 🔐 安全与加密
- **文件加密存储** - 支持文件加密保存，保护隐私数据
- **JWT 认证** - 基于 Token 的安全认证机制
- **API Key 访问** - 支持通过 API Key 访问系统 API

### 🚀 高级功能
- **异步离线下载** - 支持将远程资源异步下载到网盘中
- **冷数据归档** - 自动归档压缩长期未访问的文件，节省存储空间
- **API 服务** - 提供 RESTful API，可作为对象存储服务使用
- **第三方存储接入** - 计划支持阿里云盘、百度网盘、天翼云盘等第三方存储
- **离线下载** - 支持将远程资源异步下载到网盘中
- **种子下载** - 支持将种子或磁力链内容一键下载至网盘

### 🤖 AI能力
- **智能识别** - 通过内置的MCP服务，你可以接入大模型实现文件智能归档
- **AI检索** - 内置的MCP服务支持大模型检索网盘内文件
- **内容保存到网盘** -AI生成的内容可以直接保存到网盘
- **AI阅读和生成摘要** - 大模型可以直接读取网盘中的文件内容，并生成内容摘要

### 📱 多端支持
- **Web 端** - 功能完整的网页客户端
- **Android 端** - 原生安卓应用支持

## 🛠️ 技术栈

- **后端框架**: Gin (高性能 HTTP Web 框架)
- **ORM**: GORM (强大的 Go ORM 库)
- **数据库**: MySQL / SQLite (支持多种数据库)
- **认证**: JWT (JSON Web Token)
- **日志**: Logrus (结构化日志库)
- **配置**: TOML (配置文件格式)
- **文件类型识别**: Mimetype
- **哈希算法**: BLAKE3 (高性能哈希)

## 📋 环境要求

- Go 1.25 或更高版本
- MySQL 5.7+ 或 SQLite 3
- 磁盘空间根据存储需求而定

## 🎛️ 配置需求
### windows7以上或macOS或Linux
- 2核CPU或更高性能
- 2GB 内存或更高
- 磁盘空间根据存储需求而定

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://gitee.com/MR-wind/my-obj.git
cd myobj
```

### 2. 安装依赖

```bash
go mod download
```

### 3. 配置系统

编辑 `config.toml` 文件，配置数据库、服务器等信息：

```toml
[server]
host = "0.0.0.0"    # 监听地址
port = 8080         # 监听端口
ssl = false         # 是否启用SSL

[database]
type = "sqlite"     # 数据库类型: mysql 或 sqlite
host = "./libs/my_obj.db"  # SQLite 数据库文件路径
# 使用 MySQL 时配置以下选项
# host = "localhost"
# port = 3306
# user = "root"
# password = "password"
# db_name = "my_obj"

[auth]
secret = "your-secret-key"  # JWT 密钥
api_key = true              # 是否启用 API Key
jwt_expire = 2              # Token 有效期(小时)

[log]
level = "debug"      # 日志级别
log_path = "./logs/" # 日志路径
```

### 4. 初始化数据库

```bash
# 编译 CLI 工具
go build -o bin/cli.exe src/cmd/cli/main.go

# 执行数据库迁移
./bin/cli.exe -migrate
```

### 5. 启动服务器

```bash
# 编译服务器
go build -o bin/server.exe src/cmd/server/main.go

# 运行服务器
./bin/server.exe
```

服务器启动后会输出详细的初始化日志：
- ✅ 配置加载状态
- ✅ 日志系统初始化
- ✅ 数据库连接状态
- ✅ 路由注册信息
- ✅ 服务器监听地址

访问 `http://localhost:8080` 或 `http://您的IP地址:8080` 即可使用系统。

## 📚 使用指南

### CLI 工具命令

MyObj 提供了强大的命令行工具用于系统管理：

```bash
# 查看帮助信息
./bin/cli.exe -help

# 查看版本信息
./bin/cli.exe -version

# 数据库操作
./bin/cli.exe -migrate     # 执行数据库迁移
./bin/cli.exe -seed        # 初始化测试数据(如果支持)
./bin/cli.exe -rollback    # 回滚数据库(如果支持)

# 用户管理
./bin/cli.exe -create-user "username:password:email@example.com"
./bin/cli.exe -list-users
./bin/cli.exe -delete-user "username"
```

### API 使用

#### 用户认证

```bash
# 用户登录获取 Token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 使用 Token 访问受保护的 API
curl -X GET http://localhost:8080/api/v1/user/info \
  -H "Authorization: Bearer <your-token>"
```

#### 文件操作

```bash
# 上传文件
curl -X POST http://localhost:8080/api/v1/file/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "file=@/path/to/file"

# 下载文件
curl -X GET http://localhost:8080/api/v1/file/download/:file_id \
  -H "Authorization: Bearer <your-token>" \
  -o downloaded_file

# 创建分享链接
curl -X POST http://localhost:8080/api/v1/share/create \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{"file_id":"123","expire_hours":24}'
```

#### API Key 访问

```bash
# 使用 API Key 访问
curl -X GET http://localhost:8080/api/v1/file/list \
  -H "X-API-Key: your-api-key"
```

### 配置文件查找机制

系统会按以下顺序自动查找配置文件：
1. 当前工作目录
2. 可执行文件所在目录
3. 向上遍历项目根目录(最多 5 层)

如果未找到配置文件，系统会给出明确的错误提示。

## 📁 项目结构

```
myobj/
├── src/
│   ├── cmd/                      # 程序入口
│   │   ├── server/               # Web 服务器主程序
│   │   │   └── main.go
│   │   └── cli/                  # 命令行工具
│   │       └── main.go
│   ├── config/                   # 配置管理
│   │   └── config.go
│   ├── internal/                 # 内部实现（不对外暴露）
│   │   ├── api/                  # API 层
│   │   │   ├── handlers/         # 请求处理器
│   │   │   ├── middleware/       # 中间件
│   │   │   └── routers/          # 路由配置
│   │   └── repository/           # 数据访问层
│   │       ├── database/         # 数据库连接实现
│   │       │   ├── interface.go  # 数据库接口定义
│   │       │   ├── mysql.go      # MySQL 实现
│   │       │   ├── sqlite.go     # SQLite 实现
│   │       │   └── database_log.go
│   │       └── impl/             # Repository 实现
│   │           ├── user_repo.go          # 用户仓储
│   │           ├── file_info_repo.go     # 文件信息仓储
│   │           ├── file_chunk_repo.go    # 文件分片仓储
│   │           ├── user_files_repo.go    # 用户文件关联仓储
│   │           ├── virtual_path_repo.go  # 虚拟路径仓储
│   │           ├── share_repo.go         # 分享仓储
│   │           ├── disk_repo.go          # 磁盘配额仓储
│   │           ├── group_repo.go         # 用户组仓储
│   │           ├── group_power_repo.go   # 组权限仓储
│   │           ├── power_repo.go         # 权限仓储
│   │           ├── api_key_repo.go       # API Key 仓储
│   │           └── factory.go            # 仓储工厂
│   ├── pkg/                      # 公开包（对外暴露）
│   │   ├── models/               # 数据模型
│   │   │   ├── user_info.go      # 用户模型
│   │   │   ├── file_info.go      # 文件信息模型
│   │   │   ├── file_chunk.go     # 文件分片模型
│   │   │   ├── user_files.go     # 用户文件模型
│   │   │   ├── virtual_path.go   # 虚拟路径模型
│   │   │   ├── shares.go         # 分享模型
│   │   │   ├── disk.go           # 磁盘配额模型
│   │   │   ├── groups.go         # 用户组模型
│   │   │   ├── group_power.go    # 组权限模型
│   │   │   ├── power.go          # 权限模型
│   │   │   └── api_key.go        # API Key 模型
│   │   ├── repository/           # Repository 接口
│   │   │   └── interface.go
│   │   ├── auth/                 # 认证相关
│   │   │   ├── jwt.go            # JWT 认证
│   │   │   └── api_key.go        # API Key 认证
│   │   ├── util/                 # 工具函数
│   │   │   ├── blake3_hasher.go  # BLAKE3 哈希
│   │   │   ├── file_util.go      # 文件工具
│   │   │   ├── file_enc_util.go  # 文件加密工具
│   │   │   ├── disk_util.go      # 磁盘工具
│   │   │   ├── psw_util.go       # 密码工具
│   │   │   ├── rsa_util.go       # RSA 加密工具
│   │   │   ├── time_util.go      # 时间工具
│   │   │   └── imag_thumbnail_util.go # 缩略图生成
│   │   ├── hash/                 # 文件哈希
│   │   │   └── file_hash.go
│   │   ├── preview/              # 文件预览
│   │   │   ├── image_preview.go  # 图片预览
│   │   │   └── video_preview.go  # 视频预览
│   │   ├── share/                # 分享功能
│   │   │   └── share_link.go
│   │   ├── task/                 # 任务调度
│   │   │   └── task_scheduler.go
│   │   ├── logger/               # 日志系统
│   │   │   └── logger.go
│   │   └── custom_type/          # 自定义类型
│   │       └── time_type.go
│   ├── storage/                  # 存储驱动
│   │   ├── driver/               # 存储驱动接口与实现
│   │   │   ├── interface.go      # 驱动接口
│   │   │   └── local.go          # 本地存储驱动
│   │   └── encrypt/              # 加密存储
│   │       └── file_encrypt.go
│   └── tests/                    # 测试代码
│       ├── repository_crud_test.go      # Repository CRUD 测试
│       ├── blake3_hasher_test.go        # BLAKE3 哈希测试
│       ├── file_enc_util_test.go        # 文件加密测试
│       ├── file_enc_benchmark_test.go   # 加密性能测试
│       ├── rsa_util_test.go             # RSA 工具测试
│       └── utils_test.go                # 工具函数测试
├── web/                          # 前端代码
├── examples/                     # 示例代码
│   └── repository_example.md
├── config.toml                   # 配置文件
├── go.mod                        # Go 模块定义
├── go.sum                        # 依赖校验
└── README.md                     # 项目说明文档
```

## 🧪 测试

项目包含完善的测试用例，所有测试代码统一存放在 `src/tests/` 目录中。

### 运行测试

```bash
# 运行所有测试
go test ./src/tests/...

# 运行指定测试文件
go test ./src/tests/repository_crud_test.go

# 运行测试并显示详细输出
go test -v ./src/tests/...

# 运行性能测试
go test -bench=. ./src/tests/file_enc_benchmark_test.go

# 查看测试覆盖率
go test -cover ./src/tests/...
```

### 测试模块说明

- **repository_crud_test.go** - 数据库 Repository 的 CRUD 操作测试
- **blake3_hasher_test.go** - BLAKE3 哈希算法测试
- **file_enc_util_test.go** - 文件加密/解密功能测试
- **file_enc_benchmark_test.go** - 文件加密性能基准测试
- **rsa_util_test.go** - RSA 加密工具测试
- **utils_test.go** - 通用工具函数测试

## 🤝 贡献指南

我们欢迎所有形式的贡献！无论是新功能、Bug 修复、文档改进还是问题反馈。

### 如何贡献

1. **Fork 本项目**
   
   点击项目页面右上角的 Fork 按钮，将项目 Fork 到你的账号下。

2. **克隆你的 Fork**
   
   ```bash
   git clone https://github.com/your-username/myobj.git
   cd myobj
   ```

3. **创建特性分支**
   
   ```bash
   git checkout -b feature/your-feature-name
   # 或
   git checkout -b fix/your-bugfix-name
   ```

4. **进行开发**
   
   - 遵循项目的代码规范
   - 添加必要的测试用例
   - 确保所有测试通过
   - 更新相关文档

5. **提交更改**
   
   ```bash
   git add .
   git commit -m "feat: 添加某某功能" # 或 "fix: 修复某某问题"
   ```

6. **推送到你的 Fork**
   
   ```bash
   git push origin feature/your-feature-name
   ```

7. **创建 Pull Request**
   
   在 GitHub 上打开你的 Fork，点击 "New Pull Request" 按钮，填写 PR 描述并提交。

### 提交信息规范

我们使用约定式提交（Conventional Commits）规范：

- `feat:` 新功能
- `fix:` Bug 修复
- `docs:` 文档更新
- `style:` 代码格式调整（不影响功能）
- `refactor:` 代码重构
- `test:` 测试相关
- `chore:` 构建/工具相关

示例：
```
feat: 添加文件批量下载功能
fix: 修复大文件上传失败的问题
docs: 更新 API 使用文档
```

### 代码规范

- 遵循 Go 语言官方代码风格指南
- 使用 `gofmt` 格式化代码
- 所有公开的函数和类型都应有注释
- 内部实现放在 `internal` 目录，对外接口放在 `pkg` 目录
- 所有测试代码统一放在 `src/tests/` 目录
- 保持代码简洁、可读性强

### 开发建议

- **功能开发前** 先查看 Issues，避免重复工作
- **重大更改前** 建议先开 Issue 讨论设计方案
- **提交 PR 前** 确保通过所有测试
- **文档同步** 功能变更时同步更新相关文档

### 报告问题

发现 Bug 或有功能建议？请创建 Issue 并提供：

- 问题的详细描述
- 复现步骤（如果是 Bug）
- 期望行为
- 实际行为
- 系统环境信息（操作系统、Go 版本等）
- 相关日志或截图

### 寻求帮助

如果你在贡献过程中遇到问题：

- 查看项目文档和示例代码
- 搜索已有的 Issues
- 在 Issue 中提问
- 联系项目维护者

## 📄 开源协议

本项目采用 MIT 协议开源，详见 [LICENSE](LICENSE) 文件。

## 🙏 致谢

感谢所有为本项目做出贡献的开发者！

## 📮 联系方式

- 项目主页: [GitHub Repository]
- Issue 跟踪: [GitHub Issues]
- 讨论区: [GitHub Discussions]

---

<p align="center">如果这个项目对你有帮助，请给我们一个 ⭐ Star！</p>