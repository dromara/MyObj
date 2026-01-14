# S3服务使用指南

## 🎯 概述

MyObj S3服务提供兼容AWS S3协议的对象存储API，可以使用MinIO SDK、AWS SDK或其他S3兼容工具进行访问。

## 📋 功能特性

### ✅ 已实现功能

- **Bucket操作**
  - ListBuckets - 列出所有Bucket
  - CreateBucket (PUT /:bucket) - 创建Bucket
  - HeadBucket (HEAD /:bucket) - 检查Bucket是否存在
  - DeleteBucket (DELETE /:bucket) - 删除空Bucket
  
- **认证机制**
  - AWS Signature V4签名验证
  - 基于API Key的访问控制
  
- **Bucket映射**
  - Bucket对应用户虚拟目录
  - 一个用户可以创建多个Bucket
  - Bucket名称符合S3命名规范

### 🚧 待实现功能

- Object操作 (PutObject, GetObject, DeleteObject等)
- ListObjects / ListObjectsV2
- Multipart Upload (大文件分片上传)
- Object元数据管理
- 版本控制

## 🚀 快速开始

### 1. 安装MinIO SDK依赖

```bash
go get github.com/minio/minio-go/v7
```

### 2. 启动S3服务

S3服务已集成到主服务中，启动服务器即可：

```bash
cd src/cmd/server
go run main.go
```

或编译后运行：

```bash
go build -o server src/cmd/server/main.go
./server
```

### 3. 配置S3服务

编辑 `config.toml`:

```toml
[s3]
# 是否启用 S3 服务
enable = true
# 区域名称
region = "us-east-1"
# 是否与主服务共用端口（true: 共用 8080，false: 使用独立端口）
share_port = true
# 独立端口（当 share_port = false 时生效）
port = 9000
# S3 API 路径前缀（留空表示根路径 /）
path_prefix = ""
```

### 4. 创建API Key

通过Web界面或CLI工具创建API Key作为S3访问凭证：

```bash
# 使用CLI工具（假设有实现）
./myobj-cli user create-api-key <username>
```

或通过Web管理界面创建。

### 5. 使用MinIO SDK测试

```go
package main

import (
    "context"
    "log"
    
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

func main() {
    // 初始化客户端
    client, err := minio.New("localhost:8080", &minio.Options{
        Creds:  credentials.NewStaticV4("your-access-key-id", "your-secret-key", ""),
        Secure: false,
        Region: "us-east-1",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // 列出所有Bucket
    buckets, err := client.ListBuckets(ctx)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, bucket := range buckets {
        log.Printf("Bucket: %s, Created: %v\n", bucket.Name, bucket.CreationDate)
    }
    
    // 创建Bucket
    err = client.MakeBucket(ctx, "my-bucket", minio.MakeBucketOptions{
        Region: "us-east-1",
    })
    if err != nil {
        log.Fatal(err)
    }
    
    log.Println("Bucket created successfully")
}
```

## 📦 架构设计

```
src/s3_server/
├── types/              # S3协议类型定义
│   ├── common.go       # 通用类型
│   └── errors.go       # 错误码定义
├── auth/               # 认证模块
│   └── signature.go    # AWS Signature V4验证
├── middleware/         # 中间件
│   └── auth.go         # S3认证中间件
├── handler/            # HTTP处理器
│   └── s3_handler.go   # S3 API处理器
├── router/             # 路由配置
│   └── s3_router.go    # S3路由
└── service/            # 业务逻辑
    └── bucket_service.go  # Bucket服务
```

## 🔧 配置说明

### Bucket命名规范

符合AWS S3 Bucket命名规范：
- 长度在3-63个字符之间
- 只能包含小写字母、数字、点(.)和连字符(-)
- 必须以字母或数字开头和结尾
- 不能包含连续的点
- 不能是IP地址格式

### 认证方式

使用AWS Signature V4签名机制：
- Access Key ID: 对应MyObj的API Key
- Secret Access Key: 对应API Key的私钥
- 签名计算方式与AWS S3完全兼容

## 📊 数据库表结构

### s3_buckets
Bucket信息表，每个Bucket对应一个用户虚拟目录

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 主键 |
| bucket_name | varchar(63) | Bucket名称 |
| user_id | varchar(36) | 用户ID |
| region | varchar(32) | 区域 |
| virtual_path_id | int | 虚拟路径ID |
| created_at | datetime | 创建时间 |
| updated_at | datetime | 更新时间 |

### s3_object_metadata
对象元数据表（待实现Object操作后使用）

### s3_multipart_uploads
分片上传会话表（待实现Multipart Upload后使用）

### s3_multipart_parts
分片信息表（待实现Multipart Upload后使用）

## 🧪 测试

### 运行测试

```bash
# 安装MinIO SDK
go get github.com/minio/minio-go/v7

# 运行S3兼容性测试
go test ./src/tests/s3_minio_sdk_test.go -v
```

### 使用s3cmd测试

```bash
# 配置s3cmd
s3cmd --configure

# 设置：
# Access Key: your-access-key-id
# Secret Key: your-secret-key
# Default Region: us-east-1
# S3 Endpoint: localhost:8080
# DNS-style bucket: No

# 列出Bucket
s3cmd ls

# 创建Bucket
s3cmd mb s3://test-bucket

# 上传文件（待实现）
s3cmd put file.txt s3://test-bucket/
```

## 🔍 故障排查

### 常见问题

**1. 签名验证失败**
```
Error: SignatureDoesNotMatch
```
解决方案：
- 检查Access Key和Secret Key是否正确
- 确认时间同步（签名计算依赖时间戳）
- 查看日志中的详细错误信息

**2. Bucket已存在**
```
Error: BucketAlreadyExists
```
解决方案：
- Bucket名称在用户空间内必须唯一
- 使用不同的Bucket名称

**3. Bucket名称不合法**
```
Error: InvalidBucketName
```
解决方案：
- 检查Bucket名称是否符合S3命名规范
- 只使用小写字母、数字、点和连字符
- 长度3-63个字符

## 📝 开发计划

### Phase 1: Bucket操作 ✅
- [x] ListBuckets
- [x] CreateBucket
- [x] HeadBucket
- [x] DeleteBucket
- [x] AWS Signature V4认证

### Phase 2: Object基础操作 🚧
- [ ] PutObject
- [ ] GetObject
- [ ] HeadObject
- [ ] DeleteObject
- [ ] ListObjects / ListObjectsV2

### Phase 3: Multipart Upload 📋
- [ ] InitiateMultipartUpload
- [ ] UploadPart
- [ ] CompleteMultipartUpload
- [ ] AbortMultipartUpload
- [ ] ListParts

### Phase 4: 高级特性 📋
- [ ] CopyObject
- [ ] Object版本控制
- [ ] 对象生命周期管理
- [ ] Bucket策略

## 🤝 贡献

欢迎提交PR和Issue！

## 📄 许可证

Apache License 2.0
