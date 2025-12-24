package upload

import (
	"context"
	"fmt"
	"io"
	"myobj/src/config"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/hash"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/preview"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FileUploadData 文件上传参数
type FileUploadData struct {
	// 临时文件路径
	TempFilePath string `json:"temp_file_path"`
	// 文件名
	FileName string `json:"file_name"`
	// 文件大小
	FileSize int64 `json:"file_size"`
	// 文件hash签名
	ChunkSignature string `json:"chunk_signature"`
	// 文件分片hash 第一
	FirstChunkHash string `json:"first_chunk_hash"`
	// 文件分片hash 第二
	SecondChunkHash string `json:"second_chunk_hash"`
	// 文件分片hash 第三
	ThirdChunkHash string `json:"third_chunk_hash"`
	// 是否需要加密
	IsEnc bool `json:"is_enc"`
	// 是否分块上传
	IsChunk bool `json:"is_chunk"`
	// 分块数量
	ChunkCount int `json:"chunk_count"`
	// 文件虚拟路径
	VirtualPath string `json:"virtual_path"`
	// 上传用户ID
	UserID string `json:"user_id"`
	// 文件加密密码（明文）
	FilePassword string `json:"file_password"`
}

// ProcessUploadedFile 处理已上传的文件
// 参数:
//   - data: 文件上传数据
//   - repoFactory: 数据库仓储工厂
//
// 返回:
//   - fileID: 生成的文件ID
//   - err: 错误信息
func ProcessUploadedFile(data *FileUploadData, repoFactory *impl.RepositoryFactory) (fileID string, err error) {
	ctx := context.Background()

	// 调试：检查初始临时文件
	if tempInfo, err := os.Stat(data.TempFilePath); err == nil {
		logger.LOG.Debug("开始处理文件", "TempFilePath", data.TempFilePath, "临时文件大小", tempInfo.Size(), "期望大小", data.FileSize)
	} else {
		return "", fmt.Errorf("临时文件不存在: %s, %w", data.TempFilePath, err)
	}

	// 确保无论成功失败都清理临时文件
	defer func() {
		cleanupTempFiles(data)
	}()

	// 1. 合并分片（如果是分片上传）
	mergedFilePath := data.TempFilePath
	if data.IsChunk {
		mergedPath, err := mergeChunks(data)
		if err != nil {
			return "", fmt.Errorf("合并分片失败: %w", err)
		}
		mergedFilePath = mergedPath
	}

	// 2. 检测文件MIME类型
	mimeType, err := detectMimeType(mergedFilePath)
	if err != nil {
		return "", fmt.Errorf("检测文件类型失败: %w", err)
	}

	// 3. 并行计算全量hash和生成缩略图（如果需要）
	type asyncResult struct {
		fullHash      string
		thumbnailPath string
		err           error
	}
	resultChan := make(chan asyncResult, 2)
	var wg sync.WaitGroup

	// 3.1 异步计算全量hash
	wg.Add(1)
	go func() {
		defer wg.Done()
		hasher := hash.NewFastBlake3Hasher()
		fullHash, _, err := hasher.ComputeFileHash(mergedFilePath)
		resultChan <- asyncResult{fullHash: fullHash, err: err}
	}()

	// 3.2 异步生成缩略图（如果需要）
	var needThumbnail bool
	if config.CONFIG.File.Thumbnail && isImage(mimeType) {
		needThumbnail = true
		wg.Add(1)
		go func() {
			defer wg.Done()
			// 临时缩略图路径
			tempThumbnail := mergedFilePath + ".thumbnail.jpg"
			err := preview.GenerateImageThumbnail(mergedFilePath, tempThumbnail, 300)
			resultChan <- asyncResult{thumbnailPath: tempThumbnail, err: err}
		}()
	}

	// 等待异步任务完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果
	var fullHash, tempThumbnailPath string
	for result := range resultChan {
		if result.err != nil {
			return "", fmt.Errorf("异步处理失败: %w", result.err)
		}
		if result.fullHash != "" {
			fullHash = result.fullHash
		}
		if result.thumbnailPath != "" {
			tempThumbnailPath = result.thumbnailPath
		}
	}

	// 4. 选择存储磁盘（按剩余空间最大原则）
	disk, err := selectBestDisk(ctx, repoFactory, data.FileSize)
	if err != nil {
		return "", fmt.Errorf("选择存储磁盘失败: %w", err)
	}

	// 5. 生成文件ID和存储路径
	fileID = uuid.Must(uuid.NewV7()).String()
	virtualFileName := util.GenerateUniqueFilename()
	fileNameWithoutExt := strings.TrimSuffix(data.FileName, filepath.Ext(data.FileName))

	// 存储目录: {DataPath}/data/{原文件名不带后缀}/
	storageDir := filepath.Join(disk.DataPath, "data", fileNameWithoutExt)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return "", fmt.Errorf("创建存储目录失败: %w", err)
	}

	// 6. 判断是否需要分片存储（超大文件）
	threshold := int64(config.CONFIG.File.BigFileThreshold) * 1024 * 1024 * 1024 // GB转字节
	needChunkStorage := data.FileSize > threshold

	// 7. 文件加密（如果需要）
	var finalFilePath string
	var fileEncHash string
	if data.IsEnc {
		// 验证用户是否提供了加密密码
		if data.FilePassword == "" {
			return "", fmt.Errorf("加密文件必须提供密码")
		}

		// 查询用户信息（用于验证密码和获取用户ID作为盐）
		user, err := repoFactory.User().GetByID(ctx, data.UserID)
		if err != nil {
			return "", fmt.Errorf("查询用户信息失败: %w", err)
		}
		if user.FilePassword == "" {
			return "", fmt.Errorf("用户未设置文件加密密码")
		}

		// 验证用户输入的密码是否正确
		if !util.CheckPassword(user.FilePassword, data.FilePassword) {
			return "", fmt.Errorf("文件密码错误")
		}

		// 使用PBKDF2从明文密码和用户ID派生加密密钥
		// 这确保相同的密码+用户ID总是生成相同的密钥
		encryptionKey := util.DeriveEncryptionKey(data.FilePassword, data.UserID)
		logger.LOG.Debug("派生加密密钥", "userID", data.UserID, "keyLength", len(encryptionKey))

		// 加密文件
		encryptedPath := mergedFilePath + ".enc"
		crypto := util.NewFileCrypto(encryptionKey)
		if err := crypto.EncryptFile(mergedFilePath, encryptedPath); err != nil {
			return "", fmt.Errorf("文件加密失败: %w", err)
		}

		// 计算加密文件的hash
		encHasher := hash.NewFastBlake3Hasher()
		fileEncHash, _, err = encHasher.ComputeFileHash(encryptedPath)
		if err != nil {
			return "", fmt.Errorf("计算加密文件hash失败: %w", err)
		}
		logger.LOG.Debug("加密文件hash计算完成", "fileEncHash", fileEncHash)

		finalFilePath = encryptedPath
		// 加密后的临时文件会在cleanupTempFiles中一并清理（整个临时目录）
	} else {
		finalFilePath = mergedFilePath
	}

	// 8. 存储文件（根据是否需要分片）
	var chunks []*models.FileChunk
	var mainFilePath string
	var actualFileSize int64 // 实际文件大小

	if needChunkStorage {
		// 超大文件分片存储
		chunks, mainFilePath, err = splitAndStoreFile(finalFilePath, storageDir, virtualFileName, fileID, config.CONFIG.File.BigChunkSize)
		if err != nil {
			return "", fmt.Errorf("分片存储失败: %w", err)
		}
		// 计算实际文件大小（所有分片的总和）
		for _, chunk := range chunks {
			actualFileSize += int64(chunk.ChunkSize)
		}
	} else {
		// 普通文件直接存储
		mainFilePath = filepath.Join(storageDir, virtualFileName+".data")

		// 记录源文件大小用于调试
		srcInfo, err := os.Stat(finalFilePath)
		if err != nil {
			return "", fmt.Errorf("获取源文件信息失败: %w", err)
		}
		actualFileSize = srcInfo.Size() // 使用实际文件大小
		logger.LOG.Debug("准备复制文件", "源文件", finalFilePath, "目标文件", mainFilePath, "源文件大小", srcInfo.Size())

		if err := copyFile(finalFilePath, mainFilePath); err != nil {
			return "", fmt.Errorf("存储文件失败: %w", err)
		}

		// 验证复制后的文件大小
		dstInfo, err := os.Stat(mainFilePath)
		if err != nil {
			return "", fmt.Errorf("获取目标文件信息失败: %w", err)
		}
		logger.LOG.Debug("文件复制完成", "目标文件大小", dstInfo.Size())

		if dstInfo.Size() != srcInfo.Size() {
			return "", fmt.Errorf("文件复制后大小不一致: 源文件=%d, 目标文件=%d", srcInfo.Size(), dstInfo.Size())
		}
	}

	// 9. 存储缩略图（如果生成了）
	var thumbnailPath string
	if needThumbnail && tempThumbnailPath != "" {
		thumbnailPath = filepath.Join(storageDir, virtualFileName+".jpg")
		if err := copyFile(tempThumbnailPath, thumbnailPath); err != nil {
			logger.LOG.Warn("存储缩略图失败", "error", err)
			thumbnailPath = "" // 缩略图失败不影响主流程
		}
		// 临时缩略图会在cleanupTempFiles中一并清理（整个临时目录）
	}

	// 10. 使用数据库事务保证数据一致性
	// 如果是加密文件，加密文件路径就是主文件路径
	var encFilePath string
	if data.IsEnc {
		encFilePath = mainFilePath // 加密文件存储为.data文件
	}

	fileInfo := &models.FileInfo{
		ID:              fileID,
		Name:            data.FileName,
		RandomName:      virtualFileName,
		Size:            int(actualFileSize), // 使用实际计算的文件大小
		Mime:            mimeType,
		ThumbnailImg:    thumbnailPath,
		Path:            mainFilePath,
		FileHash:        fullHash,
		FileEncHash:     fileEncHash, // 加密文件的hash
		ChunkSignature:  data.ChunkSignature,
		FirstChunkHash:  data.FirstChunkHash,
		SecondChunkHash: data.SecondChunkHash,
		ThirdChunkHash:  data.ThirdChunkHash,
		HasFullHash:     true,
		IsEnc:           data.IsEnc,
		IsChunk:         needChunkStorage,
		ChunkCount:      len(chunks),
		EncPath:         encFilePath, // 加密文件的最终存储路径
		CreatedAt:       custom_type.Now(),
		UpdatedAt:       custom_type.Now(),
	}

	// 将虚拟路径转换为路径ID
	// 如果 VirtualPath 已经是路径ID（纯数字字符串），直接使用
	// 如果是路径字符串（如 "/home/"），则调用 getVirtualPathID 获取或创建
	var virtualPathID string
	if data.VirtualPath == "" {
		// 空路径，使用根目录
		rootPath, err := repoFactory.VirtualPath().GetRootPath(ctx, data.UserID)
		if err != nil {
			return "", fmt.Errorf("获取根目录失败: %w", err)
		}
		virtualPathID = fmt.Sprintf("%d", rootPath.ID)
	} else if matched, _ := regexp.MatchString(`^\d+$`, data.VirtualPath); matched {
		// 纯数字字符串，说明已经是路径ID，直接使用
		virtualPathID = data.VirtualPath
	} else {
		// 路径字符串，需要获取或创建路径
		var err error
		virtualPathID, err = getVirtualPathID(ctx, data.UserID, data.VirtualPath, repoFactory)
		if err != nil {
			return "", fmt.Errorf("获取虚拟路径ID失败: %w", err)
		}
	}

	userFile := &models.UserFiles{
		UserID:      data.UserID,
		FileID:      fileID,
		IsPublic:    false,         // 默认私有
		VirtualPath: virtualPathID, // 存储路径ID而不是路径字符串
		FileName:    data.FileName,
		CreatedAt:   custom_type.Now(),
		UfID:        uuid.NewString(),
	}

	// 开启数据库事务，确保所有数据库操作的原子性
	err = repoFactory.DB().Transaction(func(tx *gorm.DB) error {
		// 创建基于事务的仓储工厂
		txFactory := repoFactory.WithTx(tx)

		// 10.1 写入文件信息
		if err := txFactory.FileInfo().Create(ctx, fileInfo); err != nil {
			return fmt.Errorf("写入文件信息失败: %w", err)
		}

		// 10.2 写入分片信息（如果是分片存储）
		if len(chunks) > 0 {
			if err := txFactory.FileChunk().BatchCreate(ctx, chunks); err != nil {
				return fmt.Errorf("写入分片信息失败: %w", err)
			}
		}

		// 10.3 写入用户文件关联
		if err := txFactory.UserFiles().Create(ctx, userFile); err != nil {
			return fmt.Errorf("写入用户文件关联失败: %w", err)
		}

		// 10.4 更新用户剩余空间
		user, err := txFactory.User().GetByID(ctx, data.UserID)
		if err != nil {
			return fmt.Errorf("查询用户信息失败: %w", err)
		}
		if user.Space > 0 { // 如果不是无限空间
			user.FreeSpace -= actualFileSize // 使用实际文件大小
			if err := txFactory.User().Update(ctx, user); err != nil {
				return fmt.Errorf("更新用户剩余空间失败: %w", err)
			}
		}

		return nil // 事务成功，自动提交
	})

	if err != nil {
		// 事务回滚，需要清理已创建的文件
		cleanupProcessedFiles(mainFilePath, thumbnailPath, chunks)
		return "", err
	}

	// 10.5 写入.info文件（保存hash信息）
	if err := writeInfoFile(mainFilePath, fullHash, fileEncHash); err != nil {
		logger.LOG.Warn("写入.info文件失败", "error", err)
		// .info文件写入失败不影响主流程
	}

	logger.LOG.Info("文件处理完成", "fileID", fileID, "fileName", data.FileName, "size", actualFileSize)
	return fileID, nil
}

// mergeChunks 合并分片文件
func mergeChunks(data *FileUploadData) (string, error) {
	// 获取临时目录（应该是磁盘temp目录下的文件名子目录）
	tempDir := filepath.Dir(data.TempFilePath)
	mergedPath := filepath.Join(tempDir, "merged_"+filepath.Base(data.FileName))

	mergedFile, err := os.Create(mergedPath)
	if err != nil {
		return "", fmt.Errorf("创建合并文件失败: %w", err)
	}
	defer mergedFile.Close()

	// 按索引顺序合并分片
	for i := 0; i < data.ChunkCount; i++ {
		chunkPath := filepath.Join(tempDir, fmt.Sprintf("%d.chunk.data", i))
		chunkFile, err := os.Open(chunkPath)
		if err != nil {
			return "", fmt.Errorf("打开分片文件失败 [%d]: %w", i, err)
		}

		if _, err := io.Copy(mergedFile, chunkFile); err != nil {
			chunkFile.Close()
			return "", fmt.Errorf("合并分片失败 [%d]: %w", i, err)
		}
		chunkFile.Close()
	}

	return mergedPath, nil
}

// detectMimeType 检测文件MIME类型
func detectMimeType(filePath string) (string, error) {
	// 确保文件句柄立即释放 手动管理文件打开和关闭
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	mime, err := mimetype.DetectReader(file)
	if err != nil {
		return "", fmt.Errorf("检测MIME类型失败: %w", err)
	}
	return mime.String(), nil
}

// isImage 判断MIME类型是否为图片
func isImage(mimeType string) bool {
	return strings.HasPrefix(mimeType, "image/")
}

// selectBestDisk 选择剩余空间最大的磁盘
func selectBestDisk(ctx context.Context, repoFactory *impl.RepositoryFactory, fileSize int64) (*models.Disk, error) {
	disks, err := repoFactory.Disk().List(ctx, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("查询磁盘列表失败: %w", err)
	}

	if len(disks) == 0 {
		return nil, fmt.Errorf("没有可用的存储磁盘")
	}

	// 选择剩余空间最大且能容纳文件的磁盘
	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1

	for _, disk := range disks {
		// 磁盘大小单位是GB，需要转换为字节
		freeSpaceBytes := int64(disk.Size) * 1024 * 1024 * 1024 // GB转字节

		if freeSpaceBytes >= fileSize && freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}

	if bestDisk == nil {
		return nil, fmt.Errorf("没有足够空间的磁盘")
	}

	return bestDisk, nil
}

// splitAndStoreFile 分片存储大文件
func splitAndStoreFile(filePath, storageDir, virtualFileName, fileID string, chunkSizeGB int) ([]*models.FileChunk, string, error) {
	chunkSize := int64(chunkSizeGB) * 1024 * 1024 * 1024 // GB转字节

	file, err := os.Open(filePath)
	if err != nil {
		return nil, "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var chunks []*models.FileChunk
	var chunkIndex uint32 = 0
	buffer := make([]byte, chunkSize)

	for {
		n, err := file.Read(buffer)
		if err != nil && err != io.EOF {
			return nil, "", fmt.Errorf("读取文件失败: %w", err)
		}
		if n == 0 {
			break
		}

		// 分片文件路径
		chunkFileName := fmt.Sprintf("%s_%d.data", virtualFileName, chunkIndex)
		chunkPath := filepath.Join(storageDir, chunkFileName)

		// 写入分片
		if err := os.WriteFile(chunkPath, buffer[:n], 0644); err != nil {
			return nil, "", fmt.Errorf("写入分片失败: %w", err)
		}

		// 计算分片hash
		chunkHash := hash.ComputeBytes(buffer[:n])

		// 记录分片信息
		chunk := &models.FileChunk{
			ID:         uuid.Must(uuid.NewV7()).String(),
			FileID:     fileID,
			ChunkPath:  chunkPath,
			ChunkSize:  uint64(n),
			ChunkHash:  chunkHash,
			ChunkIndex: chunkIndex,
		}
		chunks = append(chunks, chunk)
		chunkIndex++
	}

	// 主文件路径返回第一个分片的路径
	mainPath := ""
	if len(chunks) > 0 {
		mainPath = chunks[0].ChunkPath
	}

	return chunks, mainPath, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return err
	}

	return destFile.Sync()
}

// cleanupTempFiles 清理临时文件和临时目录
func cleanupTempFiles(data *FileUploadData) {
	if data.TempFilePath == "" {
		return
	}

	// 获取临时目录（应该是磁盘temp目录下的文件名子目录）
	tempDir := filepath.Dir(data.TempFilePath)

	// Windows系统下文件句柄释放有延迟，需要重试删除整个临时目录
	cleanupDirWithRetry := func(dirPath string, maxRetries int) {
		for i := 0; i < maxRetries; i++ {
			err := os.RemoveAll(dirPath)
			if err == nil || os.IsNotExist(err) {
				logger.LOG.Info("清理临时目录成功", "path", dirPath)
				return
			}
			// 如果是文件被占用错误，等待一下再重试
			if i < maxRetries-1 {
				time.Sleep(time.Millisecond * 200) // 等待200ms
			} else {
				// 最后一次尝试失败，记录警告
				logger.LOG.Warn("清理临时目录失败", "path", dirPath, "error", err, "retries", maxRetries)
			}
		}
	}

	// 清理整个临时目录（包含所有分片、合并文件、加密临时文件等）
	cleanupDirWithRetry(tempDir, 5)
}

// FileHashInfo 文件hash信息JSON结构
type FileHashInfo struct {
	FileHash    string `json:"file_hash"`     // 原文件hash
	FileEncHash string `json:"file_enc_hash"` // 加密文件hash
}

// writeInfoFile 写入.info文件（保存hash信息的JSON格式）
func writeInfoFile(dataFilePath, fileHash, fileEncHash string) error {
	// 生成.info文件路径：将.data后缀替换为.info
	infoFilePath := strings.TrimSuffix(dataFilePath, ".data") + ".info"

	// 创建JSON数据
	jsonData := fmt.Sprintf(`{"file_hash":"%s","file_enc_hash":"%s"}`, fileHash, fileEncHash)

	// 写入文件
	if err := os.WriteFile(infoFilePath, []byte(jsonData), 0644); err != nil {
		return fmt.Errorf("写入.info文件失败: %w", err)
	}

	logger.LOG.Debug("写入.info文件成功", "path", infoFilePath)
	return nil
}

// cleanupProcessedFiles 清理已处理的文件（数据库操作失败时回滚）
func cleanupProcessedFiles(mainFilePath, thumbnailPath string, chunks []*models.FileChunk) {
	// 清理主文件
	if mainFilePath != "" {
		if err := os.Remove(mainFilePath); err != nil && !os.IsNotExist(err) {
			logger.LOG.Warn("清理主文件失败", "path", mainFilePath, "error", err)
		}

		// 清理.info文件
		infoPath := strings.TrimSuffix(mainFilePath, ".data") + ".info"
		if err := os.Remove(infoPath); err != nil && !os.IsNotExist(err) {
			logger.LOG.Warn("清理.info文件失败", "path", infoPath, "error", err)
		}
	}

	// 清理缩略图
	if thumbnailPath != "" {
		if err := os.Remove(thumbnailPath); err != nil && !os.IsNotExist(err) {
			logger.LOG.Warn("清理缩略图失败", "path", thumbnailPath, "error", err)
		}
	}

	// 清理分片文件
	for _, chunk := range chunks {
		if chunk.ChunkPath != "" {
			if err := os.Remove(chunk.ChunkPath); err != nil && !os.IsNotExist(err) {
				logger.LOG.Warn("清理分片文件失败", "path", chunk.ChunkPath, "error", err)
			}
		}
	}

	// 清理存储目录（如果为空）
	if mainFilePath != "" {
		storageDir := filepath.Dir(mainFilePath)
		// 尝试删除目录，如果不为空则会失败，这是预期的
		_ = os.Remove(storageDir)
	}

	logger.LOG.Info("已清理处理失败的文件", "mainFilePath", mainFilePath)
}

// getVirtualPathID 获取虚拟路径的ID（如果路径不存在则创建）
func getVirtualPathID(ctx context.Context, userID, fullPath string, repoFactory *impl.RepositoryFactory) (string, error) {
	// 分割路径为各层级
	parts := strings.Split(strings.Trim(fullPath, "/"), "/")
	if len(parts) == 0 {
		return "", fmt.Errorf("无效的虚拟路径: %s", fullPath)
	}

	// 首先获取用户的根目录（home），作为第一级子目录的父级
	rootPath, err := repoFactory.VirtualPath().GetRootPath(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("获取根目录失败: %w", err)
	}
	var parentID = fmt.Sprintf("%d", rootPath.ID) // 使用根目录的ID作为第一级子目录的父级ID
	var lastPathID string

	// 逐层查找或创建虚拟路径
	for _, part := range parts {
		if part == "" {
			continue
		}

		currentPath := "/" + part

		// 查询当前层级路径是否存在
		existingPaths, err := repoFactory.VirtualPath().ListByUserID(ctx, userID, 0, 1000)
		if err != nil {
			return "", fmt.Errorf("查询虚拟路径失败: %w", err)
		}

		// 查找是否已存在当前路径且父级匹配的记录
		var existingPath *models.VirtualPath
		for _, vp := range existingPaths {
			if vp.Path == currentPath && vp.ParentLevel == parentID {
				existingPath = vp
				break
			}
		}

		if existingPath != nil {
			// 路径已存在，使用该路径的ID作为下一层的父级ID
			parentID = fmt.Sprintf("%d", existingPath.ID)
			lastPathID = parentID
		} else {
			// 路径不存在，创建新记录
			newPath := &models.VirtualPath{
				UserID:      userID,
				Path:        currentPath,
				IsFile:      false,
				IsDir:       true,
				ParentLevel: parentID,
				CreatedTime: custom_type.Now(),
				UpdateTime:  custom_type.Now(),
			}

			if err := repoFactory.VirtualPath().Create(ctx, newPath); err != nil {
				// 可能是并发创建导致的重复，再次查询
				existingPaths, queryErr := repoFactory.VirtualPath().ListByUserID(ctx, userID, 0, 1000)
				if queryErr != nil {
					return "", fmt.Errorf("创建虚拟路径失败: %w, 查询失败: %w", err, queryErr)
				}
				for _, vp := range existingPaths {
					if vp.Path == currentPath && vp.ParentLevel == parentID {
						existingPath = vp
						break
					}
				}
				if existingPath != nil {
					parentID = fmt.Sprintf("%d", existingPath.ID)
					lastPathID = parentID
				} else {
					return "", fmt.Errorf("创建虚拟路径失败且无法查询到已创建的路径")
				}
			} else {
				parentID = fmt.Sprintf("%d", newPath.ID)
				lastPathID = parentID
			}
		}
	}

	if lastPathID == "" {
		return "", fmt.Errorf("无法获取虚拟路径ID: %s", fullPath)
	}

	return lastPathID, nil
}
