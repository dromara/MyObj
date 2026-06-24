package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	"myobj/src/pkg/upload"
	"myobj/src/pkg/util"
)

// CloudTransferService 云盘文件转存服务
type CloudTransferService struct {
	factory      *impl.RepositoryFactory
	providers    map[string]cloud.CloudProvider
	aliyunTokens *aliyun.AliyunTokenStore
}

// NewCloudTransferService 创建云盘转存服务
func NewCloudTransferService(factory *impl.RepositoryFactory) *CloudTransferService {
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

	return &CloudTransferService{
		factory:      factory,
		providers:    providers,
		aliyunTokens: aliyun.NewAliyunTokenStore(),
	}
}

// GetRepository 获取仓储工厂
func (s *CloudTransferService) GetRepository() *impl.RepositoryFactory {
	return s.factory
}

// TransferFromShareResult 转存结果
type TransferFromShareResult struct {
	TotalFiles   int      `json:"total_files"`
	SuccessFiles int      `json:"success_files"`
	FailedFiles  int      `json:"failed_files"`
	FailedNames  []string `json:"failed_names,omitempty"`
}

// TransferFromShare 直接从分享链接转存文件到本地
func (s *CloudTransferService) TransferFromShare(ctx context.Context, provider, shareID, shareURL, sharePwd string, fileIDs []string, userID string) (*TransferFromShareResult, error) {
	// 获取云盘提供者
	p, ok := s.providers[provider]
	if !ok {
		return nil, fmt.Errorf("不支持的云盘类型: %s", provider)
	}

	// 如果是阿里云盘，设置access_token
	if provider == string(cloud.ProviderAliyun) {
		if ap, ok := p.(*aliyun.AliyunProvider); ok {
			if token, err := s.aliyunTokens.Get(userID); err == nil {
				ap.SetAccessToken(token.AccessToken)
			}
		}
	}

	// 从URL提取folderID（如果有）
	folderID := extractFolderID(shareURL)

	// 递归获取所有文件（带路径信息）
	allFiles, err := s.listAllFiles(ctx, p, shareID, folderID)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}

	// 如果指定了文件ID，只转存指定的文件
	var targetFiles []CloudFileWithPath
	if len(fileIDs) > 0 {
		idSet := make(map[string]bool)
		for _, id := range fileIDs {
			idSet[id] = true
		}
		for _, f := range allFiles {
			if idSet[f.FileID] {
				targetFiles = append(targetFiles, f)
			}
		}
	} else {
		targetFiles = allFiles
	}

	if len(targetFiles) == 0 {
		return &TransferFromShareResult{}, nil
	}

	result := &TransferFromShareResult{
		TotalFiles: len(targetFiles),
	}

	// 为每个文件创建下载任务，写入 download_task 表
	now := custom_type.Now()
	for _, file := range targetFiles {
		taskID := fmt.Sprintf("cloud-%d", time.Now().UnixNano())
		downloadURL, _ := p.GetDownloadLink(ctx, shareID, file.FileID)

		url := ""
		if downloadURL != nil {
			url = downloadURL.URL
		}

		// 保留目录结构：VirtualPath = / + 相对路径的目录部分
		virtualPath := "/"
		if file.RelativePath != "" {
			// 提取目录部分，例如 "A/B/小说.txt" -> "/A/B"
			dir := filepath.Dir(file.RelativePath)
			if dir != "." {
				virtualPath = "/" + dir
			}
		}

		task := &models.DownloadTask{
			ID:             taskID,
			UserID:         userID,
			FileID:         file.FileID,
			FileName:       file.Name,
			FileSize:       file.Size,
			DownloadedSize: 0,
			Progress:       0,
			Type:           7, // 7 = 云盘转存
			URL:            url,
			Path:           "",
			VirtualPath:    virtualPath,
			State:          0, // 等待中
			CreateTime:     now,
			UpdateTime:     now,
		}

		if err := s.factory.DB().Create(task).Error; err != nil {
			logger.LOG.Error("创建下载任务失败", "error", err, "file", file.Name)
			result.FailedFiles++
			result.FailedNames = append(result.FailedNames, file.Name)
			continue
		}

		// 后台执行下载
		go s.executeDownloadTask(task, file.ShareFile, shareID, p)

		result.SuccessFiles++
	}

	return result, nil
}

// executeDownloadTask 执行单个下载任务
func (s *CloudTransferService) executeDownloadTask(task *models.DownloadTask, file cloud.ShareFile, shareID string, p cloud.CloudProvider) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger.LOG.Info("开始执行云盘转存任务", "taskID", task.ID, "file", task.FileName)

	// 更新状态为下载中
	s.updateTaskState(task.ID, 1, 0, "")

	// 获取下载链接
	downloadInfo, err := p.GetDownloadLink(ctx, shareID, file.FileID)
	if err != nil {
		logger.LOG.Error("获取下载链接失败", "error", err, "file", file.Name)
		s.updateTaskState(task.ID, 4, 0, "获取下载链接失败: "+err.Error())
		return
	}

	// 创建临时文件
	tmpPath := fmt.Sprintf("./obj_temp/%s_%s", task.ID, file.Name)

	// 下载文件（带进度更新，使用各网盘特定的请求头）
	if err := s.downloadWithProgress(ctx, downloadInfo.URL, tmpPath, task.ID, file.Size, downloadInfo.Headers); err != nil {
		logger.LOG.Error("下载文件失败", "error", err, "file", file.Name)
		s.updateTaskState(task.ID, 4, 0, "下载失败: "+err.Error())
		os.Remove(tmpPath)
		return
	}

	// 保存到本地存储（保留目录结构）
	_, err = upload.ProcessUploadedFile(&upload.FileUploadData{
		TempFilePath: tmpPath,
		FileName:     file.Name,
		FileSize:     file.Size,
		UserID:       task.UserID,
		VirtualPath:  task.VirtualPath,
	}, s.factory)
	if err != nil {
		logger.LOG.Error("保存文件失败", "error", err, "file", file.Name)
		s.updateTaskState(task.ID, 4, 0, "保存失败: "+err.Error())
		return
	}

	// 更新状态为完成
	s.updateTaskState(task.ID, 3, 100, "")
	logger.LOG.Info("云盘转存任务完成", "taskID", task.ID, "file", file.Name)
}

// downloadWithProgress 带进度更新的下载
func (s *CloudTransferService) downloadWithProgress(ctx context.Context, url, filePath, taskID string, totalSize int64, headers map[string]string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	// 设置通用请求头
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	// 设置各网盘特定的请求头
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("下载失败: status=%d", resp.StatusCode)
	}

	out, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer out.Close()

	buf := make([]byte, 32*1024)
	var downloaded int64
	lastUpdate := time.Now()

	for {
		n, readErr := resp.Body.Read(buf)
		if n > 0 {
			if _, writeErr := out.Write(buf[:n]); writeErr != nil {
				return writeErr
			}
			downloaded += int64(n)

			// 每秒更新一次进度
			if time.Since(lastUpdate) > time.Second {
				progress := 0
				if totalSize > 0 {
					progress = int(downloaded * 100 / totalSize)
				}
				s.updateTaskProgress(taskID, downloaded, progress)
				lastUpdate = time.Now()
			}
		}
		if readErr != nil {
			if readErr == io.EOF {
				break
			}
			return readErr
		}
	}

	// 最终更新
	s.updateTaskProgress(taskID, downloaded, 100)
	return nil
}

// updateTaskState 更新任务状态
func (s *CloudTransferService) updateTaskState(taskID string, state int, progress int, errMsg string) {
	updates := map[string]interface{}{
		"state":        state,
		"progress":     progress,
		"error_msg":    errMsg,
		"update_time":  time.Now(),
	}
	if state == 3 || state == 4 { // 完成或失败
		now := time.Now()
		updates["finish_time"] = &now
	}
	s.factory.DB().Table("download_task").Where("id = ?", taskID).Updates(updates)
}

// updateTaskProgress 更新任务进度
func (s *CloudTransferService) updateTaskProgress(taskID string, downloadedSize int64, progress int) {
	s.factory.DB().Table("download_task").Where("id = ?", taskID).Updates(map[string]interface{}{
		"downloaded_size": downloadedSize,
		"progress":        progress,
		"update_time":     time.Now(),
	})
}

// CloudFileWithPath 带路径的云盘文件
type CloudFileWithPath struct {
	cloud.ShareFile
	RelativePath string // 相对路径，如 "A/B/小说.txt"
}

// listAllFiles 递归获取所有文件，保留目录结构
func (s *CloudTransferService) listAllFiles(ctx context.Context, p cloud.CloudProvider, shareID, parentFileID string) ([]CloudFileWithPath, error) {
	return s.listAllFilesWithPrefix(ctx, p, shareID, parentFileID, "")
}

// listAllFilesWithPrefix 递归获取所有文件，带路径前缀
func (s *CloudTransferService) listAllFilesWithPrefix(ctx context.Context, p cloud.CloudProvider, shareID, parentFileID, prefix string) ([]CloudFileWithPath, error) {
	files, err := p.ListShareFiles(ctx, shareID, parentFileID)
	if err != nil {
		return nil, err
	}

	var allFiles []CloudFileWithPath
	for _, f := range files {
		if f.IsDir {
			// 递归获取子目录文件，路径前缀加上当前目录名
			subPrefix := f.Name
			if prefix != "" {
				subPrefix = prefix + "/" + f.Name
			}
			subFiles, err := s.listAllFilesWithPrefix(ctx, p, shareID, f.FileID, subPrefix)
			if err != nil {
				logger.LOG.Error("获取子目录文件失败", "error", err, "dir", f.Name)
				continue
			}
			allFiles = append(allFiles, subFiles...)
		} else {
			// 文件的相对路径
			relativePath := f.Name
			if prefix != "" {
				relativePath = prefix + "/" + f.Name
			}
			allFiles = append(allFiles, CloudFileWithPath{
				ShareFile:    f,
				RelativePath: relativePath,
			})
		}
	}

	return allFiles, nil
}

// extractFolderID 从URL中提取folderID
func extractFolderID(shareURL string) string {
	if !strings.Contains(shareURL, "/folder/") {
		return ""
	}
	parts := strings.Split(shareURL, "/folder/")
	if len(parts) < 2 {
		return ""
	}
	folderID := parts[1]
	if idx := strings.Index(folderID, "?"); idx > 0 {
		folderID = folderID[:idx]
	}
	if idx := strings.Index(folderID, "#"); idx > 0 {
		folderID = folderID[:idx]
	}
	return folderID
}

// TransferFile 转存单个文件到本地存储
func (s *CloudTransferService) TransferFile(ctx context.Context, taskID, fileID int, userID string) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Minute)
	defer cancel()

	// 1. 获取任务信息
	task, err := s.factory.CloudTask().GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}
	if task.UserID != userID {
		return fmt.Errorf("无权访问此任务")
	}

	// 2. 获取任务文件信息
	taskFile, err := s.factory.CloudTaskFile().GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("获取任务文件失败: %w", err)
	}
	if taskFile.TaskID != taskID {
		return fmt.Errorf("文件不属于此任务")
	}
	if taskFile.IsDir {
		return fmt.Errorf("不支持转存目录，请使用转存任务接口")
	}

	// 3. 获取云盘提供者
	provider, ok := s.providers[task.Provider]
	if !ok {
		return fmt.Errorf("不支持的云盘类型: %s", task.Provider)
	}

	// 4. 下载并保存文件
	localFileID, err := s.downloadAndSave(ctx, provider, task.ShareID, taskFile, userID, task.TargetPath)
	if err != nil {
		// 更新文件状态为失败
		_ = s.factory.CloudTaskFile().UpdateStatus(ctx, fileID, models.CloudFileStatusFailed, err.Error())
		return fmt.Errorf("下载并保存文件失败: %w", err)
	}

	// 5. 更新文件状态为完成
	if err := s.factory.CloudTaskFile().UpdateLocalPath(ctx, fileID, "", localFileID); err != nil {
		logger.LOG.Warn("更新任务文件状态失败", "error", err, "fileID", fileID)
	}

	logger.LOG.Info("单文件转存成功", "taskID", taskID, "fileID", fileID, "localFileID", localFileID)
	return nil
}

// TransferTask 转存任务所有文件到本地存储
func (s *CloudTransferService) TransferTask(ctx context.Context, taskID int, userID string) error {
	// 1. 获取任务信息
	task, err := s.factory.CloudTask().GetByID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取任务失败: %w", err)
	}
	if task.UserID != userID {
		return fmt.Errorf("无权访问此任务")
	}

	// 2. 更新任务状态为处理中
	if err := s.factory.CloudTask().UpdateStatus(ctx, taskID, models.CloudTaskStatusProcessing, ""); err != nil {
		logger.LOG.Warn("更新任务状态失败", "error", err)
	}

	// 3. 获取待处理文件列表
	pendingFiles, err := s.factory.CloudTaskFile().ListPendingByTaskID(ctx, taskID)
	if err != nil {
		return fmt.Errorf("获取待处理文件列表失败: %w", err)
	}

	if len(pendingFiles) == 0 {
		_ = s.factory.CloudTask().UpdateStatus(ctx, taskID, models.CloudTaskStatusCompleted, "")
		return nil
	}

	// 4. 获取云盘提供者
	provider, ok := s.providers[task.Provider]
	if !ok {
		_ = s.factory.CloudTask().UpdateStatus(ctx, taskID, models.CloudTaskStatusFailed, fmt.Sprintf("不支持的云盘类型: %s", task.Provider))
		return fmt.Errorf("不支持的云盘类型: %s", task.Provider)
	}

	// 5. 逐个处理文件
	successCount := 0
	failedCount := 0
	for _, taskFile := range pendingFiles {
		// 跳过目录
		if taskFile.IsDir {
			continue
		}

		// 更新文件状态为下载中
		_ = s.factory.CloudTaskFile().UpdateStatus(ctx, taskFile.ID, models.CloudFileStatusDownloading, "")

		localFileID, err := s.downloadAndSave(ctx, provider, task.ShareID, taskFile, userID, task.TargetPath)
		if err != nil {
			logger.LOG.Error("转存文件失败", "error", err, "taskID", taskID, "fileID", taskFile.ID, "fileName", taskFile.FileName)
			_ = s.factory.CloudTaskFile().UpdateStatus(ctx, taskFile.ID, models.CloudFileStatusFailed, err.Error())
			failedCount++
			continue
		}

		// 更新文件状态为完成
		if err := s.factory.CloudTaskFile().UpdateLocalPath(ctx, taskFile.ID, "", localFileID); err != nil {
			logger.LOG.Warn("更新任务文件状态失败", "error", err, "fileID", taskFile.ID)
		}
		successCount++

		// 更新任务进度
		_ = s.factory.CloudTask().UpdateProgress(ctx, taskID, successCount, failedCount)
	}

	// 6. 更新任务最终状态
	finalStatus := models.CloudTaskStatusCompleted
	errorMsg := ""
	if failedCount > 0 && successCount == 0 {
		finalStatus = models.CloudTaskStatusFailed
		errorMsg = fmt.Sprintf("全部失败，成功 %d，失败 %d", successCount, failedCount)
	} else if failedCount > 0 {
		errorMsg = fmt.Sprintf("部分完成，成功 %d，失败 %d", successCount, failedCount)
	}
	_ = s.factory.CloudTask().UpdateStatus(ctx, taskID, finalStatus, errorMsg)
	_ = s.factory.CloudTask().UpdateProgress(ctx, taskID, successCount, failedCount)

	logger.LOG.Info("任务转存完成", "taskID", taskID, "success", successCount, "failed", failedCount)
	return nil
}

// downloadAndSave 下载云盘文件并保存到本地存储
func (s *CloudTransferService) downloadAndSave(ctx context.Context, provider cloud.CloudProvider, shareID string, taskFile *models.CloudTaskFile, userID, targetPath string) (string, error) {
	// 1. 获取下载链接
	downloadInfo, err := provider.GetDownloadLink(ctx, shareID, taskFile.FileID)
	if err != nil {
		return "", fmt.Errorf("获取下载链接失败: %w", err)
	}

	// 2. 选择存储磁盘，创建临时目录
	disks, err := s.factory.Disk().List(ctx, 0, 1000)
	if err != nil {
		return "", fmt.Errorf("查询磁盘列表失败: %w", err)
	}
	if len(disks) == 0 {
		return "", fmt.Errorf("没有可用的存储磁盘")
	}

	// 选择剩余空间最大且能容纳文件的磁盘
	var bestDisk *models.Disk
	var maxFreeSpace int64 = -1
	for _, disk := range disks {
		freeSpaceBytes, err := util.GetDiskFreeSpaceByPath(disk.DataPath)
		if err != nil {
			continue
		}
		if freeSpaceBytes >= taskFile.FileSize && freeSpaceBytes > maxFreeSpace {
			maxFreeSpace = freeSpaceBytes
			bestDisk = disk
		}
	}
	if bestDisk == nil {
		return "", fmt.Errorf("没有足够空间的磁盘")
	}

	// 3. 创建临时目录并下载文件
	tempDir := util.BuildTempDir(bestDisk.DataPath, "cloud_transfer")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("创建临时目录失败: %w", err)
	}

	tempFilePath := filepath.Join(tempDir, taskFile.FileName)
	if err := s.downloadFile(ctx, downloadInfo.URL, tempFilePath); err != nil {
		os.RemoveAll(tempFilePath)
		return "", fmt.Errorf("下载文件失败: %w", err)
	}

	// 4. 使用 upload.ProcessUploadedFile 保存到本地存储
	// 确定虚拟路径
	virtualPath := targetPath
	if virtualPath == "" {
		virtualPath = ""
	}

	uploadData := &upload.FileUploadData{
		TempFilePath: tempFilePath,
		FileName:     taskFile.FileName,
		FileSize:     taskFile.FileSize,
		UserID:       userID,
		VirtualPath:  virtualPath,
		SkipCleanup:  false,
	}

	localFileID, err := upload.ProcessUploadedFile(uploadData, s.factory)
	if err != nil {
		return "", fmt.Errorf("保存文件到本地存储失败: %w", err)
	}

	return localFileID, nil
}

// downloadFile 下载文件到本地路径
func (s *CloudTransferService) downloadFile(ctx context.Context, url, destPath string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	client := &http.Client{Timeout: 30 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，HTTP状态码: %d", resp.StatusCode)
	}

	outFile, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("创建本地文件失败: %w", err)
	}
	defer outFile.Close()

	written, err := io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("写入文件失败: %w", err)
	}

	logger.LOG.Info("文件下载完成", "path", destPath, "size", written)
	return nil
}

// GetTransferStatus 获取转存状态
func (s *CloudTransferService) GetTransferStatus(ctx context.Context, taskID int, userID string) (map[string]interface{}, error) {
	task, err := s.factory.CloudTask().GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取任务失败: %w", err)
	}
	if task.UserID != userID {
		return nil, fmt.Errorf("无权访问此任务")
	}

	fileCounts, _ := s.factory.CloudTaskFile().CountByStatus(ctx, taskID)

	now := custom_type.Now()
	return map[string]interface{}{
		"task_id":        task.ID,
		"provider":       task.Provider,
		"share_id":       task.ShareID,
		"status":         task.Status,
		"status_text":    getStatusText(task.Status),
		"file_count":     task.FileCount,
		"total_size":     task.TotalSize,
		"success_count":  task.SuccessCount,
		"failed_count":   task.FailedCount,
		"target_path":    task.TargetPath,
		"error_message":  task.ErrorMsg,
		"created_at":     task.CreatedAt,
		"updated_at":     task.UpdatedAt,
		"completed_at":   task.CompletedAt,
		"file_counts":    fileCounts,
		"now":            now,
	}, nil
}
