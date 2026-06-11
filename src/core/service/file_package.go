package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/download"
	"myobj/src/pkg/enum"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/models"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// 打包任务状态管理
var packageTasks sync.Map // key: packageID, value: *PackageTask

func init() {
	// 定期清理已完成/失败超过1小时的打包任务
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			packageTasks.Range(func(key, value interface{}) bool {
				task := value.(*PackageTask)
				task.mu.Lock()
				isTerminal := task.Status == "ready" || task.Status == "failed"
				createdAt := task.CreatedAt
				filePath := task.FilePath
				task.mu.Unlock()
				if isTerminal && time.Since(createdAt) > time.Hour {
					// 清理临时文件
					if filePath != "" {
						tempDir := filepath.Dir(filePath)
						os.RemoveAll(tempDir)
					}
					packageTasks.Delete(key)
				}
				return true
			})
		}
	}()
}

// PackageTask 打包任务
type PackageTask struct {
	PackageID   string
	PackageName string
	UserID      string
	FileIDs     []string
	Status      string // creating, ready, failed
	Progress    int    // 0-100
	TotalSize   int64
	CreatedSize int64
	FilePath    string
	ErrorMsg    string
	CreatedAt   time.Time // 创建时间，用于清理过期任务
	mu          sync.Mutex
	cleanupOnce sync.Once
}

// CreatePackage 创建打包下载任务
func (f *FileService) CreatePackage(req *request.PackageCreateRequest, userID string) (*models.JsonResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 批量验证文件权限（前端传递的是 uf_id）
	userFilesMap, err := f.factory.UserFiles().BatchGetByUserIDAndUfIDs(ctx, userID, req.FileIDs)
	if err != nil {
		return nil, fmt.Errorf("批量查询用户文件失败: %v", err)
	}
	for _, fileID := range req.FileIDs {
		if _, ok := userFilesMap[fileID]; !ok {
			return nil, fmt.Errorf("文件不存在或无权限: %s", fileID)
		}
	}

	// 生成打包ID
	packageID := uuid.New().String()

	// 设置打包名称
	packageName := req.PackageName
	if packageName == "" {
		packageName = fmt.Sprintf("files_%d.zip", time.Now().Unix())
	}
	if !strings.HasSuffix(packageName, ".zip") {
		packageName += ".zip"
	}

	// 创建打包任务
	task := &PackageTask{
		PackageID:   packageID,
		PackageName: packageName,
		UserID:      userID,
		FileIDs:     req.FileIDs,
		Status:      "creating",
		Progress:    0,
		TotalSize:   0,
		CreatedSize: 0,
		CreatedAt:   time.Now(),
	}
	packageTasks.Store(packageID, task)

	// 批量计算总大小，避免N+1查询
	fileIDs := make([]string, 0, len(userFilesMap))
	for _, uf := range userFilesMap {
		fileIDs = append(fileIDs, uf.FileID)
	}
	if len(fileIDs) > 0 {
		fileInfoMap, batchErr := f.factory.FileInfo().BatchGetByIDs(ctx, fileIDs)
		if batchErr != nil {
			logger.LOG.Warn("批量获取文件信息失败", "error", batchErr)
		} else {
			for _, uf := range userFilesMap {
				if fi, ok := fileInfoMap[uf.FileID]; ok && fi != nil {
					task.TotalSize += int64(fi.Size)
				}
			}
		}
	}

	// 异步创建压缩包（使用独立context，不继承超时）
	go f.createZipPackage(context.Background(), task)

	return models.NewJsonResponse(200, "创建成功", response.PackageCreateResponse{
		PackageID:   packageID,
		PackageName: packageName,
		Status:      task.Status,
		Progress:    task.Progress,
		TotalSize:   task.TotalSize,
	}), nil
}

// createZipPackage 异步创建ZIP压缩包
func (f *FileService) createZipPackage(ctx context.Context, task *PackageTask) {
	defer func() {
		if r := recover(); r != nil {
			task.mu.Lock()
			task.Status = "failed"
			task.ErrorMsg = fmt.Sprintf("打包失败: %v", r)
			task.mu.Unlock()
			logger.LOG.Error("打包任务异常", "packageID", task.PackageID, "error", r)
		}
	}()

	// 创建临时目录
	tempDir := filepath.Join(os.TempDir(), "package_"+task.PackageID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		task.mu.Lock()
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("创建临时目录失败: %v", err)
		task.mu.Unlock()
		return
	}
	// 注意：不要在这里立即删除临时目录，因为文件需要保留供下载使用
	// 文件会在下载完成后或任务过期后清理

	// 创建ZIP文件
	zipPath := filepath.Join(tempDir, task.PackageName)
	zipFile, err := os.Create(zipPath)
	if err != nil {
		task.mu.Lock()
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("创建ZIP文件失败: %v", err)
		task.mu.Unlock()
		return
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// 批量预加载用户文件和文件信息，避免N+1查询
	zipUserFilesMap, err := f.factory.UserFiles().BatchGetByUserIDAndUfIDs(ctx, task.UserID, task.FileIDs)
	if err != nil {
		task.mu.Lock()
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("批量查询用户文件失败: %v", err)
		task.mu.Unlock()
		return
	}
	zipFileIDs := make([]string, 0, len(zipUserFilesMap))
	for _, uf := range zipUserFilesMap {
		zipFileIDs = append(zipFileIDs, uf.FileID)
	}
	zipFileInfoMap, err := f.factory.FileInfo().BatchGetByIDs(ctx, zipFileIDs)
	if err != nil {
		task.mu.Lock()
		task.Status = "failed"
		task.ErrorMsg = fmt.Sprintf("批量查询文件信息失败: %v", err)
		task.mu.Unlock()
		return
	}

	// 逐个添加文件到ZIP
	totalFiles := len(task.FileIDs)
	for i, fileID := range task.FileIDs {
		// 更新进度
		task.mu.Lock()
		task.Progress = int((i + 1) * 100 / totalFiles)
		task.mu.Unlock()

		// 从预加载的map中获取用户文件（前端传递的是 uf_id）
		userFile, ok := zipUserFilesMap[fileID]
		if !ok {
			logger.LOG.Warn("获取用户文件失败，文件不存在或无权限", "fileID", fileID)
			continue
		}

		// 从预加载的map中获取文件信息
		fileInfo, ok := zipFileInfoMap[userFile.FileID]
		if !ok {
			logger.LOG.Warn("获取文件信息失败", "fileID", fileID)
			continue
		}

		// 准备文件下载（处理加密和分片）
		// 注意：PrepareLocalFileDownload 需要的是 file_id（file_info表的ID），不是 uf_id
		downloadResult, err := download.PrepareLocalFileDownload(
			ctx,
			userFile.FileID, // 使用 file_info 表的 ID
			task.UserID,
			tempDir,
			f.factory,
			&download.LocalFileDownloadOptions{},
		)
		if err != nil {
			logger.LOG.Warn("准备文件下载失败", "fileID", fileID, "error", err)
			continue
		}

		// 打开文件
		sourceFile, err := os.Open(downloadResult.TempFilePath)
		if err != nil {
			logger.LOG.Warn("打开文件失败", "fileID", fileID, "error", err)
			continue
		}

		// 创建ZIP中的文件条目
		zipEntry, err := zipWriter.Create(userFile.FileName)
		if err != nil {
			sourceFile.Close()
			logger.LOG.Warn("创建ZIP条目失败", "fileID", fileID, "error", err)
			continue
		}

		// 复制文件内容到ZIP
		written, err := io.Copy(zipEntry, sourceFile)
		if err != nil {
			sourceFile.Close()
			logger.LOG.Warn("复制文件到ZIP失败", "fileID", fileID, "error", err)
			continue
		}

		sourceFile.Close()
		task.mu.Lock()
		task.CreatedSize += written
		task.mu.Unlock()

		// 清理临时文件（如果是PrepareLocalFileDownload创建的临时文件）
		// 注意：PrepareLocalFileDownload 创建的临时文件在磁盘的 temp 目录下
		// 如果需要立即清理，可以通过提取路径判断
		if downloadResult.TempFilePath != fileInfo.Path && strings.Contains(downloadResult.TempFilePath, "temp") {
			// 提取临时目录并清理（不使用 defer，避免在循环内延迟执行）
			tmpDir := filepath.Dir(downloadResult.TempFilePath)
			if strings.Contains(tmpDir, "temp") {
				os.RemoveAll(tmpDir)
			}
		}
	}

	// 完成打包
	task.mu.Lock()
	task.Status = "ready"
	task.Progress = 100
	task.FilePath = zipPath
	task.mu.Unlock()

	logger.LOG.Info("打包完成", "packageID", task.PackageID, "filePath", zipPath)

	// 为每个文件创建下载任务记录
	f.createDownloadTasksForPackage(ctx, task)
}

// GetPackageProgress 获取打包进度
func (f *FileService) GetPackageProgress(packageID, userID string) (*models.JsonResponse, error) {
	value, ok := packageTasks.Load(packageID)
	if !ok {
		return nil, fmt.Errorf("打包任务不存在")
	}

	task := value.(*PackageTask)
	if task.UserID != userID {
		return nil, fmt.Errorf("无权限访问该打包任务")
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	return models.NewJsonResponse(200, "查询成功", response.PackageProgressResponse{
		PackageID:   task.PackageID,
		Status:      task.Status,
		Progress:    task.Progress,
		TotalSize:   task.TotalSize,
		CreatedSize: task.CreatedSize,
		ErrorMsg:    task.ErrorMsg,
	}), nil
}

// DownloadPackage 下载打包文件
func (f *FileService) DownloadPackage(packageID, userID string) (string, string, error) {
	value, ok := packageTasks.Load(packageID)
	if !ok {
		return "", "", fmt.Errorf("打包任务不存在")
	}

	task := value.(*PackageTask)
	if task.UserID != userID {
		return "", "", fmt.Errorf("无权限访问该打包任务")
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	if task.Status != "ready" {
		return "", "", fmt.Errorf("打包任务未完成，状态: %s", task.Status)
	}

	if task.FilePath == "" {
		return "", "", fmt.Errorf("打包文件路径不存在")
	}

	// 检查文件是否存在
	if _, err := os.Stat(task.FilePath); os.IsNotExist(err) {
		return "", "", fmt.Errorf("打包文件不存在")
	}

	// 下载完成后，异步清理文件（延迟5分钟，给用户足够时间下载）
	go func() {
		time.Sleep(5 * time.Minute)
		task.cleanupOnce.Do(func() {
			task.mu.Lock()
			defer task.mu.Unlock()
			if task.FilePath != "" {
				// 删除文件
				if err := os.Remove(task.FilePath); err != nil {
					logger.LOG.Warn("删除打包文件失败", "packageID", packageID, "filePath", task.FilePath, "error", err)
				} else {
					logger.LOG.Info("打包文件已清理", "packageID", packageID, "filePath", task.FilePath)
				}
				// 删除临时目录
				tempDir := filepath.Dir(task.FilePath)
				if err := os.RemoveAll(tempDir); err != nil {
					logger.LOG.Warn("删除临时目录失败", "packageID", packageID, "tempDir", tempDir, "error", err)
				}
				task.FilePath = ""
			}
			// 从任务列表中移除
			packageTasks.Delete(packageID)
		})
	}()

	return task.FilePath, task.PackageName, nil
}

// createDownloadTasksForPackage 为打包中的每个文件创建下载任务记录
func (f *FileService) createDownloadTasksForPackage(ctx context.Context, task *PackageTask) {
	// 批量预加载用户文件和文件信息，避免N+1查询
	dlUserFilesMap, err := f.factory.UserFiles().BatchGetByUserIDAndUfIDs(ctx, task.UserID, task.FileIDs)
	if err != nil {
		logger.LOG.Error("批量查询用户文件失败", "packageID", task.PackageID, "error", err)
		return
	}
	dlFileIDs := make([]string, 0, len(dlUserFilesMap))
	for _, uf := range dlUserFilesMap {
		dlFileIDs = append(dlFileIDs, uf.FileID)
	}
	dlFileInfoMap, err := f.factory.FileInfo().BatchGetByIDs(ctx, dlFileIDs)
	if err != nil {
		logger.LOG.Error("批量查询文件信息失败", "packageID", task.PackageID, "error", err)
		return
	}

	for _, fileID := range task.FileIDs {
		// 从预加载的map中获取用户文件（前端传递的是 uf_id）
		userFile, ok := dlUserFilesMap[fileID]
		if !ok {
			logger.LOG.Warn("获取用户文件失败，文件不存在或无权限", "fileID", fileID)
			continue
		}

		// 从预加载的map中获取文件信息
		fileInfo, ok := dlFileInfoMap[userFile.FileID]
		if !ok {
			logger.LOG.Warn("获取文件信息失败", "fileID", fileID)
			continue
		}

		// 创建下载任务记录
		taskID := uuid.Must(uuid.NewV7()).String()
		downloadTask := &models.DownloadTask{
			ID:               taskID,
			UserID:           task.UserID,
			Type:             enum.DownloadTaskTypePackage.Value(),
			URL:              task.PackageID, // 在URL字段存储打包ID
			FileName:         userFile.FileName,
			FileSize:         int64(fileInfo.Size),
			FileID:           userFile.FileID,
			VirtualPath:      "", // 打包下载不需要虚拟路径
			EnableEncryption: false,
			State:            enum.DownloadTaskStateFinished.Value(), // 直接设置为已完成
			Progress:         100,
			DownloadedSize:   int64(fileInfo.Size),
			CreateTime:       custom_type.Now(),
			UpdateTime:       custom_type.Now(),
			FinishTime:       custom_type.Now(),
		}

		if err := f.factory.DownloadTask().Create(ctx, downloadTask); err != nil {
			logger.LOG.Error("创建打包下载任务记录失败", "fileID", fileID, "error", err)
			continue
		}

		logger.LOG.Info("创建打包下载任务记录成功", "taskID", taskID, "fileName", userFile.FileName, "packageID", task.PackageID)
	}
}
