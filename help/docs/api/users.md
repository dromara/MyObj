# 用户管理 API

## 获取用户信息

```bash
GET /api/user/info
Authorization: Bearer <token>
```

## 更新用户信息

```bash
PUT /api/user/info
Authorization: Bearer <token>
Content-Type: application/json

{
  "nickname": "新昵称",
  "email": "new@example.com"
}
```

## 修改密码

```bash
POST /api/user/password
Authorization: Bearer <token>
Content-Type: application/json

{
  "old_password": "oldpass",
  "new_password": "newpass"
}
```

## API Key 管理

- `GET /api/user/api-keys` - 获取 API Key 列表
- `POST /api/user/api-keys` - 创建新的 API Key
- `DELETE /api/user/api-keys/:id` - 删除 API Key

详细 API 文档请查看 Swagger：`http://localhost:8080/swagger/index.html`
