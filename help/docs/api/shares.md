# 分享功能 API

## 创建分享链接

```bash
POST /api/share/create
Authorization: Bearer <token>
Content-Type: application/json

{
  "file_id": "123",
  "expire_hours": 24,
  "password": "share123"
}
```

## 获取分享列表

```bash
GET /api/share/list
Authorization: Bearer <token>
```

## 删除分享

```bash
DELETE /api/share/:id
Authorization: Bearer <token>
```

## 访问分享链接

```bash
GET /api/share/access/:share_id
```

### 带密码的分享

```bash
POST /api/share/access/:share_id
Content-Type: application/json

{
  "password": "share123"
}
```

详细 API 文档请查看 Swagger：`http://localhost:8080/swagger/index.html`
