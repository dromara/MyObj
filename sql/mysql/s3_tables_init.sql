-- MySQL S3服务相关表初始化脚本
-- 用于在现有数据库中单独添加S3服务相关的表
-- 如果您已经运行过 mysql_init.sql，则不需要再执行此脚本

-- 注意：执行此脚本前请确保已创建数据库并使用
-- USE myobj;

-- 开始事务
START TRANSACTION;

-- ================================
-- 删除已存在的S3表（如果存在）
-- ================================
DROP TABLE IF EXISTS `s3_object_encryption`;
DROP TABLE IF EXISTS `s3_encryption_keys`;
DROP TABLE IF EXISTS `s3_bucket_lifecycle`;
DROP TABLE IF EXISTS `s3_bucket_policy`;
DROP TABLE IF EXISTS `s3_object_acl`;
DROP TABLE IF EXISTS `s3_bucket_acl`;
DROP TABLE IF EXISTS `s3_bucket_cors`;
DROP TABLE IF EXISTS `s3_multipart_parts`;
DROP TABLE IF EXISTS `s3_multipart_uploads`;
DROP TABLE IF EXISTS `s3_object_metadata`;
DROP TABLE IF EXISTS `s3_buckets`;

-- ================================
-- 创建S3服务相关表
-- ================================

-- S3存储桶表
CREATE TABLE `s3_buckets` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '存储桶ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称（符合S3命名规范）',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `region` VARCHAR(32) DEFAULT 'us-east-1' COMMENT '区域',
    `virtual_path_id` INT NOT NULL COMMENT '关联到虚拟路径ID',
    `versioning` VARCHAR(16) DEFAULT 'Disabled' COMMENT '版本控制状态：Enabled/Suspended/Disabled',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_bucket_user` (`bucket_name`, `user_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_virtual_path_id` (`virtual_path_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3存储桶表';

-- S3对象元数据表
CREATE TABLE `s3_object_metadata` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '对象元数据ID',
    `file_id` VARCHAR(36) DEFAULT NULL COMMENT '关联FileInfo.ID（DeleteMarker时为空）',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `object_key` VARCHAR(1024) NOT NULL COMMENT 'S3对象键名',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `etag` VARCHAR(64) DEFAULT NULL COMMENT 'MD5或BLAKE3哈希（DeleteMarker时为空）',
    `storage_class` VARCHAR(32) DEFAULT 'STANDARD' COMMENT '存储类别',
    `content_type` VARCHAR(256) DEFAULT NULL COMMENT '内容类型',
    `user_metadata` TEXT DEFAULT NULL COMMENT 'JSON格式存储x-amz-meta-*',
    `tags` TEXT DEFAULT NULL COMMENT 'JSON格式存储对象标签',
    `version_id` VARCHAR(36) DEFAULT NULL COMMENT '版本控制ID',
    `is_latest` BOOLEAN DEFAULT TRUE COMMENT '是否为最新版本',
    `is_delete_marker` BOOLEAN DEFAULT FALSE COMMENT '是否为删除标记',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_bucket_key` (`bucket_name`, `object_key`(255)),
    KEY `idx_file_id` (`file_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_version_id` (`version_id`),
    KEY `idx_is_latest` (`is_latest`),
    KEY `idx_is_delete_marker` (`is_delete_marker`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3对象元数据表';

-- S3分片上传会话表
CREATE TABLE `s3_multipart_uploads` (
    `upload_id` VARCHAR(64) NOT NULL COMMENT '上传会话ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `object_key` VARCHAR(1024) NOT NULL COMMENT '对象键名',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `metadata` TEXT DEFAULT NULL COMMENT 'JSON格式元数据',
    `status` VARCHAR(32) DEFAULT 'in-progress' COMMENT '状态：in-progress/completed/aborted',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`upload_id`),
    KEY `idx_bucket_name` (`bucket_name`),
    KEY `idx_object_key` (`object_key`(255)),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3分片上传会话表';

-- S3分片信息表
CREATE TABLE `s3_multipart_parts` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '分片ID',
    `upload_id` VARCHAR(64) NOT NULL COMMENT '上传会话ID',
    `part_number` INT NOT NULL COMMENT '分片编号',
    `etag` VARCHAR(64) NOT NULL COMMENT 'ETag',
    `size` BIGINT NOT NULL COMMENT '分片大小',
    `chunk_path` VARCHAR(512) DEFAULT NULL COMMENT '临时分片路径',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    KEY `idx_upload_part` (`upload_id`, `part_number`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3分片信息表';

-- S3 Bucket CORS配置表
CREATE TABLE `s3_bucket_cors` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'CORS配置ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `cors_config` TEXT NOT NULL COMMENT 'JSON格式存储CORS规则',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_bucket_cors` (`bucket_name`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3 Bucket CORS配置表';

-- S3 Bucket ACL配置表
CREATE TABLE `s3_bucket_acl` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'ACL配置ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `acl_config` TEXT NOT NULL COMMENT 'JSON格式存储ACL配置',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_bucket_acl` (`bucket_name`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3 Bucket ACL配置表';

-- S3 Object ACL配置表
CREATE TABLE `s3_object_acl` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'ACL配置ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `object_key` VARCHAR(1024) NOT NULL COMMENT '对象键名',
    `version_id` VARCHAR(36) DEFAULT NULL COMMENT '版本ID（支持版本控制）',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `acl_config` TEXT NOT NULL COMMENT 'JSON格式存储ACL配置',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_object_acl` (`bucket_name`, `object_key`(255), `version_id`, `user_id`(36)),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3 Object ACL配置表';

-- S3 Bucket Policy配置表
CREATE TABLE `s3_bucket_policy` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'Policy配置ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `policy_json` TEXT NOT NULL COMMENT 'JSON格式存储Bucket Policy',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_bucket_policy` (`bucket_name`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3 Bucket Policy配置表';

-- S3 Bucket Lifecycle配置表
CREATE TABLE `s3_bucket_lifecycle` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'Lifecycle配置ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `lifecycle_json` TEXT NOT NULL COMMENT 'JSON格式存储Lifecycle规则',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_bucket_lifecycle` (`bucket_name`, `user_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3 Bucket Lifecycle配置表';

-- S3加密密钥表
CREATE TABLE `s3_encryption_keys` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '密钥ID',
    `key_id` VARCHAR(64) NOT NULL COMMENT '密钥ID（用于标识）',
    `key_data` TEXT NOT NULL COMMENT '加密后的密钥数据（base64）',
    `algorithm` VARCHAR(32) DEFAULT 'AES256' COMMENT '加密算法（AES256等）',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `idx_key_id` (`key_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3加密密钥表';

-- S3对象加密元数据表
CREATE TABLE `s3_object_encryption` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '加密元数据ID',
    `bucket_name` VARCHAR(63) NOT NULL COMMENT 'Bucket名称',
    `object_key` VARCHAR(1024) NOT NULL COMMENT '对象键名',
    `version_id` VARCHAR(36) DEFAULT NULL COMMENT '版本ID（支持版本控制）',
    `user_id` VARCHAR(36) NOT NULL COMMENT '用户ID',
    `encryption_type` VARCHAR(32) NOT NULL COMMENT '加密类型：SSE-S3, SSE-C, SSE-KMS',
    `algorithm` VARCHAR(32) DEFAULT 'AES256' COMMENT '加密算法：AES256等',
    `key_id` VARCHAR(64) DEFAULT NULL COMMENT '密钥ID（SSE-S3或SSE-KMS）',
    `encrypted_key` TEXT DEFAULT NULL COMMENT '加密的密钥（SSE-C时使用）',
    `iv` VARCHAR(64) DEFAULT NULL COMMENT '初始化向量（base64）',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    KEY `idx_bucket_key` (`bucket_name`, `object_key`(255), `version_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='S3对象加密元数据表';

-- 提交事务
COMMIT;

-- ================================
-- 验证表创建结果
-- ================================
SELECT '=== S3服务相关表创建完成 ===' AS info;
SHOW TABLES LIKE 's3_%';

SELECT '=== 已创建以下11张S3相关表 ===' AS info;
SELECT 's3_buckets - S3存储桶表' AS info
UNION ALL SELECT 's3_object_metadata - S3对象元数据表'
UNION ALL SELECT 's3_multipart_uploads - S3分片上传会话表'
UNION ALL SELECT 's3_multipart_parts - S3分片信息表'
UNION ALL SELECT 's3_bucket_cors - S3 Bucket CORS配置表'
UNION ALL SELECT 's3_bucket_acl - S3 Bucket ACL配置表'
UNION ALL SELECT 's3_object_acl - S3 Object ACL配置表'
UNION ALL SELECT 's3_bucket_policy - S3 Bucket Policy配置表'
UNION ALL SELECT 's3_bucket_lifecycle - S3 Bucket Lifecycle配置表'
UNION ALL SELECT 's3_encryption_keys - S3加密密钥表'
UNION ALL SELECT 's3_object_encryption - S3对象加密元数据表';
