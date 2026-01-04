@echo off
REM Windows编译Mac版本
REM 构建前端、后端，打包所有必要文件到bin目录

setlocal enabledelayedexpansion

echo ========================================
echo Windows 编译 Mac 版本
echo ========================================
echo.

REM 设置目标平台
SET GOOS=darwin
SET GOARCH=amd64
SET CGO_ENABLED=0

REM 清理旧的bin目录
echo [1/5] 清理旧的构建文件...
if exist bin (
    rmdir /s /q bin
)
mkdir bin

REM 构建前端
echo.
echo [2/5] 构建前端项目...
cd webview
if exist dist (
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

REM 复制前端产物
echo.
echo [3/5] 复制前端产物到bin目录...
xcopy /E /I /Y webview\dist bin\webview\dist
if %ERRORLEVEL% NEQ 0 (
    echo 复制前端产物失败！
    exit /b 1
)

REM 编译后端
echo.
echo [4/5] 编译后端（Mac amd64）...
go build -ldflags="-s -w" -o bin\server src\cmd\server\main.go
if %ERRORLEVEL% NEQ 0 (
    echo 后端编译失败！
    exit /b 1
)
echo 后端编译成功！

REM 编译CLI工具
go build -ldflags="-s -w" -o bin\cli src\cmd\cli\main.go
if %ERRORLEVEL% NEQ 0 (
    echo CLI工具编译失败！
    exit /b 1
)
echo CLI工具编译成功！

REM 复制必要的依赖和配置文件
echo.
echo [5/5] 复制依赖和配置文件...
if exist libs (
    xcopy /E /I /Y libs bin\libs
)
if exist templates (
    xcopy /E /I /Y templates bin\templates
)
copy /Y config.toml bin\config.toml
if exist docs (
    xcopy /E /I /Y docs bin\docs
)

REM 创建启动说明文件
echo.
echo 创建启动说明文件...
(
echo MyObj 文件存储系统 - Mac版本
echo.
echo 使用说明：
echo 1. 给server添加执行权限：chmod +x server
echo 2. 确保config.toml配置正确
echo 3. 运行 ./server 启动服务
echo 4. 默认访问地址：http://localhost:8080
echo.
echo 注意事项：
echo - 首次运行会自动创建数据库
echo - 日志文件在 logs 目录下
echo - 上传文件存储在 obj_data 目录下
echo.
echo 支持架构：
echo - Intel Mac: amd64
) > bin\README.txt

echo.
echo ========================================
echo 构建完成！
echo 输出目录: bin\
echo 目标平台: Mac amd64
echo ========================================
echo.
echo 可以将bin目录复制到Mac系统运行
pause
