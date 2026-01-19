# 在RuoYi-Plus 框架里集成

本文档介绍如何在 RuoYi-Vue-Plus 或 RuoYi-Cloud-Plus 中集成 MyObj S3 服务。

## 📋 概述

RuoYi-Plus 框架（包括 RuoYi-Vue-Plus 和 RuoYi-Cloud-Plus）采用插件式开发，OSS 模块代码位于 `ruoyi-common/ruoyi-common-oss` 中。MyObj 完全兼容 AWS S3 协议，可以直接作为 OSS 存储服务使用。

**注意**：RuoYi-Plus 框架与传统的 RuoYi 框架有本质区别，本文档仅适用于 RuoYi-Plus 系列框架。

**重要特性**：
- ✅ MyObj S3 支持自动创建存储桶，无需手动创建
- ✅ 使用 AWS SDK 异步客户端，性能优异
- ✅ 支持大文件分片上传和断点续传
- ✅ 完全兼容 RuoYi-Plus 框架现有的 OSS 架构

## 🚀 快速集成

### 1. 修改 OssClient 类

编辑 `ruoyi-common/ruoyi-common-oss/src/main/java/org/dromara/common/oss/core/OssClient.java`，取消注释 ACL 相关代码：

**第 151 行**（`upload(Path filePath, ...)` 方法）：

```java
// 修改前（被注释）
//.acl(getAccessPolicy().getObjectCannedACL())

// 修改后（取消注释）
.acl(getAccessPolicy().getObjectCannedACL())
```

**第 202 行**（`upload(InputStream inputStream, ...)` 方法）：

```java
// 修改前（被注释）
//.acl(getAccessPolicy().getObjectCannedACL())

// 修改后（取消注释）
.acl(getAccessPolicy().getObjectCannedACL())
```

### 2. 配置 OSS 存储

在 RuoYi-Plus 管理后台配置 OSS 存储，参考[官方文档](https://plus-doc.dromara.org/#/ruoyi-vue-plus/framework/basic/oss?id=%e9%85%8d%e7%bd%aeoss)：

1. 登录系统管理后台
2. 进入 **系统管理** → **对象存储** → **配置管理**
3. 点击 **新增** 按钮
4. 填写配置信息：

| 配置项 | 字段名 | 说明 | 示例值 |
|--------|--------|------|--------|
| **配置key** | `configKey` | 配置标识（唯一），用于区分不同的 OSS 配置 | `myobj` |
| **AccessKey** | `accessKey` | MyObj API Key ID | `your-access-key-id` |
| **SecretKey** | `secretKey` | MyObj API Key Secret | `your-secret-key` |
| **桶名称** | `bucketName` | Bucket 名称（MyObj 会自动创建，无需手动创建） | `my-bucket` |
| **前缀** | `prefix` | 文件路径前缀（可选），如 `uploads/` | `uploads/` |
| **访问站点** | `endpoint` | MyObj S3 服务地址 | `localhost:9000` 或 `your-domain:9000` |
| **自定义域名** | `domain` | 自定义域名（可选），用于 CDN 加速 | `https://cdn.your-domain.com` |
| **是否https** | `isHttps` | 是否使用 HTTPS | `0`（否）或 `1`（是） |
| **域** | `region` | 区域名称 | `us-east-1` |
| **桶权限类型** | `accessPolicy` | 访问策略 | `0`（私有）、`1`（公开读写）、`2`（公开读） |
| **是否默认** | `status` | 是否设为默认存储 | `0`（是）或 `1`（否） |
| **备注** | `remark` | 配置说明（可选） | `MyObj S3 存储配置` |

**配置说明**：
- 配置信息存储在数据库表 `sys_oss_config` 中
- 系统启动时会自动加载所有配置到 Redis 缓存
- 支持多个 OSS 配置，通过 `configKey` 区分
- 修改配置后会自动刷新缓存，无需重启服务

### 3. 配置详解

#### 访问站点（Endpoint）

- **开发环境**：`localhost:9000`（不需要协议前缀，系统会自动添加）
- **生产环境**：`your-domain:9000` 或 `s3.your-domain.com`
- **注意**：不要包含 `http://` 或 `https://`，由 `isHttps` 字段控制

#### 桶名称（Bucket Name）

- **自动创建**：MyObj 支持自动创建存储桶，配置后首次上传文件时会自动创建
- **命名规范**：需符合 S3 命名规范
  - 长度：3-63 个字符
  - 只能包含：小写字母、数字、点(.)和连字符(-)
  - 必须以字母或数字开头和结尾
  - 不能包含连续的点
  - 不能是 IP 地址格式

#### 域（Region）

- **默认值**：`us-east-1`
- **配置位置**：可在 MyObj 的 `config.toml` 中配置 `[s3].region`
- **说明**：MyObj 使用统一的区域配置，此处填写与 MyObj 配置一致即可

#### 桶权限类型（Access Policy）

RuoYi-Plus 框架将权限类型映射到 S3 的 Canned ACL：

- **0 - 私有（PRIVATE）**：映射到 S3 的 `private`，文件仅创建者可见，需要认证才能访问
- **1 - 公开读写（PUBLIC_READ_WRITE）**：映射到 S3 的 `public-read-write`，文件公开可访问和修改，无需认证
- **2 - 公开读（PUBLIC_READ）**：映射到 S3 的 `public-read`，文件公开可访问，但不可修改，无需认证

**注意**：MyObj S3 服务支持标准的 S3 Canned ACL，包括：
- `private` - 私有访问
- `public-read` - 公开读
- `public-read-write` - 公开读写
- `authenticated-read` - 认证用户读
- `bucket-owner-read` - Bucket 所有者读
- `bucket-owner-full-control` - Bucket 所有者完全控制

RuoYi-Plus 框架通过 `AccessPolicyType` 枚举将这三个值映射到对应的 S3 ACL。

#### 前缀（Prefix）

- **作用**：为所有上传的文件添加路径前缀
- **示例**：设置 `uploads/` 后，文件路径为 `uploads/2024/01/15/xxx.jpg`
- **可选**：不设置则直接使用日期路径

#### 自定义域名（Domain）

- **作用**：用于 CDN 加速或自定义访问地址
- **格式**：`https://cdn.your-domain.com` 或 `http://cdn.your-domain.com`
- **注意**：如果设置了自定义域名，文件 URL 会使用该域名而不是 endpoint

## 📝 代码示例

### 获取 OSS 客户端

RuoYi-Plus 支持两种方式获取 OSS 客户端：

```java
import org.dromara.common.oss.factory.OssFactory;

// 方式1：获取默认 OSS 客户端（status=0 的配置）
OssClient ossClient = OssFactory.instance();

// 方式2：根据 configKey 获取指定的 OSS 客户端
OssClient ossClient = OssFactory.instance("myobj");
```

### 文件上传

```java
import org.dromara.common.oss.factory.OssFactory;
import org.dromara.common.oss.entity.UploadResult;
import org.springframework.web.multipart.MultipartFile;

// 获取 OSS 客户端实例
OssClient ossClient = OssFactory.instance();

// 方式1：上传文件（自动生成路径）
UploadResult result = ossClient.uploadSuffix(
    file.getInputStream(), 
    ".jpg", 
    file.getSize(), 
    file.getContentType()
);

// 方式2：指定完整路径上传
UploadResult result = ossClient.upload(
    file.getInputStream(),
    "custom/path/image.jpg",
    file.getSize(),
    file.getContentType()
);

// 获取文件 URL
String fileUrl = result.getUrl();
String filename = result.getFilename();
String eTag = result.getETag();
```

### 文件下载

```java
// 下载文件到临时目录
Path tempFile = ossClient.fileDownload(filePath);

// 或下载到输出流
ossClient.download(key, outputStream, contentLength -> {
    System.out.println("文件大小: " + contentLength);
});
```

### 文件删除

```java
// 删除文件
ossClient.delete(filePath);
```

### 预签名 URL

```java
import java.time.Duration;

// 生成下载预签名 URL（有效期 1 小时）
String downloadUrl = ossClient.createPresignedGetUrl(
    objectKey, 
    Duration.ofHours(1)
);

// 生成上传预签名 URL（有效期 1 小时）
String uploadUrl = ossClient.createPresignedPutUrl(
    objectKey, 
    Duration.ofHours(1),
    null  // 元数据（可选）
);
```

## 🔧 依赖说明

RuoYi-Plus 项目已包含必要的依赖，无需额外添加：

```xml
<!-- AWS SDK for Java 2.x -->
<dependency>
    <groupId>software.amazon.awssdk</groupId>
    <artifactId>s3</artifactId>
</dependency>

<!-- Netty HTTP 客户端 -->
<dependency>
    <groupId>software.amazon.awssdk</groupId>
    <artifactId>netty-nio-client</artifactId>
</dependency>

<!-- S3 传输管理器 -->
<dependency>
    <groupId>software.amazon.awssdk</groupId>
    <artifactId>s3-transfer-manager</artifactId>
</dependency>
```

## 🔍 故障排查

### 连接失败

**错误信息**：
```
Unable to execute HTTP request: Connection refused
```

**解决方案**：
- 检查 MyObj S3 服务是否启动
- 检查 `访问站点` 配置是否正确
- 检查网络连接和防火墙设置

### 签名验证失败

**错误信息**：
```
The request signature we calculated does not match the signature you provided
```

**解决方案**：
- 检查 `AccessKey` 和 `SecretKey` 是否正确
- 确认时间同步（签名计算依赖时间戳）
- 检查 MyObj API Key 是否有效

### 存储桶不存在

**错误信息**：
```
The specified bucket does not exist
```

**解决方案**：
- MyObj 支持自动创建存储桶，首次上传文件时会自动创建
- 如果仍然报错，检查桶名称是否符合 S3 命名规范
- 确认 API Key 有创建存储桶的权限

### 路径风格问题

**说明**：
- RuoYi-Plus 的 `OssClient` 会自动判断是否需要使用路径风格
- 对于非云服务商（如 MyObj、MinIO），会自动启用路径风格访问
- 无需手动配置

## 📊 最佳实践

### 1. 多环境配置

建议为不同环境配置不同的 OSS 存储：

- **开发环境**：使用本地 MyObj 实例
- **测试环境**：使用测试服务器
- **生产环境**：使用生产服务器，配置自定义域名

### 2. 文件路径规范

RuoYi-Plus 已实现自动路径生成，格式为：

```
{前缀}/{日期}/{UUID}{后缀}
例如：uploads/2024/01/15/abc123.jpg
```

### 3. 大文件上传

RuoYi-Plus 使用 `S3TransferManager`，自动支持：
- 大文件分片上传
- 断点续传
- 上传进度监听

### 4. 安全建议

- 生产环境使用 HTTPS
- 定期更换 API Key
- 使用私有权限（Access Policy = 0）存储敏感文件
- 公开文件使用公开读权限（Access Policy = 2），避免误修改
- 配置自定义域名时使用 CDN 加速

## 📚 配置管理说明

### 配置存储

- **数据库表**：`sys_oss_config`
- **实体类**：`org.dromara.system.domain.SysOssConfig`
- **缓存**：配置信息存储在 Redis 中，键为 `sys_oss_config:{configKey}`
- **默认配置**：`status=0` 的配置为默认配置，存储在 `sys_oss_config:default_config_key`

### 配置初始化

系统启动时，`SysOssConfigServiceImpl#init()` 方法会：
1. 从数据库加载所有 OSS 配置
2. 将配置信息存入 Redis 缓存
3. 设置默认配置
4. 发布缓存刷新通知

### 配置切换

- 支持多个 OSS 配置同时存在
- 通过 `configKey` 区分不同的配置
- 修改配置后会自动刷新缓存，无需重启服务
- 可以通过修改 `status` 字段切换默认配置

## 🔗 相关文档

- [S3 协议使用指南](/guide/s3)
- [S3 API 文档](/api/s3)
- [RuoYi-Plus 官方文档](https://plus-doc.dromara.org/)
- [RuoYi-Plus OSS 配置文档](https://plus-doc.dromara.org/#/ruoyi-vue-plus/framework/basic/oss?id=%e9%85%8d%e7%bd%aeoss)
