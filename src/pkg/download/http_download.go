package download

import (
	"context"
	"fmt"
	"io"
	"myobj/src/internal/repository/impl"
	"myobj/src/pkg/custom_type"
	"myobj/src/pkg/enum"
	"myobj/src/pkg/logger"
	"myobj/src/pkg/upload"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// 下载任务管理器
var (
	downloadTasks   = make(map[string]context.CancelFunc)
	downloadTasksMu sync.RWMutex
)

// HTTPDownloadOptions HTTP下载配置
type HTTPDownloadOptions struct {
	EnableEncryption bool   // 是否加密存储
	VirtualPath      string // 虚拟保存路径
	MaxRetries       int    // 最大重试次数
	ChunkSize        int64  // 分片大小（字节），默认10MB
	MaxConcurrent    int    // 最大并发数，默认4
	Timeout          int    // 超时时间（秒），默认300
	FilePassword     string //加密文件密码（加密存储必备）
}

// HTTPDownloadResult HTTP下载结果
type HTTPDownloadResult struct {
	FileID   string // 上传成功的文件ID
	FileName string // 文件名
	FileSize int64  // 文件大小
	Error    string // 错误信息（如果有）
}

// chunkInfo 分片下载信息
type chunkInfo struct {
	Index      int   // 分片索引
	Start      int64 // 起始位置
	End        int64 // 结束位置
	RetryCount int   // 重试次数
}

// downloadProgress 下载进度管理器
type downloadProgress struct {
	TaskID             string
	TotalSize          int64
	DownloadedSize     int64
	LastDownloadedSize int64 // 上次下载量，用于计算实时速度
	Speed              int64
	LastUpdate         time.Time
	SpeedHistory       []int64 // 速度历史记录（最多10条），用于平滑显示
	RepoFactory        *impl.RepositoryFactory
	mu                 sync.RWMutex
}

// newDownloadProgress 创建进度管理器
func newDownloadProgress(taskID string, totalSize int64, repoFactory *impl.RepositoryFactory) *downloadProgress {
	return &downloadProgress{
		TaskID:             taskID,
		TotalSize:          totalSize,
		LastDownloadedSize: 0,
		LastUpdate:         time.Now(),
		SpeedHistory:       make([]int64, 0, 10),
		RepoFactory:        repoFactory,
	}
}

// updateProgress 更新下载进度（计算实时速度）
func (dp *downloadProgress) updateProgress(downloaded int64) {
	dp.mu.Lock()
	defer dp.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(dp.LastUpdate).Seconds()

	// 每秒更新一次速度
	if elapsed >= 1.0 {
		// 计算实时速度（增量/时间差）
		sizeDiff := downloaded - dp.LastDownloadedSize
		if sizeDiff > 0 && elapsed > 0 {
			currentSpeed := int64(float64(sizeDiff) / elapsed)

			// 添加到速度历史记录（最多保留10条）
			dp.SpeedHistory = append(dp.SpeedHistory, currentSpeed)
			if len(dp.SpeedHistory) > 10 {
				dp.SpeedHistory = dp.SpeedHistory[1:]
			}

			// 计算平均速度（平滑显示）
			var totalSpeed int64 = 0
			validCount := 0
			for _, speed := range dp.SpeedHistory {
				if speed >= 0 {
					totalSpeed += speed
					validCount++
				}
			}
			if validCount > 0 {
				dp.Speed = totalSpeed / int64(validCount)
			} else {
				dp.Speed = currentSpeed
			}
		} else if sizeDiff == 0 {
			// 如果没有下载进度，速度设为0
			dp.Speed = 0
		}

		dp.LastUpdate = now
		dp.LastDownloadedSize = downloaded
		dp.DownloadedSize = downloaded

		// 更新数据库
		ctx := context.Background()
		task, err := dp.RepoFactory.DownloadTask().GetByID(ctx, dp.TaskID)
		if err != nil {
			logger.LOG.Error("获取下载任务失败", "taskID", dp.TaskID, "error", err)
			return
		}

		task.DownloadedSize = dp.DownloadedSize
		task.Speed = dp.Speed
		if dp.TotalSize > 0 {
			task.Progress = int(float64(dp.DownloadedSize) / float64(dp.TotalSize) * 100)
		}
		task.UpdateTime = custom_type.Now()

		if err := dp.RepoFactory.DownloadTask().Update(ctx, task); err != nil {
			logger.LOG.Error("更新下载任务进度失败", "taskID", dp.TaskID, "error", err)
		}
	}
}

// DownloadHTTP 下载HTTP/HTTPS文件
// 参数:
//   - taskID: 下载任务ID
//   - url: 下载地址
//   - userID: 用户ID
//   - tempDir: 临时目录
//   - repoFactory: 数据库仓储工厂
//   - opts: 下载配置选项
//
// 返回:
//   - result: 下载结果
//   - err: 错误信息
func DownloadHTTP(
	taskID string,
	url string,
	userID string,
	tempDir string,
	repoFactory *impl.RepositoryFactory,
	opts *HTTPDownloadOptions,
) (*HTTPDownloadResult, error) {
	// 创建可取消的context
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 注册取消函数
	downloadTasksMu.Lock()
	downloadTasks[taskID] = cancel
	downloadTasksMu.Unlock()

	// 确保清理
	defer func() {
		downloadTasksMu.Lock()
		delete(downloadTasks, taskID)
		downloadTasksMu.Unlock()
	}()

	// 使用默认配置
	if opts == nil {
		opts = &HTTPDownloadOptions{
			EnableEncryption: false,
			VirtualPath:      "/离线下载/",
			MaxRetries:       3,
			ChunkSize:        10 * 1024 * 1024, // 10MB
			MaxConcurrent:    4,
			Timeout:          300,
		}
	}

	// 1. 获取文件信息
	logger.LOG.Info("开始获取文件信息", "url", url)
	fileInfo, supportRange, err := getFileInfo(url, opts.Timeout)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	logger.LOG.Info("文件信息获取成功",
		"fileName", fileInfo.FileName,
		"fileSize", fileInfo.FileSize,
		"supportRange", supportRange,
	)

	// 2. 更新任务信息
	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		return nil, fmt.Errorf("获取下载任务失败: %w", err)
	}

	task.FileName = fileInfo.FileName
	task.FileSize = fileInfo.FileSize
	task.SupportRange = supportRange
	task.State = enum.DownloadTaskStateDownloading.Value()
	task.UpdateTime = custom_type.Now()
	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		return nil, fmt.Errorf("更新任务信息失败: %w", err)
	}

	// 3. 创建临时目录
	sessionDir := filepath.Join(tempDir, fmt.Sprintf("http_%s", taskID))
	if err := os.MkdirAll(sessionDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %w", err)
	}
	// 注意：不在defer中删除，以支持断点续传

	// 4. 下载文件
	filePath := filepath.Join(sessionDir, fileInfo.FileName)
	progress := newDownloadProgress(taskID, fileInfo.FileSize, repoFactory)

	if supportRange && fileInfo.FileSize > opts.ChunkSize {
		// 支持断点续传，使用多线程下载
		logger.LOG.Info("使用多线程下载", "chunkSize", opts.ChunkSize, "concurrent", opts.MaxConcurrent)
		err = downloadWithRange(ctx, url, filePath, fileInfo.FileSize, opts, progress, taskID, repoFactory)
	} else {
		// 不支持断点续传或文件较小，直接下载
		logger.LOG.Info("使用单线程下载")
		err = downloadDirect(ctx, url, filePath, opts.Timeout, progress, taskID, repoFactory)
	}

	if err != nil {
		// 更新任务为失败状态
		task.State = enum.DownloadTaskStateFailed.Value()
		task.ErrorMsg = err.Error()
		task.UpdateTime = custom_type.Now()
		repoFactory.DownloadTask().Update(ctx, task)
		logger.LOG.Error("文件下载失败", "taskID", taskID, "error", err)

		// 如果不是用户暂停或取消，则清理临时文件
		if !strings.Contains(err.Error(), "任务已取消") {
			if removeErr := os.RemoveAll(sessionDir); removeErr != nil {
				logger.LOG.Warn("清理临时目录失败", "path", sessionDir, "error", removeErr)
			}
		}
		return nil, fmt.Errorf("文件下载失败: %w", err)
	}

	logger.LOG.Info("文件下载完成", "fileName", fileInfo.FileName, "size", fileInfo.FileSize)

	// 5. 确保虚拟路径存在
	if err := ensureVirtualPath(ctx, userID, opts.VirtualPath, repoFactory); err != nil {
		// 清理临时文件
		os.RemoveAll(sessionDir)
		return nil, fmt.Errorf("创建虚拟路径失败: %w", err)
	}

	// 6. 上传文件到系统
	uploadData := &upload.FileUploadData{
		TempFilePath: filePath,
		FileName:     fileInfo.FileName,
		FileSize:     fileInfo.FileSize,
		VirtualPath:  opts.VirtualPath,
		UserID:       userID,
		IsEnc:        opts.EnableEncryption,
		IsChunk:      false,
		FilePassword: opts.FilePassword,
	}

	fileID, err := upload.ProcessUploadedFile(uploadData, repoFactory)
	if err != nil {
		// 更新任务为失败状态
		task.State = enum.DownloadTaskStateFailed.Value()
		task.ErrorMsg = fmt.Sprintf("上传文件失败: %v", err)
		task.UpdateTime = custom_type.Now()
		repoFactory.DownloadTask().Update(ctx, task)
		logger.LOG.Error("上传文件失败", "taskID", taskID, "error", err)
		// 清理临时文件
		os.RemoveAll(sessionDir)
		return nil, fmt.Errorf("上传文件失败: %w", err)
	}

	// 清理临时文件（上传成功后）
	if removeErr := os.RemoveAll(sessionDir); removeErr != nil {
		logger.LOG.Warn("清理临时目录失败", "path", sessionDir, "error", removeErr)
	}

	// 7. 更新任务为完成状态
	task.FileID = fileID
	task.State = enum.DownloadTaskStateFinished.Value()
	task.Progress = 100
	task.UpdateTime = custom_type.Now()
	task.FinishTime = custom_type.Now()
	task.DownloadedSize = task.FileSize
	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务完成状态失败", "taskID", taskID, "error", err)
	}

	logger.LOG.Info("离线下载任务完成", "taskID", taskID, "fileID", fileID)

	return &HTTPDownloadResult{
		FileID:   fileID,
		FileName: fileInfo.FileName,
		FileSize: fileInfo.FileSize,
	}, nil
}

// fileInfoResult 文件信息结果
type fileInfoResult struct {
	FileName string
	FileSize int64
}

// getFileInfo 获取文件信息（文件名和大小）
func getFileInfo(url string, timeout int) (*fileInfoResult, bool, error) {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("服务器返回错误: %d", resp.StatusCode)
	}

	// 获取文件大小
	fileSize := resp.ContentLength

	// 检测是否支持断点续传
	supportRange := resp.Header.Get("Accept-Ranges") == "bytes"

	// 获取文件名
	fileName := extractFileName(url, resp.Header.Get("Content-Disposition"))

	return &fileInfoResult{
		FileName: fileName,
		FileSize: fileSize,
	}, supportRange, nil
}

// extractFileName 从URL或Content-Disposition中提取文件名
func extractFileName(url, contentDisposition string) string {
	// 优先从Content-Disposition获取
	if contentDisposition != "" {
		// 1. 优先处理 filename*=UTF-8''xxx (RFC 5987 格式)
		if idx := strings.Index(contentDisposition, "filename*=UTF-8''"); idx >= 0 {
			value := contentDisposition[idx+len("filename*=UTF-8''"):]
			// 移除分号后的内容（如果有）
			if semicolonIdx := strings.Index(value, ";"); semicolonIdx > 0 {
				value = value[:semicolonIdx]
			}
			// URL解码
			if decoded, err := neturl.QueryUnescape(value); err == nil {
				fileName := sanitizeFileName(decoded)
				if fileName != "" {
					return fileName
				}
			}
		}

		// 2. 处理 filename="xxx" 或 filename=xxx
		if idx := strings.Index(contentDisposition, "filename="); idx >= 0 {
			value := contentDisposition[idx+len("filename="):]
			// 移除分号后的内容（如果有）
			if semicolonIdx := strings.Index(value, ";"); semicolonIdx > 0 {
				value = value[:semicolonIdx]
			}
			// 移除引号和空格
			value = strings.Trim(value, " \"")
			// 如果值不为空且不是 filename*= 的开头（避免重复处理）
			if value != "" && !strings.HasPrefix(value, "UTF-8''") {
				fileName := sanitizeFileName(value)
				if fileName != "" {
					return fileName
				}
			}
		}
	}

	// 从URL中提取
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		lastPart := parts[len(parts)-1]
		// 移除查询参数
		if idx := strings.Index(lastPart, "?"); idx > 0 {
			lastPart = lastPart[:idx]
		}
		if lastPart != "" {
			fileName := sanitizeFileName(lastPart)
			if fileName != "" {
				return fileName
			}
		}
	}

	// 使用默认文件名
	return fmt.Sprintf("未命名文件_%s", time.Now().Format("20060102150405"))
}

// sanitizeFileName 清理文件名，移除Windows不允许的字符
func sanitizeFileName(fileName string) string {
	// Windows不允许的字符: < > : " / \ | ? *
	invalidChars := []string{"<", ">", ":", "\"", "/", "\\", "|", "?", "*"}
	result := fileName

	// 移除所有非法字符
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// 移除控制字符（ASCII 0-31）
	var builder strings.Builder
	for _, r := range result {
		if r > 31 && r != 127 { // 保留可打印字符，排除DEL (127)
			builder.WriteRune(r)
		}
	}
	result = builder.String()

	// 移除首尾空格和点号（Windows不允许）
	result = strings.Trim(result, " .")

	// 如果结果为空，返回默认名称
	if result == "" {
		return fmt.Sprintf("未命名文件_%s", time.Now().Format("20060102150405"))
	}

	// Windows保留名称检查
	reservedNames := []string{"CON", "PRN", "AUX", "NUL", "COM1", "COM2", "COM3", "COM4", "COM5", "COM6", "COM7", "COM8", "COM9", "LPT1", "LPT2", "LPT3", "LPT4", "LPT5", "LPT6", "LPT7", "LPT8", "LPT9"}
	upperName := strings.ToUpper(result)
	for _, reserved := range reservedNames {
		if upperName == reserved || strings.HasPrefix(upperName, reserved+".") {
			return fmt.Sprintf("未命名文件_%s", time.Now().Format("20060102150405"))
		}
	}

	return result
}

// downloadDirect 直接下载（不支持断点续传）
func downloadDirect(
	ctx context.Context,
	url string,
	filePath string,
	timeout int,
	progress *downloadProgress,
	taskID string,
	repoFactory *impl.RepositoryFactory,
) error {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("下载请求失败: %w", err)
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

	// 带进度的复制
	buffer := make([]byte, 32*1024) // 32KB缓冲区
	var downloaded int64

	for {
		// 检查context是否取消
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
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}

	return nil
}

// downloadWithRange 使用断点续传多线程下载
func downloadWithRange(
	ctx context.Context,
	url string,
	filePath string,
	totalSize int64,
	opts *HTTPDownloadOptions,
	progress *downloadProgress,
	taskID string,
	repoFactory *impl.RepositoryFactory,
) error {
	// 检查是否存在未完成的下载文件
	var file *os.File
	var err error
	var existingSize int64

	if fileInfo, statErr := os.Stat(filePath); statErr == nil {
		// 文件已存在，从断点续传
		existingSize = fileInfo.Size()
		logger.LOG.Info("检测到未完成的下载，从断点续传",
			"filePath", filePath,
			"existingSize", existingSize,
			"totalSize", totalSize,
		)
		file, err = os.OpenFile(filePath, os.O_RDWR, 0644)
		if err != nil {
			return fmt.Errorf("打开文件失败: %w", err)
		}
	} else {
		// 创建新文件
		file, err = os.Create(filePath)
		if err != nil {
			return fmt.Errorf("创建文件失败: %w", err)
		}
		// 预分配文件空间
		if err := file.Truncate(totalSize); err != nil {
			file.Close()
			return fmt.Errorf("预分配文件空间失败: %w", err)
		}
		existingSize = 0
	}
	defer file.Close()

	// 计算分片
	chunks := calculateChunks(totalSize, opts.ChunkSize)
	logger.LOG.Info("分片计算完成", "总分片数", len(chunks))

	// 过滤已下载的分片
	var pendingChunks []chunkInfo
	for _, chunk := range chunks {
		if chunk.End < existingSize {
			// 该分片已完成
			continue
		} else if chunk.Start < existingSize {
			// 该分片部分完成，调整起始位置
			chunk.Start = existingSize
		}
		pendingChunks = append(pendingChunks, chunk)
	}

	if len(pendingChunks) == 0 {
		logger.LOG.Info("文件已下载完成")
		return nil
	}

	logger.LOG.Info("待下载分片",
		"总数", len(chunks),
		"待下载", len(pendingChunks),
		"已下载", len(chunks)-len(pendingChunks),
	)

	// 并发下载控制
	var wg sync.WaitGroup
	sem := make(chan struct{}, opts.MaxConcurrent) // 并发控制
	errChan := make(chan error, len(pendingChunks))
	var downloadedBytes int64 = existingSize // 从已下载的大小开始

	// 更新初始进度
	progress.updateProgress(existingSize)

	// 下载每个分片
	for i := range pendingChunks {
		// 检查context是否取消
		select {
		case <-ctx.Done():
			return fmt.Errorf("任务已取消")
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // 获取信号量

		go func(chunk *chunkInfo) {
			defer wg.Done()
			defer func() { <-sem }() // 释放信号量

			// 重试机制
			for retry := 0; retry <= opts.MaxRetries; retry++ {
				// 检查context是否取消
				select {
				case <-ctx.Done():
					return // 任务已取消，退出
				default:
				}

				if retry > 0 {
					logger.LOG.Warn("重试下载分片",
						"chunk", chunk.Index,
						"retry", retry,
						"maxRetries", opts.MaxRetries,
					)
					time.Sleep(time.Duration(retry) * 2 * time.Second) // 指数退避
				}

				err := downloadChunk(ctx, url, file, chunk, &downloadedBytes, progress, opts.Timeout)
				if err == nil {
					return // 下载成功
				}

				if retry == opts.MaxRetries {
					// 所有重试都失败
					errChan <- fmt.Errorf("分片 %d 下载失败: %w", chunk.Index, err)
					return
				}
			}
		}(&pendingChunks[i])
	}

	// 等待所有分片下载完成
	wg.Wait()
	close(errChan)

	// 检查是否有错误
	for err := range errChan {
		if err != nil {
			return err
		}
	}

	logger.LOG.Info("所有分片下载完成")
	return nil
}

// calculateChunks 计算分片信息
func calculateChunks(totalSize, chunkSize int64) []chunkInfo {
	var chunks []chunkInfo
	var start int64

	for start < totalSize {
		end := start + chunkSize - 1
		if end >= totalSize {
			end = totalSize - 1
		}

		chunks = append(chunks, chunkInfo{
			Index: len(chunks),
			Start: start,
			End:   end,
		})

		start = end + 1
	}

	return chunks
}

// downloadChunk 下载单个分片
func downloadChunk(
	ctx context.Context,
	url string,
	file *os.File,
	chunk *chunkInfo,
	downloadedBytes *int64,
	progress *downloadProgress,
	timeout int,
) error {
	client := &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置Range头
	rangeHeader := fmt.Sprintf("bytes=%d-%d", chunk.Start, chunk.End)
	req.Header.Set("Range", rangeHeader)

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusPartialContent && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("服务器返回错误: %d", resp.StatusCode)
	}

	// 读取数据并写入文件
	buffer := make([]byte, 32*1024) // 32KB缓冲区
	offset := chunk.Start

	for {
		// 检查context是否取消
		select {
		case <-ctx.Done():
			return fmt.Errorf("任务已取消")
		default:
		}

		n, err := resp.Body.Read(buffer)
		if n > 0 {
			// 写入指定位置
			if _, writeErr := file.WriteAt(buffer[:n], offset); writeErr != nil {
				return fmt.Errorf("写入文件失败: %w", writeErr)
			}
			offset += int64(n)
			atomic.AddInt64(downloadedBytes, int64(n))
			progress.updateProgress(atomic.LoadInt64(downloadedBytes))
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("读取数据失败: %w", err)
		}
	}

	return nil
}

// PauseDownload 暂停下载任务
func PauseDownload(taskID string, repoFactory *impl.RepositoryFactory) error {
	ctx := context.Background()

	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.State != enum.DownloadTaskStateDownloading.Value() {
		return fmt.Errorf("任务状态不允许暂停")
	}

	// 取消下载任务的context
	downloadTasksMu.RLock()
	cancel, exists := downloadTasks[taskID]
	downloadTasksMu.RUnlock()

	if exists && cancel != nil {
		cancel() // 取消context，停止所有goroutine
		logger.LOG.Info("已取消下载任务的goroutine", "taskID", taskID)
	}

	task.State = enum.DownloadTaskStatePaused.Value()
	task.UpdateTime = custom_type.Now()

	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务状态失败", "taskID", taskID, "error", err)
		return fmt.Errorf("暂停任务失败: %w", err)
	}

	logger.LOG.Info("任务已暂停", "taskID", taskID)
	return nil
}

// ResumeDownload 恢复下载任务
func ResumeDownload(taskID string, userID string, tempDir string, repoFactory *impl.RepositoryFactory) error {
	ctx := context.Background()

	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.State != enum.DownloadTaskStatePaused.Value() {
		return fmt.Errorf("任务状态不允许恢复")
	}

	task.State = enum.DownloadTaskStateDownloading.Value()
	task.UpdateTime = custom_type.Now()

	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务状态失败", "taskID", taskID, "error", err)
		return fmt.Errorf("恢复任务失败: %w", err)
	}

	// 重新启动下载（异步）
	go func() {
		opts := &HTTPDownloadOptions{
			EnableEncryption: false, // HTTP离线下载不加密
			VirtualPath:      task.VirtualPath,
			MaxRetries:       3,
			ChunkSize:        10 * 1024 * 1024,
			MaxConcurrent:    4,
			Timeout:          300,
		}
		_, err := DownloadHTTP(taskID, task.URL, userID, tempDir, repoFactory, opts)
		if err != nil {
			logger.LOG.Error("恢复下载失败", "taskID", taskID, "error", err)
		}
	}()

	logger.LOG.Info("任务已恢复", "taskID", taskID)
	return nil
}

// CancelDownload 取消下载任务
func CancelDownload(taskID string, repoFactory *impl.RepositoryFactory) error {
	ctx := context.Background()

	task, err := repoFactory.DownloadTask().GetByID(ctx, taskID)
	if err != nil {
		logger.LOG.Error("获取下载任务失败", "taskID", taskID, "error", err)
		return fmt.Errorf("获取任务失败: %w", err)
	}

	if task.State == enum.DownloadTaskStateFinished.Value() {
		return fmt.Errorf("任务已完成，无法取消")
	}

	// 取消下载任务的context
	downloadTasksMu.RLock()
	cancel, exists := downloadTasks[taskID]
	downloadTasksMu.RUnlock()

	if exists && cancel != nil {
		cancel() // 取消context，停止所有goroutine
		logger.LOG.Info("已取消下载任务的goroutine", "taskID", taskID)
	}

	task.State = enum.DownloadTaskStateFailed.Value()
	task.ErrorMsg = "用户取消下载"
	task.UpdateTime = custom_type.Now()

	if err := repoFactory.DownloadTask().Update(ctx, task); err != nil {
		logger.LOG.Error("更新任务状态失败", "taskID", taskID, "error", err)
		return fmt.Errorf("取消任务失败: %w", err)
	}

	logger.LOG.Info("任务已取消", "taskID", taskID)
	return nil
}
