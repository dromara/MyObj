# 构建脚本 - PowerShell版本
# 用于编译服务端和CLI工具

Write-Host "开始编译项目..." -ForegroundColor Green

# 编译服务端
Write-Host "`n编译服务端..." -ForegroundColor Cyan
go build -tags "nosqlite" -o bin/server.exe ./src/cmd/server
if ($LASTEXITCODE -eq 0) {
    Write-Host "服务端编译成功！" -ForegroundColor Green
    $size = (Get-Item bin/server.exe).Length / 1MB
    Write-Host "文件大小: $([math]::Round($size, 2)) MB" -ForegroundColor Gray
} else {
    Write-Host "服务端编译失败！" -ForegroundColor Red
    exit 1
}

# 编译CLI工具
Write-Host "`n编译CLI工具..." -ForegroundColor Cyan
go build -tags "nosqlite" -o bin/cli.exe ./src/cmd/cli
if ($LASTEXITCODE -eq 0) {
    Write-Host "CLI工具编译成功！" -ForegroundColor Green
    $size = (Get-Item bin/cli.exe).Length / 1MB
    Write-Host "文件大小: $([math]::Round($size, 2)) MB" -ForegroundColor Gray
} else {
    Write-Host "CLI工具编译失败！" -ForegroundColor Red
    exit 1
}

Write-Host "`n所有编译完成！" -ForegroundColor Green
Write-Host "输出目录: bin/" -ForegroundColor Gray
