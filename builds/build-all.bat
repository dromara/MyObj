@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

:: ========================================
:: MyObj 全平台交叉编译脚本
:: 从 Windows 一次性编译所有平台所有架构
:: 输出文件名与自动升级系统匹配: server-{os}-{arch}{ext}
:: ========================================

echo.
echo ========================================
echo   MyObj 全平台交叉编译
echo ========================================
echo.

:: 获取项目根目录
set "PROJECT_ROOT=%~dp0.."
cd /d "%PROJECT_ROOT%"

:: 读取版本号
set "VERSION=unknown"
if exist VERSION (
    set /p VERSION=<VERSION
    for %%i in (!VERSION!) do set "VERSION=%%i"
)
echo 当前版本: !VERSION!
echo.

:: 清理输出目录
echo [1/4] 清理旧的构建文件...
if exist dist rmdir /s /q dist
mkdir dist

:: 构建前端（只构建一次）
echo.
echo [2/4] 构建前端项目...
cd webview
if exist dist rmdir /s /q dist
call npm install
if !errorlevel! neq 0 (
    echo 前端依赖安装失败！
    cd ..
    exit /b 1
)
call npm run build
if !errorlevel! neq 0 (
    echo 前端构建失败！
    cd ..
    exit /b 1
)
echo 前端构建成功！
cd ..

:: 定义所有目标平台
:: 格式: GOOS GOARCH 后缀
set TARGETS=windows amd64 .exe windows arm64 .exe linux amd64  linux arm64  darwin amd64  darwin arm64 

:: 编译所有平台
echo.
echo [3/4] 交叉编译所有平台...

set "COUNT=0"
set "TOTAL=6"

for %%i in (%TARGETS%) do (
    set /a COUNT+=1
    if !COUNT! equ 1 set "GOOS=%%i"
    if !COUNT! equ 2 set "GOARCH=%%i"
    if !COUNT! equ 3 (
        set "EXT=%%i"

        echo.
        echo --- [!COUNT!/3] 编译 !GOOS!/!GOARCH! !EXT! ---

        set "OUT_NAME=server-!GOOS!-!GOARCH!!EXT!"
        set "OUT_DIR=dist/!GOOS!-!GOARCH!"
        mkdir "!OUT_DIR!" 2>nul

        :: 编译后端
        set "CGO_ENABLED=0"
        set "GOOS=!GOOS!"
        set "GOARCH=!GOARCH!"
        go build -ldflags="-s -w" -o "!OUT_DIR!/!OUT_NAME!" src/cmd/server/main.go
        if !errorlevel! neq 0 (
            echo 编译失败: !GOOS!/!GOARCH!
            exit /b 1
        )
        echo 后端编译成功: !OUT_NAME!

        :: 复制前端产物
        xcopy /e /i /q /y webview\dist "!OUT_DIR!\webview\dist" >nul

        :: 复制必要文件
        if exist libs xcopy /e /i /q /y libs "!OUT_DIR!\libs" >nul
        if exist templates xcopy /e /i /q /y templates "!OUT_DIR!\templates" >nul
        copy /y VERSION "!OUT_DIR!\VERSION" >nul 2>nul

        :: 复制配置示例
        if exist config.toml (
            copy /y config.toml "!OUT_DIR!\config.toml" >nul
        ) else (
            echo # 请复制 config.toml 到此目录并修改配置> "!OUT_DIR!\config.toml.example"
        )

        :: 打包
        echo 打包中...
        if "!GOOS!"=="windows" (
            powershell -Command "Compress-Archive -Path '!OUT_DIR!\*' -DestinationPath 'dist/!GOOS!-!GOARCH!.zip' -Force"
        ) else (
            powershell -Command "tar -czf 'dist/!GOOS!-!GOARCH!.tar.gz' -C '!OUT_DIR!' ."
        )
        echo 打包完成: !GOOS!-!GOARCH!

        set "COUNT=0"
    )
)

:: 生成升级用的二进制文件（直接以 server-{os}-{arch} 命名，不含前端）
echo.
echo [4/4] 生成升级用二进制文件...
mkdir dist\upgrade-bin 2>nul

for %%i in (%TARGETS%) do (
    set /a COUNT+=1
    if !COUNT! equ 1 set "GOOS=%%i"
    if !COUNT! equ 2 set "GOARCH=%%i"
    if !COUNT! equ 3 (
        set "EXT=%%i"
        set "OUT_NAME=server-!GOOS!-!GOARCH!!EXT!"
        set "UPGRADE_NAME=server-!GOOS!-!GOARCH!!EXT!"

        set "CGO_ENABLED=0"
        set "GOOS=!GOOS!"
        set "GOARCH=!GOARCH!"
        go build -ldflags="-s -w" -o "dist\upgrade-bin\!UPGRADE_NAME!" src/cmd/server\main.go
        if !errorlevel! neq 0 (
            echo 编译失败: !GOOS!/!GOARCH!
            exit /b 1
        )
        echo 升级文件: !UPGRADE_NAME!

        set "COUNT=0"
    )
)

echo.
echo ========================================
echo   构建完成！
echo ========================================
echo.
echo 输出目录: dist\
echo.
echo 完整包（含前端）:
echo   dist\windows-amd64.zip
echo   dist\windows-arm64.zip
echo   dist\linux-amd64.tar.gz
echo   dist\linux-arm64.tar.gz
echo   dist\darwin-amd64.tar.gz
echo   dist\darwin-arm64.tar.gz
echo.
echo 升级二进制（上传到 Gitee Release）:
echo   dist\upgrade-bin\server-windows-amd64.exe
echo   dist\upgrade-bin\server-windows-arm64.exe
echo   dist\upgrade-bin\server-linux-amd64
echo   dist\upgrade-bin\server-linux-arm64
echo   dist\upgrade-bin\server-darwin-amd64
echo   dist\upgrade-bin\server-darwin-arm64
echo.
echo 提示: 将 upgrade-bin 目录下的文件上传到 Gitee Release 即可被自动升级系统识别
echo.
