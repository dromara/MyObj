# MyObj 跨平台编译脚本

本目录包含了在不同操作系统上编译 MyObj 项目的脚本。

## 脚本说明

### Windows 平台脚本 (.bat)

在 Windows 系统上使用以下脚本：

- **windows-build-windows.bat** - 编译 Windows 版本
- **windows-build-linux.bat** - 编译 Linux 版本
- **windows-build-mac.bat** - 编译 Mac 版本

使用方法：
```cmd
# 双击运行或在命令行中执行
windows-build-windows.bat
```

### Mac 平台脚本 (.sh)

在 Mac 系统上使用以下脚本：

- **mac-build-mac.sh** - 编译 Mac 版本
- **mac-build-linux.sh** - 编译 Linux 版本
- **mac-build-windows.sh** - 编译 Windows 版本

使用方法：
```bash
# 添加执行权限
chmod +x mac-build-mac.sh

# 运行脚本
./mac-build-mac.sh
```

### Linux 平台脚本 (.sh)

在 Linux 系统上使用以下脚本：

- **linux-build-linux.sh** - 编译 Linux 版本
- **linux-build-mac.sh** - 编译 Mac 版本
- **linux-build-windows.sh** - 编译 Windows 版本

使用方法：
```bash
# 添加执行权限
chmod +x linux-build-linux.sh

# 运行脚本
./linux-build-linux.sh
```

## 编译过程

所有脚本都会执行以下步骤：

1. **清理旧文件** - 删除旧的 bin 目录
2. **构建前端** - 编译 Vue.js 前端项目
3. **复制前端产物** - 将前端构建文件复制到 bin/webview/dist
4. **编译后端** - 编译 Go 服务端和 CLI 工具
5. **复制依赖** - 复制所有必要的配置文件和依赖

## 输出目录

编译完成后，所有文件都在 `bin/` 目录下，包含：

```
bin/
├── server 或 server.exe    # 服务端程序
├── cli 或 cli.exe          # CLI 工具
├── webview/                # 前端资源
│   └── dist/
├── libs/                   # 数据库文件（如有）
├── templates/              # HTML 模板
├── docs/                   # API 文档（如有）
├── config.toml            # 配置文件
└── README.txt             # 使用说明
```

## 部署说明

### Windows 部署

1. 将 `bin/` 目录复制到目标 Windows 系统
2. 修改 `config.toml` 配置
3. 双击运行 `server.exe`

### Linux/Mac 部署

1. 将 `bin/` 目录复制到目标系统
2. 修改 `config.toml` 配置
3. 添加执行权限：`chmod +x server`
4. 运行服务：`./server`

## 注意事项

### 前提条件

- **Node.js** - 用于编译前端（需要 npm）
- **Go** - 用于编译后端（建议 Go 1.25+）
- **网络连接** - 首次运行需要下载 npm 依赖

### 编译选项

- **CGO_ENABLED=0** - 跨平台编译时禁用 CGO，确保二进制文件可移植
- **CGO_ENABLED=1** - Windows 编译 Windows 时启用 CGO，支持 SQLite
- **-ldflags="-s -w"** - 减小二进制文件体积

### 配置文件

编译后需要修改 `bin/config.toml` 中的配置：

- 数据库路径（使用相对路径，如 `./libs/my_obj.db`）
- 服务器监听地址和端口
- 日志目录（使用相对路径，如 `./logs/`）
- 文件存储目录（使用相对路径，如 `./obj_data/`）

### 跨平台兼容性

- 路径分隔符在配置文件中统一使用 `/`
- SQLite 数据库文件可跨平台使用
- 确保目标系统有足够的权限创建目录和文件

## 常见问题

### 1. 前端编译失败

```bash
# 先安装依赖
cd webview
npm install
cd ..
```

### 2. Go 编译失败

```bash
# 确保 Go modules 是最新的
go mod tidy
go mod download
```

### 3. 权限问题（Linux/Mac）

```bash
# 给脚本添加执行权限
chmod +x builds/*.sh
```

### 4. 跨平台编译的二进制无法运行

- 检查目标平台的 GOOS 和 GOARCH 设置是否正确
- 确保 CGO_ENABLED=0 用于跨平台编译

## 技术支持

如遇问题，请查看：
- 项目 README.md
- 部署说明.md
- GitHub Issues
