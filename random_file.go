package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func main() {
	// 提示用户输入文件大小（GB）
	fmt.Print("请输入要生成的文件大小（GB）: ")
	var sizeGBStr string
	_, err := fmt.Scanln(&sizeGBStr)
	if err != nil {
		fmt.Printf("读取输入失败: %v\n", err)
		return
	}

	// 转换为浮点数
	sizeGB, err := strconv.ParseFloat(sizeGBStr, 64)
	if err != nil {
		fmt.Printf("无效的数字格式: %v\n", err)
		return
	}

	// 验证大小
	if sizeGB <= 0 {
		fmt.Println("文件大小必须大于0")
		return
	}

	if sizeGB > 1024 {
		fmt.Print("警告：文件大小超过1TB，确定继续吗？(y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("已取消操作")
			return
		}
	}

	// 计算字节数
	sizeBytes := int64(sizeGB * 1024 * 1024 * 1024)

	// 生成随机文件名
	fileName := generateRandomFileName()
	filePath := filepath.Join(".", fileName)

	fmt.Printf("正在生成文件: %s\n", fileName)
	fmt.Printf("文件大小: %.2f GB (约 %s)\n", sizeGB, formatBytes(sizeBytes))
	fmt.Println("开始生成...")

	// 创建文件
	startTime := time.Now()
	err = createLargeFile(filePath, sizeBytes)
	if err != nil {
		fmt.Printf("生成文件失败: %v\n", err)
		// 清理可能已创建的部分文件
		os.Remove(filePath)
		return
	}

	elapsedTime := time.Since(startTime)

	// 获取文件实际大小
	fileInfo, _ := os.Stat(filePath)
	actualSize := fileInfo.Size()

	fmt.Println("\n=== 生成完成 ===")
	fmt.Printf("文件名: %s\n", fileName)
	fmt.Printf("文件路径: %s\n", filePath)
	fmt.Printf("实际大小: %s\n", formatBytes(actualSize))
	fmt.Printf("耗时: %v\n", elapsedTime)
	fmt.Printf("平均速度: %.2f MB/s\n",
		float64(actualSize)/(1024*1024)/elapsedTime.Seconds())
}

// 生成随机文件名（16字节十六进制）
func generateRandomFileName() string {
	bytes := make([]byte, 8)
	rand.Read(bytes)

	// 使用当前时间戳和随机数组合
	timestamp := time.Now().UnixNano()
	return fmt.Sprintf("testfile_%d_%x", timestamp, bytes)
}

// 创建大文件
func createLargeFile(filePath string, size int64) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用缓冲区提高写入性能
	const bufferSize = 64 * 1024 // 64KB
	buffer := make([]byte, bufferSize)

	// 填充随机数据到缓冲区
	_, err = rand.Read(buffer)
	if err != nil {
		return err
	}

	// 计算需要写入的次数
	remaining := size
	bytesWritten := int64(0)

	// 显示进度
	progressTicker := time.NewTicker(time.Second)
	defer progressTicker.Stop()

	go func() {
		for range progressTicker.C {
			progress := float64(bytesWritten) / float64(size) * 100
			fmt.Printf("\r进度: %.1f%% (已写入: %s)",
				progress, formatBytes(bytesWritten))
		}
	}()

	// 写入文件
	for remaining > 0 {
		writeSize := bufferSize
		if remaining < bufferSize {
			writeSize = int(remaining)
		}

		n, err := file.Write(buffer[:writeSize])
		if err != nil {
			return err
		}

		bytesWritten += int64(n)
		remaining -= int64(n)
	}

	fmt.Printf("\r进度: 100%% (已写入: %s)\n", formatBytes(bytesWritten))
	return nil
}

// 格式化字节大小为易读格式
func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB",
		float64(bytes)/float64(div), "KMGTPE"[exp])
}
