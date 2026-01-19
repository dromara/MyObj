# CLI 工具使用指南

MyObj CLI 是一个功能强大的命令行管理工具,用于管理 MyObj 文件存储系统。无论您是系统管理员还是开发者,通过 CLI 工具都可以轻松完成用户管理、系统配置等常见操作。

## 📚 目录

- [什么是 CLI 工具](#什么是-cli-工具)
- [本机部署使用](#本机部署使用)
  - [编译 CLI 工具](#1-编译-cli-工具)
  - [基础用法](#2-基础用法)
  - [用户管理](#3-用户管理)
  - [用户组管理](#4-用户组管理)
  - [系统信息查询](#5-系统信息查询)
- [Docker 部署使用](#docker-部署使用)
  - [进入容器](#1-进入容器)
  - [使用 CLI 工具](#2-使用-cli-工具)
  - [常见场景](#3-常见场景)
- [常见问题](#常见问题)
- [命令速查表](#命令速查表)

## 什么是 CLI 工具

CLI (Command Line Interface) 工具是 MyObj 提供的命令行管理程序,它可以:

- ✅ **管理用户** - 查看、修改、封禁/解封用户
- ✅ **重置密码** - 快速重置用户密码
- ✅ **管理用户组** - 查看用户组信息和成员
- ✅ **查看系统状态** - 实时查看系统运行情况
- ✅ **交互式操作** - 友好的交互式界面,操作安全便捷

:::tip 为什么需要 CLI 工具?
当用户忘记密码、需要批量管理用户或者排查系统问题时,CLI 工具能够提供比 Web 界面更直接、更高效的管理方式。
:::

---

## 本机部署使用

### 1. 编译 CLI 工具

如果您是从源码部署 MyObj,首先需要编译 CLI 工具。（源代码中已包含各个系统的跨平台编译脚本，可一键运行）

#### Windows 系统

```powershell
# 在项目根目录下执行
go build -buildvcs=false -o cli.exe .\src\cmd\cli
```

编译成功后,会在当前目录生成 `cli.exe` 文件。

#### Linux / macOS 系统

```bash
# 在项目根目录下执行
go build -buildvcs=false -o cli ./src/cmd/cli

# 添加执行权限
chmod +x cli
```

编译成功后,会在当前目录生成 `cli` 文件。

:::warning 注意事项
- CLI 工具需要读取 `config.toml` 配置文件,请确保在项目根目录下执行
- CLI 工具会自动连接数据库,请确保数据库配置正确
- 首次运行时会自动初始化配置和数据库连接
:::

### 2. 基础用法

#### 查看帮助信息

```bash
# 查看全局帮助
./cli --help

# 查看版本信息
./cli --version

# 查看特定命令的帮助
./cli user --help
./cli group --help
./cli system --help
```

**输出示例:**

```
╔══════════════════════════════════════════╗
║          MyObj CLI 管理工具              ║
╚══════════════════════════════════════════╝
ℹ 正在初始化...
✔ 初始化完成

NAME:
   MyObj CLI - MyObj 系统管理工具

USAGE:
   cli [global options] command [command options] [arguments...]

VERSION:
   1.0.0

COMMANDS:
   user, u     用户管理
   group, g    组管理
   system, sys 系统信息
   help, h     显示命令列表或单个命令的帮助
```

#### 命令结构说明

CLI 工具的命令结构非常简单:

```
cli [命令] [子命令] [参数]
```

**示例:**

```bash
cli user list          # 列出所有用户
cli user detail admin  # 查看 admin 用户详情
```

:::tip 使用技巧
- 每个命令都有**简短别名**,例如 `user` 可以简写为 `u`,`list` 可以简写为 `ls`
- 所有危险操作(如重置密码、封禁用户)都会要求**交互式确认**
- 命令输出采用**彩色表格**展示,清晰易读
:::

### 3. 用户管理

#### 3.1 查看所有用户

列出系统中的所有用户,以表格形式展示关键信息。

```bash
./cli user list
# 或使用简写
./cli u ls
```

**输出示例:**

```
╔══════════════════════════════════════════╗
║          MyObj CLI 管理工具              ║
╚══════════════════════════════════════════╝
ℹ 正在初始化...
✔ 初始化完成

⠋ 正在获取用户列表...

┌──────────────┬──────────┬──────────┬─────────────────────┬─────┬────────┬───────────┬────────────┐
│ ID           │ 用户名   │ 昵称     │ 邮箱                │ 组ID│ 状态   │ 空间(GB)  │ 创建时间   │
├──────────────┼──────────┼──────────┼─────────────────────┼─────┼────────┼───────────┼────────────┤
│ a1b2c3d4...  │ admin    │ 管理员   │ admin@example.com   │ 1   │ 正常   │ 100.00    │ 2024-01-01 │
│ e5f6g7h8...  │ user01   │ 测试用户 │ user01@example.com  │ 2   │ 正常   │ 50.00     │ 2024-01-02 │
│ i9j0k1l2...  │ user02   │ 用户二   │ user02@example.com  │ 2   │ 封禁   │ 50.00     │ 2024-01-03 │
└──────────────┴──────────┴──────────┴─────────────────────┴─────┴────────┴───────────┴────────────┘

ℹ 共 3 个用户
```

**表格字段说明:**

- **ID**: 用户唯一标识符(显示前 8 位)
- **用户名**: 登录用户名
- **昵称**: 用户显示名称
- **邮箱**: 用户邮箱地址
- **组ID**: 所属用户组编号
- **状态**: 用户当前状态(正常/封禁)
- **空间(GB)**: 分配的存储空间大小
- **创建时间**: 账号创建日期

#### 3.2 查看用户详情

查看指定用户的详细信息,包括存储空间使用情况。

```bash
./cli user detail <用户名>
# 或使用简写
./cli u info <用户名>
```

**使用示例:**

```bash
./cli user detail admin
```

**输出示例:**

```
══════════════════════════════════════════
 用户详情
══════════════════════════════════════════

  用户ID: a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6
  用户名: admin
  昵称: 管理员
  邮箱: admin@example.com
  手机: 13800138000
  用户组: 管理员 (ID: 1)
  状态: 正常
  总空间: 100.00 GB
  剩余空间: 85.23 GB
  创建时间: 2024-01-01 10:30:00
```

#### 3.3 重置用户密码

快速重置用户的登录密码,适用于用户忘记密码的场景。

```bash
./cli user reset-password <用户名> <新密码>
# 或使用简写
./cli u pwd <用户名> <新密码>
```

**使用示例:**

```bash
./cli user reset-password admin 123456
```

**交互过程:**

```
╔══════════════════════════════════════════╗
║          MyObj CLI 管理工具              ║
╚══════════════════════════════════════════╝
ℹ 正在初始化...
✔ 初始化完成

? 确定要重置用户 'admin' 的密码吗? (y/N) y
✔ 用户 'admin' 的密码已重置
```

:::warning 安全提示
- 新密码长度**不能少于 6 位**
- 重置密码会**立即生效**,用户需使用新密码登录
- 操作前会要求**确认**,输入 `y` 继续,输入 `N` 取消
- 建议设置足够复杂的密码以保证安全
:::

#### 3.4 修改用户组

将用户分配到不同的用户组,用户组决定了用户的权限和配额。

```bash
./cli user change-group <用户名>
# 或使用简写
./cli u chgrp <用户名>
```

**使用示例:**

```bash
./cli user change-group testuser
```

**交互过程:**

```
╔══════════════════════════════════════════╗
║          MyObj CLI 管理工具              ║
╚══════════════════════════════════════════╝
ℹ 正在初始化...
✔ 初始化完成

? 请选择用户 'testuser' 的新用户组: 
  ❯ 管理员 (ID:1)
    普通用户 (ID:2)
    VIP用户 (ID:3)

? 确定将用户 'testuser' 从组 2 改为 '管理员' (ID:1) 吗? (y/N) y
✔ 用户 'testuser' 已从组 2 变更为 '管理员' (ID:1)
```

:::tip 用户组说明
- 用户组控制用户的**权限范围**和**存储配额**
- 修改用户组后,用户的权限会**立即生效**
- 可以使用方向键 ↑↓ 选择用户组
:::

#### 3.5 封禁用户

暂停用户的所有权限,封禁后用户无法登录和访问系统。

```bash
./cli user ban <用户名>
```

**使用示例:**

```bash
./cli user ban baduser
```

**交互过程:**

```
? 确定要封禁用户 'baduser' 吗? (y/N) y
✔ 用户 'baduser' 已被封禁
```

:::warning 封禁效果
- 封禁后用户**无法登录**
- 用户的**所有 API 调用**都会被拒绝
- 用户的**文件不会被删除**,只是暂时无法访问
- 可以使用 `unban` 命令解除封禁
:::

#### 3.6 解封用户

恢复被封禁用户的正常权限。

```bash
./cli user unban <用户名>
```

**使用示例:**

```bash
./cli user unban baduser
```

**交互过程:**

```
? 确定要解封用户 'baduser' 吗? (y/N) y
✔ 用户 'baduser' 已解封
```

#### 3.7 踢出用户登录

强制清除用户的所有登录会话,用户需要重新登录。

```bash
./cli user kick <用户名>
```

**使用示例:**

```bash
./cli user kick admin
```

**交互过程:**

```
? 确定要踢出用户 'admin' 的所有登录会话吗? (y/N) y
✔ 用户 'admin' (ID: xxx) 的所有登录会话已被清除
ℹ 注意:为确保完全清除,已清空所有缓存
```

:::info 使用场景
- 用户账号可能**被盗用**,需要立即踢出所有会话
- 用户在**多个设备**登录,需要统一注销
- 系统维护时需要**清空所有会话**
:::

### 4. 用户组管理

#### 4.1 查看所有用户组

列出系统中的所有用户组,并统计每个组的用户数量。

```bash
./cli group list
# 或使用简写
./cli g ls
```

**输出示例:**

```
⠋ 正在获取组列表...

┌─────┬──────────────┬──────────┬───────────┬─────────┬────────────┐
│ ID  │ 组名         │ 默认组   │ 空间(GB)  │ 用户数  │ 创建时间   │
├─────┼──────────────┼──────────┼───────────┼─────────┼────────────┤
│ 1   │ 管理员       │ 否       │ 0.00      │ 2       │ 2024-01-01 │
│ 2   │ 普通用户     │ 是       │ 50.00     │ 15      │ 2024-01-01 │
│ 3   │ VIP用户      │ 否       │ 200.00    │ 5       │ 2024-01-02 │
└─────┴──────────────┴──────────┴───────────┴─────────┴────────────┘

ℹ 共 3 个用户组
```

**表格字段说明:**

- **ID**: 用户组编号
- **组名**: 用户组名称
- **默认组**: 新注册用户是否默认加入此组
- **空间(GB)**: 该组用户的默认存储空间(0 表示无限制)
- **用户数**: 当前属于该组的用户数量
- **创建时间**: 用户组创建日期

:::tip 用户组规则
- **默认组**只能有一个,新用户注册时会自动加入默认组
- **管理员组**(ID=1)不能设置为默认组,确保安全性
- 空间为 **0** 表示该组用户拥有**无限存储空间**
:::

### 5. 系统信息查询

#### 5.1 查看系统配置

显示系统的基本配置信息。

```bash
./cli system info
# 或使用简写
./cli sys info
```

**输出示例:**

```
══════════════════════════════════════════
 系统信息
══════════════════════════════════════════

  数据库类型: SQLITE
  缓存类型: REDIS
  应用名称: MyObj CLI
  应用版本: 1.0.0
```

#### 5.2 查看系统统计

显示系统的运行统计数据,包括用户数、组数等。

```bash
./cli system stats
# 或使用简写
./cli sys stats
```

**输出示例:**

```
⠋ 正在统计系统数据...

══════════════════════════════════════════
 系统统计
══════════════════════════════════════════

  总用户数: 22
  正常用户: 20
  封禁用户: 2
  用户组数: 3
```

---

## Docker 部署使用

如果您使用 Docker 部署 MyObj,CLI 工具已经内置在 Docker 镜像中,无需单独编译。

### 1. 进入容器

首先需要进入 MyObj 容器的 Shell 环境。

#### 使用 Docker Compose

```bash
# 进入 myobj 容器
docker-compose exec myobj sh
```

#### 使用 Docker 命令

```bash
# 查看容器名称
docker ps | grep myobj

# 进入容器
docker exec -it myobj-server sh
```

成功进入后,您会看到类似这样的提示符:

```
/app #
```

### 2. 使用 CLI 工具

在容器内,CLI 工具的二进制文件位于 `/app` 目录,但由于编译时只构建了服务端,CLI 工具需要临时编译使用。

:::warning Docker 环境注意事项
Docker 镜像中**默认只包含服务端**程序,如需使用 CLI 工具,建议在**构建镜像时一并编译**,或者从宿主机挂载编译好的 CLI 工具。
:::

#### 方式一:宿主机执行(推荐)

在宿主机上编译 CLI 工具,然后通过 `docker exec` 直接执行:

```bash
# 在宿主机项目目录下
go build -buildvcs=false -o bin/cli ./src/cmd/cli

# 通过 docker cp 复制到容器
docker cp bin/cli myobj-server:/app/

# 在容器中执行
docker exec -it myobj-server /app/cli user list
```

#### 方式二:修改 Dockerfile(一劳永逸)

编辑 `Dockerfile`,在构建阶段同时编译 CLI 工具:

```dockerfile {20-21}
# 构建阶段
FROM golang:1.25-alpine AS builder

WORKDIR /build
RUN apk add --no-cache git gcc g++ musl-dev
COPY go.mod go.sum ./
RUN go mod download
COPY . .

# 构建 server 和 cli
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o myobj ./src/cmd/server/main.go
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o cli ./src/cmd/cli/main.go

# 运行阶段
FROM alpine:latest
WORKDIR /app
# ...(其他配置)

# 复制可执行文件
COPY --from=builder /build/myobj .
COPY --from=builder /build/cli .

# ...(其他配置)
```

重新构建镜像:

```bash
docker-compose build
docker-compose up -d
```

### 3. 常见场景

#### 场景 1: 重置管理员密码

如果忘记管理员密码,可以通过容器内的 CLI 工具重置:

```bash
# 方式一:进入容器后执行
docker-compose exec myobj sh
/app # ./cli user reset-password admin newpassword123

# 方式二:直接在宿主机执行
docker exec -it myobj-server ./cli user pwd admin newpassword123
```

#### 场景 2: 查看用户列表

```bash
# 查看所有用户
docker exec -it myobj-server ./cli user list
```

#### 场景 3: 封禁恶意用户

```bash
# 封禁用户
docker exec -it myobj-server ./cli user ban malicious_user

# 查看用户详情确认状态
docker exec -it myobj-server ./cli user detail malicious_user
```

#### 场景 4: 系统健康检查

```bash
# 查看系统统计
docker exec -it myobj-server ./cli system stats

# 查看系统配置
docker exec -it myobj-server ./cli system info
```

---

## 常见问题

### Q1: 执行 CLI 工具时提示 "配置加载失败"

**原因:** CLI 工具无法找到 `config.toml` 配置文件。

**解决方案:**

```bash
# 确保在项目根目录执行
cd /path/to/myobj
./cli user list

# 或者使用完整路径
/path/to/myobj/cli user list
```

### Q2: 执行 CLI 工具时提示 "数据库连接失败"

**原因:** 数据库配置不正确或数据库服务未启动。

**解决方案:**

1. 检查 `config.toml` 中的数据库配置:

```toml
[database]
type = "sqlite"
host = "./libs/my_obj.db"  # SQLite: 确保路径正确

# 或 MySQL 配置
# type = "mysql"
# host = "localhost"
# port = 3306
# user = "root"
# password = "your-password"
# db_name = "my_obj"
```

2. 如果使用 MySQL,确保数据库服务已启动:

```bash
# Linux
sudo systemctl status mysql

# macOS
brew services list | grep mysql
```

### Q3: Docker 容器中没有 CLI 工具

**原因:** 默认的 Dockerfile 只构建了服务端程序。

**解决方案:**

参考上面的 [方式二:修改 Dockerfile](#方式二修改-dockerfile一劳永逸) 部分,重新构建包含 CLI 工具的镜像。

### Q4: CLI 工具提示权限不足

**原因:** 文件或数据库文件权限不正确。

**解决方案:**

```bash
# Linux / macOS
chmod -R 755 libs logs obj_data obj_temp
chown -R $USER:$USER libs logs obj_data obj_temp
```

### Q5: 修改配置后 CLI 不生效

**原因:** CLI 每次运行都会重新加载配置,但可能需要重启服务。

**解决方案:**

```bash
# 本机部署:重启服务
pkill myobj
./myobj

# Docker 部署:重启容器
docker-compose restart myobj
```

---

## 命令速查表

快速查找常用命令:

### 用户管理

| 命令 | 简写 | 说明 | 示例 |
|------|------|------|------|
| `user list` | `u ls` | 列出所有用户 | `./cli u ls` |
| `user detail <user>` | `u info <user>` | 查看用户详情 | `./cli u info admin` |
| `user reset-password <user> <pwd>` | `u pwd <user> <pwd>` | 重置密码 | `./cli u pwd admin 123456` |
| `user change-group <user>` | `u chgrp <user>` | 修改用户组 | `./cli u chgrp testuser` |
| `user ban <user>` | - | 封禁用户 | `./cli user ban baduser` |
| `user unban <user>` | - | 解封用户 | `./cli user unban baduser` |
| `user kick <user>` | - | 踢出登录 | `./cli user kick admin` |

### 用户组管理

| 命令 | 简写 | 说明 | 示例 |
|------|------|------|------|
| `group list` | `g ls` | 列出所有用户组 | `./cli g ls` |

### 系统信息

| 命令 | 简写 | 说明 | 示例 |
|------|------|------|------|
| `system info` | `sys info` | 查看系统配置 | `./cli sys info` |
| `system stats` | `sys stats` | 查看系统统计 | `./cli sys stats` |

### Docker 容器操作

| 命令 | 说明 |
|------|------|
| `docker-compose exec myobj sh` | 进入容器 Shell |
| `docker exec -it myobj-server /app/cli user list` | 宿主机直接执行 CLI 命令 |
| `docker cp cli myobj-server:/app/` | 复制 CLI 工具到容器 |

---

## 总结

MyObj CLI 工具为系统管理提供了强大而便捷的命令行接口。无论是日常的用户管理,还是紧急情况下的密码重置,CLI 工具都能帮您快速完成任务。

### 快速上手三步走

1. **编译工具**: `go build -buildvcs=false -o cli ./src/cmd/cli`
2. **查看帮助**: `./cli --help`
3. **开始使用**: `./cli user list`

:::tip 最佳实践
- 使用**短别名**提高效率,如 `u ls` 代替 `user list`
- 重要操作前先用 `detail` 命令**确认用户信息**
- 利用**交互式界面**安全地完成敏感操作
- Docker 环境建议**预先编译 CLI 到镜像**中
:::

如有任何问题,欢迎查阅 [README](https://gitee.com/dromara/my-obj/blob/master/README.md) 或提交 [Issue](https://gitee.com/dromara/my-obj/issues)。
