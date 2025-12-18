package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"myobj/src/pkg/upload"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/goccy/go-json"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// 全局上传锁，用于防止同一文件的并发处理
var uploadLocks sync.Map     // key: userID+fileName, value: *sync.Mutex
var processingFiles sync.Map // key: userID+fileName, value: bool (标记文件是否正在处理)

// FileService 文件服务
type FileService struct {
	factory    *impl.RepositoryFactory
	cacheLocal cache.Cache
}

func NewFileService(factory *impl.RepositoryFactory, cacheLocal cache.Cache) *FileService {
	return &FileService{
		factory:    factory,
		cacheLocal: cacheLocal,
	}
}
func (f *FileService) GetRepository() *impl.RepositoryFactory {
	return f.factory
}

// Precheck 文件预检查
func (f *FileService) Precheck(req *request.UploadPrecheckRequest, c cache.Cache) (*models.JsonResponse, error) {
	ctx := context.Background()
	user, err := f.factory.User().GetByID(ctx, req.UserID)
	if err != nil {
		logger.LOG.Error("获取用户信息失败", "error", err, "userID", req.UserID)
		return nil, err
	}
	// 检查用户可用空间 如果不是无限空间，且可用空间不足
	if user.Space > 0 && user.FreeSpace < req.FileSize {
		return models.NewJsonResponse(400, "用户可用空间不足", nil), nil
	}
	signature, err := f.factory.FileInfo().GetByChunkSignature(ctx, req.ChunkSignature, req.FileSize)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("查询文件签名失败", "error", err, "chunkSignature", req.ChunkSignature)
		return nil, err
	}
	if len(req.FilesMd5) >= 3 {
		if signature.FirstChunkHash == req.FilesMd5[0] && signature.SecondChunkHash == req.FilesMd5[1] && signature.ThirdChunkHash == req.FilesMd5[2] {
			userFile := &models.UserFiles{
				UserID:      user.ID,
				FileID:      signature.ID,
				FileName:    req.FileName,
				VirtualPath: req.PathID,
				Public:      false,
				CreatedAt:   custom_type.Now(),
				UfID:        uuid.NewString(),
			}
			err := f.factory.UserFiles().Create(context.Background(), userFile)
			if err != nil {
				logger.LOG.Error("创建用户文件失败", "error", err, "userID", req.UserID, "fileID", signature.ID, "fileName", req.FileName)
				return nil, err
			}
			return models.NewJsonResponse(200, "秒传成功", nil), nil
		}
	} else {
		if signature.FileHash == req.FilesMd5[0] {
			userFile := &models.UserFiles{
				UserID:      user.ID,
				FileID:      signature.ID,
				FileName:    req.FileName,
				VirtualPath: req.PathID,
				Public:      false,
				CreatedAt:   custom_type.Now(),
				UfID:        uuid.NewString(),
			}
			err := f.factory.UserFiles().Create(context.Background(), userFile)
			if err != nil {
				logger.LOG.Error("创建用户文件失败", "error", err, "userID", req.UserID, "fileID", signature.ID, "fileName", req.FileName)
				return nil, err
			}
			return models.NewJsonResponse(200, "秒传成功", nil), nil
		}
	}
	//无法触发秒传，但可上传，返回校验ID
	key := fmt.Sprintf("fileUpload:%s", user.ID)
	uid := uuid.New().String()
	res := new(response.FilePrecheckResponse)
	res.PrecheckID = uid
	chunks, err := f.factory.UploadChunk().GetByUserIDAndFileName(ctx, user.ID, req.FileName)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("获取文件分片失败", "error", err, "chunkSignature", req.ChunkSignature)
		return nil, err
	}
	// chunks 是数组，需要遍历
	for _, chunk := range chunks {
		res.Md5 = append(res.Md5, chunk.Md5)
	}
	// 序列化为JSON字符串存储到Redis
	resJSON, err := json.Marshal(res)
	if err != nil {
		logger.LOG.Error("序列化预检响应失败", "error", err)
		return nil, err
	}
	err = c.Set(key, string(resJSON), 12*60*60) // 12小时内可用的校验
	if err != nil {
		logger.LOG.Error("缓存设置失败", "error", err, "key", key)
		return nil, err
	}
	// 保存原始请求数据到缓存，供上传时使用
	reqKey := fmt.Sprintf("fileUploadReq:%s", user.ID)
	reqJSON, err := json.Marshal(req)
	if err != nil {
		logger.LOG.Error("序列化预检请求失败", "error", err)
		return nil, err
	}
	if err := c.Set(reqKey, string(reqJSON), 12*60*60); err != nil {
		logger.LOG.Error("保存上传请求失败", "error", err, "key", reqKey)
		return nil, err
	}
	return models.NewJsonResponse(201, "预检通过", uid), nil
}

// SearchUserFiles 搜索当前用户的文件
func (f *FileService) SearchUserFiles(req *request.FileSearchRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// 搜索用户文件
	userFiles, err := f.factory.UserFiles().SearchUserFiles(ctx, userID, req.Keyword, offset, pageSize)
	if err != nil {
		logger.LOG.Error("搜索用户文件失败", "error", err, "userID", userID, "keyword", req.Keyword)
		return nil, err
	}
	// 获取文件详情
	fileIDs := make([]string, 0, len(userFiles))
	for _, uf := range userFiles {
		fileIDs = append(fileIDs, uf.FileID)
	}

	files := make([]*models.FileInfo, 0, len(fileIDs))
	for _, fileID := range fileIDs {
		file, err := f.factory.FileInfo().GetByID(ctx, fileID)
		if err != nil {
			continue
		}
		files = append(files, file)
	}

	// 统计总数
	total, err := f.factory.UserFiles().CountUserFilesByKeyword(ctx, userID, req.Keyword)
	if err != nil {
		logger.LOG.Error("统计用户文件数量失败", "error", err, "userID", userID, "keyword", req.Keyword)
		return nil, err
	}

	result := map[string]interface{}{
		"files": files,
		"total": total,
	}
	return models.NewJsonResponse(200, "搜索成功", result), nil
}

// SearchPublicFiles 搜索公开文件（广场）
func (f *FileService) SearchPublicFiles(req *request.FileSearchRequest) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 默认分页参数
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	var userFiles []*models.UserFiles
	var total int64
	var err error

	if req.Keyword != "" {
		// 根据关键词搜索
		userFiles, err = f.factory.UserFiles().SearchPublicFiles(ctx, req.Keyword, offset, pageSize)
		if err != nil {
			logger.LOG.Error("搜索公开文件失败", "error", err, "keyword", req.Keyword)
			return nil, err
		}
		total, err = f.factory.UserFiles().CountPublicFilesByKeyword(ctx, req.Keyword)
	} else {
		// 获取所有公开文件
		userFiles, err = f.factory.UserFiles().ListPublicFiles(ctx, offset, pageSize)
		if err != nil {
			logger.LOG.Error("获取公开文件列表失败", "error", err)
			return nil, err
		}
		total, err = f.factory.UserFiles().CountPublicFiles(ctx)
	}

	if err != nil {
		logger.LOG.Error("统计公开文件数量失败", "error", err)
		return nil, err
	}

	// 获取文件详情和用户信息
	type FileWithOwner struct {
		*models.FileInfo
		OwnerName string `json:"owner_name"`
	}

	resultFiles := make([]*FileWithOwner, 0, len(userFiles))
	for _, uf := range userFiles {
		file, err := f.factory.FileInfo().GetByID(ctx, uf.FileID)
		if err != nil {
			continue
		}

		// 获取用户名
		user, err := f.factory.User().GetByID(ctx, uf.UserID)
		ownerName := "Unknown"
		if err == nil && user != nil {
			ownerName = user.UserName
		}

		resultFiles = append(resultFiles, &FileWithOwner{
			FileInfo:  file,
			OwnerName: ownerName,
		})
	}

	result := map[string]interface{}{
		"files": resultFiles,
		"total": total,
	}
	return models.NewJsonResponse(200, "搜索成功", result), nil
}

// GetFileList 获取文件列表（我的文件页面）
func (f *FileService) GetFileList(req *request.FileListRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 处理虚拟路径ID，空或为0时使用根目录
	var currentPathID int
	var currentPath *models.VirtualPath
	var err error

	if req.VirtualPath == "" || req.VirtualPath == "0" {
		// 查询用户根目录
		currentPath, err = f.factory.VirtualPath().GetRootPath(ctx, userID)
		if err != nil {
			logger.LOG.Error("获取根目录失败", "error", err, "userID", userID)
			return nil, fmt.Errorf("获取根目录失败: %w", err)
		}
		currentPathID = currentPath.ID
	} else {
		// 解析虚拟路径ID
		pathID := 0
		_, err := fmt.Sscanf(req.VirtualPath, "%d", &pathID)
		if err != nil {
			logger.LOG.Error("解析虚拟路径ID失败", "error", err, "virtualPath", req.VirtualPath)
			return nil, fmt.Errorf("无效的路径ID: %w", err)
		}
		currentPathID = pathID
		// 查询当前路径信息
		currentPath, err = f.factory.VirtualPath().GetByID(ctx, currentPathID)
		if err != nil {
			logger.LOG.Error("查询路径信息失败", "error", err, "pathID", currentPathID)
			return nil, fmt.Errorf("路径不存在: %w", err)
		}
	}

	// 查询总数（子目录 + 文件）
	folderCount, err := f.factory.VirtualPath().CountSubFoldersByParentID(ctx, userID, currentPathID)
	if err != nil {
		logger.LOG.Error("统计子目录数量失败", "error", err, "userID", userID, "pathID", currentPathID)
		return nil, err
	}
	// 文件表中virtual_path字段存的是路径ID（字符串格式）
	virtualPathIDStr := fmt.Sprintf("%d", currentPathID)
	fileCount, err := f.factory.FileInfo().CountByVirtualPath(ctx, userID, virtualPathIDStr)
	if err != nil {
		logger.LOG.Error("统计文件数量失败", "error", err, "userID", userID, "virtualPath", virtualPathIDStr)
		return nil, err
	}
	totalCount := folderCount + fileCount

	// 计算分页偏移量
	offset := (req.Page - 1) * req.PageSize

	// 优先返回文件夹
	var folders []*models.VirtualPath
	var files []*models.FileInfo

	if offset < int(folderCount) {
		// 当前页包含文件夹
		folderLimit := req.PageSize
		if offset+req.PageSize > int(folderCount) {
			folderLimit = int(folderCount) - offset
		}

		folders, err = f.factory.VirtualPath().ListSubFoldersByParentID(ctx, userID, currentPathID, offset, folderLimit)
		if err != nil {
			logger.LOG.Error("查询子目录列表失败", "error", err, "userID", userID, "pathID", currentPathID)
			return nil, err
		}

		// 如果还有剩余空间，查询文件
		remaining := req.PageSize - len(folders)
		if remaining > 0 {
			files, err = f.factory.FileInfo().ListByVirtualPath(ctx, userID, virtualPathIDStr, 0, remaining)
			if err != nil {
				logger.LOG.Error("查询文件列表失败", "error", err, "userID", userID, "virtualPath", virtualPathIDStr)
				return nil, err
			}
		}
	} else {
		// 当前页只包含文件
		fileOffset := offset - int(folderCount)
		files, err = f.factory.FileInfo().ListByVirtualPath(ctx, userID, virtualPathIDStr, fileOffset, req.PageSize)
		if err != nil {
			logger.LOG.Error("查询文件列表失败", "error", err, "userID", userID, "virtualPath", virtualPathIDStr)
			return nil, err
		}
	}

	// 获取面包屑导航（只展示当前、上级、上上级）
	breadcrumbs, err := f.buildBreadcrumbs(ctx, currentPath)
	if err != nil {
		logger.LOG.Error("构建面包屑导航失败", "error", err, "pathID", currentPath.ID)
		return nil, err
	}

	// 构造响应
	resp := &response.FileListResponse{
		Breadcrumbs: breadcrumbs,
		CurrentPath: fmt.Sprintf("%d", currentPathID),
		Folders:     make([]*response.FolderItem, 0, len(folders)),
		Files:       make([]*response.FileItem, 0, len(files)),
		Total:       totalCount,
		Page:        req.Page,
		PageSize:    req.PageSize,
	}

	// 转换文件夹数据
	for _, folder := range folders {
		resp.Folders = append(resp.Folders, &response.FolderItem{
			ID:          folder.ID,
			Name:        folder.Path,
			Path:        fmt.Sprintf("%d", folder.ID),
			CreatedTime: folder.CreatedTime,
		})
	}

	// 转换文件数据
	for _, file := range files {
		uf, _ := f.factory.UserFiles().GetByUserIDAndFileID(ctx, userID, file.ID)
		resp.Files = append(resp.Files, &response.FileItem{
			FileID:       file.ID,
			FileName:     file.Name,
			FileSize:     file.Size,
			MimeType:     file.Mime,
			IsEnc:        file.IsEnc,
			HasThumbnail: file.ThumbnailImg != "",
			CreatedAt:    file.CreatedAt,
			UfID:         uf.UfID,
		})
	}

	return models.NewJsonResponse(200, "获取成功", resp), nil
}

// buildBreadcrumbs 构建面包屑导航（只展示当前、上级、上上级）
func (f *FileService) buildBreadcrumbs(ctx context.Context, currentPath *models.VirtualPath) ([]response.Breadcrumb, error) {
	breadcrumbs := []response.Breadcrumb{}

	// 添加当前目录
	breadcrumbs = append(breadcrumbs, response.Breadcrumb{
		ID:   currentPath.ID,
		Name: currentPath.Path,
		Path: fmt.Sprintf("%d", currentPath.ID),
	})

	// 获取上级目录（如果存在）
	if currentPath.ParentLevel != "" {
		parentID := 0
		_, err := fmt.Sscanf(currentPath.ParentLevel, "%d", &parentID)
		if err == nil && parentID > 0 {
			parent, err := f.factory.VirtualPath().GetByID(ctx, parentID)
			if err == nil {
				// 在开头插入上级目录
				breadcrumbs = append([]response.Breadcrumb{{
					ID:   parent.ID,
					Name: parent.Path,
					Path: fmt.Sprintf("%d", parent.ID),
				}}, breadcrumbs...)

				// 获取上上级目录（如果存在）
				if parent.ParentLevel != "" {
					grandParentID := 0
					_, err := fmt.Sscanf(parent.ParentLevel, "%d", &grandParentID)
					if err == nil && grandParentID > 0 {
						grandParent, err := f.factory.VirtualPath().GetByID(ctx, grandParentID)
						if err == nil {
							// 在开头插入上上级目录
							breadcrumbs = append([]response.Breadcrumb{{
								ID:   grandParent.ID,
								Name: grandParent.Path,
								Path: fmt.Sprintf("%d", grandParent.ID),
							}}, breadcrumbs...)
						}
					}
				}
			}
		}
	}

	return breadcrumbs, nil
}

// MakeDir 创建目录
func (f *FileService) MakeDir(req *request.MakeDirRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	path, err := f.factory.VirtualPath().GetByPath(ctx, userID, req.DirPath)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.LOG.Error("获取目录失败", "error", err)
		return nil, err
	}
	if path != nil {
		logger.LOG.Error("目录已存在", "path", req.DirPath)
		return models.NewJsonResponse(400, "目录已存在", nil), nil
	}
	//转换为int
	parentLevel, err := strconv.Atoi(req.ParentLevel)
	if err != nil {
		logger.LOG.Error("参数错误", "error", err)
		return nil, err
	}
	virtualPath := &models.VirtualPath{
		UserID:      userID,
		Path:        req.DirPath,
		CreatedTime: custom_type.Now(),
		UpdateTime:  custom_type.Now(),
	}
	if parentLevel > 0 {
		vp, err := f.factory.VirtualPath().GetByID(ctx, parentLevel)
		if err != nil {
			logger.LOG.Error("获取上级目录失败", "error", err)
			return nil, err
		}
		virtualPath.ParentLevel = fmt.Sprintf("%d", vp.ID)
	}
	err = f.factory.VirtualPath().Create(ctx, virtualPath)
	if err != nil {
		logger.LOG.Error("创建目录失败", "error", err)
		return nil, err
	}
	return models.NewJsonResponse(200, "创建目录成功", nil), nil
}

// MoveFile 移动文件
func (f *FileService) MoveFile(req *request.MoveFileRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()
	userFile, err := f.factory.UserFiles().GetByUserIDAndFileID(ctx, userID, req.FileID)
	if err != nil {
		logger.LOG.Error("获取文件失败", "error", err)
		return nil, err
	}
	userFile.UserID = userID
	userFile.VirtualPath = req.TargetPath
	err = f.factory.UserFiles().Update(ctx, userFile)
	if err != nil {
		logger.LOG.Error("移动文件失败", "error", err)
		return nil, err
	}
	return models.NewJsonResponse(200, "移动文件成功", nil), nil
}

// GetVirtualPath 获取虚拟路径
func (f *FileService) GetVirtualPath(userID string) (*models.JsonResponse, error) {
	user, err := f.factory.VirtualPath().GetPathByUser(context.Background(), userID)
	if err != nil {
		logger.LOG.Error("获取虚拟路径失败", "error", err)
		return nil, err
	}
	return models.NewJsonResponse(200, "获取虚拟路径成功", user), nil
}

// DeleteFiles 删除文件（移动到回收站）
func (f *FileService) DeleteFiles(req *request.DeleteFileRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	successCount := 0
	failedCount := 0
	var errors []string

	for _, fileID := range req.FileIDs {
		// 验证用户是否拥有该文件
		userFile, err := f.factory.UserFiles().GetByUserIDAndFileID(ctx, userID, fileID)
		if err != nil {
			logger.LOG.Warn("用户不拥有该文件", "userID", userID, "fileID", fileID)
			errors = append(errors, fmt.Sprintf("文件 %s 不存在或无权访问", fileID))
			failedCount++
			continue
		}

		// 检查是否已在回收站
		_, err = f.factory.Recycled().GetByUserIDAndFileID(ctx, userID, fileID)
		if err == nil {
			logger.LOG.Warn("文件已在回收站", "fileID", fileID)
			errors = append(errors, fmt.Sprintf("文件 %s 已在回收站中", fileID))
			failedCount++
			continue
		}

		// 在事务中执行：1. 软删除 user_files、 2. 创建回收站记录
		err = f.factory.DB().Transaction(func(tx *gorm.DB) error {
			txFactory := f.factory.WithTx(tx)

			// 软删除 user_files 记录
			if err := tx.Where("user_id = ? AND file_id = ?", userID, fileID).Delete(&models.UserFiles{}).Error; err != nil {
				return fmt.Errorf("软删除用户文件失败: %w", err)
			}

			// 创建回收站记录
			recycled := &models.Recycled{
				ID:        uuid.Must(uuid.NewV7()).String(),
				FileID:    fileID,
				UserID:    userID,
				CreatedAt: custom_type.Now(),
			}

			if err := txFactory.Recycled().Create(ctx, recycled); err != nil {
				return fmt.Errorf("创建回收站记录失败: %w", err)
			}

			return nil
		})

		if err != nil {
			logger.LOG.Error("删除文件失败", "error", err, "fileID", fileID, "userID", userID)
			errors = append(errors, fmt.Sprintf("删除文件 %s 失败: %v", fileID, err))
			failedCount++
			continue
		}

		successCount++
		logger.LOG.Info("文件已移动到回收站", "fileID", fileID, "userID", userID, "fileName", userFile.FileName)
	}

	message := fmt.Sprintf("成功删除 %d 个文件", successCount)
	if failedCount > 0 {
		message = fmt.Sprintf("%s，失败 %d 个", message, failedCount)
	}

	result := map[string]interface{}{
		"success": successCount,
		"failed":  failedCount,
	}
	if len(errors) > 0 {
		result["errors"] = errors
	}

	return models.NewJsonResponse(200, message, result), nil
}

// UploadFile 文件上传处理
func (f *FileService) UploadFile(req *request.FileUploadRequest, file multipart.File, header *multipart.FileHeader, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 1. 从缓存获取预检信息
	cacheKey := fmt.Sprintf("fileUpload:%s", userID)
	precheckData, err := f.cacheLocal.Get(cacheKey)
	if err != nil {
		logger.LOG.Error("获取预检信息失败", "error", err, "precheckID", req.PrecheckID)
		return nil, fmt.Errorf("预检信息已过期或不存在")
	}

	// 2. 反序列化预检响应数据
	var precheckResp response.FilePrecheckResponse
	// 统一处理：Redis和LocalCache都可能返回string或对象
	switch v := precheckData.(type) {
	case *response.FilePrecheckResponse:
		// LocalCache直接返回对象指针
		precheckResp = *v
	case string:
		// Redis返回JSON字符串，需要反序列化
		if err := json.Unmarshal([]byte(v), &precheckResp); err != nil {
			logger.LOG.Error("反序列化预检信息失败", "error", err, "data", v)
			return nil, fmt.Errorf("预检信息格式错误")
		}
	default:
		logger.LOG.Error("预检信息类型错误", "type", fmt.Sprintf("%T", v))
		return nil, fmt.Errorf("预检信息类型错误")
	}

	if precheckResp.PrecheckID != req.PrecheckID {
		return nil, fmt.Errorf("无效的预检ID")
	}

	// 3. 选择合适的磁盘（按剩余空间最大原则）
	// 获取预检请求中的文件大小
	var fileSize int64
	reqCacheKey := fmt.Sprintf("fileUploadReq:%s", userID)
	reqData, err := f.cacheLocal.Get(reqCacheKey)
	if err != nil {
		logger.LOG.Error("获取预检请求失败", "error", err)
		return nil, fmt.Errorf("无法获取原始上传请求信息")
	}

	// 反序列化预检请求数据以获取文件大小
	var precheckReq request.UploadPrecheckRequest
	switch v := reqData.(type) {
	case *request.UploadPrecheckRequest:
		precheckReq = *v
	case string:
		if err := json.Unmarshal([]byte(v), &precheckReq); err != nil {
			logger.LOG.Error("反序列化预检请求失败", "error", err)
			return nil, fmt.Errorf("预检请求信息格式错误")
		}
	default:
		logger.LOG.Error("预检请求类型错误", "type", fmt.Sprintf("%T", v))
		return nil, fmt.Errorf("预检请求信息类型错误")
	}
	fileSize = precheckReq.FileSize

	// 选择最佳磁盘
	disks, err := f.factory.Disk().List(ctx, 0, 1000)
	if err != nil {
		logger.LOG.Error("查询磁盘列表失败", "error", err)
		return nil, fmt.Errorf("查询磁盘列表失败: %w", err)
	}
	if len(disks) == 0 {
		return nil, fmt.Errorf("没有可用的存储磁盘")
	}

	// 选择剩余空间最大且能容纳文件的磁盘
	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1
	for _, disk := range disks {
		freeSpaceBytes := int64(disk.Size) * 1024 * 1024 * 1024 // GB转字节
		if freeSpaceBytes >= fileSize && freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}
	if bestDisk == nil {
		return nil, fmt.Errorf("没有足够空间的磁盘")
	}

	// 4. 在选中磁盘的temp目录下创建临时目录：{DiskPath}/temp/{fileName}_{sessionID}/
	// 参考下载时的临时目录管理方式
	sessionID := req.PrecheckID[:8] // 使用预检ID的前8位作为会话ID
	// 使用文件名（去除扩展名）+ sessionID作为子目录名
	fileNameWithoutExt := precheckReq.FileName
	if idx := strings.LastIndex(precheckReq.FileName, "."); idx != -1 {
		fileNameWithoutExt = precheckReq.FileName[:idx]
	}
	tempBaseDir := filepath.Join(bestDisk.DataPath, "temp", fmt.Sprintf("%s_%s", fileNameWithoutExt, sessionID))
	if err := os.MkdirAll(tempBaseDir, 0755); err != nil {
		logger.LOG.Error("创建临时目录失败", "error", err, "path", tempBaseDir)
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	logger.LOG.Info("创建临时目录", "path", tempBaseDir, "diskPath", bestDisk.DataPath)

	// 3. 判断是否为分片上传
	isChunkUpload := req.ChunkIndex != nil && req.TotalChunks != nil

	if isChunkUpload {
		// 分片上传处理
		return f.handleChunkUpload(ctx, req, file, header, userID, tempBaseDir, &precheckResp)
	} else {
		// 小文件直传处理
		return f.handleSingleUpload(ctx, req, file, header, userID, tempBaseDir, &precheckResp)
	}
}

// handleChunkUpload 处理分片上传
func (f *FileService) handleChunkUpload(ctx context.Context, req *request.FileUploadRequest, file multipart.File, header *multipart.FileHeader, userID, tempBaseDir string, precheckResp *response.FilePrecheckResponse) (*models.JsonResponse, error) {
	chunkIndex := *req.ChunkIndex
	totalChunks := *req.TotalChunks

	// 1. 保存分片文件
	chunkPath := filepath.Join(tempBaseDir, fmt.Sprintf("%d.chunk.data", chunkIndex))
	chunkFile, err := os.Create(chunkPath)
	if err != nil {
		return nil, fmt.Errorf("创建分片文件失败: %w", err)
	}
	defer chunkFile.Close()

	if _, err := io.Copy(chunkFile, file); err != nil {
		return nil, fmt.Errorf("保存分片文件失败: %w", err)
	}

	logger.LOG.Info("分片上传成功", "chunkIndex", chunkIndex, "totalChunks", totalChunks, "userID", userID)

	// 2. 使用锁保护分片计数和删除操作，防止并发竞争
	lockKey := userID + ":" + header.Filename
	mutexVal, _ := uploadLocks.LoadOrStore(lockKey, &sync.Mutex{})
	mutex := mutexVal.(*sync.Mutex)

	mutex.Lock()
	// 注意：不使用defer，因为我们需要在文件处理前手动释放锁

	// 3. 删除 UploadChunk 表中对应的 MD5 记录（在锁保护下）
	// 注意：这里删除只是为了清理数据，不用于统计进度
	if req.ChunkMD5 != "" {
		// 查找匹配的 UploadChunk 记录
		chunks, err := f.factory.UploadChunk().ListByUserID(ctx, userID, 0, 1000)
		if err == nil {
			for _, chunk := range chunks {
				if chunk.Md5 == req.ChunkMD5 && chunk.FileName == header.Filename {
					if err := f.factory.UploadChunk().Delete(ctx, chunk.ChunkID); err != nil {
						logger.LOG.Warn("删除UploadChunk记录失败", "error", err, "chunkID", chunk.ChunkID)
					} else {
						logger.LOG.Debug("删除UploadChunk记录成功", "chunkID", chunk.ChunkID, "md5", req.ChunkMD5)
					}
					break
				}
			}
		}
	}

	// 4. 检查是否所有分片都已上传完成（通过检查临时目录中的文件数量）
	// 重要：不依赖UploadChunk表，而是直接检查磁盘上的分片文件
	uploadedChunkCount := 0
	for i := 0; i < totalChunks; i++ {
		chunkPath := filepath.Join(tempBaseDir, fmt.Sprintf("%d.chunk.data", i))
		if _, err := os.Stat(chunkPath); err == nil {
			uploadedChunkCount++
		}
	}

	remaining := int64(totalChunks - uploadedChunkCount)
	logger.LOG.Debug("分片上传进度", "chunkIndex", chunkIndex, "uploadedChunkCount", uploadedChunkCount, "totalChunks", totalChunks, "remaining", remaining, "fileName", header.Filename)

	// 5. 如果还有分片未完成，释放锁并返回成功响应
	if remaining > 0 {
		mutex.Unlock() // 释放锁
		return models.NewJsonResponse(200, "分片上传成功", map[string]interface{}{
			"chunk_index": chunkIndex,
			"uploaded":    totalChunks - int(remaining),
			"total":       totalChunks,
			"is_complete": false,
		}), nil
	}

	// 6. 所有分片上传完成，检查是否已经有其他请求在处理
	if _, isProcessing := processingFiles.LoadOrStore(lockKey, true); isProcessing {
		// 已经有其他请求在处理此文件
		mutex.Unlock()
		logger.LOG.Info("文件已被其他请求处理", "fileName", header.Filename)
		return models.NewJsonResponse(200, "文件处理中", map[string]interface{}{
			"is_complete": false,
			"message":     "文件正在处理中",
		}), nil
	}

	// 7. 标记为正在处理，现在可以释放锁了
	mutex.Unlock()
	uploadLocks.Delete(lockKey)

	// 确保处理完成后删除处理标记
	defer processingFiles.Delete(lockKey)

	logger.LOG.Info("所有分片上传完成，开始处理文件", "userID", userID, "fileName", header.Filename)

	// 获取预检请求中的原始数据
	var precheckReq request.UploadPrecheckRequest
	reqCacheKey := fmt.Sprintf("fileUploadReq:%s", userID)
	reqData, err := f.cacheLocal.Get(reqCacheKey)
	if err != nil {
		logger.LOG.Error("获取预检请求失败", "error", err)
		return nil, fmt.Errorf("无法获取原始上传请求信息")
	}

	// 反序列化预检请求数据
	switch v := reqData.(type) {
	case *request.UploadPrecheckRequest:
		// LocalCache直接返回对象指针
		precheckReq = *v
	case string:
		// Redis返回JSON字符串，需要反序列化
		if err := json.Unmarshal([]byte(v), &precheckReq); err != nil {
			logger.LOG.Error("反序列化预检请求失败", "error", err, "data", v)
			return nil, fmt.Errorf("预检请求信息格式错误")
		}
	default:
		logger.LOG.Error("预检请求类型错误", "type", fmt.Sprintf("%T", v))
		return nil, fmt.Errorf("预检请求信息类型错误")
	}

	// 构造上传数据
	uploadData := &upload.FileUploadData{
		TempFilePath:    filepath.Join(tempBaseDir, "0.chunk.data"), // 第一个分片路径作为基础
		FileName:        header.Filename,
		FileSize:        precheckReq.FileSize,
		ChunkSignature:  precheckReq.ChunkSignature,
		FirstChunkHash:  precheckReq.FilesMd5[0],
		SecondChunkHash: precheckReq.FilesMd5[1],
		ThirdChunkHash:  precheckReq.FilesMd5[2],
		IsEnc:           req.IsEnc,
		IsChunk:         true,
		ChunkCount:      totalChunks,
		VirtualPath:     precheckReq.PathID,
		UserID:          userID,
		FilePassword:    req.FilePassword, // 添加加密密码
	}

	fileID, err := upload.ProcessUploadedFile(uploadData, f.factory)
	if err != nil {
		logger.LOG.Error("处理上传文件失败", "error", err)
		return nil, fmt.Errorf("文件处理失败: %w", err)
	}

	// 6. 清除缓存
	f.cacheLocal.Delete(fmt.Sprintf("fileUpload:%s", userID))
	f.cacheLocal.Delete(reqCacheKey)

	logger.LOG.Info("文件上传完成", "fileID", fileID, "fileName", header.Filename)
	return models.NewJsonResponse(200, "上传成功", map[string]interface{}{
		"file_id":     fileID,
		"is_complete": true,
	}), nil
}

// handleSingleUpload 处理小文件直传
func (f *FileService) handleSingleUpload(ctx context.Context, req *request.FileUploadRequest, file multipart.File, header *multipart.FileHeader, userID, tempBaseDir string, precheckResp *response.FilePrecheckResponse) (*models.JsonResponse, error) {
	// 1. 保存临时文件
	tempFilePath := filepath.Join(tempBaseDir, "upload.tmp")
	tempFile, err := os.Create(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tempFile.Close()

	if _, err := io.Copy(tempFile, file); err != nil {
		return nil, fmt.Errorf("保存文件失败: %w", err)
	}

	logger.LOG.Info("小文件上传成功", "fileName", header.Filename, "size", header.Size, "userID", userID)

	// 2. 获取预检请求中的原始数据
	var precheckReq request.UploadPrecheckRequest
	cacheKey := fmt.Sprintf("fileUploadReq:%s", userID)
	reqData, err := f.cacheLocal.Get(cacheKey)
	if err != nil {
		logger.LOG.Error("获取预检请求失败", "error", err)
		return nil, fmt.Errorf("无法获取原始上传请求信息")
	}

	// 反序列化预检请求数据
	switch v := reqData.(type) {
	case *request.UploadPrecheckRequest:
		// LocalCache直接返回对象指针
		precheckReq = *v
	case string:
		// Redis返回JSON字符串，需要反序列化
		if err := json.Unmarshal([]byte(v), &precheckReq); err != nil {
			logger.LOG.Error("反序列化预检请求失败", "error", err, "data", v)
			return nil, fmt.Errorf("预检请求信息格式错误")
		}
	default:
		logger.LOG.Error("预检请求类型错误", "type", fmt.Sprintf("%T", v))
		return nil, fmt.Errorf("预检请求信息类型错误")
	}

	// 3. 构造上传数据
	uploadData := &upload.FileUploadData{
		TempFilePath:   tempFilePath,
		FileName:       header.Filename,
		FileSize:       header.Size,
		ChunkSignature: precheckReq.ChunkSignature,
		IsEnc:          req.IsEnc,
		IsChunk:        false,
		VirtualPath:    precheckReq.PathID,
		UserID:         userID,
		FilePassword:   req.FilePassword, // 添加加密密码
	}

	// 设置hash信息（如果有）
	if len(precheckReq.FilesMd5) > 0 {
		uploadData.FirstChunkHash = precheckReq.FilesMd5[0]
		if len(precheckReq.FilesMd5) > 1 {
			uploadData.SecondChunkHash = precheckReq.FilesMd5[1]
		}
		if len(precheckReq.FilesMd5) > 2 {
			uploadData.ThirdChunkHash = precheckReq.FilesMd5[2]
		}
	}

	// 4. 调用 ProcessUploadedFile
	fileID, err := upload.ProcessUploadedFile(uploadData, f.factory)
	if err != nil {
		logger.LOG.Error("处理上传文件失败", "error", err)
		return nil, fmt.Errorf("文件处理失败: %w", err)
	}

	// 5. 清除缓存
	f.cacheLocal.Delete(fmt.Sprintf("fileUpload:%s", userID))
	f.cacheLocal.Delete(cacheKey)

	logger.LOG.Info("文件上传完成", "fileID", fileID, "fileName", header.Filename)
	return models.NewJsonResponse(200, "上传成功", map[string]interface{}{
		"file_id":     fileID,
		"is_complete": true,
	}), nil
}
