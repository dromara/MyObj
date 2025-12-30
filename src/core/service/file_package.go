package service

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"myobj/src/core/domain/request"
	"myobj/src/core/domain/response"
	"myobj/src/pkg/download"
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

// PackageTask 打包任务
type PackageTask struct {
	PackageID   string
	PackageName string
	UserID      string
	FileIDs     []string
	Status      string    // creating, ready, failed
	Progress    int       // 0-100
	TotalSize   int64
	CreatedSize int64
	FilePath    string
	ErrorMsg    string
	CreatedAt   time.Time // 创建时间，用于清理过期任务
	mu          sync.Mutex
}

// CreatePackage 创建打包下载任务
func (f *FileService) CreatePackage(req *request.PackageCreateRequest, userID string) (*models.JsonResponse, error) {
	ctx := context.Background()

	// 验证文件权限（前端传递的是 uf_id）
	for _, fileID := range req.FileIDs {
		userFile, err := f.factory.UserFiles().GetByUserIDAndUfID(ctx, userID, fileID)
		if err != nil {
			return nil, fmt.Errorf("文件不存在或无权限: %s", fileID)
		}
		if userFile.UserID != userID {
			return nil, fmt.Errorf("无权限访问文件: %s", fileID)
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

	// 异步创建压缩包
	go f.createZipPackage(ctx, task)

	// 计算总大小（前端传递的是 uf_id）
	for _, fileID := range req.FileIDs {
		userFile, _ := f.factory.UserFiles().GetByUserIDAndUfID(ctx, userID, fileID)
		if userFile != nil {
			fileInfo, _ := f.factory.FileInfo().GetByID(ctx, userFile.FileID)
			if fileInfo != nil {
				task.TotalSize += int64(fileInfo.Size)
			}
		}
	}

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

	// 逐个添加文件到ZIP
	totalFiles := len(task.FileIDs)
	for i, fileID := range task.FileIDs {
		// 更新进度
		task.mu.Lock()
		task.Progress = int((i + 1) * 100 / totalFiles)
		task.mu.Unlock()

		// 获取用户文件（前端传递的是 uf_id）
		userFile, err := f.factory.UserFiles().GetByUserIDAndUfID(ctx, task.UserID, fileID)
		if err != nil {
			logger.LOG.Warn("获取用户文件失败", "fileID", fileID, "error", err)
			continue
		}

		// 获取文件信息
		fileInfo, err := f.factory.FileInfo().GetByID(ctx, userFile.FileID)
		if err != nil {
			logger.LOG.Warn("获取文件信息失败", "fileID", fileID, "error", err)
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
		// 这里暂时不清理，由系统定期清理或下载完成后清理
		// 如果需要立即清理，可以通过提取路径判断
		if downloadResult.TempFilePath != fileInfo.Path && strings.Contains(downloadResult.TempFilePath, "temp") {
			// 提取临时目录并清理
			tempDir := filepath.Dir(downloadResult.TempFilePath)
			if strings.Contains(tempDir, "temp") {
				defer os.RemoveAll(tempDir)
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
		// 再次检查任务状态，如果还是 ready，则清理文件
		task.mu.Lock()
		if task.Status == "ready" && task.FilePath != "" {
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
			// 从任务列表中移除
			packageTasks.Delete(packageID)
		}
		task.mu.Unlock()
	}()

	return task.FilePath, task.PackageName, nil
}

