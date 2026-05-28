package download

import (
	"context"
	"fmt"
	"io"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/cloudsync"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/enum"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/upload"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CloudDownloadOptions 云盘下载配置
type CloudDownloadOptions struct {
	Provider           string // 云盘类型（quark, baidu, aliyun等）
	Cookie             string // 云盘凭据
	FileID             string // 云盘文件ID
	EnableEncryption   bool   // 是否加密存储
	VirtualPath        string // 虚拟保存路径
	FilePassword       string // 加密文件密码
	Timeout            int    // 超时时间（秒），默认300
	OnCredentialUpdate func(string)
}

// DownloadCloud 从云盘下载文件到 MyObj
func DownloadCloud(
	taskID string,
	userID string,
	tempDir string,
	repoFactory *impl.RepositoryFactory,
	opts *CloudDownloadOptions,
) (*HTTPDownloadResult, error) {
	// 创建可取消的 context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册取消函数（复用全局 downloadTasks map）
	downloadTasksMu.Lock()
	downloadTasks[taskID] = cancel
	downloadTasksMu.Unlock()

	defer func() {
		downloadTasksMu.Lock()
		delete(downloadTasks, taskID)
		downloadTasksMu.Unlock()
	}()

	if opts == nil {
		return nil, fmt.Errorf("下载配置不能为空")
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 300
	}

	// 1. 获取云盘提供者
	provider, err := cloudsync.OpenProvider(opts.Provider, opts.Cookie, cloudsync.SessionOptions{
		OnCredentialUpdate: opts.OnCredentialUpdate,
	})
	if err != nil {
		updateTaskFailed(repoFactory, taskID, err.Error())
		return nil, err
	}

	// 2. 获取下载链接
	logger.LOG.Info("获取云盘下载链接", "taskID", taskID, "provider", opts.Provider, "fileID", opts.FileID)
	link, err := provider.GetDownloadLink(opts.FileID)
	if err != nil {
		errMsg := fmt.Sprintf("获取下载链接失败: %v", err)
		if strings.Contains(err.Error(), "Cookie") || strings.Contains(err.Error(), "cookie") {
			errMsg = "Cookie已过期或无效，请重新获取"
		}
		updateTaskFailed(repoFactory, taskID, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	if link.ExpiresAt != nil && time.Now().After(*link.ExpiresAt) {
		errMsg := "下载链接已过期，请重新获取"
		updateTaskFailed(repoFactory, taskID, errMsg)
		return nil, fmt.Errorf(errMsg)
	}

	// 3. 更新任务信息
	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取下载任务失败: %w", err)
	}

	// 优先使用已有的文件名和大小（来自请求），下载链接可能不包含这些信息
	fileName := task.FileName
	if fileName == "" {
		fileName = link.FileName
	}
	if fileName == "" {
		fileName = extractFileNameFromURL(link.DownloadURL)
	}
	fileSize := task.FileSize
	if fileSize == 0 {
		fileSize = link.Size
	}

	task.FileName = fileName
	task.State = enum.DownloadTaskStateDownloading.Value()
	task.UpdateTime = custom_type.Now()
	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		return nil, fmt.Errorf("更新任务信息失败: %w", err)
	}

	// 4. 创建临时目录
	sessionDir := filepath.Join(tempDir, fmt.Sprintf("cloud_%s", taskID))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		updateTaskFailed(repoFactory, taskID, "创建临时目录失败")
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}

	// 5. 下载文件（按 MustProxy / Headers 策略选择客户端）
	filePath := filepath.Join(sessionDir, fileName)
	progress := newDownloadProgress(taskID, fileSize, repoFactory)

	downloadClient := buildCloudDownloadClient(opts.Timeout, link)
	err = downloadWithHeaders(ctx, link.DownloadURL, filePath, progress, link.Headers, downloadClient)

	if err != nil {
		if !strings.Contains(err.Error(), "任务已取消") {
			updateTaskFailed(repoFactory, taskID, err.Error())
			os.RemoveAll(sessionDir)
		}
		return nil, fmt.Errorf("文件下载失败: %w", err)
	}

	if stat, err := os.Stat(filePath); err == nil && stat.Size() > 0 {
		fileSize = stat.Size()
		task.FileSize = fileSize
		task.UpdateTime = custom_type.Now()
		_ = repoFactory.DownloadTask().Update(ctx, task)
	}

	logger.LOG.Info("云盘文件下载完成", "taskID", taskID, "fileName", fileName, "size", fileSize)

	// 6. 确保虚拟路径存在
	if err := ensureVirtualPath(ctx, userID, opts.VirtualPath, repoFactory); err != nil {
		os.RemoveAll(sessionDir)
		updateTaskFailed(repoFactory, taskID, "创建虚拟路径失败")
		return nil, fmt.Errorf("创建虚拟路径失败: %w", err)
	}

	// 7. 上传文件到系统
	uploadData := &upload.FileUploadData{
		TempFilePath: filePath,
		FileName:     fileName,
		FileSize:     fileSize,
		VirtualPath:  opts.VirtualPath,
		UserID:       userID,
		IsEnc:        opts.EnableEncryption,
		IsChunk:      false,
		FilePassword: opts.FilePassword,
	}

	fileID, err := upload.ProcessUploadedFile(uploadData, repoFactory)
	if err != nil {
		updateTaskFailed(repoFactory, taskID, fmt.Sprintf("上传文件失败: %v", err))
		os.RemoveAll(sessionDir)
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 清理临时文件
	os.RemoveAll(sessionDir)

	// 8. 更新任务为完成状态
	task.FileID = fileID
	task.State = enum.DownloadTaskStateFinished.Value()
	task.Progress = 100
	task.DownloadedSize = task.FileSize
	task.UpdateTime = custom_type.Now()
	task.FinishTime = custom_type.Now()
	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务完成状态失败", "taskID", taskID, "error", err)
	}

	logger.LOG.Info("云盘下载任务完成", "taskID", taskID, "fileID", fileID)

	return &HTTPDownloadResult{
		FileID:   fileID,
		FileName: fileName,
		FileSize: fileSize,
	}, nil
}

// extractFileNameFromURL 从 URL 中提取文件名
func extractFileNameFromURL(rawURL string) string {
	u, err := neturl.Parse(rawURL)
	if err != nil {
		return "未知文件"
	}
	path := u.Path
	parts := strings.Split(path, "/")
	if len(parts) > 0 {
		name := parts[len(parts)-1]
		if name != "" {
			return name
		}
	}
	return "未知文件"
}

// buildCloudDownloadClient 根据链接策略构建 HTTP 客户端
func buildCloudDownloadClient(timeoutSec int, link *cloudsync.CloudDownloadLink) *http.Client {
	if timeoutSec <= 0 {
		timeoutSec = 300
	}
	client := &http.Client{Timeout: time.Duration(timeoutSec) * time.Second}
	if link == nil {
		return client
	}
	if link.MustProxy || len(link.Headers) > 0 {
		client.CheckRedirect = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	}
	return client
}

// downloadWithHeaders 带自定义请求头的下载（通用，适用于任何云盘 Provider）
func downloadWithHeaders(
	ctx context.Context,
	url string,
	filePath string,
	progress *downloadProgress,
	headers map[string]string,
	client *http.Client,
) error {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
	}

	if resp.StatusCode == http.StatusFound || resp.StatusCode == http.StatusMovedPermanently || resp.StatusCode == http.StatusTemporaryRedirect || resp.StatusCode == http.StatusSeeOther {
		location := resp.Header.Get("Location")
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if location != "" {
			return downloadWithHeaders(ctx, location, filePath, progress, headers, client)
		}
		return fmt.Errorf("下载失败，重定向缺少 Location，状态码: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建文件失败: %w", err)
	}
	defer file.Close()

	buffer := make([]byte, 256*1024)
	var downloaded int64

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("任务已取消")
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return fmt.Errorf("写入文件失败: %w", writeErr)
			}
			downloaded += int64(n)
			progress.updateProgress(downloaded)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}

	return nil
}

// updateTaskFailed 更新任务为失败状态
func updateTaskFailed(repoFactory *impl.RepositoryFactory, taskID, errMsg string) {
	ctx := context.Background()
	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return
	}
	task.State = enum.DownloadTaskStateFailed.Value()
	task.ErrorMsg = errMsg
	task.UpdateTime = custom_type.Now()
	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务失败状态失败", "taskID", taskID, "error", err)
	}
}
