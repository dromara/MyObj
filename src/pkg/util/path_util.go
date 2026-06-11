package util

import (
	"fmt"
	"path/filepath"

	"github.com/google/uuid"
)

// BuildTempDir 构建临时目录路径
func BuildTempDir(diskPath, prefix string) string {
	id := uuid.New().String()[:8]
	return filepath.Join(diskPath, "temp", fmt.Sprintf("%s_%s", prefix, id))
}

// ChunkPath 构建分片文件路径
func ChunkPath(dir string, index int) string {
	return filepath.Join(dir, fmt.Sprintf("%d.chunk.data", index))
}
