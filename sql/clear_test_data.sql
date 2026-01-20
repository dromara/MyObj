-- 清空测试数据脚本
-- 保留：组(groups)、权限(power)、组权限(group_power)
-- 删除：其他所有表的数据
-- 重置：自增主键

-- 注意：执行此脚本前请务必备份数据库！

-- 开始事务
BEGIN TRANSACTION;

-- ================================
-- 1. 删除用户相关数据
-- ================================
DELETE FROM user_info;
DELETE FROM api_key;

-- ================================
-- 2. 删除文件相关数据
-- ================================
DELETE FROM user_files;
DELETE FROM file_info;
DELETE FROM file_chunk;
DELETE FROM virtual_path;

-- ================================
-- 3. 删除上传下载任务数据
-- ================================
DELETE FROM upload_task;
DELETE FROM upload_chunk;
DELETE FROM download_task;

-- ================================
-- 4. 删除分享和回收站数据
-- ================================
DELETE FROM shares;
DELETE FROM recycled;

-- ================================
-- 5. 删除磁盘和系统配置数据
-- ================================
DELETE FROM disk;
DELETE FROM sys_config;

-- ================================
-- 6. 删除S3服务相关数据
-- ================================
DELETE FROM s3_object_encryption;
DELETE FROM s3_encryption_keys;
DELETE FROM s3_bucket_lifecycle;
DELETE FROM s3_bucket_policy;
DELETE FROM s3_object_acl;
DELETE FROM s3_bucket_acl;
DELETE FROM s3_bucket_cors;
DELETE FROM s3_multipart_parts;
DELETE FROM s3_multipart_uploads;
DELETE FROM s3_object_metadata;
DELETE FROM s3_buckets;

-- ================================
-- 7. 重置自增主键
-- ================================
-- SQLite 的自增主键重置
DELETE FROM sqlite_sequence WHERE name IN (
    'user_info',
    'api_key',
    'user_files',
    'file_info',
    'file_chunk',
    'virtual_path',
    'upload_task',
    'upload_chunk',
    'download_task',
    'shares',
    'recycled',
    'disk',
    'sys_config',
    's3_buckets',
    's3_object_metadata',
    's3_multipart_parts',
    's3_bucket_cors',
    's3_bucket_acl',
    's3_object_acl',
    's3_bucket_policy',
    's3_bucket_lifecycle',
    's3_encryption_keys',
    's3_object_encryption'
);

-- ================================
-- 8. 保留的表（不做任何操作）
-- ================================
-- groups (组表) - 保留
-- power (权限表) - 保留
-- group_power (组权限关联表) - 保留

-- 提交事务
COMMIT;

-- 验证数据清除结果
SELECT '=== 数据清除完成，以下是保留的数据 ===' AS info;
SELECT 'groups 表记录数:' AS info, COUNT(*) AS count FROM groups;
SELECT 'power 表记录数:' AS info, COUNT(*) AS count FROM power;
SELECT 'group_power 表记录数:' AS info, COUNT(*) AS count FROM group_power;

SELECT '=== 以下表已清空 ===' AS info;
SELECT 'user_info 表记录数:' AS info, COUNT(*) AS count FROM user_info;
SELECT 'file_info 表记录数:' AS info, COUNT(*) AS count FROM file_info;
SELECT 'user_files 表记录数:' AS info, COUNT(*) AS count FROM user_files;
SELECT 'download_task 表记录数:' AS info, COUNT(*) AS count FROM download_task;
SELECT 'upload_task 表记录数:' AS info, COUNT(*) AS count FROM upload_task;
SELECT 'shares 表记录数:' AS info, COUNT(*) AS count FROM shares;
SELECT 'recycled 表记录数:' AS info, COUNT(*) AS count FROM recycled;
SELECT 'disk 表记录数:' AS info, COUNT(*) AS count FROM disk;
