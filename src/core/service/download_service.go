package service

import (
	"context"
	"fmt"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/download"
	"myobj/src/pkg/enum"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// DownloadService 下载服务
type DownloadService struct {
	factory *impl.RepositoryFactory
	tempDir string // 临时目录
}

func NewDownloadService(factory *impl.RepositoryFactory) *DownloadService {
	// 选择最大磁盘创建临时目录
	tempDir := "./obj_temp/downloads" // 默认值

	ctx := context.Background()
	disk, err := factory.Disk().GetBigDisk(ctx)
	if err == nil && disk != nil {
		// 在最大磁盘的data_path下创建 temp 目录
		tempDir = filepath.Join(disk.DataPath, "temp", "downloads")
		logger.LOG.Info("使用最大磁盘创建临时目录", "disk", disk.DiskPath, "tempDir", tempDir)
	} else {
		logger.LOG.Warn("获取最大磁盘失败，使用默认临时目录", "error", err)
	}

	// 确保临时目录存在
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		logger.LOG.Error("创建临时目录失败", "tempDir", tempDir, "error", err)
	} else {
		logger.LOG.Info("临时目录初始化成功", "tempDir", tempDir)
	}

	return &DownloadService{
		factory: factory,
		tempDir: tempDir,
	}
}

func (d *DownloadService) GetRepository() *impl.RepositoryFactory {
	return d.factory
}

// CreateOfflineDownload 创建离线下载任务
func (d *DownloadService) CreateOfflineDownload(req *request.CreateOfflineDownloadRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 1. 验证用户是否存在
	_, err := d.factory.User().GetByID(ctx, userID)
	if err != nil {
		logger.LOG.Error("获取用户信息失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("用户不存在")
	}

	// 2. 设置默认虚拟路径
	virtualPath := req.VirtualPath
	if virtualPath == "" {
		virtualPath = "/离线下载/"
	}

	if req.EnableEncryption {
		if req.FilePassword == "" {
			return nil, fmt.Errorf("加密存储密码不能为空")
		}
	}

	// 3. 创建下载任务记录
	taskID := uuid.Must(uuid.NewV7()).String()
	task := &models.DownloadTask{
		ID:               taskID,
		UserID:           userID,
		Type:             enum.DownloadTaskTypeHttp.Value(),
		URL:              req.URL,
		VirtualPath:      virtualPath,
		EnableEncryption: req.EnableEncryption,
		State:            enum.DownloadTaskStateInit.Value(),
		TargetDir:        d.tempDir,
		CreateTime:       custom_type.Now(),
		UpdateTime:       custom_type.Now(),
	}

	if err := d.factory.DownloadTask().Create(ctx, task); err != nil {
		logger.LOG.Error("创建下载任务失败", "error", err, "userID", userID, "url", req.URL)
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 4. 异步启动下载任务
	go func() {
		opts := &download.HTTPDownloadOptions{
			EnableEncryption: req.EnableEncryption,
			VirtualPath:      virtualPath,
			MaxRetries:       3,
			ChunkSize:        10 * 1024 * 1024, // 10MB
			MaxConcurrent:    4,
			Timeout:          300,
			FilePassword:     req.FilePassword,
		}

		_, err := download.DownloadHTTP(taskID, req.URL, userID, d.tempDir, d.factory, opts)
		if err != nil {
			logger.LOG.Error("离线下载失败", "taskID", taskID, "error", err)
		} else {

		}
	}()

	logger.LOG.Info("离线下载任务已创建", "taskID", taskID, "userID", userID, "url", req.URL)

	// 返回任务信息
	taskResp := d.convertTaskToResponse(task)
	return models.NewJsonResponse(200, "任务创建成功", taskResp), nil
}

// GetTaskList 获取下载任务列表
func (d *DownloadService) GetTaskList(req *request.DownloadTaskListRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	offset := (req.Page - 1) * req.PageSize

	var tasks []*models.DownloadTask
	var total int64
	var err error

	// 判断是否指定了类型过滤
	hasTypeFilter := req.Type >= 0
	hasStateFilter := req.State >= 0

	if hasStateFilter && hasTypeFilter {
		// 按状态和类型查询
		tasks, err = d.factory.DownloadTask().ListByStateAndType(ctx, userID, req.State, req.Type, offset, req.PageSize)
		if err != nil {
			logger.LOG.Error("查询下载任务失败", "error", err, "userID", userID, "state", req.State, "type", req.Type)
			return nil, fmt.Errorf("查询任务失败: %w", err)
		}
		total, err = d.factory.DownloadTask().CountByStateAndType(ctx, userID, req.State, req.Type)
	} else if hasStateFilter {
		// 只按状态查询
		tasks, err = d.factory.DownloadTask().ListByState(ctx, userID, req.State, offset, req.PageSize)
		if err != nil {
			logger.LOG.Error("查询下载任务失败", "error", err, "userID", userID, "state", req.State)
			return nil, fmt.Errorf("查询任务失败: %w", err)
		}
		total, err = d.factory.DownloadTask().CountByState(ctx, userID, req.State)
	} else if hasTypeFilter {
		// 只按类型查询
		tasks, err = d.factory.DownloadTask().ListByType(ctx, userID, req.Type, offset, req.PageSize)
		if err != nil {
			logger.LOG.Error("查询下载任务失败", "error", err, "userID", userID, "type", req.Type)
			return nil, fmt.Errorf("查询任务失败: %w", err)
		}
		total, err = d.factory.DownloadTask().CountByType(ctx, userID, req.Type)
	} else {
		// 查询所有任务
		tasks, err = d.factory.DownloadTask().ListByUserID(ctx, userID, offset, req.PageSize)
		if err != nil {
			logger.LOG.Error("查询下载任务失败", "error", err, "userID", userID)
			return nil, fmt.Errorf("查询任务失败: %w", err)
		}
		total, err = d.factory.DownloadTask().Count(ctx, userID)
	}

	if err != nil {
		logger.LOG.Error("统计下载任务失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("统计任务失败: %w", err)
	}

	// 转换为响应格式
	taskResponses := make([]*response.DownloadTaskResponse, 0, len(tasks))
	for _, task := range tasks {
		t := d.convertTaskToResponse(task)

		// 只有已完成的任务才有 FileID，需要查询 user_files 获取 uf_id
		if task.State == enum.DownloadTaskStateFinished.Value() && task.FileID != "" {
			userFile, err := d.factory.UserFiles().GetByUserIDAndFileID(ctx, userID, task.FileID)
			if err != nil {
				logger.LOG.Warn("获取用户文件信息失败", "error", err, "fileID", task.FileID, "userID", userID)
				// 不阻断整个列表，继续处理下一个任务
				t.FileID = task.FileID // 使用原始 FileID
			} else {
				t.FileID = userFile.UfID // 返回 uf_id
			}
		} else {
			// 未完成的任务，返回空字符串
			t.FileID = ""
		}

		taskResponses = append(taskResponses, t)
	}

	result := &response.DownloadTaskListResponse{
		Tasks:    taskResponses,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	return models.NewJsonResponse(200, "查询成功", result), nil
}

// PauseTask 暂停下载任务
func (d *DownloadService) PauseTask(req *request.TaskOperationRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证任务是否属于该用户
	task, err := d.factory.DownloadTask().GetByID(ctx, req.TaskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "error", err, "taskID", req.TaskID)
		return nil, fmt.Errorf("任务不存在")
	}

	if task.UserID != userID {
		logger.LOG.Warn("用户尝试操作他人任务", "userID", userID, "taskID", req.TaskID, "taskOwner", task.UserID)
		return nil, fmt.Errorf("无权操作此任务")
	}

	// 根据任务类型调用不同的暂停函数
	if task.Type == enum.DownloadTaskTypeBtp.Value() || task.Type == enum.DownloadTaskTypeMagnet.Value() {
		// 种子/磁力链下载
		if err := download.PauseTorrentDownload(req.TaskID, d.factory); err != nil {
			logger.LOG.Error("暂停种子下载任务失败", "error", err, "taskID", req.TaskID)
			return nil, fmt.Errorf("暂停任务失败: %w", err)
		}
	} else {
		// HTTP/FTP/SFTP等其他类型下载
		if err := download.PauseDownload(req.TaskID, d.factory); err != nil {
			logger.LOG.Error("暂停下载任务失败", "error", err, "taskID", req.TaskID)
			return nil, fmt.Errorf("暂停任务失败: %w", err)
		}
	}

	logger.LOG.Info("下载任务已暂停", "taskID", req.TaskID, "userID", userID)
	return models.NewJsonResponse(200, "任务已暂停", nil), nil
}

// ResumeTask 恢复下载任务
func (d *DownloadService) ResumeTask(req *request.TaskOperationRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证任务是否属于该用户
	task, err := d.factory.DownloadTask().GetByID(ctx, req.TaskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "error", err, "taskID", req.TaskID)
		return nil, fmt.Errorf("任务不存在")
	}

	if task.UserID != userID {
		logger.LOG.Warn("用户尝试操作他人任务", "userID", userID, "taskID", req.TaskID, "taskOwner", task.UserID)
		return nil, fmt.Errorf("无权操作此任务")
	}

	// 根据任务类型调用不同的恢复函数
	if task.Type == enum.DownloadTaskTypeBtp.Value() || task.Type == enum.DownloadTaskTypeMagnet.Value() {
		// 种子/磁力链下载
		if err := download.ResumeTorrentDownload(req.TaskID, userID, d.tempDir, d.factory); err != nil {
			logger.LOG.Error("恢复种子下载任务失败", "error", err, "taskID", req.TaskID)
			return nil, fmt.Errorf("恢复任务失败: %w", err)
		}
	} else {
		// HTTP/FTP/SFTP等其他类型下载
		if err := download.ResumeDownload(req.TaskID, userID, d.tempDir, d.factory); err != nil {
			logger.LOG.Error("恢复下载任务失败", "error", err, "taskID", req.TaskID)
			return nil, fmt.Errorf("恢复任务失败: %w", err)
		}
	}

	logger.LOG.Info("下载任务已恢复", "taskID", req.TaskID, "userID", userID)
	return models.NewJsonResponse(200, "任务已恢复", nil), nil
}

// CancelTask 取消下载任务
func (d *DownloadService) CancelTask(req *request.TaskOperationRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证任务是否属于该用户
	task, err := d.factory.DownloadTask().GetByID(ctx, req.TaskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "error", err, "taskID", req.TaskID)
		return nil, fmt.Errorf("任务不存在")
	}

	if task.UserID != userID {
		logger.LOG.Warn("用户尝试操作他人任务", "userID", userID, "taskID", req.TaskID, "taskOwner", task.UserID)
		return nil, fmt.Errorf("无权操作此任务")
	}

	// 根据任务类型调用不同的取消函数
	if task.Type == enum.DownloadTaskTypeBtp.Value() || task.Type == enum.DownloadTaskTypeMagnet.Value() {
		// 种子/磁力链下载
		if err := download.CancelTorrentDownload(req.TaskID, d.factory); err != nil {
			logger.LOG.Error("取消种子下载任务失败", "error", err, "taskID", req.TaskID)
			return nil, fmt.Errorf("取消任务失败: %w", err)
		}
	} else {
		// HTTP/FTP/SFTP等其他类型下载
		if err := download.CancelDownload(req.TaskID, d.factory); err != nil {
			logger.LOG.Error("取消下载任务失败", "error", err, "taskID", req.TaskID)
			return nil, fmt.Errorf("取消任务失败: %w", err)
		}
	}

	logger.LOG.Info("下载任务已取消", "taskID", req.TaskID, "userID", userID)
	return models.NewJsonResponse(200, "任务已取消", nil), nil
}

// DeleteTask 删除下载任务
func (d *DownloadService) DeleteTask(req *request.DeleteTaskRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证任务是否属于该用户
	task, err := d.factory.DownloadTask().GetByID(ctx, req.TaskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "error", err, "taskID", req.TaskID)
		return nil, fmt.Errorf("任务不存在")
	}

	if task.UserID != userID {
		logger.LOG.Warn("用户尝试删除他人任务", "userID", userID, "taskID", req.TaskID, "taskOwner", task.UserID)
		return nil, fmt.Errorf("无权删除此任务")
	}

	// 只能删除已完成或失败的任务
	if task.State != enum.DownloadTaskStateFinished.Value() && task.State != enum.DownloadTaskStateFailed.Value() {
		return nil, fmt.Errorf("只能删除已完成或失败的任务")
	}

	// 删除任务前，先清理临时文件（如果存在）
	if task.Path != "" && download.IsTempPath(task.Path) {
		logger.LOG.Info("删除任务时清理临时文件", "taskID", req.TaskID, "path", task.Path)
		if err := os.RemoveAll(task.Path); err != nil {
			logger.LOG.Warn("清理临时文件失败", "error", err, "path", task.Path)
			// 清理失败不影响删除任务
		}
	}

	// 如果是种子下载任务，清理对应的临时目录
	if task.Type == enum.DownloadTaskTypeBtp.Value() || task.Type == enum.DownloadTaskTypeMagnet.Value() {
		torrentTempDir := filepath.Join(d.tempDir, fmt.Sprintf("torrent_%s", req.TaskID))
		if _, err := os.Stat(torrentTempDir); err == nil {
			logger.LOG.Info("删除种子下载临时目录", "taskID", req.TaskID, "path", torrentTempDir)
			if err := os.RemoveAll(torrentTempDir); err != nil {
				logger.LOG.Warn("清理种子临时目录失败", "error", err, "path", torrentTempDir)
			}
		}
	}

	if err := d.factory.DownloadTask().Delete(ctx, req.TaskID); err != nil {
		logger.LOG.Error("删除下载任务失败", "error", err, "taskID", req.TaskID)
		return nil, fmt.Errorf("删除任务失败: %w", err)
	}

	logger.LOG.Info("下载任务已删除", "taskID", req.TaskID, "userID", userID)
	return models.NewJsonResponse(200, "任务已删除", nil), nil
}

// convertTaskToResponse 转换任务模型为响应格式
func (d *DownloadService) convertTaskToResponse(task *models.DownloadTask) *response.DownloadTaskResponse {
	stateText := d.getStateText(task.State)
	typeText := d.getTypeText(task.Type)

	return &response.DownloadTaskResponse{
		ID:             task.ID,
		URL:            task.URL,
		FileName:       task.FileName,
		FileSize:       task.FileSize,
		DownloadedSize: task.DownloadedSize,
		Progress:       task.Progress,
		Speed:          task.Speed,
		Type:           task.Type,
		TypeText:       typeText,
		State:          task.State,
		StateText:      stateText,
		VirtualPath:    task.VirtualPath,
		SupportRange:   task.SupportRange,
		ErrorMsg:       task.ErrorMsg,
		FileID:         task.FileID,
		CreateTime:     task.CreateTime,
		UpdateTime:     task.UpdateTime,
		FinishTime:     task.FinishTime,
	}
}

// getStateText 获取状态文本
func (d *DownloadService) getStateText(state int) string {
	switch state {
	case enum.DownloadTaskStateInit.Value():
		return "等待中"
	case enum.DownloadTaskStateDownloading.Value():
		return "下载中"
	case enum.DownloadTaskStatePaused.Value():
		return "已暂停"
	case enum.DownloadTaskStateFinished.Value():
		return "已完成"
	case enum.DownloadTaskStateFailed.Value():
		return "失败"
	default:
		return "未知"
	}
}

// getTypeText 获取类型文本
func (d *DownloadService) getTypeText(taskType int) string {
	switch taskType {
	case enum.DownloadTaskTypeHttp.Value():
		return "HTTP"
	case enum.DownloadTaskTypeFTP.Value():
		return "FTP"
	case enum.DownloadTaskTypeSFTP.Value():
		return "SFTP"
	case enum.DownloadTaskTypeS3.Value():
		return "S3"
	case enum.DownloadTaskTypeBtp.Value():
		return "种子"
	case enum.DownloadTaskTypeMagnet.Value():
		return "磁力链接"
	case enum.DownloadTaskTypeLocal.Value():
		return "本地文件"
	case enum.DownloadTaskTypeLocalFile.Value():
		return "网盘下载"
	default:
		return "未知"
	}
}

// CreateLocalFileDownload 创建网盘文件下载任务
func (d *DownloadService) CreateLocalFileDownload(req *request.CreateLocalFileDownloadRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 1. 验证用户是否存在
	_, err := d.factory.User().GetByID(ctx, userID)
	if err != nil {
		logger.LOG.Error("获取用户信息失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("用户不存在")
	}

	// 2. 验证文件是否存在
	userFile, err := d.factory.UserFiles().GetByUfID(ctx, req.FileID)
	if err != nil {
		logger.LOG.Error("获取用户文件信息失败", "error", err, "fileID", req.FileID)
		return nil, err
	}
	if !userFile.IsPublic && userFile.UserID != userID {
		logger.LOG.Warn("用户尝试下载非公开文件", "userID", userID, "fileID", req.FileID)
		return nil, fmt.Errorf("无权下载此文件")
	}
	fileInfo, err := d.factory.FileInfo().GetByID(ctx, userFile.FileID)
	if err != nil {
		logger.LOG.Error("文件不存在", "error", err, "fileID", req.FileID)
		return nil, fmt.Errorf("文件不存在")
	}

	// 3. 创建下载任务记录
	taskID := uuid.Must(uuid.NewV7()).String()
	task := &models.DownloadTask{
		ID:               taskID,
		UserID:           userID,
		Type:             enum.DownloadTaskTypeLocalFile.Value(),
		URL:              req.FileID, // 存储 uf_id 在URL字段
		FileName:         fileInfo.Name,
		FileSize:         int64(fileInfo.Size),
		VirtualPath:      "",    // 网盘下载不需要虚拟路径
		EnableEncryption: false, // 网盘文件下载不加密存储（文件本身可能已加密）
		State:            enum.DownloadTaskStateInit.Value(),
		TargetDir:        d.tempDir,
		CreateTime:       custom_type.Now(),
		UpdateTime:       custom_type.Now(),
	}

	if err := d.factory.DownloadTask().Create(ctx, task); err != nil {
		logger.LOG.Error("创建下载任务失败", "error", err, "userID", userID, "fileID", req.FileID)
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	// 保存真实的 file_id，用于异步任务
	realFileID := userFile.FileID

	// 4. 异步准备下载文件（解密+合并）
	go func() {
		// 更新任务状态为准备中
		task.State = enum.DownloadTaskStateDownloading.Value()
		task.UpdateTime = custom_type.Now()
		d.factory.DownloadTask().Update(context.Background(), task)

		opts := &download.LocalFileDownloadOptions{
			FilePassword: req.FilePassword,
		}

		result, err := download.PrepareLocalFileDownload(
			context.Background(),
			realFileID, // 使用真实的 file_id
			userID,
			d.tempDir,
			d.factory,
			opts,
		)

		if err != nil {
			// 准备失败
			task.State = enum.DownloadTaskStateFailed.Value()
			task.ErrorMsg = err.Error()
			task.UpdateTime = custom_type.Now()
			d.factory.DownloadTask().Update(context.Background(), task)
			logger.LOG.Error("准备下载文件失败", "taskID", taskID, "error", err)
			return
		}

		// 准备完成，更新任务状态为已完成（网盘文件下载准备完成即可下载）
		task.State = enum.DownloadTaskStateFinished.Value() // state=3 表示准备完成，可下载
		task.Progress = 100
		task.DownloadedSize = result.FileSize
		task.Path = result.TempFilePath // 存储临时文件路径
		task.UpdateTime = custom_type.Now()
		task.FinishTime = custom_type.Now()
		d.factory.DownloadTask().Update(context.Background(), task)

		logger.LOG.Info("网盘文件下载准备完成", "taskID", taskID, "realFileID", realFileID, "ufID", req.FileID, "tempPath", result.TempFilePath)
	}()

	logger.LOG.Info("网盘文件下载任务已创建", "taskID", taskID, "userID", userID, "ufID", req.FileID, "realFileID", realFileID)

	// 返回任务信息
	return models.NewJsonResponse(200, "任务创建成功", map[string]interface{}{
		"task_id":   taskID,
		"file_name": fileInfo.Name,
		"file_size": fileInfo.Size,
	}), nil
}

// ParseTorrent 解析种子/磁力链
func (d *DownloadService) ParseTorrent(req *request.ParseTorrentRequest) (*models.JsonResponse, error) {
	// 调用解析功能（超时120秒）
	result, err := download.ParseTorrent(req.Content, 120)
	if err != nil {
		logger.LOG.Error("解析种子失败", "error", err)
		return nil, fmt.Errorf("解析失败: %w", err)
	}

	// 转换为响应格式
	files := make([]response.TorrentFileInfo, 0, len(result.Files))
	for _, f := range result.Files {
		files = append(files, response.TorrentFileInfo{
			Index: f.Index,
			Name:  f.Name,
			Size:  f.Size,
			Path:  f.Path,
		})
	}

	resp := &response.ParseTorrentResponse{
		Name:      result.Name,
		InfoHash:  result.InfoHash,
		Files:     files,
		TotalSize: result.TotalSize,
	}

	logger.LOG.Info("种子解析成功",
		"name", result.Name,
		"infoHash", result.InfoHash,
		"fileCount", len(files),
		"totalSize", result.TotalSize,
	)

	return models.NewJsonResponse(200, "解析成功", resp), nil
}

// StartTorrentDownload 开始种子/磁力链下载
func (d *DownloadService) StartTorrentDownload(req *request.StartTorrentDownloadRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 1. 验证用户是否存在
	_, err := d.factory.User().GetByID(ctx, userID)
	if err != nil {
		logger.LOG.Error("获取用户信息失败", "error", err, "userID", userID)
		return nil, fmt.Errorf("用户不存在")
	}

	// 2. 解析种子获取元数据
	parseResult, err := download.ParseTorrent(req.Content, 120)
	if err != nil {
		logger.LOG.Error("解析种子失败", "error", err)
		return nil, fmt.Errorf("解析种子失败: %w", err)
	}

	// 3. 验证文件索引
	for _, idx := range req.FileIndexes {
		if idx < 0 || idx >= len(parseResult.Files) {
			return nil, fmt.Errorf("文件索引无效: %d", idx)
		}
	}

	// 4. 设置默认虚拟路径
	virtualPath := req.VirtualPath
	if virtualPath == "" {
		virtualPath = filepath.Join("离线下载/")
	}

	// 验证加密存储密码
	if req.EnableEncryption {
		if req.FilePassword == "" {
			return nil, fmt.Errorf("加密存储密码不能为空")
		}
	}

	// 5. 为每个文件创建下载任务
	taskIDs := make([]string, 0, len(req.FileIndexes))
	for _, fileIndex := range req.FileIndexes {
		fileInfo := parseResult.Files[fileIndex]
		taskID := uuid.Must(uuid.NewV7()).String()

		// 判断任务类型（磁力链或种子）
		taskType := enum.DownloadTaskTypeBtp.Value()
		if strings.HasPrefix(req.Content, "magnet:") {
			taskType = enum.DownloadTaskTypeMagnet.Value()
		}

		task := &models.DownloadTask{
			ID:               taskID,
			UserID:           userID,
			Type:             taskType,
			URL:              req.Content, // 存储种子内容或磁力链
			FileName:         fileInfo.Name,
			FileSize:         fileInfo.Size,
			VirtualPath:      virtualPath,
			EnableEncryption: req.EnableEncryption,
			InfoHash:         parseResult.InfoHash,
			FileIndex:        fileIndex,
			TorrentName:      parseResult.Name,
			State:            enum.DownloadTaskStateInit.Value(),
			TargetDir:        d.tempDir,
			CreateTime:       custom_type.Now(),
			UpdateTime:       custom_type.Now(),
		}

		if err := d.factory.DownloadTask().Create(ctx, task); err != nil {
			logger.LOG.Error("创建下载任务失败", "error", err, "userID", userID, "fileIndex", fileIndex)
			return nil, fmt.Errorf("创建任务失败: %w", err)
		}

		taskIDs = append(taskIDs, taskID)

		// 异步启动下载任务
		go func(tid string, fIndex int) {
			opts := &download.TorrentSingleFileDownloadOptions{
				MaxConcurrentPeers: 200, // 提高并发连接数以加速下载
				EnableEncryption:   req.EnableEncryption,
				VirtualPath:        virtualPath,
				TorrentName:        parseResult.Name,
				InfoHash:           parseResult.InfoHash,
				FilePassword:       req.FilePassword,
			}

			logger.LOG.Debug("启动种子下载任务", "taskID", tid, "tempDir", d.tempDir)

			fileID, err := download.DownloadTorrentSingleFile(
				context.Background(),
				tid,
				req.Content,
				fIndex,
				userID,
				d.tempDir,
				d.factory,
				opts,
			)

			if err != nil {
				logger.LOG.Error("种子文件下载失败", "taskID", tid, "error", err)
				// 获取最新任务状态，防止覆盖暂停状态
				task, _ := d.factory.DownloadTask().GetByID(context.Background(), tid)
				if task != nil && task.State != enum.DownloadTaskStatePaused.Value() {
					// 只有当任务不是暂停状态时，才更新为失败
					task.State = enum.DownloadTaskStateFailed.Value()
					task.ErrorMsg = err.Error()
					task.UpdateTime = custom_type.Now()
					d.factory.DownloadTask().Update(context.Background(), task)
				}
			} else {
				// 获取最新任务状态，防止覆盖暂停状态
				task, _ := d.factory.DownloadTask().GetByID(context.Background(), tid)
				if task != nil && task.State != enum.DownloadTaskStatePaused.Value() {
					// 只有当任务不是暂停状态时，才更新为完成
					task.FileID = fileID
					task.State = enum.DownloadTaskStateFinished.Value()
					task.Progress = 100
					task.UpdateTime = custom_type.Now()
					task.FinishTime = custom_type.Now()
					d.factory.DownloadTask().Update(context.Background(), task)
				}
				logger.LOG.Info("种子文件下载完成", "taskID", tid, "fileID", fileID)
			}
		}(taskID, fileIndex)
	}

	logger.LOG.Info("种子下载任务已创建",
		"torrentName", parseResult.Name,
		"infoHash", parseResult.InfoHash,
		"taskCount", len(taskIDs),
		"userID", userID,
	)

	// 返回任务信息
	resp := &response.StartTorrentDownloadResponse{
		TaskIDs:     taskIDs,
		TorrentName: parseResult.Name,
		TaskCount:   len(taskIDs),
	}

	return models.NewJsonResponse(200, "任务创建成功", resp), nil
}
