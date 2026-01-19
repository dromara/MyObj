# 文件操作 API

## 获取文件列表

```bash
GET /api/file/list?path=/&page=1&pageSize=20
Authorization: Bearer <token>
```

## 上传文件

### 简单上传

```bash
POST /api/file/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <file>
virtual_path: /
```

### 加密上传

```bash
POST /api/file/upload
Authorization: Bearer <token>
Content-Type: multipart/form-data

file: <file>
virtual_path: /
is_enc: true
file_password: mypassword
```

## 下载文件

```bash
GET /api/file/download/:file_id
Authorization: Bearer <token>
```

### 下载加密文件

```bash
GET /api/file/download/:file_id?password=mypassword
Authorization: Bearer <token>
```

## 删除文件

```bash
DELETE /api/file/:file_id
Authorization: Bearer <token>
```

## 文件操作

- `PUT /api/file/:id/rename` - 重命名文件
- `PUT /api/file/:id/move` - 移动文件
- `POST /api/file/copy` - 复制文件

详细 API 文档请查看 Swagger：`http://localhost:8080/swagger/index.html`
