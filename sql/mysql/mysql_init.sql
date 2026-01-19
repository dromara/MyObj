-- MySQL 数据库初始化脚本
-- 包含建表语句和初始数据
-- 基于 SQLite 数据库结构和 clear_test_data.sql 脚本生成

-- 注意：执行此脚本前请确保已创建数据库
-- CREATE DATABASE IF NOT EXISTS myobj CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;
-- USE myobj;

-- 开始事务
START TRANSACTION;

-- ================================
-- 1. 删除已存在的表（如果存在）
-- ================================
DROP TABLE IF EXISTS `group_power`;
DROP TABLE IF EXISTS `power`;
DROP TABLE IF EXISTS `groups`;
DROP TABLE IF EXISTS `user_files`;
DROP TABLE IF EXISTS `file_chunk`;
DROP TABLE IF EXISTS `virtual_path`;
DROP TABLE IF EXISTS `upload_chunk`;
DROP TABLE IF EXISTS `upload_task`;
DROP TABLE IF EXISTS `download_task`;
DROP TABLE IF EXISTS `shares`;
DROP TABLE IF EXISTS `recycled`;
DROP TABLE IF EXISTS `disk`;
DROP TABLE IF EXISTS `sys_config`;
DROP TABLE IF EXISTS `api_key`;
DROP TABLE IF EXISTS `user_info`;
DROP TABLE IF EXISTS `file_info`;
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
-- 2. 创建基础权限表
-- ================================

-- 权限表
CREATE TABLE `power` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '权限ID',
    `name` VARCHAR(255) NOT NULL COMMENT '权限名称',
    `description` TEXT NOT NULL COMMENT '权限描述',
    `characteristic` TEXT NOT NULL COMMENT '权限特征',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='权限表';

-- 组表
CREATE TABLE `groups` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '组ID',
    `name` VARCHAR(255) NOT NULL COMMENT '组名称',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `group_default` INT NOT NULL COMMENT '是否为默认组 0-否 1-是',
    `space` BIGINT DEFAULT NULL COMMENT '组默认可用存储空间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组表';

-- 组权限关联表
CREATE TABLE `group_power` (
    `group_id` INT NOT NULL COMMENT '组ID',
    `power_id` INT NOT NULL COMMENT '权限ID',
    PRIMARY KEY (`group_id`, `power_id`),
    KEY `idx_power_id` (`power_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='组权限关联表';

-- ================================
-- 3. 创建用户相关表
-- ================================

-- 用户信息表
CREATE TABLE `user_info` (
    `id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `name` VARCHAR(255) NOT NULL COMMENT '用户昵称',
    `user_name` VARCHAR(255) NOT NULL COMMENT '用户名',
    `password` TEXT NOT NULL COMMENT '用户密码',
    `email` TEXT NOT NULL COMMENT '用户邮箱',
    `phone` VARCHAR(20) NOT NULL COMMENT '用户手机号',
    `group_id` INT NOT NULL COMMENT '用户组ID',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `space` BIGINT DEFAULT NULL COMMENT '用户可用存储空间',
    `file_password` TEXT DEFAULT NULL COMMENT '用户文件密码',
    `free_space` BIGINT DEFAULT NULL COMMENT '用户剩余存储空间',
    `state` INT NOT NULL DEFAULT 0 COMMENT '用户状态 0正常 1禁用',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_group_id` (`group_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户信息表';

-- API密钥表
CREATE TABLE `api_key` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT 'API密钥ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `key` TEXT NOT NULL COMMENT 'API密钥',
    `expires_at` DATETIME DEFAULT NULL COMMENT '过期时间',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `private_key` TEXT NOT NULL COMMENT '私钥',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='API密钥表';

-- ================================
-- 4. 创建文件相关表
-- ================================

-- 文件信息表
CREATE TABLE `file_info` (
    `id` VARCHAR(64) NOT NULL COMMENT '文件ID',
    `name` VARCHAR(255) NOT NULL COMMENT '文件原名',
    `random_name` VARCHAR(255) NOT NULL COMMENT '文件存储名（随机生成）',
    `size` BIGINT NOT NULL COMMENT '文件大小',
    `mime` VARCHAR(255) NOT NULL COMMENT '文件MIME类型',
    `thumbnail_img` TEXT DEFAULT NULL COMMENT '缩略图路径',
    `path` TEXT DEFAULT NULL COMMENT '文件实际存储路径',
    `file_hash` TEXT NOT NULL COMMENT '文件哈希值（全量hash）',
    `file_enc_hash` TEXT DEFAULT NULL COMMENT '加密文件哈希值',
    `chunk_signature` TEXT DEFAULT NULL COMMENT '分片签名（快速预检）',
    `first_chunk_hash` TEXT DEFAULT NULL COMMENT '第一个分片hash',
    `second_chunk_hash` TEXT DEFAULT NULL COMMENT '第二个分片hash',
    `third_chunk_hash` TEXT DEFAULT NULL COMMENT '第三个分片hash',
    `has_full_hash` BOOLEAN DEFAULT FALSE COMMENT '是否已计算全量hash',
    `is_enc` BOOLEAN DEFAULT FALSE COMMENT '是否加密',
    `is_chunk` BOOLEAN NOT NULL COMMENT '是否分块存储',
    `chunk_count` INT DEFAULT NULL COMMENT '分块数量',
    `enc_path` TEXT NOT NULL COMMENT '加密文件路径',
    `created_at` DATETIME DEFAULT NULL COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_file_hash` (`file_hash`(255)),
    KEY `idx_chunk_signature` (`chunk_signature`(255)),
    KEY `idx_mime` (`mime`),
    KEY `idx_name` (`name`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件信息表';

-- 用户文件关联表
CREATE TABLE `user_files` (
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `file_id` VARCHAR(64) NOT NULL COMMENT '文件ID',
    `file_name` TEXT NOT NULL COMMENT '文件名',
    `virtual_path` TEXT NOT NULL COMMENT '虚拟路径',
    `public` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否公开',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `deleted_at` DATETIME DEFAULT NULL COMMENT '删除时间',
    `uf_id` VARCHAR(64) NOT NULL COMMENT '用户文件ID',
    PRIMARY KEY (`uf_id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='用户文件关联表';

-- 文件分片表
CREATE TABLE `file_chunk` (
    `id` VARCHAR(64) NOT NULL COMMENT '分片ID',
    `file_id` VARCHAR(64) NOT NULL COMMENT '文件ID',
    `chunk_path` TEXT NOT NULL COMMENT '分片文件路径',
    `chunk_size` BIGINT NOT NULL COMMENT '分片文件大小',
    `chunk_hash` TEXT NOT NULL COMMENT '分片文件哈希',
    `chunk_index` INT NOT NULL COMMENT '分片文件索引',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='文件分片表';

-- 虚拟路径表
CREATE TABLE `virtual_path` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `path` TEXT NOT NULL COMMENT '虚拟路径',
    `is_file` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否为文件',
    `is_dir` BOOLEAN NOT NULL DEFAULT TRUE COMMENT '是否为目录',
    `parent_level` TEXT DEFAULT NULL COMMENT '父级层级信息',
    `created_time` DATETIME NOT NULL COMMENT '创建时间',
    `update_time` DATETIME NOT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='虚拟路径表';

-- ================================
-- 5. 创建上传下载任务表
-- ================================

-- 上传任务表
CREATE TABLE `upload_task` (
    `id` VARCHAR(64) NOT NULL COMMENT '任务ID',
    `user_id` VARCHAR(64) DEFAULT NULL COMMENT '用户ID',
    `file_name` TEXT NOT NULL COMMENT '文件名',
    `file_size` BIGINT NOT NULL COMMENT '文件大小（字节）',
    `chunk_size` BIGINT NOT NULL DEFAULT 5242880 COMMENT '分片大小（字节，默认5MB）',
    `total_chunks` INT NOT NULL COMMENT '总分片数',
    `uploaded_chunks` INT DEFAULT 0 COMMENT '已上传分片数',
    `chunk_signature` TEXT DEFAULT NULL COMMENT '文件hash签名（用于秒传检测）',
    `path_id` TEXT DEFAULT NULL COMMENT '路径ID',
    `temp_dir` TEXT DEFAULT NULL COMMENT '临时目录路径',
    `status` VARCHAR(20) DEFAULT 'pending' COMMENT '任务状态（pending/uploading/completed/failed/aborted）',
    `error_message` TEXT DEFAULT NULL COMMENT '错误信息',
    `create_time` DATETIME DEFAULT NULL COMMENT '创建时间',
    `update_time` DATETIME DEFAULT NULL COMMENT '更新时间',
    `expire_time` DATETIME DEFAULT NULL COMMENT '过期时间（7天后自动清理）',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上传任务表（支持断点续传）';

-- 上传分片表
CREATE TABLE `upload_chunk` (
    `chunk_id` VARCHAR(64) NOT NULL COMMENT '分片ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `file_name` TEXT NOT NULL COMMENT '文件名',
    `file_size` INT DEFAULT NULL COMMENT '文件大小',
    `md5` TEXT DEFAULT NULL COMMENT 'MD5',
    `path_id` TEXT DEFAULT NULL COMMENT '路径ID',
    PRIMARY KEY (`chunk_id`),
    KEY `idx_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='上传分片表';

-- 下载任务表
CREATE TABLE `download_task` (
    `id` VARCHAR(64) NOT NULL COMMENT '任务ID',
    `user_id` VARCHAR(64) DEFAULT NULL COMMENT '用户ID',
    `file_id` VARCHAR(64) DEFAULT NULL COMMENT '文件ID',
    `file_name` TEXT DEFAULT NULL COMMENT '文件名',
    `file_size` BIGINT DEFAULT NULL COMMENT '文件大小',
    `downloaded_size` BIGINT DEFAULT 0 COMMENT '已下载大小',
    `progress` INT DEFAULT 0 COMMENT '下载进度 (0-100)',
    `speed` BIGINT DEFAULT 0 COMMENT '下载速度 (字节/秒)',
    `type` INT NOT NULL COMMENT '任务类型',
    `url` TEXT DEFAULT NULL COMMENT '下载URL',
    `path` TEXT DEFAULT NULL COMMENT '下载路径',
    `virtual_path` TEXT DEFAULT NULL COMMENT '虚拟路径',
    `state` INT DEFAULT NULL COMMENT '任务状态',
    `error_msg` TEXT DEFAULT NULL COMMENT '错误信息',
    `target_dir` TEXT DEFAULT NULL COMMENT '目标临时目录',
    `support_range` BOOLEAN DEFAULT FALSE COMMENT '是否支持断点续传',
    `enable_encryption` BOOLEAN DEFAULT FALSE COMMENT '是否加密存储',
    `info_hash` TEXT DEFAULT NULL COMMENT '种子InfoHash（BT/磁力链任务）',
    `file_index` INT DEFAULT NULL COMMENT '种子内文件索引（BT/磁力链任务）',
    `torrent_name` TEXT DEFAULT NULL COMMENT '种子名称（BT/磁力链任务）',
    `create_time` DATETIME DEFAULT NULL COMMENT '创建时间',
    `update_time` DATETIME DEFAULT NULL COMMENT '更新时间',
    `finish_time` DATETIME DEFAULT NULL COMMENT '完成时间',
    PRIMARY KEY (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_info_hash` (`info_hash`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='下载任务表';

-- ================================
-- 6. 创建分享和回收站表
-- ================================

-- 分享表
CREATE TABLE `shares` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '分享记录ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `file_id` VARCHAR(64) NOT NULL COMMENT '文件ID',
    `token` TEXT NOT NULL COMMENT '分享令牌',
    `expires_at` DATETIME NOT NULL COMMENT '分享过期时间',
    `password_hash` TEXT NOT NULL COMMENT '访问密码哈希',
    `download_count` INT NOT NULL DEFAULT 0 COMMENT '下载次数统计',
    `created_at` DATETIME NOT NULL COMMENT '分享创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='分享表';

-- 回收站表
CREATE TABLE `recycled` (
    `id` VARCHAR(64) NOT NULL COMMENT '回收站ID',
    `file_id` VARCHAR(64) NOT NULL COMMENT '文件ID',
    `user_id` VARCHAR(64) NOT NULL COMMENT '用户ID',
    `created_at` DATETIME NOT NULL COMMENT '删除时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_user_id` (`user_id`),
    KEY `idx_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='回收站表';

-- ================================
-- 7. 创建磁盘和系统配置表
-- ================================

-- 磁盘表
CREATE TABLE `disk` (
    `id` VARCHAR(64) NOT NULL COMMENT '磁盘ID',
    `size` INT NOT NULL COMMENT '磁盘总大小',
    `disk_path` TEXT NOT NULL COMMENT '磁盘路径',
    `data_path` TEXT NOT NULL COMMENT '数据存储路径',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_disk_path` (`disk_path`(255))
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='磁盘信息表';

-- 系统配置表
CREATE TABLE `sys_config` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '配置ID',
    `key` VARCHAR(255) NOT NULL COMMENT '配置键',
    `value` TEXT NOT NULL COMMENT '配置值',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_key` (`key`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='系统配置表';

-- ================================
-- 8. 创建S3服务相关表
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

-- ================================
-- 9. 插入初始数据
-- ================================

-- 插入组数据
INSERT INTO `groups` (`id`, `name`, `created_at`, `group_default`, `space`) VALUES
(1, 'administer', '2025-11-10 23:04:08', 0, NULL),
(2, 'user', '2025-11-15 23:23:29', 1, 500);

-- 插入权限数据
INSERT INTO `power` (`id`, `name`, `description`, `created_at`, `characteristic`) VALUES
(1, '用户查看', '查看系统所有用户', '2025-11-09 22:35:22', 'user:get'),
(2, '用户修改', '修改系统用户信息', '2025-11-09 22:35:50', 'user:update'),
(3, '用户删除', '删除系统用户', '2025-11-09 22:36:07', 'user:delete'),
(4, '用户停用', '暂停用户所有功能', '2025-11-09 22:36:26', 'user:state'),
(5, '用户空间分配', '分配用户可用空间大小', '2025-11-09 22:36:58', 'user:space'),
(6, '挂载磁盘', '挂载系统可用磁盘', '2025-11-09 23:35:06', 'disk:mount'),
(7, '删除挂载磁盘', '删除已经挂载的磁盘', '2025-11-10 00:27:35', 'disk:delete'),
(8, '查看挂载磁盘', '查看已经挂载磁盘的信息', '2025-11-10 00:27:59', 'disk:get'),
(9, '上传文件', '上传文件到磁盘', '2025-11-10 23:08:13', 'file:upload'),
(10, '重命名文件', '重命名磁盘文件', '2025-11-10 23:08:28', 'file:rechristen'),
(11, '分享文件', '创建文件分享链接', '2025-11-10 23:08:47', 'file:share'),
(12, '下载文件', '下载磁盘中的文件', '2025-11-10 23:11:02', 'file:download'),
(13, '离线下载', '离线下载文件到磁盘', '2025-11-10 23:13:30', 'file:offLine'),
(14, '文件保险箱', '加密文件的上传修改下载', '2025-11-10 23:15:34', 'file:insurance'),
(15, '文件预览', '查看文件和预览支持格式的文件', '2025-11-10 23:15:48', 'file:preview'),
(16, '创建目录', '创建文件目录', '2025-11-10 23:16:34', 'dir:create'),
(17, '删除目录', '删除已经存在的目录', '2025-11-10 23:16:48', 'dir:delete'),
(18, '创建apikey', '创建当前用户权限的apikey', '2025-11-10 23:18:35', 'apikey:create'),
(19, '删除apikey', '删除当前用户已存在的apikey', '2025-11-10 23:57:52', 'apikey:delete'),
(20, '修改其他用户信息', '修改其他用户信息，包括密码', '2025-11-12 20:52:19', 'user:update:else'),
(21, '用户密码修改', '修改用户自身密码', '2025-11-13 01:23:28', 'user:update:password'),
(22, '用户文件密码', '设置，修改文件密码', '2025-11-13 19:14:46', 'file:update:filePassword'),
(23, '移动文件', '移动文件至其他虚拟目录', '2025-11-18 01:17:59', 'file:move'),
(24, '删除文件', '删除文件（移动到回收站）', '2025-12-11 19:02:02', 'file:delete'),
(25, 'WebDAV访问', '允许通过WebDAV协议访问文件系统', '2025-12-30 07:34:05', 'webdav:access');

-- 插入组权限关联数据
INSERT INTO `group_power` (`group_id`, `power_id`) VALUES
(1, 1),
(1, 2),
(1, 3),
(1, 4),
(1, 5),
(1, 6),
(1, 7),
(1, 8),
(1, 9),
(1, 10),
(1, 11),
(1, 12),
(1, 13),
(1, 14),
(1, 15),
(1, 16),
(1, 17),
(1, 18),
(1, 19),
(1, 20),
(1, 21),
(1, 22),
(1, 23),
(1, 24),
(1, 25),
(2, 9),
(2, 10),
(2, 11),
(2, 12),
(2, 13),
(2, 14),
(2, 15),
(2, 16),
(2, 17),
(2, 18),
(2, 19),
(2, 22),
(2, 23),
(2, 24),
(2, 25);

-- 提交事务
COMMIT;

-- ================================
-- 10. 验证初始数据
-- ================================
SELECT '=== 数据库初始化完成，以下是初始数据统计 ===' AS info;
SELECT 'groups 表记录数:' AS info, COUNT(*) AS count FROM `groups`;
SELECT 'power 表记录数:' AS info, COUNT(*) AS count FROM `power`;
SELECT 'group_power 表记录数:' AS info, COUNT(*) AS count FROM `group_power`;

SELECT '=== 所有表已创建 ===' AS info;
SHOW TABLES;
