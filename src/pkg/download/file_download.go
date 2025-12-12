package download

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/hash"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// PrepareFileForDownload 准备文件用于下载
//
// 处理流程:
//  1. 检查文件是否存在（数据库记录和磁盘文件）
//  2. 如果是分片文件，检查所有分片是否完整
//  3. 根据文件状态进行处理:
//     - 分片+加密: 合并分片 → 解密 → 校验原始hash → 返回临时文件路径
//     - 仅分片:   合并分片 → 校验原始hash → 返回临时文件路径
//     - 仅加密:   解密 → 校验原始hash → 返回临时文件路径
//     - 普通文件: 校验原始hash → 直接返回原始路径（不复制）
//  4. 通过.info文件和数据库进行双重hash校验
//
// 参数:
//   - fileID: 文件ID
//   - tempDir: 临时目录路径（用于存放处理后的文件）
//   - repoFactory: 数据库仓储工厂
//
// 返回:
//   - filePath: 可以直接读取的文件路径
//   - fileInfo: 文件信息（供调用方使用）
//   - err: 错误信息
//
// 注意:
//   - 临时目录中生成的文件由调用方负责清理
//   - 如果文件未加密未分片，返回的是原始文件路径，不会复制到临时目录
func PrepareFileForDownload(fileID string, tempDir string, repoFactory *impl.RepositoryFactory) (filePath string, fileInfo *models.FileInfo, err error) {
	ctx := context.Background()

	// 1. 查询数据库中的文件信息
	fileInfo, err = repoFactory.FileInfo().GetByID(ctx, fileID)
	if err != nil {
		return "", nil, fmt.Errorf("文件记录不存在: %w", err)
	}

	// 2. 检查主文件是否存在
	if _, err := os.Stat(fileInfo.Path); os.IsNotExist(err) {
		return "", nil, fmt.Errorf("文件不存在: %s", fileInfo.Path)
	}

	// 3. 根据文件类型选择处理策略
	var processedPath string

	if fileInfo.IsChunk && fileInfo.IsEnc {
		// 场景1: 分片+加密 → 合并分片 → 解密 → 校验
		processedPath, err = handleChunkedEncryptedFile(ctx, fileInfo, tempDir, repoFactory)
		if err != nil {
			return "", nil, err
		}
	} else if fileInfo.IsChunk {
		// 场景2: 仅分片 → 合并分片 → 校验
		processedPath, err = handleChunkedFile(ctx, fileInfo, tempDir, repoFactory)
		if err != nil {
			return "", nil, err
		}
	} else if fileInfo.IsEnc {
		// 场景3: 仅加密 → 解密 → 校验
		processedPath, err = handleEncryptedFile(ctx, fileInfo, tempDir, repoFactory)
		if err != nil {
			return "", nil, err
		}
	} else {
		// 场景4: 普通文件 → 直接校验 → 返回原始路径
		if err := verifyFileHash(fileInfo.Path, fileInfo.FileHash, "原始文件"); err != nil {
			return "", nil, err
		}
		// 二次校验：通过.info文件
		if err := verifyInfoFile(fileInfo.Path, fileInfo.FileHash, fileInfo.FileEncHash); err != nil {
			logger.LOG.Warn(".info文件校验失败", "error", err)
			// .info文件校验失败不影响主流程（可能文件是老数据）
		}
		processedPath = fileInfo.Path // 不复制，直接返回原始路径
	}

	logger.LOG.Info("文件准备完成", "fileID", fileID, "path", processedPath)
	return processedPath, fileInfo, nil
}

// handleChunkedEncryptedFile 处理分片+加密的文件
// 流程: 合并分片 → 解密 → 校验原始hash
func handleChunkedEncryptedFile(ctx context.Context, fileInfo *models.FileInfo, tempDir string, repoFactory *impl.RepositoryFactory) (string, error) {
	// 1. 查询所有分片信息
	chunks, err := repoFactory.FileChunk().GetByFileID(ctx, fileInfo.ID)
	if err != nil {
		return "", fmt.Errorf("查询分片信息失败: %w", err)
	}

	if len(chunks) == 0 {
		return "", fmt.Errorf("分片信息为空")
	}

	if len(chunks) != fileInfo.ChunkCount {
		return "", fmt.Errorf("分片数量不匹配: 期望%d个，实际%d个", fileInfo.ChunkCount, len(chunks))
	}

	// 2. 检查所有分片文件是否存在
	for _, chunk := range chunks {
		if _, err := os.Stat(chunk.ChunkPath); os.IsNotExist(err) {
			return "", fmt.Errorf("分片文件不存在: %s", chunk.ChunkPath)
		}
	}

	// 3. 合并分片到临时文件（加密状态）
	mergedEncPath := filepath.Join(tempDir, fmt.Sprintf("%s_merged.enc", fileInfo.ID))
	if err := mergeChunkFiles(chunks, mergedEncPath); err != nil {
		return "", fmt.Errorf("合并分片失败: %w", err)
	}
	defer os.Remove(mergedEncPath) // 清理中间文件

	// 4. 校验加密文件的hash（如果有记录）
	if fileInfo.FileEncHash != "" {
		if err := verifyFileHash(mergedEncPath, fileInfo.FileEncHash, "加密文件"); err != nil {
			return "", err
		}
	}

	// 5. 解密文件
	user, err := getUserByFileID(ctx, fileInfo.ID, repoFactory)
	if err != nil {
		return "", err
	}

	decryptedPath := filepath.Join(tempDir, fmt.Sprintf("%s_decrypted", fileInfo.ID))
	crypto := util.NewFileCrypto(user.FilePassword)
	if err := crypto.DecryptFile(mergedEncPath, decryptedPath); err != nil {
		return "", fmt.Errorf("文件解密失败: %w", err)
	}

	// 6. 校验解密后的原始文件hash
	if err := verifyFileHash(decryptedPath, fileInfo.FileHash, "解密后文件"); err != nil {
		return "", err
	}

	logger.LOG.Debug("分片+加密文件处理完成", "path", decryptedPath)
	return decryptedPath, nil
}

// handleChunkedFile 处理仅分片的文件
// 流程: 合并分片 → 校验原始hash
func handleChunkedFile(ctx context.Context, fileInfo *models.FileInfo, tempDir string, repoFactory *impl.RepositoryFactory) (string, error) {
	// 1. 查询所有分片信息
	chunks, err := repoFactory.FileChunk().GetByFileID(ctx, fileInfo.ID)
	if err != nil {
		return "", fmt.Errorf("查询分片信息失败: %w", err)
	}

	if len(chunks) == 0 {
		return "", fmt.Errorf("分片信息为空")
	}

	if len(chunks) != fileInfo.ChunkCount {
		return "", fmt.Errorf("分片数量不匹配: 期望%d个，实际%d个", fileInfo.ChunkCount, len(chunks))
	}

	// 2. 检查所有分片文件是否存在
	for _, chunk := range chunks {
		if _, err := os.Stat(chunk.ChunkPath); os.IsNotExist(err) {
			return "", fmt.Errorf("分片文件不存在: %s", chunk.ChunkPath)
		}
	}

	// 3. 合并分片到临时文件
	mergedPath := filepath.Join(tempDir, fmt.Sprintf("%s_merged", fileInfo.ID))
	if err := mergeChunkFiles(chunks, mergedPath); err != nil {
		return "", fmt.Errorf("合并分片失败: %w", err)
	}

	// 4. 校验合并后的文件hash
	if err := verifyFileHash(mergedPath, fileInfo.FileHash, "合并后文件"); err != nil {
		return "", err
	}

	// 5. 二次校验：通过.info文件（使用主文件路径）
	if err := verifyInfoFile(fileInfo.Path, fileInfo.FileHash, fileInfo.FileEncHash); err != nil {
		logger.LOG.Warn(".info文件校验失败", "error", err)
	}

	logger.LOG.Debug("分片文件处理完成", "path", mergedPath)
	return mergedPath, nil
}

// handleEncryptedFile 处理仅加密的文件
// 流程: 解密 → 校验原始hash
func handleEncryptedFile(ctx context.Context, fileInfo *models.FileInfo, tempDir string, repoFactory *impl.RepositoryFactory) (string, error) {
	// 1. 校验加密文件的hash（如果有记录）
	if fileInfo.FileEncHash != "" {
		if err := verifyFileHash(fileInfo.EncPath, fileInfo.FileEncHash, "加密文件"); err != nil {
			return "", err
		}
	}

	// 2. 查询用户加密密码
	user, err := getUserByFileID(ctx, fileInfo.ID, repoFactory)
	if err != nil {
		return "", err
	}

	// 3. 解密文件到临时目录
	decryptedPath := filepath.Join(tempDir, fmt.Sprintf("%s_decrypted", fileInfo.ID))
	crypto := util.NewFileCrypto(user.FilePassword)
	if err := crypto.DecryptFile(fileInfo.EncPath, decryptedPath); err != nil {
		return "", fmt.Errorf("文件解密失败: %w", err)
	}

	// 4. 校验解密后的原始文件hash
	if err := verifyFileHash(decryptedPath, fileInfo.FileHash, "解密后文件"); err != nil {
		return "", err
	}

	// 5. 二次校验：通过.info文件
	if err := verifyInfoFile(fileInfo.Path, fileInfo.FileHash, fileInfo.FileEncHash); err != nil {
		logger.LOG.Warn(".info文件校验失败", "error", err)
	}

	logger.LOG.Debug("加密文件处理完成", "path", decryptedPath)
	return decryptedPath, nil
}

// mergeChunkFiles 合并分片文件
func mergeChunkFiles(chunks []*models.FileChunk, outputPath string) error {
	// 按分片索引排序
	sort.Slice(chunks, func(i, j int) bool {
		return chunks[i].ChunkIndex < chunks[j].ChunkIndex
	})

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer outputFile.Close()

	// 按顺序合并分片
	for _, chunk := range chunks {
		chunkFile, err := os.Open(chunk.ChunkPath)
		if err != nil {
			return fmt.Errorf("打开分片文件失败 [%d]: %w", chunk.ChunkIndex, err)
		}

		if _, err := io.Copy(outputFile, chunkFile); err != nil {
			chunkFile.Close()
			return fmt.Errorf("复制分片数据失败 [%d]: %w", chunk.ChunkIndex, err)
		}
		chunkFile.Close()
	}

	return nil
}

// verifyFileHash 校验文件hash值
func verifyFileHash(filePath, expectedHash, fileDesc string) error {
	hasher := hash.NewFastBlake3Hasher()
	actualHash, _, err := hasher.ComputeFileHash(filePath)
	if err != nil {
		return fmt.Errorf("计算%shash失败: %w", fileDesc, err)
	}

	if actualHash != expectedHash {
		return fmt.Errorf("%shash校验失败: 期望=%s, 实际=%s", fileDesc, expectedHash, actualHash)
	}

	logger.LOG.Debug(fmt.Sprintf("%shash校验通过", fileDesc), "hash", actualHash)
	return nil
}

// verifyInfoFile 通过.info文件进行二次hash校验
func verifyInfoFile(dataFilePath, expectedFileHash, expectedFileEncHash string) error {
	// 生成.info文件路径
	infoFilePath := strings.TrimSuffix(dataFilePath, ".data") + ".info"

	// 读取.info文件
	infoData, err := os.ReadFile(infoFilePath)
	if err != nil {
		return fmt.Errorf("读取.info文件失败: %w", err)
	}

	// 解析JSON
	var hashInfo struct {
		FileHash    string `json:"file_hash"`
		FileEncHash string `json:"file_enc_hash"`
	}
	if err := json.Unmarshal(infoData, &hashInfo); err != nil {
		return fmt.Errorf("解析.info文件失败: %w", err)
	}

	// 校验原始文件hash
	if hashInfo.FileHash != expectedFileHash {
		return fmt.Errorf(".info文件中file_hash不匹配: info=%s, db=%s", hashInfo.FileHash, expectedFileHash)
	}

	// 校验加密文件hash（如果有）
	if expectedFileEncHash != "" && hashInfo.FileEncHash != expectedFileEncHash {
		return fmt.Errorf(".info文件中file_enc_hash不匹配: info=%s, db=%s", hashInfo.FileEncHash, expectedFileEncHash)
	}

	logger.LOG.Debug(".info文件校验通过", "path", infoFilePath)
	return nil
}

// getUserByFileID 通过文件ID查询文件所属用户（用于获取解密密码）
func getUserByFileID(ctx context.Context, fileID string, repoFactory *impl.RepositoryFactory) (*models.UserInfo, error) {
	// 直接查询user_files表
	var userFile models.UserFiles
	if err := repoFactory.DB().Where("file_id = ?", fileID).First(&userFile).Error; err != nil {
		return nil, fmt.Errorf("查询文件所属用户失败: %w", err)
	}

	// 查询用户信息
	user, err := repoFactory.User().GetByID(ctx, userFile.UserID)
	if err != nil {
		return nil, fmt.Errorf("查询用户信息失败: %w", err)
	}

	if user.FilePassword == "" {
		return nil, fmt.Errorf("用户未设置文件加密密码")
	}

	return user, nil
}
