package util

import (
	"myobj/src/pkg/logger"
	"os"
	"path/filepath"
)

// DeletePhysicalFile 删除物理文件（包含缩略图）
func DeletePhysicalFile(filePath string) error {
	if filePath == "" {
		return nil
	}
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		logger.LOG.Warn("删除物理文件失败", "path", filePath, "error", err)
		return err
	}
	// 删除缩略图
	DeleteThumbnail(filePath)
	return nil
}

// DeleteThumbnail 根据文件路径删除对应的缩略图
func DeleteThumbnail(filePath string) {
	if filePath == "" {
		return
	}
	thumbnailPath := filePath + ".thumb"
	if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
		logger.LOG.Debug("删除缩略图失败", "path", thumbnailPath, "error", err)
	}
}

// DeleteFile 删除文件（物理文件 + 缩略图）
func DeleteFile(filePath string) error {
	return DeletePhysicalFile(filePath)
}

// DeleteDirectory 删除目录（递归删除所有内容）
func DeleteDirectory(dirPath string) error {
	if dirPath == "" {
		return nil
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		fullPath := filepath.Join(dirPath, entry.Name())
		if entry.IsDir() {
			if err := DeleteDirectory(fullPath); err != nil {
				logger.LOG.Warn("删除子目录失败", "path", fullPath, "error", err)
			}
		} else {
			if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
				logger.LOG.Warn("删除文件失败", "path", fullPath, "error", err)
			}
		}
	}
	return os.Remove(dirPath)
}

// DeleteDirectoryIfEmpty 如果目录为空则删除
func DeleteDirectoryIfEmpty(dirPath string) error {
	if dirPath == "" {
		return nil
	}
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return err
	}
	if len(entries) == 0 {
		return os.Remove(dirPath)
	}
	return nil
}
