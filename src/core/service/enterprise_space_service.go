package service

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
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
	"time"

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
func (s *EnterpriseSpaceService) CreateDir(req *request.CreateSharedDirRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

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
		CreatedBy:    userID,
		CreatedAt:    custom_type.Now(),
	}

	if err := s.factory.EnterpriseSharedPath().Create(ctx, path); err != nil {
		return models.NewJsonResponse(500, "创建目录失败", nil), err
	}

	return models.NewJsonResponse(200, "创建成功", path), nil
}

// ListFiles 列出共享空间指定目录下的文件和子目录
func (s *EnterpriseSpaceService) ListFiles(req *request.SharedFileListRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

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

	// 获取文件（支持排序）
	sortOrder := req.SortOrder
	if sortOrder == "" {
		sortOrder = "DESC"
	}
	sortField := "created_at"
	switch req.SortBy {
	case "name":
		sortField = "file_name"
	case "size":
		sortField = "size"
	case "updated_at":
		sortField = "updated_at"
	}
	offset := (req.Page - 1) * req.PageSize
	files, err := s.factory.EnterpriseSharedFile().ListByPathIDWithSort(ctx, req.EnterpriseID, req.PathID, sortField, sortOrder, offset, req.PageSize)
	if err != nil {
		return models.NewJsonResponse(500, "查询文件失败", nil), err
	}

	total, err := s.factory.EnterpriseSharedFile().CountByPathID(ctx, req.EnterpriseID, req.PathID)
	if err != nil {
		return models.NewJsonResponse(500, "统计文件失败", nil), err
	}

	// 构建带用户名的目录列表
	var dirList []map[string]interface{}
	for _, p := range paths {
		item := map[string]interface{}{
			"id":         p.ID,
			"name":       p.Name,
			"parent_id":  p.ParentID,
			"created_by": p.CreatedBy,
			"created_at": p.CreatedAt,
			"updated_by": p.UpdatedBy,
			"updated_at": p.UpdatedAt,
		}
		if p.CreatedBy != "" {
			if creator, _ := s.factory.User().GetByID(ctx, p.CreatedBy); creator != nil {
				name := creator.UserName
				if creator.Name != "" {
					name = creator.Name
				}
				item["uploader_name"] = name
			}
		}
		// 查询修改人名称
		if p.UpdatedBy != "" {
			if updater, _ := s.factory.User().GetByID(ctx, p.UpdatedBy); updater != nil {
				name := updater.UserName
				if updater.Name != "" {
					name = updater.Name
				}
				item["updated_by_name"] = name
			}
		}
		dirList = append(dirList, item)
	}

	// 构建带用户名的文件列表
	var fileList []map[string]interface{}
	for _, f := range files {
		item := map[string]interface{}{
			"id":          f.ID,
			"file_id":     f.FileID,
			"file_name":   f.FileName,
			"path_id":     f.PathID,
			"uploader_id": f.UploaderID,
			"size":        f.Size,
			"created_at":  f.CreatedAt,
			"updated_by":  f.UpdatedBy,
			"updated_at":  f.UpdatedAt,
		}
		// 查询上传人名称
		if uploader, _ := s.factory.User().GetByID(ctx, f.UploaderID); uploader != nil {
			name := uploader.UserName
			if uploader.Name != "" {
				name = uploader.Name
			}
			item["uploader_name"] = name
		}
		// 查询修改人名称
		if f.UpdatedBy != "" {
			if updater, _ := s.factory.User().GetByID(ctx, f.UpdatedBy); updater != nil {
				name := updater.UserName
				if updater.Name != "" {
					name = updater.Name
				}
				item["updated_by_name"] = name
			}
		}
		fileList = append(fileList, item)
	}

	result := map[string]interface{}{
		"dirs":     dirList,
		"files":    fileList,
		"total":    total,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}

	return models.NewJsonResponse(200, "查询成功", result), nil
}

// UploadPrecheck 共享空间上传预检
func (s *EnterpriseSpaceService) UploadPrecheck(req *request.SharedUploadPrecheckRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

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
	if !enterprise.SpaceUnlimited {
		if enterprise.Space == 0 {
			return models.NewJsonResponse(400, "企业未分配存储空间，无法上传", nil), nil
		}
		if enterprise.FreeSpace < req.FileSize {
			return models.NewJsonResponse(400, "企业可用空间不足", nil), nil
		}
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
			if !enterprise.SpaceUnlimited {
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

	// 序列化预检信息后缓存（Redis 需要字符串存储）
	precheckJSON, err := json.Marshal(precheckData)
	if err != nil {
		return nil, fmt.Errorf("序列化预检信息失败: %w", err)
	}
	if err := s.cacheLocal.Set(cacheKey, string(precheckJSON), 1800); err != nil {
		return nil, fmt.Errorf("缓存预检信息失败: %w", err)
	}

	// 序列化原始请求信息后缓存
	reqCacheKey := fmt.Sprintf("sharedUploadReq:%s", precheckID)
	reqJSON, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化预检请求失败: %w", err)
	}
	if err := s.cacheLocal.Set(reqCacheKey, string(reqJSON), 1800); err != nil {
		return nil, fmt.Errorf("缓存预检请求失败: %w", err)
	}

	return models.NewJsonResponse(200, "预检成功", precheckData), nil
}

// UploadFile 处理共享空间文件上传
func (s *EnterpriseSpaceService) UploadFile(req *request.SharedFileUploadRequest, file multipart.File, header *multipart.FileHeader, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return nil, err
	}

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
	if !enterprise.SpaceUnlimited {
		if enterprise.Space == 0 {
			return nil, fmt.Errorf("企业未分配存储空间，无法上传")
		}
		if enterprise.FreeSpace < precheckReq.FileSize {
			return nil, fmt.Errorf("企业可用空间不足")
		}
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
		if !enterprise.SpaceUnlimited {
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
func (s *EnterpriseSpaceService) DeleteFile(fileID, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	physicalFileID := sharedFile.FileID
	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)

		if err := txFactory.EnterpriseSharedFile().Delete(ctx, fileID); err != nil {
			return fmt.Errorf("删除共享文件记录失败: %w", err)
		}

		if err := tx.Model(&models.Enterprise{}).
			Where("id = ? AND space_unlimited = ?", sharedFile.EnterpriseID, false).
			Update("free_space", gorm.Expr("free_space + ?", sharedFile.Size)).Error; err != nil {
			return fmt.Errorf("更新企业空间失败: %w", err)
		}

		return s.cleanupPhysicalFileIfUnreferenced(ctx, txFactory, physicalFileID)
	})

	if err != nil {
		return models.NewJsonResponse(500, "删除失败", nil), err
	}

	return models.NewJsonResponse(200, "删除成功", nil), nil
}

// DownloadFile 获取共享文件下载信息
func (s *EnterpriseSpaceService) DownloadFile(fileID, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
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
func (s *EnterpriseSpaceService) GetSpaceUsage(enterpriseID, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, enterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

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
func (s *EnterpriseSpaceService) DeleteDir(dirID int, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	dir, err := s.factory.EnterpriseSharedPath().GetByID(ctx, dirID)
	if err != nil || dir == nil {
		return models.NewJsonResponse(404, "目录不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, dir.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, dir.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	var totalSize int64
	var deletedFileIDs []string
	err = s.factory.DB().Transaction(func(tx *gorm.DB) error {
		txFactory := s.factory.WithTx(tx)
		if err := s.collectAndDeleteInTx(ctx, txFactory, dir.EnterpriseID, dirID, &totalSize, &deletedFileIDs); err != nil {
			return err
		}
		if err := s.deletePathsRecursive(ctx, txFactory, dir.EnterpriseID, dirID); err != nil {
			return err
		}
		// 恢复空间配额
		if totalSize > 0 {
			enterprise, err := txFactory.Enterprise().GetByID(ctx, dir.EnterpriseID)
			if err != nil {
				return err
			}
			if !enterprise.SpaceUnlimited {
				enterprise.FreeSpace += totalSize
				if err := txFactory.Enterprise().Update(ctx, enterprise); err != nil {
					return err
				}
			}
		}
		for _, fid := range deletedFileIDs {
			if err := s.cleanupPhysicalFileIfUnreferenced(ctx, txFactory, fid); err != nil {
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

// collectAndDeleteInTx 在事务内递归收集并删除目录下的所有文件
func (s *EnterpriseSpaceService) collectAndDeleteInTx(ctx context.Context, txFactory *impl.RepositoryFactory, enterpriseID string, pathID int, totalSize *int64, deletedFileIDs *[]string) error {
	offset := 0
	batchSize := 1000
	for {
		files, err := txFactory.EnterpriseSharedFile().ListByPathID(ctx, enterpriseID, pathID, offset, batchSize)
		if err != nil {
			return err
		}
		if len(files) == 0 {
			break
		}
		for _, f := range files {
			*totalSize += f.Size
			*deletedFileIDs = append(*deletedFileIDs, f.FileID)
			if err := txFactory.EnterpriseSharedFile().Delete(ctx, f.ID); err != nil {
				return err
			}
		}
		if len(files) < batchSize {
			break
		}
		offset += batchSize
	}
	subDirs, err := txFactory.EnterpriseSharedPath().ListByParentID(ctx, enterpriseID, pathID)
	if err != nil {
		return err
	}
	for _, d := range subDirs {
		if err := s.collectAndDeleteInTx(ctx, txFactory, enterpriseID, d.ID, totalSize, deletedFileIDs); err != nil {
			return err
		}
	}
	return nil
}

func (s *EnterpriseSpaceService) deletePathsRecursive(ctx context.Context, txFactory *impl.RepositoryFactory, enterpriseID string, pathID int) error {
	subDirs, err := txFactory.EnterpriseSharedPath().ListByParentID(ctx, enterpriseID, pathID)
	if err != nil {
		return err
	}
	for _, d := range subDirs {
		if err := s.deletePathsRecursive(ctx, txFactory, enterpriseID, d.ID); err != nil {
			return err
		}
	}
	return txFactory.EnterpriseSharedPath().Delete(ctx, pathID)
}

// RenameFile 重命名共享文件
func (s *EnterpriseSpaceService) RenameFile(fileID, newName, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil || sharedFile == nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	sharedFile.FileName = newName
	sharedFile.UpdatedBy = userID
	sharedFile.UpdatedAt = custom_type.Now()
	if err := s.factory.EnterpriseSharedFile().Update(ctx, sharedFile); err != nil {
		return models.NewJsonResponse(500, "重命名失败", nil), err
	}

	return models.NewJsonResponse(200, "重命名成功", nil), nil
}

// RenameDir 重命名共享目录
func (s *EnterpriseSpaceService) RenameDir(dirID int, newName, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	dir, err := s.factory.EnterpriseSharedPath().GetByID(ctx, dirID)
	if err != nil || dir == nil {
		return models.NewJsonResponse(404, "目录不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, dir.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, dir.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	dir.Name = newName
	dir.UpdatedBy = userID
	dir.UpdatedAt = custom_type.Now()
	if err := s.factory.EnterpriseSharedPath().Update(ctx, dir); err != nil {
		return models.NewJsonResponse(500, "重命名失败", nil), err
	}

	return models.NewJsonResponse(200, "重命名成功", nil), nil
}

// GetThumbnailPath 获取文件缩略图路径
func (s *EnterpriseSpaceService) GetThumbnailPath(fileID, contextEnterpriseID, userID string) (string, error) {
	ctx := context.Background()

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
	if err != nil || sharedFile == nil {
		return "", fmt.Errorf("文件不存在")
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return "", err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return "", err
	}

	fileInfo, err := s.factory.FileInfo().GetByID(ctx, sharedFile.FileID)
	if err != nil || fileInfo == nil {
		return "", fmt.Errorf("文件信息不存在")
	}

	if fileInfo.ThumbnailImg == "" {
		return "", fmt.Errorf("无缩略图")
	}

	return fileInfo.ThumbnailImg, nil
}

// SearchFiles 搜索企业空间文件
func (s *EnterpriseSpaceService) SearchFiles(req *request.SearchEnterpriseFilesRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, req.EnterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	offset := (req.Page - 1) * req.PageSize
	files, err := s.factory.EnterpriseSharedFile().ListByEnterpriseID(ctx, req.EnterpriseID, req.Keyword, offset, req.PageSize)
	if err != nil {
		return models.NewJsonResponse(500, "搜索失败", nil), err
	}

	total, _ := s.factory.EnterpriseSharedFile().CountByEnterpriseIDAndKeyword(ctx, req.EnterpriseID, req.Keyword)

	// 构建带用户名的文件列表
	var fileList []map[string]interface{}
	for _, f := range files {
		item := map[string]interface{}{
			"id":         f.ID,
			"file_id":    f.FileID,
			"file_name":  f.FileName,
			"path_id":    f.PathID,
			"uploader_id": f.UploaderID,
			"size":       f.Size,
			"created_at": f.CreatedAt,
			"updated_by": f.UpdatedBy,
			"updated_at": f.UpdatedAt,
		}
		// 查询上传人名称
		if uploader, _ := s.factory.User().GetByID(ctx, f.UploaderID); uploader != nil {
			name := uploader.UserName
			if uploader.Name != "" {
				name = uploader.Name
			}
			item["uploader_name"] = name
		}
		// 查询修改人名称
		if f.UpdatedBy != "" {
			if updater, _ := s.factory.User().GetByID(ctx, f.UpdatedBy); updater != nil {
				name := updater.UserName
				if updater.Name != "" {
					name = updater.Name
				}
				item["updated_by_name"] = name
			}
		}
		fileList = append(fileList, item)
	}

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"files": fileList,
		"total": total,
		"page":  req.Page,
		"pageSize": req.PageSize,
	}), nil
}

// GetPathTree 获取企业空间目录树
func (s *EnterpriseSpaceService) GetPathTree(enterpriseID, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, enterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	member, err := s.factory.EnterpriseMember().GetByEnterpriseAndUser(ctx, enterpriseID, userID)
	if err != nil || member == nil || member.Status != 0 {
		return models.NewJsonResponse(403, "您不是该企业的活跃成员", nil), err
	}

	paths, err := s.factory.EnterpriseSharedPath().GetPathTree(ctx, enterpriseID)
	if err != nil {
		return models.NewJsonResponse(500, "获取目录树失败", nil), err
	}

	// 构建树形结构
	type TreeNode struct {
		ID       int        `json:"id"`
		Name     string     `json:"name"`
		ParentID int        `json:"parent_id"`
		Children []*TreeNode `json:"children,omitempty"`
	}

	nodeMap := make(map[int]*TreeNode)
	var roots []*TreeNode

	for _, p := range paths {
		node := &TreeNode{ID: p.ID, Name: p.Name, ParentID: p.ParentID}
		nodeMap[p.ID] = node
	}

	for _, p := range paths {
		node := nodeMap[p.ID]
		if p.ParentID == 0 {
			roots = append(roots, node)
		} else if parent, ok := nodeMap[p.ParentID]; ok {
			parent.Children = append(parent.Children, node)
		}
	}

	return models.NewJsonResponse(200, "ok", roots), nil
}

// MoveFile 移动企业空间文件到目标目录
func (s *EnterpriseSpaceService) MoveFile(req *request.MoveEnterpriseFileRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, req.FileID)
	if err != nil || sharedFile == nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	targetPathID := req.TargetPath

	// 验证目标目录存在（0=根目录）
	if targetPathID != 0 {
		targetDir, err := s.factory.EnterpriseSharedPath().GetByID(ctx, targetPathID)
		if err != nil || targetDir == nil || targetDir.EnterpriseID != sharedFile.EnterpriseID {
			return models.NewJsonResponse(400, "目标目录不存在", nil), err
		}
	}

	// 检查目标目录是否有同名文件
	existing, _ := s.factory.EnterpriseSharedFile().GetByPathIDAndName(ctx, sharedFile.EnterpriseID, targetPathID, sharedFile.FileName)
	if existing != nil && existing.ID != req.FileID {
		return models.NewJsonResponse(409, "目标目录已存在同名文件", nil), nil
	}

	sharedFile.PathID = targetPathID
	sharedFile.UpdatedBy = userID
	sharedFile.UpdatedAt = custom_type.Now()
	if err := s.factory.EnterpriseSharedFile().Update(ctx, sharedFile); err != nil {
		return models.NewJsonResponse(500, "移动失败", nil), err
	}

	return models.NewJsonResponse(200, "移动成功", nil), nil
}

// PackageCreate 创建打包下载任务
func (s *EnterpriseSpaceService) PackageCreate(req *request.EnterprisePackageCreateRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, req.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	for _, fileID := range req.FileIDs {
		sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, fileID)
		if err != nil || sharedFile == nil {
			return models.NewJsonResponse(404, fmt.Sprintf("文件 %s 不存在", fileID), nil), err
		}
		if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
			return models.NewJsonResponse(403, err.Error(), nil), err
		}
	}

	packageID := uuid.New().String()
	packageName := req.PackageName
	if packageName == "" {
		packageName = fmt.Sprintf("enterprise_files_%s.zip", packageID[:8])
	}
	if filepath.Ext(packageName) == "" {
		packageName += ".zip"
	}

	// 缓存打包任务信息
	taskData := map[string]interface{}{
		"package_id":   packageID,
		"file_ids":     req.FileIDs,
		"package_name": packageName,
		"status":       "creating",
		"progress":     0,
	}
	taskJSON, _ := json.Marshal(taskData)
	s.cacheLocal.Set(fmt.Sprintf("pkg:%s", packageID), string(taskJSON), 3600)

	// 异步创建 ZIP
	go s.buildPackage(packageID, req.FileIDs, packageName)

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"package_id":   packageID,
		"package_name": packageName,
		"status":       "creating",
	}), nil
}

func (s *EnterpriseSpaceService) buildPackage(packageID string, fileIDs []string, packageName string) {
	tmpDir := filepath.Join(os.TempDir(), "enterprise_packages")
	os.MkdirAll(tmpDir, 0755)
	zipPath := filepath.Join(tmpDir, packageID+".zip")

	zipFile, err := os.Create(zipPath)
	if err != nil {
		s.updatePackageStatus(packageID, "failed", 0)
		return
	}
	defer zipFile.Close()

	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	total := len(fileIDs)
	for i, fileID := range fileIDs {
		sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(context.Background(), fileID)
		if err != nil || sharedFile == nil {
			continue
		}
		fileInfo, err := s.factory.FileInfo().GetByID(context.Background(), sharedFile.FileID)
		if err != nil || fileInfo == nil {
			continue
		}

		srcFile, err := os.Open(fileInfo.Path)
		if err != nil {
			continue
		}

		entry, err := writer.Create(sharedFile.FileName)
		if err != nil {
			srcFile.Close()
			continue
		}
		io.Copy(entry, srcFile)
		srcFile.Close()

		progress := int(float64(i+1) / float64(total) * 100)
		s.updatePackageStatus(packageID, "creating", progress)
	}

	s.updatePackageStatus(packageID, "ready", 100)
}

func (s *EnterpriseSpaceService) updatePackageStatus(packageID, status string, progress int) {
	taskData := map[string]interface{}{
		"package_id": packageID,
		"status":     status,
		"progress":   progress,
	}
	taskJSON, _ := json.Marshal(taskData)
	s.cacheLocal.Set(fmt.Sprintf("pkg:%s", packageID), string(taskJSON), 3600)
}

// PackageProgress 查询打包进度
func (s *EnterpriseSpaceService) PackageProgress(packageID string) (*models.JsonResponse, error) {
	data, err := s.cacheLocal.Get(fmt.Sprintf("pkg:%s", packageID))
	if err != nil || data == nil {
		return models.NewJsonResponse(404, "任务不存在或已过期", nil), nil
	}

	var taskInfo map[string]interface{}
	switch v := data.(type) {
	case string:
		json.Unmarshal([]byte(v), &taskInfo)
	case map[string]interface{}:
		taskInfo = v
	}

	return models.NewJsonResponse(200, "ok", taskInfo), nil
}

// GetPackageFile 获取打包文件路径
func (s *EnterpriseSpaceService) GetPackageFile(packageID string) (string, string, error) {
	data, err := s.cacheLocal.Get(fmt.Sprintf("pkg:%s", packageID))
	if err != nil || data == nil {
		return "", "", fmt.Errorf("任务不存在")
	}

	var taskInfo map[string]interface{}
	switch v := data.(type) {
	case string:
		json.Unmarshal([]byte(v), &taskInfo)
	case map[string]interface{}:
		taskInfo = v
	}

	status, _ := taskInfo["status"].(string)
	if status != "ready" {
		return "", "", fmt.Errorf("打包未完成")
	}

	zipPath := filepath.Join(os.TempDir(), "enterprise_packages", packageID+".zip")
	packageName, _ := taskInfo["package_name"].(string)
	if packageName == "" {
		packageName = "enterprise_files.zip"
	}

	return zipPath, packageName, nil
}

// ExtractCheck 检测解压冲突
func (s *EnterpriseSpaceService) ExtractCheck(req *request.EnterpriseExtractCheckRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, req.FileID)
	if err != nil || sharedFile == nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	// 获取文件信息
	fileInfo, err := s.factory.FileInfo().GetByID(ctx, sharedFile.FileID)
	if err != nil || fileInfo == nil {
		return models.NewJsonResponse(404, "文件信息不存在", nil), err
	}

	// 检查是否是 ZIP 压缩包（当前仅支持 ZIP）
	ext := strings.ToLower(filepath.Ext(sharedFile.FileName))
	if ext != ".zip" {
		return models.NewJsonResponse(400, "当前仅支持 ZIP 格式解压", nil), nil
	}

	// 打开 ZIP 读取条目名称，检查目标目录下的实际冲突
	targetPathID := req.TargetPathID
	reader, err := zip.OpenReader(fileInfo.Path)
	if err != nil {
		return models.NewJsonResponse(500, "打开压缩包失败", nil), err
	}
	defer reader.Close()

	var conflicts []string
	for _, f := range reader.File {
		if f.FileInfo().IsDir() {
			continue
		}
		cleanName := filepath.ToSlash(f.Name)
		cleanName = strings.TrimPrefix(cleanName, "/")
		cleanName = strings.ReplaceAll(cleanName, "../", "")
		cleanName = strings.ReplaceAll(cleanName, "..\\", "")
		if cleanName == "" {
			continue
		}
		existing, _ := s.factory.EnterpriseSharedFile().GetByPathIDAndName(ctx, sharedFile.EnterpriseID, targetPathID, cleanName)
		if existing != nil {
			conflicts = append(conflicts, cleanName)
		}
	}

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"has_conflict":   len(conflicts) > 0,
		"conflict_files": conflicts,
		"file_name":      sharedFile.FileName,
		"file_size":      sharedFile.Size,
		"total_files":    len(reader.File),
	}), nil
}

// ExtractStart 开始解压
func (s *EnterpriseSpaceService) ExtractStart(req *request.ExtractStartRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, req.FileID)
	if err != nil || sharedFile == nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	fileInfo, err := s.factory.FileInfo().GetByID(ctx, sharedFile.FileID)
	if err != nil || fileInfo == nil {
		return models.NewJsonResponse(404, "文件信息不存在", nil), err
	}

	taskID := uuid.New().String()

	// 缓存任务信息
	taskData := map[string]interface{}{
		"task_id":    taskID,
		"file_id":    req.FileID,
		"status":     "extracting",
		"progress":   0,
		"total":      0,
		"completed":  0,
		"current":    "",
	}
	taskJSON, _ := json.Marshal(taskData)
	s.cacheLocal.Set(fmt.Sprintf("extract:%s", taskID), string(taskJSON), 3600)

	// 异步解压
	go s.doExtract(taskID, fileInfo.Path, sharedFile.EnterpriseID, req.TargetPathID, userID, req.ConflictStrategy)

	return models.NewJsonResponse(200, "ok", map[string]interface{}{
		"task_id": taskID,
		"status":  "extracting",
	}), nil
}

func (s *EnterpriseSpaceService) doExtract(taskID, filePath, enterpriseID string, targetPathID int, userID, conflictStrategy string) {
	// 简化解压实现：仅支持 ZIP
	ext := strings.ToLower(filepath.Ext(filePath))
	if ext != ".zip" {
		s.updateExtractStatus(taskID, "failed", 0, 0, 0, "仅支持 ZIP 格式")
		return
	}

	reader, err := zip.OpenReader(filePath)
	if err != nil {
		s.updateExtractStatus(taskID, "failed", 0, 0, 0, "打开压缩包失败")
		return
	}
	defer reader.Close()

	// 查询企业信息（循环外只查一次）
	enterprise, _ := s.factory.Enterprise().GetByID(context.Background(), enterpriseID)

	total := len(reader.File)
	completed := 0

	for _, f := range reader.File {
		// 路径穿越防护：清理条目名称
		cleanName := filepath.ToSlash(f.Name)
		cleanName = strings.TrimPrefix(cleanName, "/")
		cleanName = strings.TrimPrefix(cleanName, "../")
		cleanName = strings.ReplaceAll(cleanName, "../", "")
		cleanName = strings.ReplaceAll(cleanName, "..\\", "")
		if cleanName == "" || cleanName == "." {
			completed++
			continue
		}

		if f.FileInfo().IsDir() {
			completed++
			continue
		}

		// 检查冲突
		existing, _ := s.factory.EnterpriseSharedFile().GetByPathIDAndName(
			context.Background(), enterpriseID, targetPathID, cleanName)

		if existing != nil {
			switch conflictStrategy {
			case "skip":
				completed++
				continue
			case "overwrite":
				if enterprise != nil && !enterprise.SpaceUnlimited {
					s.updateEnterpriseSpace(enterpriseID, -existing.Size)
				}
				oldPhysicalID := existing.FileID
				s.factory.EnterpriseSharedFile().Delete(context.Background(), existing.ID)
				_ = s.cleanupPhysicalFileIfUnreferenced(context.Background(), s.factory, oldPhysicalID)
			case "keep_both":
				// 保持原名，会在后面处理
			default:
				completed++
				continue
			}
		}

		// 流式写入磁盘（避免大文件 OOM）
		rc, err := f.Open()
		if err != nil {
			completed++
			continue
		}

		// 先写入临时文件，同时计算哈希
		disks, diskErr := s.factory.Disk().List(context.Background(), 0, 1000)
		if diskErr != nil || len(disks) == 0 {
			rc.Close()
			completed++
			continue
		}
		bestDisk := disks[0]

		// 用 UUID 作为临时文件名，计算完哈希后再确定最终路径
		tmpPath := filepath.Join(bestDisk.DataPath, "enterprise", enterpriseID, ".tmp", uuid.New().String())
		os.MkdirAll(filepath.Dir(tmpPath), 0755)

		h := sha256.New()
		writer, createErr := os.Create(tmpPath)
		if createErr != nil {
			rc.Close()
			completed++
			continue
		}

		fileSize, copyErr := io.Copy(writer, io.TeeReader(rc, h))
		writer.Close()
		rc.Close()
		if copyErr != nil {
			os.Remove(tmpPath)
			completed++
			continue
		}

		hashHex := hex.EncodeToString(h.Sum(nil))

		// 空间配额检查
		if enterprise != nil && !enterprise.SpaceUnlimited {
			if enterprise.FreeSpace < fileSize {
				os.Remove(tmpPath)
				s.updateExtractStatus(taskID, "failed", 0, total, completed, "企业空间不足")
				return
			}
		}

		// 检查是否已存在相同文件（去重）
		var fileRecord *models.FileInfo
		existingFile, _ := s.factory.FileInfo().GetByHash(context.Background(), hashHex)
		if existingFile != nil {
			fileRecord = existingFile
			// 已存在相同文件，删除临时文件
			os.Remove(tmpPath)
		} else {
			// 移动临时文件到最终路径
			storagePath := filepath.Join(bestDisk.DataPath, "enterprise", enterpriseID, hashHex[:2], hashHex[2:4], hashHex)
			os.MkdirAll(filepath.Dir(storagePath), 0755)
			if err := os.Rename(tmpPath, storagePath); err != nil {
				// Rename 失败则尝试复制
				if err := os.WriteFile(storagePath, []byte{}, 0644); err != nil {
					os.Remove(tmpPath)
					completed++
					continue
				}
				// 重新从 zip 写入
				os.Remove(tmpPath)
				rc2, _ := f.Open()
				if rc2 == nil {
					completed++
					continue
				}
				dst, _ := os.Create(storagePath)
				if dst == nil {
					rc2.Close()
					completed++
					continue
				}
				io.Copy(dst, rc2)
				dst.Close()
				rc2.Close()
			}

			// 检测 MIME 类型
			mimeFile, _ := os.Open(storagePath)
			var mimeStr string
			if mimeFile != nil {
				buf := make([]byte, 512)
				n, _ := mimeFile.Read(buf)
				mimeFile.Close()
				mime := mimetype.Detect(buf[:n])
				mimeStr = mime.String()
			}

			fileRecord = &models.FileInfo{
				ID:          uuid.New().String(),
				Name:        cleanName,
				RandomName:  util.GenerateUniqueFilename(),
				Size:        int(fileSize),
				Mime:        mimeStr,
				Path:        storagePath,
				FileHash:    hashHex,
				HasFullHash: true,
				CreatedAt:   custom_type.Now(),
				UpdatedAt:   custom_type.Now(),
			}
			s.factory.FileInfo().Create(context.Background(), fileRecord)
		}

		// 创建企业文件关联
		fileName := cleanName
		if conflictStrategy == "keep_both" && existing != nil {
			ext := filepath.Ext(cleanName)
			base := strings.TrimSuffix(cleanName, ext)
			fileName = fmt.Sprintf("%s_%s%s", base, uuid.New().String()[:8], ext)
		}

		sharedFile := &models.EnterpriseSharedFile{
			ID:           uuid.New().String(),
			EnterpriseID: enterpriseID,
			FileID:       fileRecord.ID,
			FileName:     fileName,
			PathID:       targetPathID,
			UploaderID:   userID,
			Size:         int64(fileRecord.Size),
			CreatedAt:    custom_type.Now(),
		}
		s.factory.EnterpriseSharedFile().Create(context.Background(), sharedFile)

		// 扣除空间配额
		if enterprise != nil && !enterprise.SpaceUnlimited {
			s.updateEnterpriseSpace(enterpriseID, int64(fileRecord.Size))
		}

		completed++
		progress := int(float64(completed) / float64(total) * 100)
		s.updateExtractStatus(taskID, "extracting", progress, total, completed, cleanName)
	}

	s.updateExtractStatus(taskID, "done", 100, total, total, "")
}

func (s *EnterpriseSpaceService) updateExtractStatus(taskID, status string, progress, total, completed int, current string) {
	taskData := map[string]interface{}{
		"task_id":   taskID,
		"status":    status,
		"progress":  progress,
		"total":     total,
		"completed": completed,
		"current":   current,
	}
	taskJSON, _ := json.Marshal(taskData)
	s.cacheLocal.Set(fmt.Sprintf("extract:%s", taskID), string(taskJSON), 3600)
}

// ExtractProgress 查询解压进度
func (s *EnterpriseSpaceService) ExtractProgress(taskID string) (*models.JsonResponse, error) {
	data, err := s.cacheLocal.Get(fmt.Sprintf("extract:%s", taskID))
	if err != nil || data == nil {
		return models.NewJsonResponse(404, "任务不存在或已过期", nil), nil
	}

	var taskInfo map[string]interface{}
	switch v := data.(type) {
	case string:
		json.Unmarshal([]byte(v), &taskInfo)
	case map[string]interface{}:
		taskInfo = v
	}

	return models.NewJsonResponse(200, "ok", taskInfo), nil
}

// CreateShare 创建企业文件分享链接（写入 shares 表）
func (s *EnterpriseSpaceService) CreateShare(req *request.CreateEnterpriseShareRequest, contextEnterpriseID, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	if err := assertEnterpriseContext(contextEnterpriseID, req.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	sharedFile, err := s.factory.EnterpriseSharedFile().GetByID(ctx, req.FileID)
	if err != nil || sharedFile == nil {
		return models.NewJsonResponse(404, "文件不存在", nil), err
	}
	if err := assertEnterpriseContext(contextEnterpriseID, sharedFile.EnterpriseID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}
	if err := s.verifyActiveMember(ctx, sharedFile.EnterpriseID, userID); err != nil {
		return models.NewJsonResponse(403, err.Error(), nil), err
	}

	fileInfo, err := s.factory.FileInfo().GetByID(ctx, sharedFile.FileID)
	if err != nil || fileInfo == nil {
		return models.NewJsonResponse(404, "物理文件不存在", nil), err
	}

	token := fmt.Sprintf("%s-%v", uuid.New().String(), util.TimeUtil{}.GetTimestamp())
	var passwordHash string
	if req.Password != "" {
		passwordHash, err = util.GeneratePassword(req.Password)
		if err != nil {
			return models.NewJsonResponse(500, "生成分享密码失败", nil), err
		}
	}

	now := custom_type.Now()
	expireAt := now.Add(30 * 24 * time.Hour)
	if req.ExpireDays > 0 {
		expireAt = now.Add(time.Duration(req.ExpireDays) * 24 * time.Hour)
	} else if req.ExpireDays == 0 {
		expireAt = now.Add(100 * 365 * 24 * time.Hour)
	}

	share := &models.Share{
		UserID:        userID,
		FileID:        fileInfo.ID,
		Token:         token,
		ExpiresAt:     expireAt,
		PasswordHash:  passwordHash,
		DownloadCount: 0,
		CreatedAt:     custom_type.Now(),
	}
	if err := s.factory.Share().Create(ctx, share); err != nil {
		return models.NewJsonResponse(500, "创建分享失败", nil), err
	}

	return models.NewJsonResponse(200, "ok", fmt.Sprintf("/api/share/download/%s", token)), nil
}

// updateEnterpriseSpace 更新企业空间用量
func (s *EnterpriseSpaceService) updateEnterpriseSpace(enterpriseID string, sizeDelta int64) {
	ctx := context.Background()
	// 原子更新，避免并发读写竞争
	s.factory.DB().WithContext(ctx).
		Model(&models.Enterprise{}).
		Where("id = ? AND space_unlimited = ?", enterpriseID, false).
		Update("free_space", gorm.Expr("free_space - ?", sizeDelta))
}
