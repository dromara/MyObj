# 认证授权

MyObj 支持两种认证方式：JWT Token 和 API Key。

## JWT Token 认证

### 用户登录

```bash
POST /api/auth/login
Content-Type: application/json

{
  "username": "admin",
  "password": "password"
}
```

响应：

```json
{
  "code": 200,
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expire": "2024-01-01T12:00:00Z"
  }
}
```

### 使用 Token

在后续请求的请求头中添加：

```
Authorization: Bearer <your-token>
```

## API Key 认证

### 创建 API Key

在用户设置页面创建 API Key。

### 使用 API Key

在请求头中添加：

```
X-API-Key: <your-api-key>
```

## Token 刷新

Token 默认有效期为 2 小时，过期后需要重新登录。

## 安全建议

1. 不要在客户端代码中硬编码 Token
2. 使用 HTTPS 传输 Token
3. 定期更换 API Key
4. 限制 API Key 的权限范围
