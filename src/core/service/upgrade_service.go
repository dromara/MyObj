package service

import (
	"encoding/json"
	"fmt"
	"io"
	"myobj/src/pkg/logger"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	giteeAPIURL  = "https://gitee.com/api/v5/repos/MR-wind/my-obj/releases/latest"
	versionFile  = "VERSION"
	backupSuffix = ".bak"
)

// UpgradeService 自动升级服务
type UpgradeService struct {
	execPath string // 当前可执行文件路径
}

// NewUpgradeService 创建升级服务
func NewUpgradeService() *UpgradeService {
	execPath, err := os.Executable()
	if err != nil {
		execPath = os.Args[0]
	}
	return &UpgradeService{execPath: execPath}
}

// GiteeRelease Gitee Release 结构
type GiteeRelease struct {
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Body    string `json:"body"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
		Size               int64  `json:"size"`
	} `json:"assets"`
}

// UpdateInfo 版本更新信息
type UpdateInfo struct {
	HasUpdate   bool   `json:"has_update"`
	CurrentVer  string `json:"current_version"`
	LatestVer   string `json:"latest_version"`
	ReleaseNote string `json:"release_note"`
	DownloadURL string `json:"download_url"`
	FileSize    int64  `json:"file_size"`
}

// GetCurrentVersion 获取当前版本号
func (s *UpgradeService) GetCurrentVersion() string {
	// 尝试从可执行文件同目录的 VERSION 文件读取
	dir := filepath.Dir(s.execPath)
	versionPath := filepath.Join(dir, versionFile)

	data, err := os.ReadFile(versionPath)
	if err != nil {
		// 尝试从工作目录读取
		data, err = os.ReadFile(versionFile)
		if err != nil {
			return "unknown"
		}
	}
	return strings.TrimSpace(string(data))
}

// CheckUpdate 检查是否有新版本
func (s *UpgradeService) CheckUpdate() (*UpdateInfo, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(giteeAPIURL)
	if err != nil {
		return nil, fmt.Errorf("获取版本信息失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Gitee API 返回状态码: %d", resp.StatusCode)
	}

	var release GiteeRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("解析版本信息失败: %w", err)
	}

	currentVer := s.GetCurrentVersion()
	latestVer := strings.TrimSpace(release.TagName)
	if strings.HasPrefix(latestVer, "v") {
		latestVer = latestVer[1:]
	}

	// 查找当前平台的下载文件
	platform := runtime.GOOS // windows, linux, darwin
	arch := runtime.GOARCH   // amd64, arm64
	assetName := s.getAssetName(platform, arch)

	var downloadURL string
	var fileSize int64
	for _, asset := range release.Assets {
		if strings.EqualFold(asset.Name, assetName) {
			downloadURL = asset.BrowserDownloadURL
			fileSize = asset.Size
			break
		}
	}

	hasUpdate := currentVer != latestVer && downloadURL != ""

	return &UpdateInfo{
		HasUpdate:   hasUpdate,
		CurrentVer:  currentVer,
		LatestVer:   latestVer,
		ReleaseNote: release.Body,
		DownloadURL: downloadURL,
		FileSize:    fileSize,
	}, nil
}

// getAssetName 根据平台和架构获取发布文件名
func (s *UpgradeService) getAssetName(platform, arch string) string {
	ext := ""
	if platform == "windows" {
		ext = ".exe"
	}
	return fmt.Sprintf("server-%s-%s%s", platform, arch, ext)
}

// PerformUpgrade 执行升级
func (s *UpgradeService) PerformUpgrade(downloadURL string) error {
	dir := filepath.Dir(s.execPath)

	// 1. 下载新版本到临时文件
	tmpFile := filepath.Join(dir, "server_new"+getExt())
	logger.LOG.Info("开始下载新版本", "url", downloadURL, "target", tmpFile)

	if err := downloadFile(downloadURL, tmpFile); err != nil {
		return fmt.Errorf("下载新版本失败: %w", err)
	}
	logger.LOG.Info("新版本下载完成", "path", tmpFile)

	// 2. 备份当前版本
	backupPath := s.execPath + backupSuffix
	if err := copyFile(s.execPath, backupPath); err != nil {
		os.Remove(tmpFile)
		return fmt.Errorf("备份当前版本失败: %w", err)
	}
	logger.LOG.Info("当前版本已备份", "backup", backupPath)

	// 3. 生成升级脚本并执行
	if err := s.createAndRunUpgradeScript(tmpFile); err != nil {
		// 回滚：恢复备份
		copyFile(backupPath, s.execPath)
		os.Remove(tmpFile)
		return fmt.Errorf("执行升级脚本失败: %w", err)
	}

	return nil
}

// createAndRunUpgradeScript 创建升级脚本并执行
func (s *UpgradeService) createAndRunUpgradeScript(newBinary string) error {
	dir := filepath.Dir(s.execPath)
	scriptPath := filepath.Join(dir, getUpgradeScriptName())
	currentBinary := s.execPath
	backupPath := s.execPath + backupSuffix

	var script string
	if runtime.GOOS == "windows" {
		script = fmt.Sprintf(`@echo off
timeout /t 2 /nobreak >nul
taskkill /f /im "%s" >nul 2>&1
timeout /t 1 /nobreak >nul
copy /y "%s" "%s" >nul
del "%s" >nul
del "%s" >nul
start "" "%s"
del "%%~f0"
`, filepath.Base(currentBinary), newBinary, currentBinary, newBinary, backupPath, currentBinary)
	} else {
		script = fmt.Sprintf(`#!/bin/bash
sleep 2
killall "%s" 2>/dev/null
sleep 1
cp -f "%s" "%s"
rm -f "%s"
rm -f "%s"
chmod +x "%s"
nohup "%s" > /dev/null 2>&1 &
rm -f "$0"
`, filepath.Base(currentBinary), newBinary, currentBinary, newBinary, backupPath, currentBinary, currentBinary)
	}

	if err := os.WriteFile(scriptPath, []byte(script), 0755); err != nil {
		return fmt.Errorf("创建升级脚本失败: %w", err)
	}

	// 启动升级脚本
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "start", "/b", scriptPath)
	} else {
		cmd = exec.Command("bash", scriptPath)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("启动升级脚本失败: %w", err)
	}

	logger.LOG.Info("升级脚本已启动，服务即将重启")
	return nil
}

// getExt 获取当前平台的可执行文件扩展名
func getExt() string {
	if runtime.GOOS == "windows" {
		return ".exe"
	}
	return ""
}

// getUpgradeScriptName 获取升级脚本名称
func getUpgradeScriptName() string {
	if runtime.GOOS == "windows" {
		return "upgrade.bat"
	}
	return "upgrade.sh"
}

// downloadFile 下载文件
func downloadFile(url, dest string) error {
	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
