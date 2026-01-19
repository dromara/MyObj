# API 概览

MyObj 提供了完整的 RESTful API 和 S3 兼容 API，支持文件管理、用户管理、分享等功能。

## 基础信息

### REST API

- **Base URL**: `http://localhost:8080/api`
- **认证方式**: JWT Token 或 API Key
- **数据格式**: JSON

### S3 API

- **Endpoint**: `http://localhost:8080` (共用端口) 或 `http://localhost:9000` (独立端口)
- **认证方式**: AWS Signature V4
- **Region**: 在 `config.toml` 中配置（默认 `us-east-1`）

## 认证

### JWT Token (REST API)

在请求头中添加：

```
Authorization: Bearer <your-token>
```

### API Key (REST API)

在请求头中添加：

```
X-API-Key: <your-api-key>
```

### AWS Signature V4 (S3 API)

使用 AWS Signature V4 签名机制，Access Key ID 和 Secret Access Key 对应 MyObj 的 API Key。

## API 分类

### REST API

MyObj 提供标准的 RESTful API：

- [认证授权](/api/authentication) - JWT 和 API Key 认证
- [文件操作](/api/files) - 文件相关 API
- [用户管理](/api/users) - 用户管理 API
- [分享功能](/api/shares) - 分享相关 API

### S3 API

MyObj 提供完整的 AWS S3 兼容 API：

- [S3 API 文档](/api/s3) - 完整的 S3 兼容 API，支持所有 S3 标准操作

## REST API 端点

### 认证相关

- `POST /api/auth/login` - 用户登录
- `POST /api/auth/register` - 用户注册
- `GET /api/auth/info` - 获取用户信息

### 文件操作

- `GET /api/file/list` - 获取文件列表
- `POST /api/file/upload` - 上传文件
- `GET /api/file/download/:id` - 下载文件
- `DELETE /api/file/:id` - 删除文件

### 用户管理

- `GET /api/user/info` - 获取用户信息
- `PUT /api/user/info` - 更新用户信息
- `POST /api/user/password` - 修改密码

### 分享功能

- `POST /api/share/create` - 创建分享链接
- `GET /api/share/list` - 获取分享列表
- `DELETE /api/share/:id` - 删除分享

## Swagger 文档

启动服务后，访问 `http://localhost:8080/swagger/index.html` 查看完整的 REST API 文档。

## 详细文档

- [认证授权](/api/authentication)
- [文件操作](/api/files)
- [用户管理](/api/users)
- [分享功能](/api/shares)
- [S3 API](/api/s3)
