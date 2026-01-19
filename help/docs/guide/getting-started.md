# 快速开始

本指南将帮助你在 5 分钟内快速上手 MyObj。

## 前提条件

- 操作系统：Windows 7+, macOS, Linux
- 内存：2GB RAM 或更高
- 磁盘：根据存储需求而定

## 方式一：使用预编译版本（推荐）

### 1. 下载最新版本

从 [Releases](https://github.com/dromara/MyObj/releases) 页面下载对应平台的二进制文件。

### 2. 解压并配置

```bash
# 解压文件
unzip myobj-{version}-{platform}.zip
cd myobj

# 编辑配置文件
vim config.toml  # Linux/Mac
notepad config.toml  # Windows
```

### 3. 启动服务

```bash
# Linux/Mac
./server

# Windows
server.exe
```

### 4. 访问系统

打开浏览器访问: `http://localhost:8080`

默认管理员账号：
- 用户名：`admin`
- 密码：首次启动时会在控制台显示，请妥善保存

## 方式二：Docker 部署

```bash
# 拉取镜像
docker pull myobj/myobj:latest

# 运行容器
docker run -d \
  --name myobj \
  -p 8080:8080 \
  -v /path/to/data:/app/obj_data \
  -v /path/to/config.toml:/app/config.toml \
  myobj/myobj:latest
```

## 下一步

- 📖 查看 [安装部署指南](/guide/installation) 了解详细安装步骤
- ⚙️ 查看 [配置说明](/guide/configuration) 了解如何配置系统
- 🎯 查看 [功能指南](/guide/features) 了解系统功能
