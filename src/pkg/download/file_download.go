package download

import (
	"context"
	"fmt"
	"io"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// LocalFileDownloadOptions 本地文件下载配置
type LocalFileDownloadOptions struct {
	FilePassword string // 文件解密密码（加密文件必需）
}

// LocalFileDownloadResult 本地文件下载结果
type LocalFileDownloadResult struct {
	TempFilePath string // 临时文件路径（已解密、已合并）
	FileName     string // 文件名
	FileSize     int64  // 文件大小
	ContentType  string // MIME类型
	IsEncrypted  bool   // 是否加密
	IsChunked    bool   // 是否分片
	Error        string // 错误信息
}

// PrepareLocalFileDownload 准备本地文件下载（解密+合并）
// 返回临时文件路径，调用者负责清理
func PrepareLocalFileDownload(
	ctx context.Context,
	fileID string,
	userID string,
	tempDir string,
	repoFactory *impl.RepositoryFactory,
	opts *LocalFileDownloadOptions,
) (*LocalFileDownloadResult, error) {
	result := &LocalFileDownloadResult{}

	// 1. 查询文件信息
	fileInfo, err := repoFactory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	result.FileName = fileInfo.Name
	result.FileSize = int64(fileInfo.Size)
	result.ContentType = fileInfo.Mime
	result.IsEncrypted = fileInfo.IsEnc
	result.IsChunked = fileInfo.IsChunk

	// 2. 验证文件权限（公开文件或用户自己的文件）
	if err := validateFilePermission(ctx, fileID, userID, repoFactory); err != nil {
		return nil, err
	}

	// 3. 判断是否需要临时文件（加密或分片）
	needTempFile := fileInfo.IsEnc || fileInfo.IsChunk

	var tempFilePath string
	var sessionTempDir string

	if !needTempFile {
		// 不需要临时文件，直接返回data路径
		result.TempFilePath = fileInfo.Path
		logger.LOG.Info("文件无需处理，直接使用data路径", "fileID", fileID, "path", fileInfo.Path)
		return result, nil
	}

	// 4. 需要临时文件，在文件所在磁盘的temp目录下创建
	diskPath := extractDiskPathFromFilePath(fileInfo.Path)
	if diskPath == "" {
		return nil, fmt.Errorf("无法提取磁盘路径: %s", fileInfo.Path)
	}

	// 创建临时目录：{磁盘路径}/temp/download_{sessionID}/
	sessionID := uuid.New().String()[:8]
	sessionTempDir = filepath.Join(diskPath, "temp", fmt.Sprintf("download_%s", sessionID))
	if err := os.MkdirAll(sessionTempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	tempFilePath = filepath.Join(sessionTempDir, fileInfo.Name)

	// 4. 处理加密文件
	if fileInfo.IsEnc {
		if opts == nil || opts.FilePassword == "" {
			// 清理临时目录
			os.RemoveAll(sessionTempDir)
			return nil, fmt.Errorf("加密文件需要提供解密密码")
		}

		// 验证密码
		user, err := repoFactory.User().GetByID(ctx, userID)
		if err != nil {
			os.RemoveAll(sessionTempDir)
			return nil, fmt.Errorf("查询用户信息失败: %w", err)
		}

		// 验证密码（使用bcrypt比对哈希值）
		if !util.CheckPassword(user.FilePassword, opts.FilePassword) {
			os.RemoveAll(sessionTempDir)
			return nil, fmt.Errorf("密码错误")
		}

		// 获取加密文件路径：优先使用EncPath，如果为空则使用Path
		// 根据设计，加密文件存储为.data文件，Path和EncPath应该都指向同一个文件
		encryptedPath := fileInfo.EncPath
		if encryptedPath == "" {
			encryptedPath = fileInfo.Path
		}

		logger.LOG.Info("开始解密文件",
			"fileID", fileID,
			"encPath", encryptedPath,
			"fileInfo.EncPath", fileInfo.EncPath,
			"fileInfo.Path", fileInfo.Path,
			"fileInfo.Name", fileInfo.Name)

		// 检查加密文件是否存在
		if _, err := os.Stat(encryptedPath); os.IsNotExist(err) {
			os.RemoveAll(sessionTempDir)
			return nil, fmt.Errorf("加密文件不存在: %s", encryptedPath)
		}

		// 使用PBKDF2从明文密码和用户ID派生加密密钥
		// 这与上传时的逻辑完全一致
		encryptionKey := util.DeriveEncryptionKey(opts.FilePassword, userID)
		logger.LOG.Debug("派生解密密钥", "userID", userID, "keyLength", len(encryptionKey))

		crypto := util.NewFileCrypto(encryptionKey)
		if err := crypto.DecryptFile(encryptedPath, tempFilePath); err != nil {
			os.RemoveAll(sessionTempDir)
			return nil, fmt.Errorf("文件解密失败: %w", err)
		}

		logger.LOG.Info("文件解密完成", "fileID", fileID, "tempPath", tempFilePath)
	} else if fileInfo.IsChunk {
		// 5. 处理分片文件（合并）
		logger.LOG.Info("开始合并分片文件", "fileID", fileID, "chunkCount", fileInfo.ChunkCount)

		if err := mergeChunkedFile(ctx, fileInfo, tempFilePath, repoFactory); err != nil {
			os.RemoveAll(sessionTempDir)
			return nil, fmt.Errorf("合并分片文件失败: %w", err)
		}

		logger.LOG.Info("分片文件合并完成", "fileID", fileID, "tempPath", tempFilePath)
	}

	result.TempFilePath = tempFilePath
	return result, nil
}

// validateFilePermission 验证文件下载权限
// userID 可以为空（未登录用户），此时只允许访问公开文件
func validateFilePermission(ctx context.Context, fileID string, userID string, repoFactory *impl.RepositoryFactory) error {
	// 如果用户已登录，先检查是否是用户自己的文件
	if userID != "" {
		userFile, err := repoFactory.UserFiles().GetByUserIDAndFileID(ctx, userID, fileID)
		if err == nil && userFile != nil {
			// 用户自己的文件，允许下载
			return nil
		}
	}

	// 检查是否为公开文件（查询所有公开文件）
	allPublicFiles, err := repoFactory.UserFiles().ListPublicFiles(ctx, 0, 1000)
	if err != nil {
		return fmt.Errorf("查询文件失败: %w", err)
	}

	for _, uf := range allPublicFiles {
		if uf.FileID == fileID && uf.IsPublic {
			// 公开文件，允许下载
			logger.LOG.Info("下载公开文件", "fileID", fileID, "ownerID", uf.UserID, "downloaderID", userID)
			return nil
		}
	}

	// 既不是自己的文件，也不是公开文件
	return fmt.Errorf("无权限下载此文件")
}

// mergeChunkedFile 合并分片文件
func mergeChunkedFile(ctx context.Context, fileInfo *models.FileInfo, outputPath string, repoFactory *impl.RepositoryFactory) error {
	// 1. 查询所有分片
	chunks, err := repoFactory.FileChunk().GetByFileID(ctx, fileInfo.ID)
	if err != nil {
		return fmt.Errorf("查询分片信息失败: %w", err)
	}

	if len(chunks) == 0 {
		return fmt.Errorf("未找到分片文件")
	}

	// 2. 按分片索引排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	})

	// 3. 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outFile.Close()

	// 4. 逐个读取分片并写入
	for _, chunk := range chunks {
		chunkFile, err := os.Open(chunk.ChunkPath)
		if err != nil {
			return fmt.Errorf("打开分片文件失败 [索引=%d]: %w", chunk.ChunkIndex, err)
		}

		if _, err := io.Copy(outFile, chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("复制分片数据失败 [索引=%d]: %w", chunk.ChunkIndex, err)
		}

		chunkFile.Close()
		logger.LOG.Debug("分片合并进度", "chunk", chunk.ChunkIndex+1, "total", len(chunks))
	}

	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	return err
}

// extractDiskPathFromFilePath 从文件路径中提取磁盘路径
// 文件路径格式: {DiskPath}/data/{原文件名不带后缀}/{文件}
// 返回: {DiskPath}
func extractDiskPathFromFilePath(filePath string) string {
	// 规范化路径
	filePath = filepath.Clean(filePath)

	// 查找 "/data/" 或 "\\data\\" 的位置
	dataIndex := strings.Index(strings.ToLower(filePath), string(filepath.Separator)+"data"+string(filepath.Separator))
	if dataIndex == -1 {
		return ""
	}

	// 返回data之前的路径
	return filePath[:dataIndex]
}

// IsTempPath 判断路径是否为临时路径（导出函数）
// 临时路径格式: {DiskPath}/temp/...
func IsTempPath(path string) bool {
	path = filepath.Clean(path)
	return strings.Contains(strings.ToLower(path), string(filepath.Separator)+"temp"+string(filepath.Separator))
}

// ServeFileWithRange 支持HTTP Range请求的文件服务
type FileRangeInfo struct {
	Start      int64
	End        int64
	TotalSize  int64
	IsRanged   bool
	StatusCode int
}

// ParseRangeHeader 解析HTTP Range头
func ParseRangeHeader(rangeHeader string, fileSize int64) (*FileRangeInfo, error) {
	info := &FileRangeInfo{
		Start:      0,
		End:        fileSize - 1,
		TotalSize:  fileSize,
		IsRanged:   false,
		StatusCode: 200,
	}

	if rangeHeader == "" {
		return info, nil
	}

	// 解析 Range: bytes=start-end
	if !strings.HasPrefix(rangeHeader, "bytes=") {
		return nil, fmt.Errorf("不支持的Range格式")
	}

	rangeStr := strings.TrimPrefix(rangeHeader, "bytes=")
	parts := strings.Split(rangeStr, "-")
	if len(parts) != 2 {
		return nil, fmt.Errorf("无效的Range格式")
	}

	// 解析起始位置
	if parts[0] != "" {
		start, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的起始位置: %w", err)
		}
		info.Start = start
	}

	// 解析结束位置
	if parts[1] != "" {
		end, err := strconv.ParseInt(parts[1], 10, 64)
		if err != nil {
			return nil, fmt.Errorf("无效的结束位置: %w", err)
		}
		info.End = end
	}

	// 验证范围
	if info.Start < 0 || info.End >= fileSize || info.Start > info.End {
		return nil, fmt.Errorf("Range超出文件范围")
	}

	info.IsRanged = true
	info.StatusCode = 206 // Partial Content
	return info, nil
}
