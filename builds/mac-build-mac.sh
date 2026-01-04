#!/bin/bash
# Mac编译Mac版本
# 构建前端、后端，打包所有必要文件到bin目录

set -e

echo "========================================"
echo "Mac 编译 Mac 版本"
echo "========================================"
echo ""

# 获取脚本所在目录的父目录（项目根目录）
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
PROJECT_ROOT="$( cd "$SCRIPT_DIR/.." && pwd )"
cd "$PROJECT_ROOT"

# 设置目标平台
export GOOS=darwin
export GOARCH=amd64
export CGO_ENABLED=0

# 清理旧的bin目录
echo "[1/5] 清理旧的构建文件..."
rm -rf bin
mkdir -p bin

# 构建前端
echo ""
echo "[2/5] 构建前端项目..."
cd webview
if [ -d "dist" ]; then
    echo "清理旧的前端构建文件..."
    rm -rf dist
fi
echo "安装前端依赖..."
npm install
if [ $? -ne 0 ]; then
    echo "前端依赖安装失败！"
    cd ..
    exit 1
fi
npm run build
if [ $? -ne 0 ]; then
    echo "前端构建失败！"
    cd ..
    exit 1
fi
echo "前端构建成功！"
cd ..

# 复制前端产物
echo ""
echo "[3/5] 复制前端产物到bin目录..."
cp -r webview/dist bin/webview/
if [ $? -ne 0 ]; then
    echo "复制前端产物失败！"
    exit 1
fi

# 编译后端
echo ""
echo "[4/5] 编译后端（Mac amd64）..."
go build -ldflags="-s -w" -o bin/server src/cmd/server/main.go
if [ $? -ne 0 ]; then
    echo "后端编译失败！"
    exit 1
fi
echo "后端编译成功！"

# 编译CLI工具
go build -ldflags="-s -w" -o bin/cli src/cmd/cli/main.go
if [ $? -ne 0 ]; then
    echo "CLI工具编译失败！"
    exit 1
fi
echo "CLI工具编译成功！"

# 给可执行文件添加执行权限
chmod +x bin/server
chmod +x bin/cli

# 复制必要的依赖和配置文件
echo ""
echo "[5/5] 复制依赖和配置文件..."
if [ -d "libs" ]; then
    cp -r libs bin/
fi
if [ -d "templates" ]; then
    cp -r templates bin/
fi
cp config.toml bin/
if [ -d "docs" ]; then
    cp -r docs bin/
fi

# 创建启动说明文件
echo ""
echo "创建启动说明文件..."
cat > bin/README.txt << 'EOF'
MyObj 文件存储系统 - Mac版本

使用说明：
1. 确保config.toml配置正确
2. 运行 ./server 启动服务
3. 默认访问地址：http://localhost:8080

注意事项：
- 首次运行会自动创建数据库
- 日志文件在 logs 目录下
- 上传文件存储在 obj_data 目录下

支持架构：
- Intel Mac: amd64
EOF

echo ""
echo "========================================"
echo "构建完成！"
echo "输出目录: bin/"
echo "目标平台: Mac amd64"
echo "========================================"
echo ""
echo "可以将bin目录复制到Mac系统直接运行"
