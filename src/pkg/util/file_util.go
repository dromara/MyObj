package util

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"myobj/src/pkg/logger"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

// GenerateUniqueFilename 生成基于时间戳的唯一文件名
func GenerateUniqueFilename() string {
	currentTime := time.Now().UnixNano()
	return fmt.Sprintf("%d_%s", currentTime, generateCode())
}

func generateCode() string {
	var code string
	for i := 0; i < 4; i++ {
		// 生成一个0-9之间的随机数字
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			logger.LOG.Error("Failed to generate random number", "error", err)
			return ""
		}
		code += num.String()
	}
	return code
}

// getMimeType 使用mimetype库检测文件的MIME类型（基于内容分析）
func getMimeType(filePath string) (string, error) {
	mime, err := mimetype.DetectFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to detect MIME type: %w", err)
	}
	return mime.String(), nil
}

// isImageByExtension 通过文件扩展名判断是否为图片类型
func isImageByExtension(filePath string) bool {
	ext := filepath.Ext(filePath)
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".tiff":
		return true
	default:
		return false
	}
}

// fileRechristen 重命名文件，保留在原目录
func fileRechristen(filePath string, newName string) error {
	dir := filepath.Dir(filePath)
	newPath := filepath.Join(dir, newName)

	err := os.Rename(filePath, newPath)
	if err != nil {
		return fmt.Errorf("failed to rename file: %w", err)
	}
	return nil
}
