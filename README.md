<div align="center">

# 🌟 MyObj

### 现代化的私有云存储解决方案

[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?style=flat&logo=go)](https://golang.org)
[![Vue Version](https://img.shields.io/badge/Vue-3.5+-4FC08D?style=flat&logo=vue.js)](https://vuejs.org)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

[功能特性](#-功能特性) • [快速开始](#-快速开始) • [文档](#-文档) • [贡献指南](#-贡献指南) • [许可证](#-许可证)

</div>

---

## 📖 项目简介

**MyObj** 是一个基于 **Go + Vue3** 开发的现代化开源私有云存储系统，为个人和家庭用户提供安全、高效、易用的文件管理服务。系统采用前后端分离架构，支持大文件上传、断点续传、秒传、文件分享、权限管理、WebDAV 协议等丰富功能，可作为企业级对象存储服务的轻量化替代方案。

### ✨ 为什么选择 MyObj？

- 🔒 **数据安全** - 文件加密存储，完全掌控自己的数据
- ⚡ **高性能** - 基于 Gin 框架和 BLAKE3 哈希算法，性能卓越
- 🎯 **功能丰富** - 支持秒传、分片上传、离线下载、种子下载等高级特性
- 🌐 **跨平台** - 支持 Web、Android，提供 WebDAV 协议对接第三方客户端
- 🤖 **AI 赋能** - 内置 MCP 服务，支持大模型智能文件管理（开发中）
- 🚀 **易于部署** - 支持 SQLite/MySQL，单文件部署，开箱即用

---

## 🎯 功能特性

### 📤 文件上传与下载

<table>
  <tr>
    <td width="50%">
      <b>🚀 智能上传</b>
      <ul>
        <li>大文件分片上传（支持 GB 级别文件）</li>
        <li>断点续传（网络中断自动恢复）</li>
        <li>秒传功能（基于 BLAKE3 哈希去重）</li>
        <li>多文件批量上传</li>
        <li>文件加密存储</li>
      </ul>
    </td>
    <td width="50%">
      <b>📥 灵活下载</b>
      <ul>
        <li>多文件批量打包下载</li>
        <li>离线下载（HTTP/HTTPS 资源）</li>
        <li>种子下载（磁力链/Torrent 文件）</li>
        <li>分享链接下载</li>
      </ul>
    </td>
  </tr>
</table>

### 👥 用户与权限管理

- ✅ **用户存储配额** - 为每个用户设置独立的存储空间限制
- ✅ **细粒度权限控制** - 支持文件级别的访问权限设置
- ✅ **用户组管理** - 基于用户组的权限策略控制
- ✅ **API Key 认证** - 支持程序化访问和第三方集成
- ✅ **JWT 安全认证** - 基于 Token 的现代化认证机制

### 🗂️ 文件管理

- 📁 **虚拟目录结构** - 灵活的文件组织方式，不暴露服务端真实目录，多用户互不干扰
- 🏷️ **文件操作** - 重命名、移动、复制、删除（回收站机制）
- 🔍 **搜索功能** - 快速搜索文件和文件夹
- 🔗 **限时分享链接** - 生成带有效期的文件分享链接，支持密码保护
- 👁️ **文件预览** - 支持图片、视频在线预览
- 🖼️ **自动缩略图** - 为图片和视频自动生成预览缩略图
- 🌐 **公开文件广场** - 用户可以将文件设为公开，供其他用户浏览

### 🔐 安全与隐私

- 🔒 **文件加密存储** - 可选择性加密敏感文件，保护隐私数据
- 🛡️ **JWT 认证** - 安全的 Token 认证机制
- 🔑 **API Key 管理** - 支持创建和管理多个 API Key
- 🗑️ **回收站机制** - 删除的文件可恢复，防止误操作
- 📊 **操作日志** - 完整的文件操作审计日志

### 🌐 WebDAV 支持

- ✅ **标准 WebDAV 协议** - 兼容主流 WebDAV 客户端
- ✅ **文件浏览与下载** - 通过 WebDAV 访问所有文件
- ✅ **目录管理** - 创建、删除、重命名文件夹
- ✅ **权限控制** - 基于用户权限的访问控制
- ✅ **多用户隔离** - 每个用户只能访问自己的文件空间

> 💡 **详细使用指南**: [WebDAV 配置文档](docs/WEBDAV_USAGE.md)

### 🤖 AI 智能能力（开发中）

通过内置的 **MCP (Model Context Protocol)** 服务，MyObj 支持与大语言模型深度集成：

- 🧠 **智能文件归档** - AI 自动识别文件类型并智能分类
- 🔍 **AI 检索** - 大模型语义搜索网盘文件
- 💾 **内容保存** - AI 生成的内容可直接保存到网盘
- 📄 **智能摘要** - 大模型可阅读文件并生成内容摘要

### 📱 多端支持

- 🌐 **Web 端** - 现代化的响应式网页界面（Vue3 + Element Plus）
- 📱 **Android 端** - 原生安卓应用（开发中）
- 💻 **WebDAV 客户端** - 支持所有兼容 WebDAV 的第三方应用

---

## 🛠️ 技术栈

### 后端技术

<table>
  <tr>
    <td><b>框架</b></td>
    <td>Gin 1.11+ (高性能 HTTP Web 框架)</td>
  </tr>
  <tr>
    <td><b>ORM</b></td>
    <td>GORM 1.31+ (强大的 Go ORM 库)</td>
  </tr>
  <tr>
    <td><b>数据库</b></td>
    <td>MySQL 5.7+ / SQLite 3 (双数据库支持)</td>
  </tr>
  <tr>
    <td><b>认证</b></td>
    <td>JWT (JSON Web Token) + API Key</td>
  </tr>
  <tr>
    <td><b>缓存</b></td>
    <td>Redis / 本地缓存 (可选)</td>
  </tr>
  <tr>
    <td><b>哈希算法</b></td>
    <td>BLAKE3 (高性能哈希，支持秒传)</td>
  </tr>
  <tr>
    <td><b>协议</b></td>
    <td>WebDAV (RFC 4918)</td>
  </tr>
  <tr>
    <td><b>BT 下载</b></td>
    <td>anacrolix/torrent (完整的 BitTorrent 客户端)</td>
  </tr>
  <tr>
    <td><b>文档</b></td>
    <td>Swagger 2.0 (自动生成 API 文档)</td>
  </tr>
</table>

### 前端技术

<table>
  <tr>
    <td><b>框架</b></td>
    <td>Vue 3.5+ (Composition API)</td>
  </tr>
  <tr>
    <td><b>UI 组件</b></td>
    <td>Element Plus 2.11+ (现代化组件库)</td>
  </tr>
  <tr>
    <td><b>构建工具</b></td>
    <td>Vite 7.2+ (极速开发体验)</td>
  </tr>
  <tr>
    <td><b>状态管理</b></td>
    <td>Pinia 3.0+ (Vue 官方推荐)</td>
  </tr>
  <tr>
    <td><b>路由</b></td>
    <td>Vue Router 4.6+</td>
  </tr>
  <tr>
    <td><b>类型支持</b></td>
    <td>TypeScript 5.9+</td>
  </tr>
  <tr>
    <td><b>哈希计算</b></td>
    <td>spark-md5 (文件秒传哈希)</td>
  </tr>
</table>

---

## 📋 系统要求

### 运行环境

- **操作系统**: Windows 7+, macOS, Linux
- **CPU**: 2 核或更高
- **内存**: 2GB RAM 或更高
- **磁盘**: 根据存储需求而定

### 开发环境

- **Go**: 1.25 或更高版本
- **Node.js**: 18.0 或更高版本
- **数据库**: MySQL 5.7+ 或 SQLite 3

---

## 🚀 快速开始

### 方式一：使用预编译版本（推荐）

1. **下载最新版本**

   从 [Releases](https://github.com/your-repo/myobj/releases) 页面下载对应平台的二进制文件。

2. **解压并配置**

   ```bash
   # 解压文件
   unzip myobj-{version}-{platform}.zip
   cd myobj
   
   # 编辑配置文件
   vim config.toml  # Linux/Mac
   notepad config.toml  # Windows
   ```

3. **启动服务**

   ```bash
   # Linux/Mac
   ./server
   
   # Windows
   server.exe
   ```

4. **访问系统**

   打开浏览器访问: `http://localhost:8080`

### 方式二：从源码构建

#### 1. 克隆项目

```bash
git clone https://gitee.com/MR-wind/my-obj.git
cd myobj
```

#### 2. 配置系统

编辑 `config.toml` 文件，配置数据库、服务器等信息：

```toml
[server]
host = "0.0.0.0"    # 监听地址（0.0.0.0 允许外部访问）
port = 8080         # 监听端口
ssl = false         # 是否启用 SSL
swagger = true      # 是否启用 Swagger API 文档

[database]
type = "sqlite"     # 数据库类型: mysql 或 sqlite
host = "./libs/my_obj.db"  # SQLite 数据库文件路径

# 使用 MySQL 时的配置示例：
# type = "mysql"
# host = "localhost"
# port = 3306
# user = "root"
# password = "your-password"
# db_name = "my_obj"

[auth]
secret = "your-secret-key"  # JWT 密钥（请修改为随机字符串）
api_key = true              # 是否启用 API Key
jwt_expire = 2              # Token 有效期（小时）

[file]
thumbnail = true            # 是否生成缩略图
big_file_threshold = 1      # 大文件分片阈值（GB）
big_chunk_size = 1          # 大文件分片大小（GB）
data_dir = "obj_data"       # 文件存储目录
temp_dir = "obj_temp"       # 临时文件目录

[webdav]
enable = true               # 是否启用 WebDAV 服务
host = "0.0.0.0"           # 监听地址
port = 8081                 # 监听端口
prefix = "/dav"             # 路径前缀

[log]
level = "debug"             # 日志级别: debug, info, warn, error
log_path = "./logs/"        # 日志路径
max_size = 10               # 日志文件最大大小（MB）
max_age = 7                 # 日志保留天数
```

#### 3. 初始化数据库

```bash
# 编译 CLI 工具
go build -o bin/cli ./src/cmd/cli

# 执行数据库迁移
./bin/cli -migrate

# Windows 用户使用:
# go build -o bin\cli.exe .\src\cmd\cli
# .\bin\cli.exe -migrate
```

#### 4. 启动开发服务器

**后端开发:**

```bash
# 安装依赖
go mod download

# 启动后端服务
go run src/cmd/server/main.go
```

**前端开发:**

```bash
# 进入前端目录
cd webview

# 安装依赖
npm install

# 启动开发服务器
npm run dev
```

前端开发服务器默认运行在 `http://localhost:5173`，API 请求会自动代理到后端 `http://localhost:8080`。

#### 5. 构建生产版本

使用 `builds/` 目录下的跨平台编译脚本：

**Windows 平台:**

```cmd
cd builds
windows-build-windows.bat     # 编译 Windows 版本
windows-build-linux.bat       # 编译 Linux 版本
windows-build-mac.bat         # 编译 Mac 版本
```

**Linux/Mac 平台:**

```bash
cd builds
chmod +x *.sh                 # 添加执行权限

./linux-build-linux.sh        # Linux: 编译 Linux 版本
./mac-build-mac.sh            # Mac: 编译 Mac 版本
```

编译完成后，所有文件位于 `bin/` 目录，可直接部署到目标服务器。

> 💡 **详细构建说明**: [编译脚本文档](builds/README.md)

---

## 📚 使用指南

### CLI 工具命令

MyObj 提供了强大的命令行工具用于系统管理：

```bash
# 查看帮助信息
./bin/cli -help

# 查看版本信息
./bin/cli -version

# 数据库操作
./bin/cli -migrate              # 执行数据库迁移（创建表结构）

# 用户管理
./bin/cli -create-user "username:password:email@example.com"
./bin/cli -list-users           # 列出所有用户
./bin/cli -delete-user "username"  # 删除指定用户
```

### API 文档

系统内置 Swagger API 文档，启动服务后访问：

```
http://localhost:8080/swagger/index.html
```

#### API 使用示例

**用户认证:**

```bash
# 用户登录获取 Token
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'

# 响应示例
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expire": "2024-01-01T12:00:00Z"
  }
}
```

**文件上传:**

```bash
# 简单文件上传
curl -X POST http://localhost:8080/api/file/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "file=@/path/to/file.pdf" \
  -F "virtual_path=/"

# 加密文件上传
curl -X POST http://localhost:8080/api/file/upload \
  -H "Authorization: Bearer <your-token>" \
  -F "file=@/path/to/secret.doc" \
  -F "virtual_path=/" \
  -F "is_enc=true" \
  -F "file_password=mypassword"
```

**文件下载:**

```bash
# 下载文件
curl -X GET http://localhost:8080/api/file/download/:file_id \
  -H "Authorization: Bearer <your-token>" \
  -o downloaded_file

# 加密文件下载（需要密码）
curl -X GET http://localhost:8080/api/file/download/:file_id?password=mypassword \
  -H "Authorization: Bearer <your-token>" \
  -o downloaded_file
```

**创建分享链接:**

```bash
curl -X POST http://localhost:8080/api/share/create \
  -H "Authorization: Bearer <your-token>" \
  -H "Content-Type: application/json" \
  -d '{
    "file_id": "123",
    "expire_hours": 24,
    "password": "share123"
  }'
```

**API Key 访问:**

```bash
# 使用 API Key 访问
curl -X GET http://localhost:8080/api/file/list \
  -H "X-API-Key: your-api-key"
```

### WebDAV 使用

详细的 WebDAV 配置和使用指南，请参阅：[WebDAV 使用文档](docs/WEBDAV_USAGE.md)

---

## 📁 项目结构

```
myobj/
├── src/                          # 后端源代码
│   ├── cmd/                      # 程序入口
│   │   ├── server/               # Web 服务器主程序
│   │   │   └── main.go
│   │   └── cli/                  # 命令行工具
│   │       └── main.go
│   ├── config/                   # 配置管理
│   │   └── config.go             # 配置文件加载与解析
│   ├── core/                     # 核心业务逻辑
│   │   ├── domain/               # 领域模型
│   │   │   ├── request/          # 请求 DTO
│   │   │   └── response/         # 响应 DTO
│   │   └── service/              # 业务服务层
│   │       ├── file_service.go   # 文件管理服务
│   │       ├── user_service.go   # 用户管理服务
│   │       ├── shares_service.go # 分享服务
│   │       └── ...
│   ├── internal/                 # 内部实现（不对外暴露）
│   │   ├── api/                  # API 层
│   │   │   ├── handlers/         # 请求处理器
│   │   │   ├── middleware/       # 中间件（认证、日志、CORS等）
│   │   │   └── routers/          # 路由配置
│   │   └── repository/           # 数据访问层
│   │       ├── database/         # 数据库实现
│   │       │   ├── mysql.go      # MySQL 驱动
│   │       │   └── sqlite.go     # SQLite 驱动
│   │       └── impl/             # Repository 实现
│   ├── pkg/                      # 公开包（可复用组件）
│   │   ├── models/               # 数据模型（对应数据库表）
│   │   ├── auth/                 # 认证模块（JWT、API Key）
│   │   ├── cache/                # 缓存模块（Redis、本地缓存）
│   │   ├── download/             # 下载模块（HTTP、Torrent）
│   │   ├── upload/               # 上传模块（分片、秒传）
│   │   ├── hash/                 # 哈希计算（BLAKE3）
│   │   ├── preview/              # 文件预览（图片、视频）
│   │   ├── share/                # 分享功能
│   │   ├── task/                 # 任务调度
│   │   ├── util/                 # 工具函数
│   │   ├── webdav/               # WebDAV 协议实现
│   │   └── logger/               # 日志系统
│   └── tests/                    # 测试代码
│       ├── repository_crud_test.go
│       ├── upload_test.go
│       └── ...
│
├── webview/                      # 前端源代码（Vue3）
│   ├── src/
│   │   ├── api/                  # API 请求封装
│   │   ├── assets/               # 静态资源
│   │   ├── components/           # 公共组件
│   │   ├── composables/          # 组合式函数
│   │   ├── layout/               # 布局组件
│   │   ├── router/               # 路由配置
│   │   ├── stores/               # Pinia 状态管理
│   │   ├── types/                # TypeScript 类型定义
│   │   ├── utils/                # 工具函数
│   │   ├── views/                # 页面组件
│   │   ├── App.vue               # 根组件
│   │   └── main.ts               # 入口文件
│   ├── dist/                     # 构建产物
│   ├── package.json
│   ├── vite.config.ts
│   └── tsconfig.json
│
├── builds/                       # 跨平台编译脚本
│   ├── windows-build-windows.bat
│   ├── linux-build-linux.sh
│   ├── mac-build-mac.sh
│   └── README.md
│
├── docs/                         # 文档
│   ├── WEBDAV_USAGE.md          # WebDAV 使用文档
│   ├── swagger.json              # Swagger API 定义
│   └── swagger.yaml
│
├── templates/                    # HTML 模板
│   ├── share_password.html      # 分享密码验证页面
│   └── 404.html
│
├── sql/                          # SQL 脚本
│   └── clear_test_data.sql
│
├── config.toml                   # 主配置文件
├── go.mod                        # Go 模块定义
├── go.sum                        # 依赖校验
└── README.md                     # 项目说明文档
```

---

## 🧪 测试

项目包含完善的测试用例，所有测试代码统一存放在 `src/tests/` 目录中。

### 运行测试

```bash
# 运行所有测试
go test ./src/tests/... -v

# 运行指定测试文件
go test ./src/tests/repository_crud_test.go -v

# 运行性能基准测试
go test -bench=. ./src/tests/file_enc_benchmark_test.go

# 查看测试覆盖率
go test -cover ./src/tests/...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./src/tests/...
go tool cover -html=coverage.out
```

### 测试模块

| 测试文件                       | 说明                            |
| ------------------------------ | ------------------------------- |
| `repository_crud_test.go`      | 数据库 Repository CRUD 操作测试 |
| `upload_test.go`               | 文件上传功能测试                |
| `instant_upload_test.go`       | 秒传功能测试                    |
| `blake3_hasher_test.go`        | BLAKE3 哈希算法测试             |
| `file_enc_util_test.go`        | 文件加密/解密功能测试           |
| `file_enc_benchmark_test.go`   | 文件加密性能基准测试            |
| `rsa_util_test.go`             | RSA 加密工具测试                |
| `utils_test.go`                | 通用工具函数测试                |

---

## 🤝 贡献指南

我们欢迎所有形式的贡献！无论是新功能、Bug 修复、文档改进还是问题反馈。

### 贡献流程

#### 1. Fork 本项目

点击项目页面右上角的 **Fork** 按钮，将项目 Fork 到你的账号下。

#### 2. 克隆你的 Fork

```bash
git clone https://github.com/your-username/myobj.git
cd myobj
```

#### 3. 创建特性分支

```bash
# 新功能分支
git checkout -b feature/your-feature-name

# 或者 Bug 修复分支
git checkout -b fix/your-bugfix-name
```

#### 4. 进行开发

- ✅ 遵循项目的代码规范
- ✅ 添加必要的测试用例
- ✅ 确保所有测试通过
- ✅ 更新相关文档

#### 5. 提交更改

```bash
git add .
git commit -m "feat: 添加某某功能"  # 或 "fix: 修复某某问题"
```

#### 6. 推送到你的 Fork

```bash
git push origin feature/your-feature-name
```

#### 7. 创建 Pull Request

在 GitHub 上打开你的 Fork，点击 **"New Pull Request"** 按钮：

1. 选择 base 分支（通常是 `main`）
2. 填写 PR 标题和详细描述
3. 说明改动的内容和原因
4. 提交 Pull Request

### 提交信息规范

我们使用 [约定式提交](https://www.conventionalcommits.org/) (Conventional Commits) 规范：

| 类型       | 说明                               | 示例                                  |
| ---------- | ---------------------------------- | ------------------------------------- |
| `feat`     | 新功能                             | `feat: 添加文件批量下载功能`          |
| `fix`      | Bug 修复                           | `fix: 修复大文件上传失败的问题`       |
| `docs`     | 文档更新                           | `docs: 更新 API 使用文档`             |
| `style`    | 代码格式调整（不影响功能）         | `style: 格式化代码缩进`               |
| `refactor` | 代码重构（不修改功能）             | `refactor: 重构文件上传逻辑`          |
| `perf`     | 性能优化                           | `perf: 优化哈希计算性能`              |
| `test`     | 测试相关                           | `test: 添加文件上传单元测试`          |
| `chore`    | 构建/工具相关                      | `chore: 更新依赖版本`                 |
| `ci`       | CI/CD 相关                         | `ci: 添加 GitHub Actions 工作流`      |

**提交信息示例:**

```bash
feat: 添加文件批量下载功能

- 实现多文件打包下载 API
- 添加打包进度查询接口
- 前端界面支持批量选择下载
```

### 代码规范

#### Go 代码规范

- ✅ 遵循 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- ✅ 使用 `gofmt` 或 `goimports` 格式化代码
- ✅ 所有公开的函数、类型、常量都必须有注释
- ✅ 注释应以函数/类型名称开头，如：`// GetUserByID retrieves a user by their ID`
- ✅ 内部实现放在 `internal/` 目录，对外接口放在 `pkg/` 目录
- ✅ 所有测试代码统一放在 `src/tests/` 目录
- ✅ 错误处理要完善，避免 panic
- ✅ 使用有意义的变量名，避免单字母变量（除循环变量）

#### Vue/TypeScript 代码规范

- ✅ 遵循 [Vue 3 风格指南](https://cn.vuejs.org/style-guide/)
- ✅ 使用 Composition API 而非 Options API
- ✅ 组件名称使用 PascalCase
- ✅ 使用 TypeScript 类型注解，避免 `any`
- ✅ 使用 ESLint 和 Prettier 格式化代码
- ✅ 组件应拆分为合理的粒度，避免过大

### 开发建议

#### 功能开发前

- 🔍 先查看 [Issues](https://github.com/your-repo/myobj/issues)，避免重复工作
- 💬 如果是重大更改，建议先开 Issue 讨论设计方案
- 📖 阅读相关代码，理解现有实现

#### 提交 PR 前

- ✅ 确保所有测试通过 (`go test ./src/tests/...`)
- ✅ 确保代码已格式化 (`gofmt -w .`)
- ✅ 更新相关文档（如果有 API 变更）
- ✅ 在 PR 描述中说明改动内容和测试情况

#### 文档同步

- 📝 功能变更时同步更新 README.md
- 📝 新增 API 时更新 Swagger 注释
- 📝 重大特性需要在 `docs/` 目录添加使用文档

### 报告问题

发现 Bug 或有功能建议？请创建 [Issue](https://github.com/your-repo/myobj/issues/new) 并提供：

- 📋 **问题的详细描述**
- 🔄 **复现步骤**（如果是 Bug）
- ✅ **期望行为**
- ❌ **实际行为**
- 💻 **系统环境信息**（操作系统、Go 版本、浏览器等）
- 📄 **相关日志或截图**

**Issue 模板示例:**

```markdown
### 问题描述
简要描述遇到的问题

### 复现步骤
1. 打开文件列表页面
2. 点击上传按钮
3. 选择大于 1GB 的文件
4. 观察上传进度

### 期望行为
文件应该能正常上传并显示进度

### 实际行为
上传到 50% 时报错并中断

### 环境信息
- OS: Windows 11
- Go 版本: 1.25
- 浏览器: Chrome 120
- MyObj 版本: v1.0.0

### 相关日志
```
ERROR: upload failed: connection timeout
```
```

### 寻求帮助

如果你在贡献过程中遇到问题：

1. 📖 查看项目文档和示例代码
2. 🔍 搜索已有的 [Issues](https://github.com/your-repo/myobj/issues)
3. 💬 在 Issue 中提问
4. 📧 联系项目维护者

---

## 🗺️ 未来规划

我们正在积极开发以下功能，欢迎贡献：

### 即将推出

- [ ] **多网盘接入** - 支持阿里云盘、百度网盘、天翼云盘等第三方存储
- [ ] **文件类型自动分类** - 智能归类文件（文档、图片、视频等）
- [ ] **文本文件在线编辑** - 支持在线预览和编辑文本文件
- [ ] **视频封面自动生成** - 视频文件自动截取封面作为缩略图
- [ ] **格式转换** - 下载时支持指定格式转换（如 HEIC → JPG）
- [ ] **冷数据归档** - 自动归档压缩长期未访问的文件，节省存储空间

### 长期目标

- [ ] **iOS 客户端** - 原生 iOS 应用
- [ ] **桌面客户端** - Electron 跨平台桌面应用
- [ ] **Docker 部署** - 提供官方 Docker 镜像
- [ ] **K8s 支持** - Helm Charts 和部署文档
- [ ] **对象存储扩展** - 支持 S3、MinIO 等对象存储后端
- [ ] **文件版本控制** - 类似 Git 的文件版本管理
- [ ] **协作功能** - 多人实时协作编辑
- [ ] **全文搜索** - 基于 ElasticSearch 的文件内容搜索

### 贡献优先功能

以下功能特别欢迎社区贡献：

- 🌐 **国际化** (i18n) - 多语言支持
- 🎨 **主题系统** - 自定义界面主题
- 📊 **统计面板** - 可视化数据统计
- 🔌 **插件系统** - 可扩展的插件架构
- 📱 **移动端适配优化** - 更好的移动端体验

> 💡 **参与讨论**: 查看 [待开发功能列表](待开发.md) 或在 [Discussions](https://github.com/your-repo/myobj/discussions) 中分享你的想法！

---

## 📄 开源协议

本项目采用 **Apache License 2.0** 开源，详见 [LICENSE](LICENSE) 文件。

### 你可以自由地：

- ✅ **商业使用** - 可用于商业项目
- ✅ **修改代码** - 可以修改源代码
- ✅ **分发软件** - 可以分发原始或修改后的代码
- ✅ **专利授权** - 提供明确的专利授权
- ✅ **私人使用** - 可以私人使用和修改

### 你需要遵守：

- 📄 **保留版权声明** - 保留原作者的版权声明
- 📄 **声明修改** - 需要在修改的文件中说明修改内容
- 📄 **包含许可证** - 分发时需要包含 Apache 2.0 许可证副本
- 📄 **保留 NOTICE 文件** - 如果项目包含 NOTICE 文件，需要一并保留

### Apache 2.0 的优势：

- 🛡️ **专利保护** - 明确授予专利使用权，避免专利诉讼
- 📜 **商业友好** - 适合企业使用，无需担心法律风险
- 🌍 **国际认可** - 被 Apache 基金会等主流开源组织广泛采用

---

## 🙏 致谢

感谢以下开源项目和技术社区的支持：

- [Gin](https://github.com/gin-gonic/gin) - 高性能 Go Web 框架
- [GORM](https://gorm.io/) - 优秀的 Go ORM 库
- [Vue.js](https://vuejs.org/) - 渐进式 JavaScript 框架
- [Element Plus](https://element-plus.org/) - 基于 Vue 3 的组件库
- [BLAKE3](https://github.com/zeebo/blake3) - 高性能哈希算法
- [anacrolix/torrent](https://github.com/anacrolix/torrent) - Go BitTorrent 库

特别感谢所有为本项目做出贡献的开发者！

<a href="https://github.com/your-repo/myobj/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=your-repo/myobj" />
</a>

---

## 📮 联系方式

- 🏠 **项目主页**: [https://github.com/your-repo/myobj](https://github.com/your-repo/myobj)
- 🐛 **问题反馈**: [GitHub Issues](https://github.com/your-repo/myobj/issues)
- 💬 **讨论区**: [GitHub Discussions](https://github.com/your-repo/myobj/discussions)
- 📧 **邮件**: your-email@example.com
- 🌐 **官方网站**: [https://myobj.example.com](https://myobj.example.com)

---

## 📊 项目统计

![Alt](https://repobeats.axiom.co/api/embed/your-repo-id.svg "Repobeats analytics image")

---

<div align="center">

### ⭐ Star History

[![Star History Chart](https://api.star-history.com/svg?repos=https://gitee.com/MR-wind/myobj&type=Date)](https://star-history.com/#https://gitee.com/MR-wind/myobj&Date)

---

**如果这个项目对你有帮助，请给我们一个 ⭐ Star！**

Made with ❤️ by MyObj Team

[⬆ 回到顶部](#-myobj)

</div>
