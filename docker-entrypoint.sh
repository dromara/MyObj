#!/bin/sh
set -e

# 检查数据目录中是否存在数据库文件
if [ ! -f "/app/libs/my_obj.db" ]; then
    echo "Database file not found in /app/libs/, copying from default..."
    # 确保目录存在
    mkdir -p /app/libs
    # 从镜像内置的默认数据库文件复制
    cp /app/default-libs/my_obj.db /app/libs/my_obj.db
    echo "Database file initialized successfully."
else
    echo "Database file already exists in /app/libs/, skipping initialization."
fi

# 执行原始命令
exec "$@"
