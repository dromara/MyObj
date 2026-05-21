-- MySQL 数据库更新脚本
-- 用于在已有数据库（运行过 mysql_init.sql）上升级表结构
-- 基于企业空间功能扩展（enterprise branch）
--
-- 执行前请备份数据库
-- USE myobj;

START TRANSACTION;

-- ================================
-- 1. 更新 user_info 表（新增字段）
-- ================================

-- 是否无限空间
ALTER TABLE `user_info`
    ADD COLUMN IF NOT EXISTS `space_unlimited` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否无限空间';

-- 当前活跃企业ID
ALTER TABLE `user_info`
    ADD COLUMN IF NOT EXISTS `current_enterprise_id` VARCHAR(255) DEFAULT NULL COMMENT '当前活跃企业ID（空=个人空间）';

-- ================================
-- 2. 更新 api_key 表（新增字段）
-- ================================

ALTER TABLE `api_key`
    ADD COLUMN IF NOT EXISTS `s3_secret_key` TEXT DEFAULT NULL COMMENT 'S3密钥（用于HMAC-SHA256签名）';

-- ================================
-- 3. 修复 disk 表（size 列 INT → BIGINT）
-- ================================

ALTER TABLE `disk`
    MODIFY COLUMN `size` BIGINT NOT NULL COMMENT '磁盘总大小（字节）';

-- ================================
-- 4. 更新 audit_log 表（新增字段和索引）
-- ================================

ALTER TABLE `audit_log`
    MODIFY COLUMN `target_type` VARCHAR(64) NOT NULL COMMENT '目标类型';

ALTER TABLE `audit_log`
    ADD COLUMN IF NOT EXISTS `enterprise_id` VARCHAR(255) DEFAULT NULL COMMENT '企业ID（企业空间操作时记录）';

ALTER TABLE `audit_log`
    ADD INDEX IF NOT EXISTS `idx_audit_enterprise_id` (`enterprise_id`);

-- ================================
-- 5. 创建企业空间相关表
-- ================================

-- 企业信息表
CREATE TABLE IF NOT EXISTS `enterprise` (
    `id` VARCHAR(255) NOT NULL COMMENT '企业ID',
    `name` VARCHAR(255) NOT NULL COMMENT '企业名称',
    `logo` TEXT COMMENT '企业Logo',
    `description` TEXT COMMENT '企业描述',
    `creator_id` VARCHAR(255) NOT NULL COMMENT '创建者用户ID',
    `space` BIGINT NOT NULL DEFAULT 0 COMMENT '企业总存储空间',
    `free_space` BIGINT NOT NULL DEFAULT 0 COMMENT '企业剩余存储空间',
    `invite_code` VARCHAR(255) DEFAULT NULL COMMENT '邀请码',
    `invite_link` TEXT COMMENT '邀请链接',
    `state` INT NOT NULL DEFAULT 0 COMMENT '企业状态 0正常 1禁用',
    `space_unlimited` BOOLEAN NOT NULL DEFAULT FALSE COMMENT '是否无限空间',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    UNIQUE KEY `uk_invite_code` (`invite_code`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业信息表';

-- 企业成员表
CREATE TABLE IF NOT EXISTS `enterprise_member` (
    `id` VARCHAR(255) NOT NULL COMMENT '成员ID',
    `enterprise_id` VARCHAR(255) NOT NULL COMMENT '企业ID',
    `user_id` VARCHAR(255) NOT NULL COMMENT '用户ID',
    `role_id` VARCHAR(255) NOT NULL COMMENT '角色ID',
    `joined_at` DATETIME NOT NULL COMMENT '加入时间',
    `status` INT NOT NULL DEFAULT 0 COMMENT '成员状态 0正常 1禁用',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_em_enterprise_id` (`enterprise_id`),
    KEY `idx_em_user_id` (`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业成员表';

-- 企业角色表
CREATE TABLE IF NOT EXISTS `enterprise_role` (
    `id` VARCHAR(255) NOT NULL COMMENT '角色ID',
    `enterprise_id` VARCHAR(255) NOT NULL COMMENT '企业ID',
    `name` VARCHAR(255) NOT NULL COMMENT '角色名称',
    `is_default` INT NOT NULL DEFAULT 0 COMMENT '是否为默认角色 0否 1是',
    `is_admin` INT NOT NULL DEFAULT 0 COMMENT '是否为管理员角色 0否 1是',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_er_enterprise_id` (`enterprise_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业角色表';

-- 企业角色权限关联表
CREATE TABLE IF NOT EXISTS `enterprise_role_power` (
    `role_id` VARCHAR(255) NOT NULL COMMENT '角色ID',
    `power_id` INT NOT NULL COMMENT '权限ID',
    PRIMARY KEY (`role_id`, `power_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业角色权限关联表';

-- 企业邀请记录表
CREATE TABLE IF NOT EXISTS `enterprise_invite` (
    `id` VARCHAR(255) NOT NULL COMMENT '邀请记录ID',
    `enterprise_id` VARCHAR(255) NOT NULL COMMENT '企业ID',
    `inviter_id` VARCHAR(255) NOT NULL COMMENT '邀请者用户ID',
    `invitee_id` VARCHAR(255) DEFAULT NULL COMMENT '被邀请者用户ID',
    `invite_code` VARCHAR(255) DEFAULT NULL COMMENT '邀请码',
    `type` INT NOT NULL COMMENT '邀请类型',
    `status` INT NOT NULL DEFAULT 0 COMMENT '邀请状态 0待接受 1已接受 2已拒绝 3已过期',
    `expire_at` DATETIME DEFAULT NULL COMMENT '过期时间',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_ei_enterprise_id` (`enterprise_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业邀请记录表';

-- 企业共享空间目录表
CREATE TABLE IF NOT EXISTS `enterprise_shared_path` (
    `id` INT NOT NULL AUTO_INCREMENT COMMENT '目录ID',
    `enterprise_id` VARCHAR(255) NOT NULL COMMENT '企业ID',
    `name` VARCHAR(255) NOT NULL COMMENT '目录名称',
    `parent_id` INT NOT NULL DEFAULT 0 COMMENT '父级目录ID',
    `created_by` VARCHAR(255) DEFAULT NULL COMMENT '创建者用户ID',
    `updated_by` VARCHAR(255) DEFAULT NULL COMMENT '更新者用户ID',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_esp_enterprise_id` (`enterprise_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业共享空间目录表';

-- 企业共享空间文件关联表
CREATE TABLE IF NOT EXISTS `enterprise_shared_file` (
    `id` VARCHAR(255) NOT NULL COMMENT '关联ID',
    `enterprise_id` VARCHAR(255) NOT NULL COMMENT '企业ID',
    `file_id` VARCHAR(255) NOT NULL COMMENT '文件ID',
    `file_name` VARCHAR(255) NOT NULL COMMENT '文件名',
    `path_id` INT NOT NULL DEFAULT 0 COMMENT '目录ID',
    `uploader_id` VARCHAR(255) NOT NULL COMMENT '上传者用户ID',
    `size` BIGINT NOT NULL COMMENT '文件大小',
    `updated_by` VARCHAR(255) DEFAULT NULL COMMENT '更新者用户ID',
    `created_at` DATETIME NOT NULL COMMENT '创建时间',
    `updated_at` DATETIME DEFAULT NULL COMMENT '更新时间',
    PRIMARY KEY (`id`),
    UNIQUE KEY `uk_id` (`id`),
    KEY `idx_esf_enterprise_id` (`enterprise_id`),
    KEY `idx_esf_file_id` (`file_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci COMMENT='企业共享空间文件关联表';

COMMIT;

-- ================================
-- 验证更新结果
-- ================================
SELECT '=== 数据库更新完成 ===' AS info;
SELECT '已更新 user_info 表（新增 space_unlimited, current_enterprise_id）' AS status
UNION ALL SELECT '已更新 api_key 表（新增 s3_secret_key）'
UNION ALL SELECT '已更新 audit_log 表（新增 enterprise_id）'
UNION ALL SELECT '已创建 enterprise 表'
UNION ALL SELECT '已创建 enterprise_member 表'
UNION ALL SELECT '已创建 enterprise_role 表'
UNION ALL SELECT '已创建 enterprise_role_power 表'
UNION ALL SELECT '已创建 enterprise_invite 表'
UNION ALL SELECT '已创建 enterprise_shared_path 表'
UNION ALL SELECT '已创建 enterprise_shared_file 表';

SHOW TABLES LIKE 'enterprise%';
