package util

import (
	"fmt"
	"myobj/src/pkg/logger"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/disk"
)

// DiskByte 常量定义
const (
	DiskByte = 1024 * 1024 * 1024
)

// DiskInfo 磁盘信息结构体
type DiskInfo struct {
	// 挂载点
	Mount string `json:"mount"`
	// 总大小 (字节)
	Total uint64 `json:"total"`
	// 已使用 (字节)
	Used uint64 `json:"used"`
	// 剩余 (字节)
	Free uint64 `json:"free"`
	// 可用 (字节)
	Avail uint64 `json:"avail"`
}

// GetDiskInfo 获取所有磁盘信息
func GetDiskInfo() ([]DiskInfo, error) {
	partitions, err := disk.Partitions(true)
	if err != nil {
		logger.LOG.Error("获取磁盘分区信息失败")
		return nil, fmt.Errorf("获取磁盘分区信息失败: %w", err)
	}
	var diskInfos []DiskInfo

	for _, partition := range partitions {
		// 跳过特殊文件系统（在某些系统上）
		if shouldSkipMount(partition.Mountpoint) {
			continue
		}

		usage, err := disk.Usage(partition.Mountpoint)
		if err != nil {
			logger.LOG.Error("获取磁盘信息失败", "mountpoint", partition.Mountpoint, "error", err)
			continue
		}

		info := DiskInfo{
			Mount: partition.Mountpoint,
			Total: usage.Total,
			Used:  usage.Used,
			Free:  usage.Free,
			Avail: usage.Free, // gopsutil中Free就是可用空间
		}

		diskInfos = append(diskInfos, info)
	}

	return diskInfos, nil
}

// DiskInfoPath 获取指定挂载点的磁盘信息
func DiskInfoPath(mount string) (*DiskInfo, error) {
	diskInfos, err := GetDiskInfo()
	if err != nil {
		return nil, err
	}

	for _, info := range diskInfos {
		if info.Mount == mount {
			return &info, nil
		}
	}

	return nil, fmt.Errorf("未找到指定路径的磁盘信息: %s", mount)
}

// GetCurrentDirectoryDiskSpace 获取当前软件所在磁盘的空间信息
func GetCurrentDirectoryDiskSpace() (*DiskInfo, error) {
	// 获取当前可执行文件路径
	execPath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("获取可执行文件路径失败: %w", err)
	}

	// 转换为绝对路径
	absPath, err := filepath.Abs(execPath)
	if err != nil {
		return nil, fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 找到该路径所在的挂载点
	diskInfos, err := GetDiskInfo()
	if err != nil {
		return nil, err
	}

	// 查找最匹配的挂载点
	mountPoint := findMountForPath(absPath, diskInfos)
	if mountPoint == "" {
		return nil, fmt.Errorf("未找到当前目录所在磁盘信息: %s", absPath)
	}

	// 获取该挂载点的磁盘信息
	return DiskInfoPath(mountPoint)
}

// findMountForPath 查找包含给定路径的挂载点
func findMountForPath(path string, diskInfos []DiskInfo) string {
	var bestMatch string

	for _, info := range diskInfos {
		mount := info.Mount

		// 标准化路径分隔符（Windows兼容）
		normalizedPath := filepath.Clean(path)

		if strings.HasPrefix(normalizedPath, mount) {
			bestMatch = mount
		}
	}

	return bestMatch
}

// shouldSkipMount 判断是否应该跳过某个挂载点
func shouldSkipMount(mountpoint string) bool {
	// 跳过虚拟文件系统和特殊挂载点
	skipMounts := []string{
		"/proc", "/sys", "/dev", "/run", "/snap",
		"/sys/fs/cgroup", "/sys/kernel/security",
	}

	for _, skip := range skipMounts {
		if strings.HasPrefix(mountpoint, skip) {
			return true
		}
	}

	// Windows下跳过没有盘符的路径
	if runtime.GOOS == "windows" && len(mountpoint) < 2 {
		return true
	}

	return false
}

// FormatBytes 格式化字节大小为可读字符串
func FormatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := uint64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
