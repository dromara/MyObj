package service

import (
	"context"
	"errors"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cache"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"strconv"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

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
	if signature.FirstChunkHash == req.FirstChunkHash && signature.SecondChunkHash == req.SecondChunkHash && signature.ThirdChunkHash == req.ThirdChunkHash {
		//TODO: 需要向数据库写入用户对应的文件信息
		return models.NewJsonResponse(200, "秒传成功", nil), nil
	}
	//无法触发秒传，但可上传，返回校验ID
	key := fmt.Sprintf("fileUpload:%s", user.ID)
	uid := uuid.New().String()
	err = c.Set(key, uid, 300) //五分钟内可用的校验
	if err != nil {
		logger.LOG.Error("缓存设置失败", "error", err, "key", key)
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
		resp.Files = append(resp.Files, &response.FileItem{
			FileID:       file.ID,
			FileName:     file.Name,
			FileSize:     file.Size,
			MimeType:     file.Mime,
			IsEnc:        file.IsEnc,
			HasThumbnail: file.ThumbnailImg != "",
			CreatedAt:    file.CreatedAt,
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
	errors := []string{}

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
