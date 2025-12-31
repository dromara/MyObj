@echo off
REM 构建脚本 - Windows Batch版本
REM 用于编译前端、服务端（Linux版本）

echo 开始构建项目...

REM 构建前端
echo.
echo 构建前端项目...
cd webview
if exist dist (
    echo 清理旧的构建文件...
    rmdir /s /q dist
)
call npm run build
if %ERRORLEVEL% NEQ 0 (
    echo 前端构建失败！
    cd ..
    exit /b 1
)
echo 前端构建成功！
cd ..

REM 编译服务端（Linux版本）
echo.
echo 编译服务端（Linux版本）...
SET CGO_ENABLED=0
SET GOOS=linux
SET GOARCH=amd64
go build -ldflags="-s -w" -o bin/server src/cmd/server/main.go
if %ERRORLEVEL% NEQ 0 (
    echo 服务端编译失败！
    exit /b 1
)
echo 服务端编译成功！

echo.
echo 所有构建完成！
echo 输出目录: bin/