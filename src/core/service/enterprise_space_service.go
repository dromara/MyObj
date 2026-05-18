package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"myobj/src/config"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/hash"
	"myobj/src/pkg/models"
	"myobj/src/pkg/preview"
	"myobj/src/pkg/util"
	"os"
	"path/filepath"
	"strings"

	"github.com/gabriel-vasile/mimetype"
	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnterpriseSpaceService struct {
	factory    *impl.RepositoryFactory
	cacheLocal cache.Cache
}

func NewEnterpriseSpaceService(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *EnterpriseSpaceService {
	return &EnterpriseSpaceService{
		factory:    factory,
		cacheLocal: cacheLocal,
	}
}

func (s *EnterpriseSpaceService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// CreateDir 创建共享目录
func (s *EnterpriseSpaceService) CreateDir(req *request.CreateSharedDirRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证企业存在
	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}
	_ = enterprise

	// 验证用户是企业活跃成员
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	// 验证父目录存在（ParentID=0 表示根目录）
	if req.ParentID != 0 {
		parent, err := s.factory.EnterpriseSharedPath().GetByID(ctx, req.ParentID)
		if err != nil || parent == nil || parent.EnterpriseID != req.EnterpriseID {
			return models.NewJsonResponse(400, "父目录不存在", nil), err
		}
	}

	path := &models.EnterpriseSharedPath{
		EnterpriseID: req.EnterpriseID,
		Name:         req.Name,
		ParentID:     req.ParentID,
		CreatedAt:    custom_type.Now(),
	}

	if err := s.factory.EnterpriseSharedPath().Create(ctx, path); err != nil {
		return models.NewJsonResponse(500, "创建目录失败", nil), err
	}

	return models.NewJsonResponse(200, "创建成功", path), nil
}

// ListFiles 列出共享空间指定目录下的文件和子目录
func (s *EnterpriseSpaceService) ListFiles(req *request.SharedFileListRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证用户是企业活跃成员
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	// 获取子目录
	paths, err := s.factory.EnterpriseSharedPath().ListByParentID(ctx, req.EnterpriseID, req.PathID)
	if err != nil {
		return models.NewJsonResponse(500, "查询目录失败", nil), err
	}

	// 获取文件
	offset := (req.Page - 1) * req.PageSize
	files, err := s.factory.EnterpriseSharedFile().ListByPathID(ctx, req.EnterpriseID, req.PathID, offset, req.PageSize)
	if err != nil {
		return models.NewJsonResponse(500, "查询文件失败", nil), err
	}

	total, err := s.factory.EnterpriseSharedFile().CountByPathID(ctx, req.EnterpriseID, req.PathID)
	if err != nil {
		return models.NewJsonResponse(500, "统计文件失败", nil), err
	}

	result := map[string]interface{}{
		"dirs":  paths,
		"files": files,
		"total": total,
		"page":  req.Page,
		"pageSize": req.PageSize,
	}

	return models.NewJsonResponse(200, "查询成功", result), nil
}

// UploadPrecheck 共享空间上传预检
func (s *EnterpriseSpaceService) UploadPrecheck(req *request.SharedUploadPrecheckRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证企业存在
	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	// 验证用户是企业活跃成员
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}
	_ = member

	// 检查企业可用空间
	if enterprise.Space > 0 && enterprise.FreeSpace < req.FileSize {
		return models.NewJsonResponse(400, "企业可用空间不足", nil), nil
	}

	// 检查磁盘空间
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil {
		return nil, err
	}
	if len(disks) == 0 {
		return models.NewJsonResponse(500, "没有可用的存储磁盘", nil), nil
	}

	var hasEnoughSpace bool
	for _, disk := range disks {
		if disk.Size >= req.FileSize {
			hasEnoughSpace = true
			break
		}
	}
	if !hasEnoughSpace {
		return models.NewJsonResponse(400, "磁盘空间不足，请联系管理员扩容", nil), nil
	}

	// 检查秒传（通过 chunk_signature + file_size 匹配已有文件）
	signature, err := s.factory.FileInfo().GetByChunkSignature(ctx, req.ChunkSignature, req.FileSize)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	if err == nil && signature != nil && signature.ID != "" && !signature.IsEnc {
		// 秒传成功：创建共享文件关联
		sharedFile := &models.EnterpriseSharedFile{
			ID:           uuid.NewString(),
			EnterpriseID: req.EnterpriseID,
			FileID:       signature.ID,
			FileName:     req.FileName,
			PathID:       req.PathID,
			UploaderID:   userID,
			Size:         req.FileSize,
			CreatedAt:    custom_type.Now(),
		}

		txErr := s.factory.DB().Transaction(func(tx *gorm.DB) error {
			txFactory := s.factory.WithTx(tx)
			if err := txFactory.EnterpriseSharedFile().Create(ctx, sharedFile); err != nil {
				return err
			}
			// 扣除企业空间
			if enterprise.Space > 0 {
				enterprise.FreeSpace -= req.FileSize
				if err := txFactory.Enterprise().Update(ctx, enterprise); err != nil {
					return err
				}
			}
			return nil
		})

		if txErr != nil {
			return models.NewJsonResponse(500, "秒传失败", nil), txErr
		}

		return models.NewJsonResponse(200, "秒传成功", nil), nil
	}

	// 生成预检ID并缓存
	precheckID := uuid.New().String()
	cacheKey := fmt.Sprintf("sharedUpload:%s", precheckID)
	precheckData := &response.FilePrecheckResponse{
		PrecheckID: precheckID,
	}

	// 缓存预检信息（30分钟过期）
	if err := s.cacheLocal.Set(cacheKey, precheckData, 1800); err != nil {
		return nil, fmt.Errorf("缓存预检信息失败: %w", err)
	}

	// 缓存原始请求信息
	reqCacheKey := fmt.Sprintf("sharedUploadReq:%s", precheckID)
	if err := s.cacheLocal.Set(reqCacheKey, req, 1800); err != nil {
		return nil, fmt.Errorf("缓存预检请求失败: %w", err)
	}

	return models.NewJsonResponse(200, "预检成功", precheckData), nil
}

// UploadFile 处理共享空间文件上传
func (s *EnterpriseSpaceService) UploadFile(req *request.SharedFileUploadRequest, file multipart.File, header *multipart.FileHeader, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 1. 从缓存获取预检信息
	reqCacheKey := fmt.Sprintf("sharedUploadReq:%s", req.PrecheckID)
	reqData, err := s.cacheLocal.Get(reqCacheKey)
	if err != nil {
		return nil, fmt.Errorf("预检信息已过期或不存在")
	}

	var precheckReq request.SharedUploadPrecheckRequest
	switch v := reqData.(type) {
	case *request.SharedUploadPrecheckRequest:
		precheckReq = *v
	case string:
		if err := json.Unmarshal([]byte(v), &precheckReq); err != nil {
			return nil, fmt.Errorf("预检请求信息格式错误")
		}
	default:
		return nil, fmt.Errorf("预检请求信息类型错误")
	}

	if precheckReq.EnterpriseID != req.EnterpriseID {
		return nil, fmt.Errorf("企业ID不匹配")
	}

	// 2. 验证企业空间
	enterprise, err := s.factory.Enterprise().GetByID(ctx, req.EnterpriseID)
	if err != nil {
		return nil, fmt.Errorf("企业不存在")
	}
	if enterprise.Space > 0 && enterprise.FreeSpace < precheckReq.FileSize {
		return nil, fmt.Errorf("企业可用空间不足")
	}

	// 3. 选择磁盘
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("查询磁盘列表失败: %w", err)
	}
	var bestDisk *models.Disk
	var maxSize int64 = -1
	for _, disk := range disks {
		if disk.Size >= precheckReq.FileSize && disk.Size > maxSize {
			maxSize = disk.Size
			bestDisk = disk
		}
	}
	if bestDisk == nil {
		return nil, fmt.Errorf("没有足够空间的磁盘")
	}

	// 4. 保存临时文件
	sessionID := req.PrecheckID[:8]
	fileNameWithoutExt := precheckReq.FileName
	if idx := strings.LastIndex(precheckReq.FileName, "."); idx != -1 {
		fileNameWithoutExt = precheckReq.FileName[:idx]
	}
	tempBaseDir := filepath.Join(bestDisk.DataPath, "temp", fmt.Sprintf("enterprise_%s_%s", fileNameWithoutExt, sessionID))
	if err := os.MkdirAll(tempBaseDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 判断是否分片上传
	isChunkUpload := req.ChunkIndex != nil && req.TotalChunks != nil

	var tempFilePath string
	if isChunkUpload {
		// 分片上传：保存分片
		chunkPath := filepath.Join(tempBaseDir, fmt.Sprintf("%d.chunk.data", *req.ChunkIndex))
		dst, err := os.Create(chunkPath)
		if err != nil {
			return nil, fmt.Errorf("创建分片文件失败: %w", err)
		}
		if _, err := io.Copy(dst, file); err != nil {
			dst.Close()
			return nil, fmt.Errorf("保存分片失败: %w", err)
		}
		dst.Close()
		tempFilePath = chunkPath

		// 更新上传进度到缓存
		progressKey := fmt.Sprintf("sharedUploadProgress:%s", req.PrecheckID)
		s.cacheLocal.Set(progressKey, map[string]interface{}{
			"chunk_index":   *req.ChunkIndex,
			"total_chunks":  *req.TotalChunks,
			"precheck_id":   req.PrecheckID,
			"file_name":     precheckReq.FileName,
			"file_size":     precheckReq.FileSize,
		}, 1800)

		// 如果不是最后一个分片，返回进度
		if *req.ChunkIndex < *req.TotalChunks-1 {
			return models.NewJsonResponse(200, "分片上传成功", map[string]interface{}{
				"chunk_index":  *req.ChunkIndex,
				"total_chunks": *req.TotalChunks,
			}), nil
		}

		// 最后一个分片：合并分片
		mergedPath := filepath.Join(tempBaseDir, "merged_"+precheckReq.FileName)
		mergedFile, err := os.Create(mergedPath)
		if err != nil {
			return nil, fmt.Errorf("创建合并文件失败: %w", err)
		}
		for i := 0; i < *req.TotalChunks; i++ {
			chunkFile, err := os.Open(filepath.Join(tempBaseDir, fmt.Sprintf("%d.chunk.data", i)))
			if err != nil {
				mergedFile.Close()
				return nil, fmt.Errorf("打开分片失败: %w", err)
			}
			if _, err := io.Copy(mergedFile, chunkFile); err != nil {
				chunkFile.Close()
				mergedFile.Close()
				return nil, fmt.Errorf("合并分片失败: %w", err)
			}
			chunkFile.Close()
		}
		mergedFile.Close()
		tempFilePath = mergedPath
	} else {
		// 单文件上传
		tempFilePath = filepath.Join(tempBaseDir, precheckReq.FileName)
		dst, err := os.Create(tempFilePath)
		if err != nil {
			return nil, fmt.Errorf("创建临时文件失败: %w", err)
		}
		if _, err := io.Copy(dst, file); err != nil {
			dst.Close()
			return nil, fmt.Errorf("保存文件失败: %w", err)
		}
		dst.Close()
	}

	// 确保清理临时文件
	defer func() {
		os.RemoveAll(tempBaseDir)
	}()

	// 5. 检测MIME类型
	mimeFile, err := os.Open(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	mimeType, err := mimetype.DetectReader(mimeFile)
	mimeFile.Close()
	if err != nil {
		return nil, fmt.Errorf("检测文件类型失败: %w", err)
	}

	// 6. 计算hash
	hasher := hash.NewFastBlake3Hasher()
	fullHash, _, err := hasher.ComputeFileHash(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("计算文件hash失败: %w", err)
	}

	// 7. 生成存储路径
	fileID := uuid.Must(uuid.NewV7()).String()
	virtualFileName := util.GenerateUniqueFilename()
	storageDir := filepath.Join(bestDisk.DataPath, "data", fileNameWithoutExt)
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}

	// 8. 存储文件
	mainFilePath := filepath.Join(storageDir, virtualFileName+".data")
	srcFile, err := os.Open(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("打开源文件失败: %w", err)
	}
	dstFile, err := os.Create(mainFilePath)
	if err != nil {
		srcFile.Close()
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	written, err := io.Copy(dstFile, srcFile)
	srcFile.Close()
	dstFile.Close()
	if err != nil {
		os.Remove(mainFilePath)
		return nil, fmt.Errorf("存储文件失败: %w", err)
	}

	// 9. 生成缩略图
	var thumbnailPath string
	if config.CONFIG.File.Thumbnail && strings.HasPrefix(mimeType.String(), "image/") {
		thumbnailPath = filepath.Join(storageDir, virtualFileName+".jpg")
		if err := preview.GenerateImageThumbnail(tempFilePath, thumbnailPath, 300); err != nil {
			thumbnailPath = ""
		}
	}

	// 10. 写入.info文件
	infoPath := strings.TrimSuffix(mainFilePath, ".data") + ".info"
	infoJson := fmt.Sprintf(`{"file_hash":"%s","file_enc_hash":""}`, fullHash)
	if err := os.WriteFile(infoPath, []byte(infoJson), 0644); err != nil {
		return nil, fmt.Errorf("写入info文件失败: %w", err)
	}

	// 11. 数据库事务
	fileInfo := &models.FileInfo{
		ID:             fileID,
		Name:           precheckReq.FileName,
		RandomName:     virtualFileName,
		Size:           int(written),
		Mime:           mimeType.String(),
		ThumbnailImg:   thumbnailPath,
		Path:           mainFilePath,
		FileHash:       fullHash,
		ChunkSignature: precheckReq.ChunkSignature,
		HasFullHash:    true,
		IsEnc:          req.IsEnc,
		IsChunk:        false,
		CreatedAt:      custom_type.Now(),
		UpdatedAt:      custom_type.Now(),
	}

	sharedFile := &models.EnterpriseSharedFile{
		ID:           uuid.NewString(),
		EnterpriseID: req.EnterpriseID,
		FileID:       fileID,
		FileName:     precheckReq.FileName,
		PathID:       req.PathID,
		UploaderID:   userID,
		Size:         written,
		CreatedAt:    custom_type.Now(),
	}

	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)

		if err := txFactory.FileInfo().Create(ctx, fileInfo); err != nil {
			return fmt.Errorf("写入文件信息失败: %w", err)
		}

		if err := txFactory.EnterpriseSharedFile().Create(ctx, sharedFile); err != nil {
			return fmt.Errorf("写入共享文件关联失败: %w", err)
		}

		// 扣除企业空间
		if enterprise.Space > 0 {
			enterprise.FreeSpace -= written
			if err := txFactory.Enterprise().Update(ctx, enterprise); err != nil {
				return fmt.Errorf("更新企业空间失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		os.Remove(mainFilePath)
		if thumbnailPath != "" {
			os.Remove(thumbnailPath)
		}
		infoPath := strings.TrimSuffix(mainFilePath, ".data") + ".info"
		os.Remove(infoPath)
		return nil, err
	}

	// 清理缓存
	s.cacheLocal.Delete(fmt.Sprintf("sharedUpload:%s", req.PrecheckID))
	s.cacheLocal.Delete(fmt.Sprintf("sharedUploadReq:%s", req.PrecheckID))
	s.cacheLocal.Delete(fmt.Sprintf("sharedUploadProgress:%s", req.PrecheckID))

	return models.NewJsonResponse(200, "上传成功", map[string]interface{}{
		"file_id":   fileID,
		"file_name": precheckReq.FileName,
		"size":      written,
	}), nil
}

// DeleteFile 删除共享文件
func (s *EnterpriseSpaceService) DeleteFile(fileID string, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}

	// 验证用户是企业活跃成员
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, sharedFile.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, sharedFile.EnterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)

		if err := txFactory.EnterpriseSharedFile().Delete(ctx, fileID); err != nil {
			return fmt.Errorf("删除共享文件记录失败: %w", err)
		}

		// 恢复企业空间
		if enterprise.Space > 0 {
			enterprise.FreeSpace += sharedFile.Size
			if err := txFactory.Enterprise().Update(ctx, enterprise); err != nil {
				return fmt.Errorf("更新企业空间失败: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return models.NewJsonResponse(500, "删除失败", nil), err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// DownloadFile 获取共享文件下载信息
func (s *EnterpriseSpaceService) DownloadFile(fileID string, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}

	// 验证用户是企业活跃成员
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, sharedFile.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	// 获取物理文件信息
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, sharedFile.FileID)
	if err != nil {
		return models.NewJsonResponse(404, "物理文件不存在", nil), err
	}

	result := map[string]interface{}{
		"file_id":    sharedFile.ID,
		"file_name":  sharedFile.FileName,
		"file_path":  fileInfo.Path,
		"mime":       fileInfo.Mime,
		"size":       sharedFile.Size,
		"is_enc":     fileInfo.IsEnc,
		"created_at": sharedFile.CreatedAt,
	}

	return models.NewJsonResponse(200, "查询成功", result), nil
}

// GetSpaceUsage 获取企业空间使用情况
func (s *EnterpriseSpaceService) GetSpaceUsage(enterpriseID string, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证用户是企业活跃成员
	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	enterprise, err := s.factory.Enterprise().GetByID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(404, "企业不存在", nil), err
	}

	usedSpace, err := s.factory.EnterpriseSharedFile().SumSizeByEnterpriseID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(500, "查询空间使用失败", nil), err
	}

	fileCount, err := s.factory.EnterpriseSharedFile().CountByEnterpriseID(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(500, "查询文件数量失败", nil), err
	}

	result := map[string]interface{}{
		"total_space":  enterprise.Space,
		"free_space":   enterprise.FreeSpace,
		"used_space":   usedSpace,
		"file_count":   fileCount,
	}

	return models.NewJsonResponse(200, "查询成功", result), nil
}

// DeleteDir 删除共享目录（递归删除子目录和文件）
func (s *EnterpriseSpaceService) DeleteDir(dirID int, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	dir, err := s.factory.EnterpriseSharedPath().GetByID(ctx, dirID)
	if err != nil || dir == nil {
		return models.NewJsonResponse(404, "目录不存在", nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, dir.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	var totalSize int64
	var fileIDs []string
	s.collectFilesRecursive(ctx, dir.EnterpriseID, dirID, &fileIDs, &totalSize)

	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)
		for _, fid := range fileIDs {
			if err := txFactory.EnterpriseSharedFile().Delete(ctx, fid); err != nil {
				return err
			}
		}
		if err := s.deletePathsRecursive(ctx, txFactory, dir.EnterpriseID, dirID); err != nil {
			return err
		}
		enterprise, err := txFactory.Enterprise().GetByID(ctx, dir.EnterpriseID)
		if err != nil {
			return err
		}
		if enterprise.Space > 0 && totalSize > 0 {
			enterprise.FreeSpace += totalSize
			if err := txFactory.Enterprise().Update(ctx, enterprise); err != nil {
				return err
			}
		}
		return nil
	})

	if err != nil {
		return models.NewJsonResponse(500, "删除目录失败", nil), err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

func (s *EnterpriseSpaceService) collectFilesRecursive(ctx context.Context, enterpriseID string, pathID int, fileIDs *[]string, totalSize *int64) {
	files, _ := s.factory.EnterpriseSharedFile().ListByPathID(ctx, enterpriseID, pathID, 0, 10000)
	for _, f := range files {
		*fileIDs = append(*fileIDs, f.ID)
		*totalSize += f.Size
	}
	subDirs, _ := s.factory.EnterpriseSharedPath().ListByParentID(ctx, enterpriseID, pathID)
	for _, d := range subDirs {
		s.collectFilesRecursive(ctx, enterpriseID, d.ID, fileIDs, totalSize)
	}
}

func (s *EnterpriseSpaceService) deletePathsRecursive(ctx context.Context, txFactory *impl.RepositoryFactory, enterpriseID string, pathID int) error {
	subDirs, _ := txFactory.EnterpriseSharedPath().ListByParentID(ctx, enterpriseID, pathID)
	for _, d := range subDirs {
		if err := s.deletePathsRecursive(ctx, txFactory, enterpriseID, d.ID); err != nil {
			return err
		}
	}
	return txFactory.EnterpriseSharedPath().Delete(ctx, pathID)
}

// RenameFile 重命名共享文件
func (s *EnterpriseSpaceService) RenameFile(fileID, newName string, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil || sharedFile == nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, sharedFile.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	sharedFile.FileName = newName
	if err := s.factory.EnterpriseSharedFile().Update(ctx, sharedFile); err != nil {
		return models.NewJsonResponse(500, "重命名失败", nil), err
	}

	return models.NewJsonResponse(200, "重命名成功", nil), nil
}

// RenameDir 重命名共享目录
func (s *EnterpriseSpaceService) RenameDir(dirID int, newName string, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	dir, err := s.factory.EnterpriseSharedPath().GetByID(ctx, dirID)
	if err != nil || dir == nil {
		return models.NewJsonResponse(404, "目录不存在", nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, dir.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	dir.Name = newName
	if err := s.factory.EnterpriseSharedPath().Update(ctx, dir); err != nil {
		return models.NewJsonResponse(500, "重命名失败", nil), err
	}

	return models.NewJsonResponse(200, "重命名成功", nil), nil
}
