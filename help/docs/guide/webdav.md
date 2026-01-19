# WebDAV 使用

MyObj 支持标准的 WebDAV 协议，可以与各种 WebDAV 客户端集成。

## 启用 WebDAV

在 `config.toml` 中配置：

```toml
[webdav]
enable = true               # 是否启用 WebDAV 服务
host = "0.0.0.0"           # 监听地址
port = 8081                 # 监听端口
prefix = "/dav"             # 路径前缀
[auth]
api_key = true #开启apikey
```

## 连接信息

- **服务器地址**：`http://your-domain:8081/dav`
- **用户名**：你的 MyObj 用户名
- **密码**：系统设置中生成的API-KEY

## 客户端配置

### Windows

#### 使用资源管理器

1. 打开"此电脑"
2. 右键 → "添加网络位置"
3. 输入 WebDAV 地址
4. 输入用户名和密码

#### 使用 RaiDrive（推荐）

1. 下载安装 [RaiDrive](https://www.raidrive.com/)
2. 添加 WebDAV 驱动器
3. 输入连接信息
4. 挂载为本地驱动器

### macOS

#### 使用 Finder

1. 打开 Finder
2. 菜单栏 → "前往" → "连接服务器"
3. 输入 `http://your-domain:8081/dav`
4. 输入用户名和密码

### Linux

#### 使用 davfs2

```bash
# 安装 davfs2
sudo apt-get install davfs2

# 挂载
sudo mount -t davfs http://your-domain:8081/dav /mnt/myobj
```

### Android

推荐使用以下应用：
- **FolderSync**：支持自动同步
- **Solid Explorer**：文件管理器
- **WebDAV Navigator**：专门的 WebDAV 客户端

### iOS

推荐使用以下应用：
- **FileBrowser**：文件管理器
- **Documents**：文档管理
- **FE File Explorer**：文件浏览器

## 功能支持

- ✅ 文件浏览
- ✅ 文件上传
- ✅ 文件下载
- ✅ 目录创建/删除
- ✅ 文件重命名
- ✅ 文件移动

## 权限说明

- 每个用户只能访问自己的文件空间
- 权限控制与 Web 端一致
- 加密文件需要通过 Web 端下载（需要密码）

## 故障排查

### 无法连接

1. 检查 WebDAV 服务是否启用
2. 检查防火墙设置
3. 确认端口未被占用
4. 验证用户名和密码

### 连接缓慢

1. 检查网络连接
2. 检查服务器性能
3. 尝试使用 HTTPS（如果已配置）

详细配置请参考：[WebDAV 使用文档](https://github.com/dromara/MyObj/blob/master/docs/WEBDAV_USAGE.md)
