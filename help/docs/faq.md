# 常见问题

## 安装部署

### Q: 首次启动后如何获取管理员密码？

A: 首次启动时，系统会在控制台输出默认管理员密码，请妥善保存。默认管理员用户名为 `admin`。

### Q: 如何修改默认端口？

A: 编辑 `config.toml` 文件，修改 `[server]` 部分的 `port` 配置项。

### Q: 支持哪些数据库？

A: 目前支持 SQLite 和 MySQL。SQLite 适合小型部署，MySQL 适合生产环境。

### Q: 如何启用 HTTPS？

A: 在 `config.toml` 中配置 `[server]` 部分：
```toml
[server]
ssl = true
ssl_cert = "/path/to/cert.pem"
ssl_key = "/path/to/key.pem"
```

## 功能使用

### Q: 如何上传大文件？

A: MyObj 支持大文件分片上传，默认阈值为 1GB。超过阈值的文件会自动分片上传，支持断点续传。

### Q: 秒传功能如何工作？

A: 系统使用 BLAKE3 哈希算法计算文件哈希值。如果服务器已存在相同哈希的文件，会直接创建文件引用，无需重新上传。

### Q: 如何创建分享链接？

A: 在文件列表中，选择要分享的文件，点击"分享"按钮，设置有效期和密码（可选），即可生成分享链接。

### Q: 如何配置 WebDAV？

A: 在 `config.toml` 中配置 `[webdav]` 部分，启用 WebDAV 服务。详细配置请查看 [WebDAV 使用指南](/guide/webdav)。

### Q: 如何启用 S3 服务？

A: 在 `config.toml` 中配置 `[s3]` 部分：
```toml
[s3]
enable = true
region = "us-east-1"
share_port = false
port = 9000
```
详细配置请查看 [S3 协议使用指南](/guide/s3)。

### Q: MyObj 会自动创建存储桶吗？

A: 是的，MyObj 支持自动创建存储桶。首次上传文件到指定的 Bucket 时，如果 Bucket 不存在，系统会自动创建，无需手动创建。

## S3 协议使用

### Q: 如何获取 S3 访问凭证？

A: 
1. 登录 Web 界面
2. 进入"设置" → "API Key"
3. 创建新的 API Key
4. Access Key ID 和 Secret Access Key 即为 S3 访问凭证

### Q: S3 服务支持哪些功能？

A: MyObj S3 服务完全兼容 AWS S3 协议，支持：
- Bucket 和 Object 操作
- Multipart Upload（大文件分片上传）
- 版本控制
- 预签名 URL
- CORS 配置
- 对象标签
- ACL（访问控制列表）
- Bucket 策略
- 生命周期管理
- 服务端加密（SSE-S3）

详细功能请查看 [S3 协议使用指南](/guide/s3)。

### Q: S3 服务端口如何配置？

A: 在 `config.toml` 中配置：
- `share_port = true`：与主服务共用 8080 端口
- `share_port = false`：使用独立端口（默认 9000）

### Q: 如何配置 S3 路径风格访问？

A: MyObj S3 默认使用路径风格访问（`http://endpoint/bucket/key`），这是自动配置的，无需手动设置。使用 AWS SDK 时，需要设置 `forcePathStyle(true)`。

### Q: S3 签名验证失败怎么办？

A: 
1. 检查 Access Key 和 Secret Key 是否正确
2. 确认系统时间同步（签名计算依赖时间戳）
3. 检查 API Key 是否有效
4. 查看日志获取详细错误信息

### Q: Bucket 名称有什么限制？

A: Bucket 名称需符合 S3 命名规范：
- 长度：3-63 个字符
- 只能包含：小写字母、数字、点(.)和连字符(-)
- 必须以字母或数字开头和结尾
- 不能包含连续的点
- 不能是 IP 地址格式

### Q: 支持哪些 ACL 类型？

A: MyObj 支持标准的 S3 Canned ACL：
- `private` - 私有访问
- `public-read` - 公开读
- `public-read-write` - 公开读写
- `authenticated-read` - 认证用户读
- `bucket-owner-read` - Bucket 所有者读
- `bucket-owner-full-control` - Bucket 所有者完全控制

## 框架集成

### Q: 如何在 RuoYi-Plus 框架中集成 MyObj S3？

A: 
1. 修改 `OssClient.java`，取消注释 ACL 相关代码（第 151 行和第 202 行）
2. 在 RuoYi-Plus 管理后台配置 OSS 存储
3. 填写 MyObj S3 的配置信息（endpoint、accessKey、secretKey 等）

详细步骤请查看 [RuoYi-Plus 框架集成指南](/guide/integration-ruoyi)。

### Q: RuoYi-Plus 中的权限类型如何映射到 S3 ACL？

A: RuoYi-Plus 框架将权限类型映射如下：
- `0` (PRIVATE) → S3 `private`
- `1` (PUBLIC_READ_WRITE) → S3 `public-read-write`
- `2` (PUBLIC_READ) → S3 `public-read`

### Q: 如何在 Java 项目中使用 MyObj S3？

A: 使用 AWS SDK for Java：
1. 添加 Maven 依赖（`software.amazon.awssdk:s3` 和 `software.amazon.awssdk:netty-nio-client`）
2. 创建 S3Client，设置 endpoint 和凭证
3. 使用 S3Client 进行文件操作

详细示例请查看 [S3 协议使用指南](/guide/s3)。

### Q: 如何在 Spring Boot 项目中集成？

A: 创建 S3Client 配置类，从配置文件读取参数并创建 Bean。详细示例请查看 [S3 协议使用指南](/guide/s3)。

## 故障排查

### Q: 上传文件失败怎么办？

A: 
1. 检查磁盘空间是否充足
2. 检查文件存储目录权限
3. 查看日志文件获取详细错误信息
4. 检查网络连接
5. 确认文件大小未超过限制

### Q: 无法访问 WebDAV？

A:
1. 确认 WebDAV 服务已启用
2. 检查防火墙设置
3. 确认端口未被占用
4. 检查用户权限
5. 查看 [WebDAV 使用指南](/guide/webdav)

### Q: 数据库连接失败？

A:
1. 检查数据库服务是否运行
2. 验证数据库连接信息是否正确
3. 确认数据库用户权限
4. 检查网络连接
5. 查看数据库日志

### Q: S3 服务连接失败？

A:
1. 检查 MyObj S3 服务是否启动
2. 检查 `config.toml` 中 S3 配置是否正确
3. 检查 endpoint 配置是否正确
4. 检查网络连接和防火墙设置
5. 查看日志文件获取详细错误信息

### Q: 存储桶不存在错误？

A:
1. MyObj 支持自动创建存储桶，首次上传时会自动创建
2. 如果仍然报错，检查桶名称是否符合 S3 命名规范
3. 确认 API Key 有创建存储桶的权限
4. 检查配置中的 bucketName 是否正确

### Q: 文件预览失败？

A:
1. 检查文件格式是否支持预览
2. 检查文件是否损坏
3. 检查浏览器是否支持该格式
4. 查看浏览器控制台错误信息
5. 查看 [文件预览指南](/guide/file-preview)

## 性能优化

### Q: 如何提高上传速度？

A:
1. 调整 `big_chunk_size` 配置项，增大分片大小
2. 使用 SSD 存储文件
3. 确保网络带宽充足
4. 启用多线程上传（如果支持）

### Q: 如何减少存储空间占用？

A:
1. 启用文件去重（秒传功能）
2. 定期清理临时文件
3. 压缩存储大文件
4. 清理回收站中的过期文件
5. 删除不需要的文件

### Q: 如何优化 S3 访问性能？

A:
1. 使用 CDN 加速（配置自定义域名）
2. 启用 HTTP/2
3. 使用连接池
4. 配置合适的超时时间
5. 使用异步客户端（如 AWS SDK 的异步客户端）

## 安全相关

### Q: 如何保护数据安全？

A:
1. 启用文件加密存储
2. 使用强密码
3. 定期备份数据
4. 启用 SSL/TLS
5. 限制 API Key 权限
6. 配置访问控制列表（ACL）
7. 使用私有权限存储敏感文件

### Q: 如何重置用户密码？

A: 使用 CLI 工具重置密码：
```bash
./myobj-cli user reset-password <username> <new-password>
```

### Q: API Key 泄露了怎么办？

A:
1. 立即删除泄露的 API Key
2. 创建新的 API Key
3. 更新所有使用该 API Key 的配置
4. 检查是否有异常访问记录
5. 必要时更换所有相关凭证

### Q: 如何配置 ACL 权限？

A:
1. 上传文件时设置 ACL（通过 `x-amz-acl` header 或 SDK 的 ACL 参数）
2. 使用 Bucket 策略控制访问
3. 在 RuoYi-Plus 中通过 `accessPolicy` 配置项设置
4. 详细说明请查看 [S3 协议使用指南](/guide/s3)

## 其他问题

### Q: 如何贡献代码？

A: 查看项目的 [贡献指南](https://github.com/dromara/MyObj/blob/master/README.md#贡献指南)。

### Q: 如何报告 Bug？

A: 在 [Gitee Issues](https://gitee.com/dromara/my-obj/issues) 创建 Issue，提供详细的问题描述和复现步骤。

### Q: 如何获取技术支持？

A: 
- 查看文档和 FAQ
- 在 [Gitee Issues](https://gitee.com/dromara/my-obj/issues) 提问
- 在 [讨论区](https://gitee.com/dromara/my-obj/discussions) 交流

### Q: 支持哪些编程语言？

A: MyObj S3 服务兼容所有支持 AWS S3 协议的 SDK，包括：
- Go（MinIO SDK、AWS SDK）
- Java（AWS SDK）
- Python（boto3）
- JavaScript/TypeScript（AWS SDK）
- .NET（AWS SDK）
- 其他语言的 S3 兼容 SDK

### Q: 如何备份数据？

A:
1. 定期备份数据库文件
2. 备份 `obj_data` 目录中的文件
3. 使用 S3 兼容的备份工具（如 rclone）备份到其他存储
4. 配置自动备份任务

### Q: 如何迁移数据？

A:
1. 备份数据库和文件数据
2. 在新服务器上安装 MyObj
3. 恢复数据库
4. 复制文件数据到新服务器的 `obj_data` 目录
5. 更新配置文件
6. 启动服务并验证
