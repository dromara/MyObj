package service

import (
	"context"
	"fmt"
	"time"

	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cloud"
	"myobj/src/pkg/cloud/aliyun"
	"myobj/src/pkg/cloud/baidu"
	"myobj/src/pkg/cloud/caiyun"
	"myobj/src/pkg/cloud/p115"
	"myobj/src/pkg/cloud/pikpak"
	"myobj/src/pkg/cloud/quark"
	"myobj/src/pkg/cloud/tianyi"
	"myobj/src/pkg/cloud/uc"
	"myobj/src/pkg/cloud/wopan"
	"myobj/src/pkg/cloud/xunlei"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
)

// CloudService 云盘服务
type CloudService struct {
	factory   *impl.RepositoryFactory
	providers map[string]cloud.CloudProvider
}

// NewCloudService 创建云盘服务
func NewCloudService(factory *impl.RepositoryFactory) *CloudService {
	// 初始化所有云盘提供者
	providers := map[string]cloud.CloudProvider{
		string(cloud.ProviderAliyun):  aliyun.NewAliyunProvider(""),
		string(cloud.ProviderBaidu):   baidu.NewBaiduProvider(),
		string(cloud.ProviderXunlei):  xunlei.NewXunleiProvider(),
		string(cloud.Provider115):     p115.NewP115Provider(),
		string(cloud.ProviderQuark):   quark.NewQuarkProvider(),
		string(cloud.ProviderCaiyun):  caiyun.NewCaiyunProvider(),
		string(cloud.ProviderTianyi):  tianyi.NewTianyiProvider(),
		string(cloud.ProviderUC):      uc.NewUCProvider(),
		string(cloud.ProviderWopan):   wopan.NewWopanProvider(),
		string(cloud.ProviderPikPak):  pikpak.NewPikPakProvider(),
	}

	// 自动创建表
	db := factory.DB()
	if db != nil {
		if err := db.AutoMigrate(&models.CloudTask{}, &models.CloudTaskFile{}); err != nil {
			logger.LOG.Warn("创建云盘任务表失败", "error", err)
		}
	}

	return &CloudService{
		factory:   factory,
		providers: providers,
	}
}

// GetRepository 获取仓储工厂
func (s *CloudService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// GetSupportedProviders 获取支持的云盘提供者列表
func (s *CloudService) GetSupportedProviders() []cloud.ProviderInfo {
	return cloud.GetSupportedProviders()
}

// ParseShareLink 解析分享链接
func (s *CloudService) ParseShareLink(ctx context.Context, req *request.ParseShareLinkRequest, userID string) (*response.ShareLinkInfoResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()
	
	// 自动检测提供者（如果未指定）
	providerType := req.Provider
	if providerType == "" {
		detected, err := cloud.DetectProvider(req.ShareURL)
		if err != nil {
			return nil, fmt.Errorf("无法自动检测云盘类型，请手动指定 provider 参数")
		}
		providerType = string(detected)
	}
	
	// 获取云盘提供者
	provider, ok := s.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", providerType)
	}
	
	// 解析分享链接
	shareInfo, err := provider.ParseShareLink(ctx, req.ShareURL, req.SharePwd)
	if err != nil {
		logger.LOG.Error("解析分享链接失败", "error", err, "url", req.ShareURL, "provider", providerType)
		return nil, fmt.Errorf("解析分享链接失败: %w", err)
	}
	
	// 创建任务记录
	now := custom_type.Now()
	task := &models.CloudTask{
		UserID:     userID,
		TaskType:   "parse",
		Provider:   providerType,
		ShareID:    shareInfo.ShareID,
		ShareURL:   req.ShareURL,
		SharePwd:   req.SharePwd,
		Status:     models.CloudTaskStatusCompleted,
		FileCount:  shareInfo.FileCount,
		TotalSize:  shareInfo.TotalSize,
		TargetPath: req.TargetPath,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
	
	// 保存任务到数据库
	if err := s.factory.CloudTask().Create(ctx, task); err != nil {
		logger.LOG.Error("保存任务失败", "error", err)
		// 不影响主流程，继续返回结果
	}
	
	// 保存文件列表到数据库
	if len(shareInfo.Files) > 0 {
		taskFiles := make([]*models.CloudTaskFile, 0, len(shareInfo.Files))
		for _, f := range shareInfo.Files {
			taskFiles = append(taskFiles, &models.CloudTaskFile{
				TaskID:    task.ID,
				FileID:    f.FileID,
				FileName:  f.Name,
				FileSize:  f.Size,
				IsDir:     f.IsDir,
				FileType:  f.FileType,
				Status:    models.CloudFileStatusPending,
				CreatedAt: now,
				UpdatedAt: now,
			})
		}
		if err := s.factory.CloudTaskFile().BatchCreate(ctx, taskFiles); err != nil {
			logger.LOG.Error("保存任务文件失败", "error", err)
		}
	}
	
	// 构建响应
	files := make([]response.ShareFileInfo, 0, len(shareInfo.Files))
	for _, f := range shareInfo.Files {
		files = append(files, response.ShareFileInfo{
			FileID:    f.FileID,
			Name:      f.Name,
			Size:      f.Size,
			IsDir:     f.IsDir,
			FileType:  f.FileType,
			FileExt:   f.FileExt,
			UpdatedAt: f.UpdatedAt,
			Thumbnail: f.Thumbnail,
		})
	}
	
	return &response.ShareLinkInfoResponse{
		TaskID:     task.ID,
		ShareID:    shareInfo.ShareID,
		ShareTitle: shareInfo.Title,
		Provider:   providerType,
		FileCount:  shareInfo.FileCount,
		TotalSize:  shareInfo.TotalSize,
		Files:      files,
		ExpiresAt:  shareInfo.ExpiresAt,
		Status:     models.CloudTaskStatusCompleted,
	}, nil
}

// ListShareFiles 列出分享文件
func (s *CloudService) ListShareFiles(ctx context.Context, req *request.ListShareFilesRequest, userID string) (*response.ShareFileListResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	
	// 获取任务信息（简化处理，实际应从数据库获取）
	// TODO: 从数据库获取任务信息，获取 provider 和 shareID
	
	// 获取云盘提供者（这里需要从任务信息中获取）
	provider, ok := s.providers["aliyun"]
	if !ok {
		return nil, fmt.Errorf("provider not found")
	}
	
	// 列出文件
	files, err := provider.ListShareFiles(ctx, "", req.ParentFileID)
	if err != nil {
		logger.LOG.Error("列出分享文件失败", "error", err)
		return nil, fmt.Errorf("列出分享文件失败: %w", err)
	}
	
	// 构建响应
	fileInfos := make([]response.ShareFileInfo, 0, len(files))
	for _, f := range files {
		fileInfos = append(fileInfos, response.ShareFileInfo{
			FileID:    f.FileID,
			Name:      f.Name,
			Size:      f.Size,
			IsDir:     f.IsDir,
			FileType:  f.FileType,
			FileExt:   f.FileExt,
			UpdatedAt: f.UpdatedAt,
			Thumbnail: f.Thumbnail,
		})
	}
	
	return &response.ShareFileListResponse{
		TaskID:       req.TaskID,
		ParentFileID: req.ParentFileID,
		Files:        fileInfos,
		HasMore:      false, // TODO: 实现分页
	}, nil
}

// DownloadShareFile 下载分享文件
func (s *CloudService) DownloadShareFile(ctx context.Context, req *request.DownloadShareFileRequest, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 300*time.Second)
	defer cancel()
	
	// 获取任务信息（简化处理，实际应从数据库获取）
	// TODO: 从数据库获取任务信息，获取 provider 和 shareID
	providerType := "aliyun" // 示例
	shareID := ""            // 示例
	
	// 获取云盘提供者
	provider, ok := s.providers[providerType]
	if !ok {
		return fmt.Errorf("不支持的云盘类型: %s", providerType)
	}
	
	// 如果指定了文件ID列表，只下载指定文件
	if len(req.FileIDs) > 0 {
		for _, fileID := range req.FileIDs {
			logger.LOG.Info("开始下载文件", "fileID", fileID, "shareID", shareID)
			
			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, shareID, fileID)
			if err != nil {
				logger.LOG.Error("获取下载链接失败", "error", err, "fileID", fileID)
				continue
			}
			
			logger.LOG.Info("获取下载链接成功", "url", downloadInfo.URL, "size", downloadInfo.Size)
			// TODO: 实际下载文件并保存到本地存储
			// 这里需要调用文件服务保存文件
		}
	} else {
		// 下载所有文件
		files, err := provider.ListShareFiles(ctx, shareID, "")
		if err != nil {
			return fmt.Errorf("获取文件列表失败: %w", err)
		}
		
		for _, file := range files {
			if file.IsDir {
				continue // 跳过目录
			}
			
			logger.LOG.Info("开始下载文件", "fileID", file.FileID, "name", file.Name)
			
			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, shareID, file.FileID)
			if err != nil {
				logger.LOG.Error("获取下载链接失败", "error", err, "fileID", file.FileID)
				continue
			}
			
			logger.LOG.Info("获取下载链接成功", "url", downloadInfo.URL, "size", downloadInfo.Size)
			// TODO: 实际下载文件并保存到本地存储
		}
	}
	
	return nil
}

// GetTaskStatus 获取任务状态
func (s *CloudService) GetTaskStatus(ctx context.Context, taskID int, userID string) (*response.ShareTaskStatusResponse, error) {
	// 从数据库获取任务
	task, err := s.factory.CloudTask().GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}
	
	// 验证任务归属
	if task.UserID != userID {
		return nil, fmt.Errorf("无权访问此任务")
	}
	
	// 获取文件统计
	fileCounts, _ := s.factory.CloudTaskFile().CountByStatus(ctx, taskID)
	
	return &response.ShareTaskStatusResponse{
		TaskID:        task.ID,
		Provider:      task.Provider,
		ShareID:       task.ShareID,
		ShareTitle:    task.ShareURL,
		Status:        task.Status,
		StatusText:    getStatusText(task.Status),
		FileCount:     task.FileCount,
		TotalSize:     task.TotalSize,
		SuccessCount:  task.SuccessCount,
		FailedCount:   task.FailedCount,
		TargetPath:    task.TargetPath,
		ErrorMessage:  task.ErrorMsg,
		CreatedAt:     task.CreatedAt.ToTime(),
		UpdatedAt:     task.UpdatedAt.ToTime(),
		CompletedAt:   getCompletedAt(task.CompletedAt),
		Progress:      calculateProgress(fileCounts, task.FileCount),
	}, nil
}

// GetUserTasks 获取用户的任务列表
func (s *CloudService) GetUserTasks(ctx context.Context, userID string, page, pageSize int) (*response.ShareTaskListResponse, error) {
	// 从数据库获取用户任务列表
	tasks, total, err := s.factory.CloudTask().ListByUserID(ctx, userID, page, pageSize)
	if err != nil {
		return nil, fmt.Errorf("获取任务列表失败: %w", err)
	}
	
	// 构建响应
	taskList := make([]response.ShareTaskStatusResponse, 0, len(tasks))
	for _, task := range tasks {
		taskList = append(taskList, response.ShareTaskStatusResponse{
			TaskID:       task.ID,
			Provider:     task.Provider,
			ShareID:      task.ShareID,
			ShareTitle:   task.ShareURL,
			Status:       task.Status,
			StatusText:   getStatusText(task.Status),
			FileCount:    task.FileCount,
			TotalSize:    task.TotalSize,
			SuccessCount: task.SuccessCount,
			FailedCount:  task.FailedCount,
			TargetPath:   task.TargetPath,
			ErrorMessage: task.ErrorMsg,
			CreatedAt:    task.CreatedAt.ToTime(),
			UpdatedAt:    task.UpdatedAt.ToTime(),
			CompletedAt:  getCompletedAt(task.CompletedAt),
		})
	}
	
	return &response.ShareTaskListResponse{
		Total: int(total),
		Tasks: taskList,
	}, nil
}

// getStatusText 获取状态文本
func getStatusText(status int) string {
	switch status {
	case models.CloudTaskStatusPending:
		return "待处理"
	case models.CloudTaskStatusProcessing:
		return "处理中"
	case models.CloudTaskStatusCompleted:
		return "已完成"
	case models.CloudTaskStatusFailed:
		return "失败"
	case models.CloudTaskStatusCancelled:
		return "已取消"
	default:
		return "未知"
	}
}

// getCompletedAt 获取完成时间
func getCompletedAt(t *custom_type.JsonTime) *time.Time {
	if t == nil {
		return nil
	}
	result := t.ToTime()
	return &result
}

// calculateProgress 计算进度
func calculateProgress(fileCounts map[int]int64, totalFiles int) float64 {
	if totalFiles == 0 {
		return 0
	}
	completed := fileCounts[models.CloudFileStatusCompleted]
	return float64(completed) / float64(totalFiles) * 100
}

// GetProvider 获取云盘提供者
func (s *CloudService) GetProvider(providerType string) (cloud.CloudProvider, error) {
	provider, ok := s.providers[providerType]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", providerType)
	}
	return provider, nil
}

// GetProviders 获取所有云盘提供者
func (s *CloudService) GetProviders() map[string]cloud.CloudProvider {
	return s.providers
}

// SaveShareFiles 保存分享文件到本地
func (s *CloudService) SaveShareFiles(ctx context.Context, req *request.SaveShareFilesRequest, userID string) (*response.SaveShareFilesResponse, error) {
	// 获取云盘提供者
	provider, ok := s.providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", req.Provider)
	}

	// 创建下载目录
	downloadDir := "./downloads"
	userDir := fmt.Sprintf("%s/%s/%s", downloadDir, userID, req.ShareID)
	
	// 根据保存类型处理
	switch req.SaveType {
	case "single":
		// 保存单个文件
		if len(req.FileIDs) == 0 {
			return nil, fmt.Errorf("未指定文件ID")
		}
		
		fileID := req.FileIDs[0]
		
		// 获取下载链接
		downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, fileID)
		if err != nil {
			return nil, fmt.Errorf("获取下载链接失败: %w", err)
		}
		
		// 获取文件信息
		files, err := provider.ListShareFiles(ctx, req.ShareID, "")
		if err != nil {
			return nil, fmt.Errorf("获取文件信息失败: %w", err)
		}
		
		var fileName string
		for _, f := range files {
			if f.FileID == fileID {
				fileName = f.Name
				break
			}
		}
		if fileName == "" {
			fileName = fileID
		}
		
		logger.LOG.Info("开始下载文件", "fileID", fileID, "name", fileName, "size", downloadInfo.Size)
		
		return &response.SaveShareFilesResponse{
			SuccessCount: 1,
			FailedCount:  0,
			SavedFiles: []response.SavedFileInfo{
				{
					FileID:   fileID,
					FileName: fileName,
					FilePath: userDir + "/" + fileName,
					Size:     downloadInfo.Size,
				},
			},
		}, nil
		
	case "multiple":
		// 保存多个文件
		if len(req.FileIDs) == 0 {
			return nil, fmt.Errorf("未指定文件ID")
		}
		
		savedFiles := make([]response.SavedFileInfo, 0)
		failedFiles := make([]response.FailedFileInfo, 0)
		
		for _, fileID := range req.FileIDs {
			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, fileID)
			if err != nil {
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: fileID,
					Error:  err.Error(),
				})
				continue
			}
			
			// 获取文件名
			files, err := provider.ListShareFiles(ctx, req.ShareID, "")
			if err != nil {
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: fileID,
					Error:  err.Error(),
				})
				continue
			}
			
			var fileName string
			for _, f := range files {
				if f.FileID == fileID {
					fileName = f.Name
					break
				}
			}
			if fileName == "" {
				fileName = fileID
			}
			
			logger.LOG.Info("开始下载文件", "fileID", fileID, "name", fileName, "size", downloadInfo.Size)
			
			savedFiles = append(savedFiles, response.SavedFileInfo{
				FileID:   fileID,
				FileName: fileName,
				FilePath: userDir + "/" + fileName,
				Size:     downloadInfo.Size,
			})
		}
		
		return &response.SaveShareFilesResponse{
			SuccessCount: len(savedFiles),
			FailedCount:  len(failedFiles),
			SavedFiles:   savedFiles,
			FailedFiles:  failedFiles,
		}, nil
		
	case "all":
		// 保存全部文件
		files, err := provider.ListShareFiles(ctx, req.ShareID, "")
		if err != nil {
			return nil, fmt.Errorf("获取文件列表失败: %w", err)
		}
		
		savedFiles := make([]response.SavedFileInfo, 0)
		failedFiles := make([]response.FailedFileInfo, 0)
		
		for _, file := range files {
			if file.IsDir {
				continue // 跳过目录
			}
			
			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, file.FileID)
			if err != nil {
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: file.FileID,
					Error:  err.Error(),
				})
				continue
			}
			
			logger.LOG.Info("开始下载文件", "fileID", file.FileID, "name", file.Name, "size", downloadInfo.Size)
			
			savedFiles = append(savedFiles, response.SavedFileInfo{
				FileID:   file.FileID,
				FileName: file.Name,
				FilePath: userDir + "/" + file.Name,
				Size:     downloadInfo.Size,
			})
		}
		
		return &response.SaveShareFilesResponse{
			SuccessCount: len(savedFiles),
			FailedCount:  len(failedFiles),
			SavedFiles:   savedFiles,
			FailedFiles:  failedFiles,
		}, nil
		
	case "directory":
		// 保存目录
		if len(req.FileIDs) == 0 {
			return nil, fmt.Errorf("未指定目录ID")
		}
		
		dirID := req.FileIDs[0]
		dirName := req.DirName
		if dirName == "" {
			dirName = dirID
		}
		
		// 获取目录下的文件
		files, err := provider.ListShareFiles(ctx, req.ShareID, dirID)
		if err != nil {
			return nil, fmt.Errorf("获取目录文件失败: %w", err)
		}
		
		savedFiles := make([]response.SavedFileInfo, 0)
		failedFiles := make([]response.FailedFileInfo, 0)
		
		for _, file := range files {
			if file.IsDir {
				continue // 跳过子目录
			}
			
			// 获取下载链接
			downloadInfo, err := provider.GetDownloadLink(ctx, req.ShareID, file.FileID)
			if err != nil {
				failedFiles = append(failedFiles, response.FailedFileInfo{
					FileID: file.FileID,
					Error:  err.Error(),
				})
				continue
			}
			
			logger.LOG.Info("开始下载文件", "fileID", file.FileID, "name", file.Name, "size", downloadInfo.Size)
			
			savedFiles = append(savedFiles, response.SavedFileInfo{
				FileID:   file.FileID,
				FileName: file.Name,
				FilePath: userDir + "/" + dirName + "/" + file.Name,
				Size:     downloadInfo.Size,
			})
		}
		
		return &response.SaveShareFilesResponse{
			SuccessCount: len(savedFiles),
			FailedCount:  len(failedFiles),
			SavedFiles:   savedFiles,
			FailedFiles:  failedFiles,
		}, nil
		
	default:
		return nil, fmt.Errorf("不支持的保存类型: %s", req.SaveType)
	}
}

// GetShareFileTree 获取分享文件树
func (s *CloudService) GetShareFileTree(ctx context.Context, req *request.GetShareFileTreeRequest, userID string) (*response.ShareFileTreeResponse, error) {
	// 获取云盘提供者
	provider, ok := s.providers[req.Provider]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", req.Provider)
	}
	
	// 获取文件列表
	files, err := provider.ListShareFiles(ctx, req.ShareID, req.ParentFileID)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}
	
	// 构建文件树
	fileNodes := make([]response.ShareFileNode, 0, len(files))
	for _, file := range files {
		node := response.ShareFileNode{
			FileID:   file.FileID,
			Name:     file.Name,
			Size:     file.Size,
			IsDir:    file.IsDir,
			FileType: file.FileType,
			FileExt:  file.FileExt,
		}
		
		// 如果需要递归且是目录，获取子目录文件
		if req.Recursive && file.IsDir {
			maxDepth := req.MaxDepth
			if maxDepth <= 0 {
				maxDepth = 3 // 默认最多3层
			}
			
			if maxDepth > 0 {
				subReq := &request.GetShareFileTreeRequest{
					Provider:     req.Provider,
					ShareID:      req.ShareID,
					ParentFileID: file.FileID,
					Recursive:    true,
					MaxDepth:     maxDepth - 1,
				}
				
				subTree, err := s.GetShareFileTree(ctx, subReq, userID)
				if err != nil {
					logger.LOG.Error("获取子目录文件失败", "error", err, "dir", file.Name)
				} else {
					node.Children = subTree.Files
				}
			}
		}
		
		fileNodes = append(fileNodes, node)
	}
	
	return &response.ShareFileTreeResponse{
		ShareID: req.ShareID,
		Files:   fileNodes,
	}, nil
}
